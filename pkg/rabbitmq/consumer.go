package rabbitmq

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailConsumer consumes from email.queue and on failure republishes to email.dlq with retry count.
func EmailConsumer(conn *amqp.Connection, smtpSend func([]byte) error) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume("email.queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for msg := range msgs {
		retryCount := 0
		if v, ok := msg.Headers["x-retry-count"]; ok {
			if n, ok := v.(int32); ok {
				retryCount = int(n)
			}
		}
		err := smtpSend(msg.Body)
		if err != nil {
			// Re-publish to DLQ with incremented retry count
			err = ch.PublishWithContext(context.Background(), "", "email.dlq", false, false, amqp.Publishing{
				Headers:     amqp.Table{"x-retry-count": retryCount + 1},
				ContentType: msg.ContentType,
				Body:        msg.Body,
				Timestamp:   time.Now(),
			})
			if err != nil {
				log.Printf("Failed to publish to DLQ: %v", err)
			}
			msg.Ack(false)
			continue
		}
		msg.Ack(false)
	}
}
