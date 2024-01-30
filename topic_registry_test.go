package sebus

import (
	"errors"
	"testing"
)

func TestTopicRegistry(t *testing.T) {
	tr := newTopicRegistry()
	topicName := "test_topic"
	sub := newSubscription(topicName, 10)

	// Test add
	tr.add(sub)
	topic, ok := tr.get(topicName)

	if !ok || len(topic.subscribers) != 1 {
		t.Errorf("Expected one subscriber for topic %s, but got %v", topicName, topic)
	}

	// Test remove
	err := errors.New("test error")
	tr.remove(sub, err)
	topic, ok = tr.get(topicName)

	if ok && len(topic.subscribers) != 0 {
		t.Errorf("Expected no subscribers for topic %s, but got %v", topicName, topic)
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v, but got %v", err, sub.Err())
	}

	// Test close
	sub = newSubscription(topicName, 10)
	tr.add(sub)
	tr.close(err)
	topic, ok = tr.get(topicName)

	if ok && len(topic.subscribers) != 0 {
		t.Errorf("Expected no subscribers for topic %s after close, but got %v", topicName, topic)
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v after close, but got %v", err, sub.Err())
	}
}

func TestTopicRegistry_MultipleSubscriptions(t *testing.T) {
	tr := newTopicRegistry()
	topicName := "test_topic"
	sub1 := newSubscription(topicName, 10)
	sub2 := newSubscription(topicName, 10)

	tr.add(sub1)
	tr.add(sub2)

	topic, ok := tr.get(topicName)
	if !ok || len(topic.subscribers) != 2 {
		t.Errorf("Expected two subscribers for topic %s, but got %v", topicName, topic)
	}

	tr.remove(sub2, nil)
	topic, ok = tr.get(topicName)

	if !ok || len(topic.subscribers) != 1 {
		t.Errorf("Expected one subscriber for topic %s, but got %v", topicName, topic)
	}

	tr.remove(sub1, nil)

	topic, ok = tr.get(topicName)
	if ok && len(topic.subscribers) != 0 {
		t.Errorf("Expected no subscribers for topic")
	}
}

func TestTopicRegistry_RemoveSubscriptionThatNotExists(t *testing.T) {
	tr := newTopicRegistry()
	topicName := "test_topic"
	sub := newSubscription(topicName, 10)

	tr.remove(sub, nil)

	_, ok := tr.get(topicName)
	if ok {
		t.Errorf("Expected topic to not exist")
	}
}
