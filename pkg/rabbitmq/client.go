package rabbitmq

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

// Publisher abstracts RabbitMQ publishing for testability
//go:generate mockgen -source=client.go -destination=mock_client.go -package=rabbitmq

// Publisher is the interface for publishing messages
// (for production and test fakes)
type Publisher interface {
	Publish(queue string, body interface{}) error
	Close() error
}

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, channel: ch}, nil
}

func (c *Client) Publish(queue string, body interface{}) error {
	q, err := c.channel.QueueDeclare(
		queue, true, false, false, false, nil,
	)
	if err != nil {
		return err
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.channel.Publish(
		"", q.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
}

func (c *Client) Close() error {
	c.channel.Close()
	return c.conn.Close()
}
