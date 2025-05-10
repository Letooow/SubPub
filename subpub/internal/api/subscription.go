package api

type SubscriptionRepo struct {
	ID int64

	subject string
	subPub  *SubPubRepo

	cb      MessageHandler
	message chan any
}

var updateID int64 = 1

func (s *SubscriptionRepo) Unsubscribe() {
	s.subPub.mu.Lock()
	close(s.message)
	delete(s.subPub.subscriptions[s.subject], s.ID)
	if len(s.subPub.subscriptions[s.subject]) == 0 {
		delete(s.subPub.subscriptions, s.subject)
	}
	s.subPub.mu.Unlock()
}

func NewSubscription(subPub *SubPubRepo, subject string, cb MessageHandler) *SubscriptionRepo {
	sb := &SubscriptionRepo{
		ID:      updateID,
		subject: subject,
		cb:      cb,
		subPub:  subPub,
		message: make(chan any),
	}
	updateID++
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
