package api

type subscriber struct {
	cb      MessageHandler
	message chan any
}

func newSubscriber(cb MessageHandler) *subscriber {
	sb := &subscriber{
		cb:      cb,
		message: make(chan any),
	}
	go func() {
		for value := range sb.message {
			done := make(chan struct{})
			go func() {

				sb.cb(value)
				close(done)
			}()
			<-done
		}
	}()
	return sb
}
