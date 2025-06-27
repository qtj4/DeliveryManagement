package ws

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn       *websocket.Conn
	Send       chan []byte
	DeliveryID string
	LastPong   time.Time
}

type Hub struct {
	Redis   *redis.Client
	Clients map[string]map[*Client]struct{} // deliveryID -> clients
	Mu      sync.RWMutex
}

func NewHub(redis *redis.Client) *Hub {
	h := &Hub{
		Redis:   redis,
		Clients: make(map[string]map[*Client]struct{}),
	}
	return h
}

func (h *Hub) Run(ctx context.Context) {
	// No-op: all pub/sub handled per delivery
}

func (h *Hub) SubscribeDelivery(ctx context.Context, deliveryID string, client *Client) {
	h.Mu.Lock()
	if h.Clients[deliveryID] == nil {
		h.Clients[deliveryID] = make(map[*Client]struct{})
		go h.redisForwarder(ctx, deliveryID)
	}
	h.Clients[deliveryID][client] = struct{}{}
	h.Mu.Unlock()
}

func (h *Hub) UnsubscribeDelivery(deliveryID string, client *Client) {
	h.Mu.Lock()
	if clients, ok := h.Clients[deliveryID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.Clients, deliveryID)
		}
	}
	h.Mu.Unlock()
}

func (h *Hub) Publish(deliveryID string, msg []byte) {
	h.Redis.Publish(context.Background(), "ws:delivery:"+deliveryID, msg)
}

func (h *Hub) redisForwarder(ctx context.Context, deliveryID string) {
	pubsub := h.Redis.Subscribe(ctx, "ws:delivery:"+deliveryID)
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-ch:
			h.Mu.RLock()
			for c := range h.Clients[deliveryID] {
				select {
				case c.Send <- []byte(m.Payload):
				default:
					// drop if blocked
				}
			}
			h.Mu.RUnlock()
		}
	}
}

func (c *Client) StartPingPong(timeout time.Duration, disconnect func()) {
	go func() {
		ticker := time.NewTicker(timeout / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if time.Since(c.LastPong) > timeout {
					disconnect()
					return
				}
				c.Conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}()
}
