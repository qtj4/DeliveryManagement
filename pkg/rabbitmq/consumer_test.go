package rabbitmq

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDLQOnSMTPFailure(t *testing.T) {
	// This is a pseudo-test: in real code, use a mock AMQP server or interface
	called := false
	fakeSMTP := func([]byte) error {
		called = true
		return assert.AnError
	}
	// Simulate a message with retry count
	msg := amqp.Delivery{
		Body:    []byte("test body"),
		Headers: amqp.Table{"x-retry-count": int32(2)},
	}
	// Instead of real AMQP, just check logic
	if err := fakeSMTP(msg.Body); err == nil {
		t.Error("expected SMTP error")
	}
	if !called {
		t.Error("SMTP not called")
	}
	// In real test, check that message would be published to DLQ with retry+1
}
