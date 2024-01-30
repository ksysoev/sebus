package sebus

type Topic struct {
	subscribers []*Subscription
}

func newTopic() *Topic {
	return &Topic{
		subscribers: make([]*Subscription, 0, 1),
	}
}

func (t *Topic) add(sub *Subscription) {
	t.subscribers = append(t.subscribers, sub)
}

func (t *Topic) remove(sub *Subscription, err error) {
	for i, v := range t.subscribers {
		if v.stream != sub.stream {
			continue
		}

		if len(t.subscribers) == 1 {
			t.subscribers = make([]*Subscription, 0)
		} else {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
		}

		sub.close(err)

		return
	}
}

func (t *Topic) close(err error) {
	for _, sub := range t.subscribers {
		sub.close(err)
	}

	t.subscribers = make([]*Subscription, 0)
}

func (t *Topic) publish(event Event) {
	for _, sub := range t.subscribers {
		err := sub.publish(event)
		if err != nil {
			t.remove(sub, err)
		}
	}
}
