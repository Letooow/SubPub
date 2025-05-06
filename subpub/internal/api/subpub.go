package api

import (
	"context"
	"errors"
	"sync"
)

type SubPubRepo struct {
	subscriptions map[string]*SubscriptionRepo
	mu            *sync.Mutex
}

func NewSubPub() SubPub {
	return &SubPubRepo{
		subscriptions: make(map[string]*SubscriptionRepo),
		mu:            new(sync.Mutex),
	}
}

func (sb *SubPubRepo) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	if cb == nil {
		return nil, errors.New("cb is nil")
	}

	sb.mu.Lock()
	if s, ok := sb.subscriptions[subject]; ok {
		subscriber := NewSubscriber(cb)
		s.subscribers = append(s.subscribers, subscriber)
	} else {
		subscriber := NewSubscriber(cb)
		sb.subscriptions[subject] = &SubscriptionRepo{
			subject:     subject,
			subscribers: []*Subscriber{subscriber},
			subPub:      sb,
		}
	}
	subscribers := sb.subscriptions[subject].subscribers
	sb.mu.Unlock()
	for i := range subscribers {
		go func() {
			for value := range subscribers[i].Message {
				done := make(chan struct{})
				go func() {
					subscribers[i].Cb(value)
					close(done)
				}()
				<-done
			}
		}()
	}
	return sb.subscriptions[subject], nil
}

func (sb *SubPubRepo) Publish(subject string, msg any) error {
	if s, ok := sb.subscriptions[subject]; ok {
		for i := range len(s.subscribers) {
			s.subscribers[i].Message <- msg
		}
	} else {
		return errors.New("no such subscription")
	}
	return nil
}

func (sb *SubPubRepo) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for _, s := range sb.subscriptions {
			s.Unsubscribe()
		}
		return nil
	}
}
