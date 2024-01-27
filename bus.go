package sebus

import (
	"sync"
)

type EventBus struct {
	subscribers map[string]DataChannelSlice
	rm          sync.RWMutex
}

type DataChannel chan any
type DataChannelSlice []DataChannel

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]DataChannelSlice),
	}
}

func (eb *EventBus) Publish(topic string, data any) {
	eb.rm.RLock()
	if chans, found := eb.subscribers[topic]; found {
		channels := append(DataChannelSlice{}, chans...)
		go func(data any, dataChannelSlice DataChannelSlice) {
			for _, ch := range dataChannelSlice {
				ch <- data
			}
		}(data, channels)
	}
	eb.rm.RUnlock()
}

func (eb *EventBus) Subscribe(topic string) <-chan any {
	eb.rm.Lock()
	ch := make(DataChannel)
	if prev, ok := eb.subscribers[topic]; ok {
		eb.subscribers[topic] = append(prev, ch)
	} else {
		eb.subscribers[topic] = append([]DataChannel{}, ch)
	}
	eb.rm.Unlock()

	return ch
}
