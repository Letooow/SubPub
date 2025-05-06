package api

type Subscriber struct {
	Cb      MessageHandler
	Message chan any
}

func NewSubscriber(cb MessageHandler) *Subscriber {
	return &Subscriber{
		Cb:      cb,
		Message: make(chan any),
	}
}
