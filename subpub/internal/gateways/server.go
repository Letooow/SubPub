package gateways

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
	gen "subpub/internal/gateways/generated"
	"time"
)

type Server struct {
	server *grpc.Server
	pubSub *PubSub
	host   string
	port   uint16
}

func NewSubPubServer(pb *PubSub, options ...func(*Server)) *Server {
	d := grpc.NewServer()
	s := &Server{
		server: d,
		pubSub: pb,
		host:   "localhost",
		port:   8080,
	}
	for _, o := range options {
		o(s)
	}

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
	gen.RegisterPubSubServer(s.server, s.pubSub)
	if err != nil {
		return err
	}
	log.Printf("listening on %s %s", lis.Addr().Network(), lis.Addr())

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err = s.server.Serve(lis); err != nil {
			return fmt.Errorf("serve error: %v", err)
		}
		return nil
	})

	go func() {
		<-ctx.Done()
		log.Println("shutting down server...")
		select {
		case <-time.After(5 * time.Second):
			err := s.pubSub.sb.Close(context.Background())
			if err != nil {
				log.Printf("close pubSub connection error: %v", err)
				return
			}
			s.server.Stop()
			log.Println("server shutdown gracefully")
		}
	}()

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
