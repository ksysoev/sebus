package sebus

import (
	"context"
	"errors"
)

type Event interface {
	Topic() string
	Data() any
}

type EventBus struct {
	ctx     context.Context
	cancel  context.CancelFunc
	newSubs chan *Subscription
	rmSubs  chan *Subscription
	events  chan Event
}

type SubscribersList []*Subscription

var ErrEventBusClosed = errors.New("eventbus is closed")

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

func (eb *EventBus) Publish(event Event) error {
	select {
	case eb.events <- event:
		return nil
	case <-eb.ctx.Done():
		return ErrEventBusClosed
	}
}

// TODO: Would be nice to subscribe to multiple topics(not sure about "all" topics subscription, probably not)
func (eb *EventBus) Subscribe(topic string, bufferSize uint) (*Subscription, error) {
	sub := newSubscription(topic, bufferSize)

	select {
	case eb.newSubs <- sub:
		return sub, nil
	case <-eb.ctx.Done():
		return nil, ErrEventBusClosed
	}
}

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
			subs, ok := registry.get(event.Topic())
			if !ok {
				continue
			}

			for _, sub := range subs {
				err := sub.publish(event)
				if err != nil {
					registry.remove(sub, err)
				}
			}
		case <-eb.ctx.Done():
			registry.close(ErrEventBusClosed)
		}
	}
}

func (eb *EventBus) Close() {
	eb.cancel()
}
