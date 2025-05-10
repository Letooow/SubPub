package gateways

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	subpub "subpub/internal/api"
	gen "subpub/internal/gateways/generated"
)

const (
	ErrWrongTypeOfMessage = "wrong message type"
)

type PubSub struct {
	gen.UnimplementedPubSubServer
	sb  subpub.SubPub
	Log *logrus.Logger
}

func NewPubSub(sb subpub.SubPub) *PubSub {
	return &PubSub{sb: sb}
}

func (pb *PubSub) Subscribe(subscribeRequest *gen.SubscribeRequest, g grpc.ServerStreamingServer[gen.Event]) error {
	msgChan := make(chan any)
	ctx, cancel := context.WithCancel(g.Context())
	defer cancel()
	subscriber, err := pb.sb.Subscribe(subscribeRequest.Key, func(msg any) {
		select {
		case msgChan <- msg:
		case <-ctx.Done():
		}
	})

	if err != nil {
		return status.Errorf(codes.Internal, "subscribe error: %v", err)
	}

	pb.Log.Infof("new subscriber ID: %v to: %s\n", subscriber, subscribeRequest.Key)

	go func() {
		<-ctx.Done()
		pb.Log.Infof("subscribe cancelled from: %s", subscribeRequest.Key)
		close(msgChan)
	}()

	errChan := make(chan error)
	go func() {
		for msg := range msgChan {
			str, ok := msg.(string)
			if !ok {
				pb.Log.Errorf("wrong type")
				errChan <- status.Errorf(codes.Internal, ErrWrongTypeOfMessage)
			}
			pb.Log.Infoln("received message:", str, "from:", subscribeRequest.Key)
			if err := g.Send(&gen.Event{Data: str}); err != nil {
				errChan <- status.Errorf(codes.Internal, "send error: %v", err)
				return
			}
		}
		errChan <- nil
	}()
	err = <-errChan
	if err != nil {
		return status.Errorf(codes.Internal, "subscribe error: %v", err)
	}
	return nil
}

func (pb *PubSub) Publish(ctx context.Context, publishRequest *gen.PublishRequest) (*emptypb.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := pb.sb.Publish(publishRequest.Key, publishRequest.Data)
		if err != nil {
			pb.Log.Infof("publish error: %v with key: %s and data: %s\n", err, publishRequest.Key, publishRequest.Data)
			return nil, status.Errorf(codes.NotFound, "publish error: %v", err)
		}
		pb.Log.Infoln("succeed publish to:", publishRequest.Key, "with data:", publishRequest.Data)
		return &emptypb.Empty{}, nil
	}
}
