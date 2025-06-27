package ws

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisPubSubIntegration(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("integration test")
	}
	ctx := context.Background()
	redis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redis.Close()

	ch := redis.Subscribe(ctx, "ws:delivery:42").Channel()
	msg := []byte("hello")
	go func() {
		time.Sleep(100 * time.Millisecond)
		redis.Publish(ctx, "ws:delivery:42", msg)
	}()
	select {
	case m := <-ch:
		if m.Payload != string(msg) {
			t.Errorf("expected %q, got %q", msg, m.Payload)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for pubsub message")
	}
}
