package api

import (
	"context"
	"errors"
)

var (
	ErrNoSuchSubscription = errors.New("no such subscription")
	ErrorNotInitSubPub    = errors.New("not init subPub")
	ErrCallBackFuncIsNil  = errors.New("call back func is nil")
)

// MessageHandler is a callback function that processed messages delivered to subscribes
type MessageHandler func(msg any)

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}

type SubPub interface {
	// Subscribe create an asynchronous queue subscriber on the given subject
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg any) error

	// Close will shut down sub-pub system
	// May be blocked by data delivery until the context is canceled
	Close(ctx context.Context) error
}
