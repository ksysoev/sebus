package sebus

import (
	"testing"
	"time"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()
	testEvent := testEvent{}

	sub, err := bus.Subscribe(testEvent.Topic(), 0)

	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	go func() {
		err := bus.Publish(testEvent)
		if err != nil {
			t.Errorf("Expected to get no error, but got %v", err)
		}
	}()

	select {
	case event := <-sub.Stream():
		if event.Data().(string) != testEvent.Data().(string) {
			t.Errorf("Expected message to be %v, but got %v", testEvent, event)
		}
	case <-time.After(time.Millisecond * 1000):
		t.Error("Did not receive message within timeout")
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	sub, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	if err := bus.Unsubscribe(sub); err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	go func() {
		err := bus.Publish(event)
		if err != nil {
			t.Errorf("Expected to get no error, but got %v", err)
		}
	}()

	select {
	case _, ok := <-sub.Stream():
		if ok {
			t.Error("Expected to get closed channel")
		}
	default:
		t.Error("Expected to get closed channel")
	}
}
