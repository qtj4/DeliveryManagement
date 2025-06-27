package handler

import (
	"context"
	"net/http"
	"time"

	"deliverymanagement/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewWSHub(redis *redis.Client) *ws.Hub {
	return ws.NewHub(redis)
}

// WebSocketHandler handles /ws/track/:deliveryID
func WebSocketHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		deliveryID := c.Param("deliveryID")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		client := &ws.Client{
			Conn:       conn,
			Send:       make(chan []byte, 256),
			DeliveryID: deliveryID,
			LastPong:   time.Now(),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		hub.SubscribeDelivery(ctx, deliveryID, client)
		defer hub.UnsubscribeDelivery(deliveryID, client)

		// Heartbeat
		client.Conn.SetPongHandler(func(string) error {
			client.LastPong = time.Now()
			return nil
		})
		client.StartPingPong(30*time.Second, func() {
			client.Conn.Close()
			cancel()
		})

		go func() {
			for msg := range client.Send {
				client.Conn.WriteMessage(websocket.TextMessage, msg)
			}
		}()

		for {
			_, _, err := client.Conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
