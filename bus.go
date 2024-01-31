package sebus

import (
	"context"
	"errors"
)

// Event represents an event that can be published to the event bus.
// It must implement the Topic() and Data() methods.
// The Topic() method must return the topic of the event.
// The Data() method must return the data of the event.
type Event interface {
	Topic() string
	Data() any
}

// EventBus represents an event bus.
// It can be used to publish events and subscribe to topics.
// It is safe to use from multiple goroutines.
type EventBus struct {
	ctx     context.Context
	cancel  context.CancelFunc
	newSubs chan *Subscription
	rmSubs  chan *Subscription
	events  chan Event
}

var ErrEventBusClosed = errors.New("eventbus is closed")

// NewEventBus creates a new instance of EventBus.
// Returns a pointer to the created EventBus.
func NewEventBus() *EventBus {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	eb := &EventBus{
		ctx:     ctx,
		cancel:  cancel,
		newSubs: make(chan *Subscription),
		rmSubs:  make(chan *Subscription),
		events:  make(chan Event),
	}

	go eb.runRouter()

	return eb
}

// Publish publishes the given event to the event bus.
// It returns an error if the event bus is closed.
// If the event bus is closed, it returns an ErrEventBusClosed error.
func (eb *EventBus) Publish(event Event) error {
	select {
	case eb.events <- event:
		return nil
	case <-eb.ctx.Done():
		return ErrEventBusClosed
	}
}

// Subscribe adds a new subscription to the event bus for the specified topic.
// It returns a pointer to the created Subscription and an error, if any.
// The bufferSize parameter specifies the size of the internal channel buffer for the subscription.
// If the event bus is closed, it returns nil and an ErrEventBusClosed error.
func (eb *EventBus) Subscribe(topic string, bufferSize uint) (*Subscription, error) {
	sub := newSubscription(topic, bufferSize)

	select {
	case eb.newSubs <- sub:
		return sub, nil
	case <-eb.ctx.Done():
		return nil, ErrEventBusClosed
	}
}

// Unsubscribe removes the given subscription from the event bus.
// It returns an error if the event bus is closed.
// If the event bus is closed, it returns an ErrEventBusClosed error.
func (eb *EventBus) Unsubscribe(sub *Subscription) error {
	select {
	case eb.rmSubs <- sub:
		return nil
	case <-eb.ctx.Done():
		return ErrEventBusClosed
	}
}

func (eb *EventBus) runRouter() {
	registry := newTopicRegistry()

	for {
		select {
		case sub := <-eb.newSubs:
			registry.add(sub)
		case sub := <-eb.rmSubs:
			registry.remove(sub, ErrSubcriptionClosed)
		case event := <-eb.events:
			// TODO: offload to goroutine, to not block event publishing.. Can we?
			topic, ok := registry.get(event.Topic())
			if !ok {
				continue
			}

			topic.publish(event)
		case <-eb.ctx.Done():
			registry.close(ErrEventBusClosed)
		}
	}
}

// Close closes the event bus.
func (eb *EventBus) Close() {
	eb.cancel()
}
