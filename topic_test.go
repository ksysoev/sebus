package sebus

import (
	"errors"
	"testing"
)

func TestTopic(t *testing.T) {
	topic := newTopic()
	sub := newSubscription("test_topic", 10)

	// Test add
	topic.add(sub)

	if len(topic.subscribers) != 1 {
		t.Errorf("Expected one subscriber, but got %d", len(topic.subscribers))
	}

	// Test remove
	err := errors.New("test error")
	topic.remove(sub, err)

	if len(topic.subscribers) != 0 {
		t.Errorf("Expected no subscribers after remove, but got %d", len(topic.subscribers))
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v, but got %v", err, sub.Err())
	}

	// Test close
	sub = newSubscription("test_topic", 10)
	topic.add(sub)
	topic.close(err)

	if len(topic.subscribers) != 0 {
		t.Errorf("Expected no subscribers after close, but got %d", len(topic.subscribers))
	}

	if sub.Err() != err {
		t.Errorf("Expected error to be %v after close, but got %v", err, sub.Err())
	}

	// Test publish
	sub = newSubscription("test_topic", 10)
	topic.add(sub)

	event := testEvent{}
	topic.publish(event)

	select {
	case msg := <-sub.stream:
		if msg.Data().(string) != event.Data().(string) {
			t.Errorf("Expected message to be %v, but got %v", event.Data(), msg.Data())
		}
	default:
		t.Error("Did not receive message after publish")
	}
}
