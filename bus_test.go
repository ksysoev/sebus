package sebus

import (
	"sync"
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

func TestEventBus_PublisToTopicWithNoSubscribers(t *testing.T) {
	bus := NewEventBus()
	event := testEvent{}

	err := bus.Publish(event)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}
}

func BenchmarkPublish(_ *testing.B) {
	startCh := make(chan int)
	bus := NewEventBus()
	topics := []string{"topic1", "topic2", "topic3", "topic4", "topic5", "topic6", "topic7", "topic8", "topic9", "topic10"}

	wg := sync.WaitGroup{}

	wg.Add(10)

	for _, topic := range topics {
		sub, _ := bus.Subscribe(topic, 10000)

		go func() {
			for i := 0; i < 10000; i++ {
				<-sub.Stream()
			}
			wg.Done()
		}()
	}

	wg.Add(1000)

	for _, topic := range topics {
		e := testEvent{topic: topic}

		for i := 0; i < 100; i++ {
			go func() {
				<-startCh

				for i := 0; i < 100; i++ {
					_ = bus.Publish(e)
				}

				wg.Done()
			}()
		}
	}

	close(startCh)
	wg.Wait()
}

func BenchmarkSubsciption(_ *testing.B) {
	startCh := make(chan int)
	bus := NewEventBus()
	topics := []string{"topic1", "topic2", "topic3", "topic4", "topic5", "topic6", "topic7", "topic8", "topic9", "topic10"}

	wg := sync.WaitGroup{}

	wg.Add(1000)

	for _, topic := range topics {
		for i := 0; i < 100; i++ {
			sub, _ := bus.Subscribe(topic, 100)

			go func() {
				for i := 0; i < 100; i++ {
					<-sub.Stream()
				}
				wg.Done()
			}()
		}
	}

	wg.Add(10)

	for _, topic := range topics {
		e := testEvent{topic: topic}

		go func() {
			<-startCh

			for i := 0; i < 100; i++ {
				_ = bus.Publish(e)
			}

			wg.Done()
		}()
	}

	time.Sleep(time.Millisecond)
	close(startCh)
	wg.Wait()
}
