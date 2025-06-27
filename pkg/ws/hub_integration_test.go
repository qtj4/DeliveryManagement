package ws

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestMultiInstancePubSub(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("integration test")
	}
	ctx := context.Background()
	redis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redis.Close()

	hub1 := NewHub(redis)
	hub2 := NewHub(redis)

	msg := []byte("cross-instance event")
	var wg sync.WaitGroup
	wg.Add(1)
	// Simulate client on hub2
	client := &Client{Send: make(chan []byte, 1), DeliveryID: "42"}
	go func() {
		hub2.SubscribeDelivery(ctx, "42", client)
		select {
		case m := <-client.Send:
			if string(m) != string(msg) {
				t.Errorf("expected %q, got %q", msg, m)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for event")
		}
		wg.Done()
	}()
	// Publish from hub1
	time.Sleep(100 * time.Millisecond)
	hub1.Publish("42", msg)
	wg.Wait()
}

func TestPingPongDisconnect(t *testing.T) {
	// This is a logic test, not a real WebSocket test
	client := &Client{LastPong: time.Now()}
	var disconnected bool
	client.StartPingPong(100*time.Millisecond, func() { disconnected = true })
	time.Sleep(300 * time.Millisecond)
	if !disconnected {
		t.Error("client should have been disconnected after missed pings")
	}
}
