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
	pubish  chan Event
}

type SubscribersList []*Subscription

var ErrEventBusClosed = errors.New("eventbus is closed")

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
	select {
	case eb.pubish <- event:
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
	//TODO: move to separate abstraction
	subRegistry := make(map[string]SubscribersList)

	for {
		select {
		case sub := <-eb.newSubs:
			//TODO: Spin up a goroutine to handle separately each topic
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
						sub.close(ErrSubscriptionBufferOverflow)
					}
				}
			}
		case <-eb.ctx.Done():
			for _, subs := range subRegistry {
				for _, sub := range subs {
					sub.close(ErrEventBusClosed)
				}
			}

			for {
				select {
				case sub := <-eb.newSubs:
					sub.close(ErrEventBusClosed)
				default:
					return
				}
			}
		}
	}
}

func (eb *EventBus) Close() {
	eb.cancel()
}

func deleteSubscription(subs SubscribersList, sub *Subscription) SubscribersList {
	for i, v := range subs {
		if v.stream == sub.stream {
			return append(subs[:i], subs[i+1:]...)
		}
	}

	return subs
}
