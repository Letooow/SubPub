package api

import (
	"context"
	"errors"
	"sync"
)

type SubPubRepo struct {
	subscriptions map[string]map[int64]*SubscriptionRepo
	mu            *sync.Mutex
}

func NewSubPub() SubPub {
	return &SubPubRepo{
		subscriptions: make(map[string]map[int64]*SubscriptionRepo),
		mu:            new(sync.Mutex),
	}
}

func (sb *SubPubRepo) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	if cb == nil {
		return nil, errors.New("cb is nil")
	}
	if sb.subscriptions == nil {
		return nil, errors.New("not init subPub")
	}

	sb.mu.Lock()
	sub := NewSubscription(sb, subject, cb)
	if s, ok := sb.subscriptions[subject]; ok {
		s[sub.ID] = sub
	} else {
		sb.subscriptions[subject] = make(map[int64]*SubscriptionRepo)
		sb.subscriptions[subject][sub.ID] = sub
	}
	sb.mu.Unlock()
	return sb.subscriptions[subject][sub.ID], nil
}

func (sb *SubPubRepo) Publish(subject string, msg any) error {
	if sb.subscriptions == nil {
		return errors.New("not init subPub")
	}
	if s, ok := sb.subscriptions[subject]; ok {
		for _, v := range s {
			v.message <- msg
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
			for _, sub := range s {
				sub.Unsubscribe()
			}
		}
		sb.mu.Lock()
		sb.subscriptions = nil
		sb.mu.Unlock()
		return nil
	}
}
