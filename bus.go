package sebus

import (
	"context"
	"errors"
	"sync"
)

type Event interface {
	Topic() string
	Data() any
}

type EventBus struct {
	ctx        context.Context
	cancel     context.CancelFunc
	newSubs    chan *Subscription
	rmSubs     chan *Subscription
	pubish     chan Event
	closeMutex sync.RWMutex
}

type SubscribersList []*Subscription

var EventBusClosedError = errors.New("eventbus is closed")

func NewEventBus() *EventBus {
	newSub := make(chan *Subscription)
	rmSub := make(chan *Subscription)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	eb := &EventBus{
		ctx:     ctx,
		cancel:  cancel,
		newSubs: newSub,
		rmSubs:  rmSub,
	}

	go eb.runRouter()

	return eb
}

func (eb *EventBus) Publish(event Event) error {
	eb.closeMutex.RLock()
	defer eb.closeMutex.RUnlock()

	if eb.ctx.Err() != nil {
		return EventBusClosedError
	}

	eb.pubish <- event

	return nil
}

func (eb *EventBus) Subscribe(topic string, bufferSize uint) (*Subscription, error) {
	eb.closeMutex.RLock()
	defer eb.closeMutex.RUnlock()

	if eb.ctx.Err() != nil {
		return nil, EventBusClosedError
	}

	sub := newSubscription(topic, bufferSize, eb.rmSubs)

	eb.newSubs <- sub

	return sub, nil
}

func (eb *EventBus) Unsubscribe(sub *Subscription) error {
	eb.closeMutex.RLock()
	defer eb.closeMutex.RUnlock()

	if eb.ctx.Err() != nil {
		return EventBusClosedError
	}

	eb.rmSubs <- sub

	return nil
}

func (eb *EventBus) runRouter() {
	subRegistry := make(map[string]SubscribersList)

	for {
		select {
		case sub := <-eb.newSubs:
			if _, found := subRegistry[sub.topic]; !found {
				subRegistry[sub.topic] = SubscribersList{}
			}
			subRegistry[sub.topic] = append(subRegistry[sub.topic], sub)
		case sub := <-eb.rmSubs:
			if _, found := subRegistry[sub.topic]; found {
				subRegistry[sub.Topic()] = deleteSubscription(subRegistry[sub.Topic()], sub)
			}
		case event := <-eb.pubish:
			// TODO: offload to goroutine, to not block event publishing.. Can we?
			if subs, found := subRegistry[event.Topic()]; found {
				for _, sub := range subs {
					select {
					case sub.stream <- event:
					default:
						subRegistry[event.Topic()] = deleteSubscription(subRegistry[event.Topic()], sub)
						sub.err = SubscriptionBufferOverflowError
						close(sub.stream)

					}
				}
			}
		case <-eb.ctx.Done():
			for _, subs := range subRegistry {
				for _, sub := range subs {
					sub.err = EventBusClosedError
					close(sub.stream)
				}
			}
			return
		}
	}
}

func (eb *EventBus) Close() error {
	eb.closeMutex.Lock()
	defer eb.closeMutex.Unlock()

	if eb.ctx.Err() != nil {
		return EventBusClosedError
	}

	eb.cancel()
	return nil
}

func deleteSubscription(subs SubscribersList, sub *Subscription) SubscribersList {
	for i, v := range subs {
		if v.stream == sub.stream {
			return append(subs[:i], subs[i+1:]...)
		}
	}

	return subs
}
