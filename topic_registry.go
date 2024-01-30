package sebus

type topicRegistry struct {
	topics map[string]*Topic
}

func newTopicRegistry() *topicRegistry {
	return &topicRegistry{
		topics: make(map[string]*Topic),
	}
}

func (tr *topicRegistry) add(sub *Subscription) {
	topic, ok := tr.topics[sub.Topic()]
	if !ok {
		topic = newTopic()
		tr.topics[sub.Topic()] = topic
	}

	topic.add(sub)
}

func (tr *topicRegistry) remove(sub *Subscription, err error) {
	subs, ok := tr.topics[sub.Topic()]
	if !ok {
		return
	}

	subs.remove(sub, err)
}

func (tr *topicRegistry) get(topic string) (*Topic, bool) {
	t, ok := tr.topics[topic]
	return t, ok
}

func (tr *topicRegistry) close(err error) {
	for name, topic := range tr.topics {
		topic.close(err)
		delete(tr.topics, name)
	}
}
