package sebus

import (
	"errors"
	"testing"
)

func TestNewSubscription(t *testing.T) {
	topic := "test_topic"
	bufferSize := uint(10)
	sub := newSubscription(topic, bufferSize)

	if sub.Topic() != topic {
		t.Errorf("Expected topic to be %s, but got %s", topic, sub.Topic())
	}

	if cap(sub.Stream()) != int(bufferSize) {
		t.Errorf("Expected buffer size to be %d, but got %d", bufferSize, cap(sub.Stream()))
	}

	if sub.Err() != nil {
		t.Errorf("Expected error to be nil, but got %v", sub.Err())
	}
}

func TestSubscription_Err(t *testing.T) {
	sub := &Subscription{
		err: ErrSubcriptionClosed,
	}

	if sub.Err() != ErrSubcriptionClosed {
		t.Errorf("Expected error to be %v, but got %v", ErrSubcriptionClosed, sub.Err())
	}
}

func TestSubscription_close(t *testing.T) {
	topic := "test_topic"
	bufferSize := uint(10)
	sub := newSubscription(topic, bufferSize)

	ErrTest := errors.New("test error")

	sub.close(ErrTest)

	if sub.Err() != ErrTest {
		t.Errorf("Expected error to be %v after closing, but got %v", ErrTest, sub.Err())
	}
}
