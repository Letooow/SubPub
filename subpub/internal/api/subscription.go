package api

type SubscriptionRepo struct {
	subject     string
	subscribers []*subscriber
	subPub      *SubPubRepo
}

func (s *SubscriptionRepo) Unsubscribe() {
	s.subPub.mu.Lock()
	for i := range s.subscribers {
		close(s.subscribers[i].message)
	}
	delete(s.subPub.subscriptions, s.subject)
	s.subPub.mu.Unlock()
}
