package sebus

import "errors"

var ErrSubcriptionClosed = errors.New("subscription is closed")
var ErrSubscriptionBufferOverflow = errors.New("subscription buffer overflow")

type Subscription struct {
	err    error
	stream chan Event
	topic  string
}

func newSubscription(topic string, bufferSize uint) *Subscription {
	return &Subscription{
		topic:  topic,
		stream: make(chan Event, bufferSize),
	}
}

func (s *Subscription) Topic() string {
	return s.topic
}

func (s *Subscription) Stream() <-chan Event {
	return s.stream
}

func (s *Subscription) Err() error {
	return s.err
}

func (s *Subscription) close(err error) {
	s.err = err
	close(s.stream)
}

func (s *Subscription) publish(event Event) error {
	select {
	case s.stream <- event:
		return nil
	default:
		return ErrSubscriptionBufferOverflow
	}
}
