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
	case <-time.After(time.Millisecond):
		t.Error("Expected to get closed channel")
	}
}

func TestEventBus_UnsubscribeAll(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	sub1, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	sub2, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	if err := bus.Unsubscribe(sub1); err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	if err := bus.Unsubscribe(sub2); err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	go func() {
		err := bus.Publish(event)
		if err != nil {
			t.Errorf("Expected to get no error, but got %v", err)
		}
	}()

	select {
	case _, ok := <-sub1.Stream():
		if ok {
			t.Error("Expected to get closed channel")
		}
	case _, ok := <-sub2.Stream():
		if ok {
			t.Error("Expected to get closed channel")
		}
	case <-time.After(time.Millisecond):
		t.Error("Expected to get closed channel")
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	sub, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	bus.Close()

	go func() {
		err := bus.Publish(event)
		if err == nil {
			t.Errorf("Expected to get error, but got %v", err)
		}
	}()

	select {
	case _, ok := <-sub.Stream():
		if ok {
			t.Error("Expected to get closed channel")
		}
	case <-time.After(time.Millisecond):
		t.Error("Expected to get closed channel")
	}
}

func TestEventBus_SubscribeOnClosedBus(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	bus.Close()

	_, err := bus.Subscribe(event.Topic(), 0)
	if err == nil {
		t.Error("Expected to get error, but got nil")
	}
}

func TestEventBus_UnsubscribeOnClosedBus(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	sub, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	bus.Close()

	err = bus.Unsubscribe(sub)
	if err == nil {
		t.Error("Expected to get error, but got nil")
	}
}

func TestEventBus_SubscriptionOverflow(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	sub, err := bus.Subscribe(event.Topic(), 0)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	err = bus.Publish(event)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	//TODO: How we can synchronize here without sleep?
	time.Sleep(time.Millisecond)

	select {
	case _, ok := <-sub.Stream():
		if ok {
			t.Error("Expected to get closed channel")
		}

		if sub.Err() != ErrSubscriptionBufferOverflow {
			t.Errorf("Expected to get error %v, but got %v", ErrSubscriptionBufferOverflow, sub.Err())
		}
	case <-time.After(time.Millisecond):
		t.Error("Expected to get closed channel")
	}
}
