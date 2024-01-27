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
	newSubs    chan Subscription
	rmSubs     chan Subscription
	pubish     chan Event
	closeMutex sync.RWMutex
}

type Subscription struct {
	topic  string
	stream chan Event
}

type SubscribersList []chan Event

var EventBusClosedError = errors.New("eventbus is closed")

func NewEventBus() *EventBus {
	newSub := make(chan Subscription)
	rmSub := make(chan Subscription)
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

	ch := make(chan Event, bufferSize)
	sub := Subscription{
		topic:  topic,
		stream: ch,
	}

	//TODO: Fix race condition if evebt bus is stoped after our context check
	eb.newSubs <- sub

	return &sub, nil
}

func (eb *EventBus) Unsubscribe(sub *Subscription) error {
	eb.closeMutex.RLock()
	defer eb.closeMutex.RUnlock()

	if eb.ctx.Err() != nil {
		return EventBusClosedError
	}

	eb.rmSubs <- *sub

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
			subRegistry[sub.topic] = append(subRegistry[sub.topic], sub.stream)
		case sub := <-eb.rmSubs:
			if _, found := subRegistry[sub.topic]; found {
				for i, v := range subRegistry[sub.topic] {
					if v == sub.stream {
						subRegistry[sub.topic] = append(subRegistry[sub.topic][:i], subRegistry[sub.topic][i+1:]...)
						break
					}
				}
			}
		case event := <-eb.pubish:
			if _, found := subRegistry[event.Topic()]; found {
				copyStreams := make(SubscribersList, len(subRegistry[event.Topic()]))
				copy(copyStreams, subRegistry[event.Topic()])

				//TODO: Find a way to provide order guarantee
				go func() {
					for _, stream := range copyStreams {
						stream <- event
					}
				}()
			}
		case <-eb.ctx.Done():
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
