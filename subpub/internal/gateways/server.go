package gateways

import (
	"context"
	"fmt"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	gen "subpub/internal/gateways/generated"
	"time"
)

type Server struct {
	server *grpc.Server
	pubSub *PubSub
	host   string
	port   uint16
	Log    *logrus.Logger
}

func NewSubPubServer(pb *PubSub, logger *logrus.Logger, options ...func(*Server)) *Server {
	logEntry := logrus.NewEntry(logger)
	grpclog.ReplaceGrpcLogger(logEntry)

	d := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpclog.UnaryServerInterceptor(logEntry)),
		grpc.ChainStreamInterceptor(
			grpclog.StreamServerInterceptor(logEntry)),
	)
	pb.Log = logger
	s := &Server{
		server: d,
		pubSub: pb,
		host:   "localhost",
		port:   8080,
		Log:    logger,
	}
	for _, o := range options {
		o(s)
	}
	gen.RegisterPubSubServer(s.server, s.pubSub)
	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}
	s.Log.Infof("listening on %s %s", lis.Addr().Network(), lis.Addr())

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err = s.server.Serve(lis); err != nil {
			return fmt.Errorf("serve error: %v", err)
		}
		return nil
	})

	go func() {
		<-ctx.Done()
		s.Log.Infof("shutting down server...")
		select {
		case <-time.After(5 * time.Second):
			err := s.pubSub.sb.Close(context.Background())
			if err != nil {
				s.Log.Errorf("close pubSub connection error: %v", err)
				return
			}
			s.server.Stop()
			s.Log.Infof("server shutdown gracefully")
		}
	}()

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
