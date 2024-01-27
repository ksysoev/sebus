package sebus

import "errors"

var ErrSubcriptionClosed = errors.New("subscription is closed")
var ErrSubscriptionBufferOverflow = errors.New("subscription buffer overflow")

type Subscription struct {
	err    error
	stream chan Event
	rmSub  chan<- *Subscription
	topic  string
}

func newSubscription(topic string, bufferSize uint, rmSub chan<- *Subscription) *Subscription {
	return &Subscription{
		topic:  topic,
		stream: make(chan Event, bufferSize),
		rmSub:  rmSub,
	}
}

func (s *Subscription) Topic() string {
	return s.topic
}

func (s *Subscription) Stream() <-chan Event {
	return s.stream
}

func (s *Subscription) Close() {
	s.rmSub <- s
	s.err = ErrSubcriptionClosed
}

func (s *Subscription) Err() error {
	return s.err
}
