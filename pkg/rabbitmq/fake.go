package rabbitmq

// FakePublisher is a test double for Publisher
// Stores published messages in-memory for assertions
type FakePublisher struct {
	Messages []PublishedMessage
}

type PublishedMessage struct {
	Queue string
	Body  interface{}
}

func (f *FakePublisher) Publish(queue string, body interface{}) error {
	f.Messages = append(f.Messages, PublishedMessage{Queue: queue, Body: body})
	return nil
}

func (f *FakePublisher) Close() error {
	return nil
}
