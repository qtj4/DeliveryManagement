package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type wsClient struct {
	deliveryId string
	conn       *websocket.Conn
}

type wsHub struct {
	mu      sync.RWMutex
	clients map[string]map[*wsClient]struct{} // deliveryId -> clients
}

var hub = &wsHub{clients: make(map[string]map[*wsClient]struct{})}

func (h *wsHub) addClient(deliveryId string, c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[deliveryId] == nil {
		h.clients[deliveryId] = make(map[*wsClient]struct{})
	}
	h.clients[deliveryId][c] = struct{}{}
}

func (h *wsHub) removeClient(deliveryId string, c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[deliveryId] != nil {
		delete(h.clients[deliveryId], c)
		if len(h.clients[deliveryId]) == 0 {
			delete(h.clients, deliveryId)
		}
	}
}

func (h *wsHub) broadcast(deliveryId string, msg interface{}) {
	h.mu.RLock()
	clients := h.clients[deliveryId]
	h.mu.RUnlock()
	for c := range clients {
		c.conn.WriteJSON(msg)
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func TrackWSHandler(c *gin.Context) {
	deliveryId := c.Param("deliveryId")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &wsClient{deliveryId: deliveryId, conn: conn}
	hub.addClient(deliveryId, client)
	defer func() {
		hub.removeClient(deliveryId, client)
		conn.Close()
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
