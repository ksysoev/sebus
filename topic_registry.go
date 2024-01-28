package sebus

type topicRegistry struct {
	topics map[string]SubscribersList
}

func newTopicRegistry() *topicRegistry {
	return &topicRegistry{
		topics: make(map[string]SubscribersList),
	}
}

func (tr *topicRegistry) add(sub *Subscription) {
	subs, ok := tr.topics[sub.Topic()]
	if !ok {
		subs = make(SubscribersList, 1)
	}

	subs = append(subs, sub)
	tr.topics[sub.Topic()] = subs
}

func (tr *topicRegistry) remove(sub *Subscription, err error) {
	subs, ok := tr.topics[sub.Topic()]
	if !ok {
		return
	}

	for i, v := range subs {
		if v.stream != sub.stream {
			continue
		}

		if len(subs) == 1 {
			delete(tr.topics, sub.Topic())
		} else {
			tr.topics[sub.Topic()] = append(subs[:i], subs[i+1:]...)
		}

		sub.close(err)

		return
	}
}

func (tr *topicRegistry) get(topic string) (SubscribersList, bool) {
	subs, ok := tr.topics[topic]
	return subs, ok
}

func (tr *topicRegistry) close(err error) {
	for topic, subs := range tr.topics {
		for _, sub := range subs {
			sub.close(err)
		}

		delete(tr.topics, topic)
	}
}
