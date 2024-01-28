package sebus

import (
	"errors"
	"testing"
)

func TestTopicRegistry(t *testing.T) {
	tr := newTopicRegistry()
	topic := "test_topic"
	sub := newSubscription(topic, 10)

	// Test add
	tr.add(sub)
	subs, ok := tr.get(topic)

	if !ok || len(subs) != 1 {
		t.Errorf("Expected one subscriber for topic %s, but got %d", topic, len(subs))
	}

	// Test remove
	err := errors.New("test error")
	tr.remove(sub, err)
	subs, ok = tr.get(topic)

	if ok || len(subs) != 0 {
		t.Errorf("Expected no subscribers for topic %s, but got %d", topic, len(subs))
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v, but got %v", err, sub.Err())
	}

	// Test close
	sub = newSubscription(topic, 10)
	tr.add(sub)
	tr.close(err)
	subs, ok = tr.get(topic)

	if ok || len(subs) != 0 {
		t.Errorf("Expected no subscribers for topic %s after close, but got %d", topic, len(subs))
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v after close, but got %v", err, sub.Err())
	}
}

func TestTopicRegistry_MultipleSubscriptions(t *testing.T) {
	tr := newTopicRegistry()
	topic := "test_topic"
	sub1 := newSubscription(topic, 10)
	sub2 := newSubscription(topic, 10)

	tr.add(sub1)
	tr.add(sub2)

	subs, ok := tr.get(topic)
	if !ok || len(subs) != 2 {
		t.Errorf("Expected two subscribers for topic %s, but got %d", topic, len(subs))
	}

	tr.remove(sub2, nil)
	subs, ok = tr.get(topic)

	if !ok || len(subs) != 1 {
		t.Errorf("Expected one subscriber for topic %s, but got %d", topic, len(subs))
	}

	tr.remove(sub1, nil)

	_, ok = tr.get(topic)
	if ok {
		t.Errorf("Expected no subscribers for topic")
	}
}

func TestTopicRegistry_RemoveSubscriptionThatNotExists(t *testing.T) {
	tr := newTopicRegistry()
	topic := "test_topic"
	sub := newSubscription(topic, 10)

	tr.remove(sub, nil)

	_, ok := tr.get(topic)
	if ok {
		t.Errorf("Expected topic to not exist")
	}
}
