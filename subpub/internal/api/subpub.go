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
	if sb.subscriptions == nil {
		return nil, errors.New("not init subpub")
	}

	sb.mu.Lock()
	if s, ok := sb.subscriptions[subject]; ok {
		sub := newSubscriber(cb)
		s.subscribers = append(s.subscribers, sub)
	} else {
		sub := newSubscriber(cb)
		sb.subscriptions[subject] = &SubscriptionRepo{
			subject:     subject,
			subscribers: []*subscriber{sub},
			subPub:      sb,
		}
	}
	sb.mu.Unlock()
	return sb.subscriptions[subject], nil
}

func (sb *SubPubRepo) Publish(subject string, msg any) error {
	if sb.subscriptions == nil {
		return errors.New("not init subpub")
	}
	if s, ok := sb.subscriptions[subject]; ok {
		for i := range len(s.subscribers) {
			s.subscribers[i].message <- msg
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
		sb.mu.Lock()
		sb.subscriptions = nil
		sb.mu.Unlock()
		return nil
	}
}
