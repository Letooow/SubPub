package gateways

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"log"
	subpub "subpub/internal/api"
	gen "subpub/internal/gateways/generated"
)

type PubSub struct {
	gen.UnimplementedPubSubServer
	sb subpub.SubPub
}

func NewPubSub(sb subpub.SubPub) *PubSub {
	return &PubSub{sb: sb}
}

func (pb *PubSub) Subscribe(subscribeRequest *gen.SubscribeRequest, g grpc.ServerStreamingServer[gen.Event]) error {
	msgChan := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := pb.sb.Subscribe(subscribeRequest.Key, func(msg any) {
		select {
		case msgChan <- msg:
		case <-ctx.Done():
		}
	})
	if err != nil {
		return status.Errorf(codes.Internal, "subscribe error: %v", err)
	}
	log.Println("new subscriber to: ", subscribeRequest.Key)
	errChan := make(chan error)
	go func() {
		for msg := range msgChan {
			str, ok := msg.(string)
			if !ok {
				log.Println("wrong type")
			}
			log.Println("received message:", str, "from:", subscribeRequest.Key)
			if err := g.Send(&gen.Event{Data: str}); err != nil {
				errChan <- status.Errorf(codes.Internal, "send error: %v", err)
				return
			}
		}
		errChan <- nil
	}()
	return <-errChan
}

func (pb *PubSub) Publish(ctx context.Context, publishRequest *gen.PublishRequest) (*emptypb.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := pb.sb.Publish(publishRequest.Key, publishRequest.Data)
		if err != nil {
			log.Printf("publish error: %v with key: %s and data: %s\n", err, publishRequest.Key, publishRequest.Data)
			return nil, status.Errorf(codes.NotFound, "publish error: %v", err)
		}
		log.Println("succeed publish to:", publishRequest.Key, "with data:", publishRequest.Data)
		return &emptypb.Empty{}, nil
	}
}

func (pb *PubSub) mustEmbedUnimplementedPubSubServer() {
	//TODO implement me
	panic("implement me")
}
