package sebus

import "errors"

var ErrSubcriptionClosed = errors.New("subscription is closed")
var ErrSubscriptionBufferOverflow = errors.New("subscription buffer overflow")

// Subscription represents a subscription to a topic.
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

// Topic returns the topic of the subscription.
func (s *Subscription) Topic() string {
	return s.topic
}

// Stream returns a channel that can be used to receive events from the subscription.
func (s *Subscription) Stream() <-chan Event {
	return s.stream
}

// Err returns the error that caused the subscription to close.
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
