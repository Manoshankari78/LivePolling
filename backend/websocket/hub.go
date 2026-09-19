package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Event struct {
	PollID     string   `json:"pollId"`
	Results    []Result `json:"results"`
	TotalVotes int64    `json:"totalVotes"`
}
type Result struct {
	OptionID string `json:"optionId"`
	Votes    int64  `json:"votes"`
}
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	pollID string
}
type Hub struct {
	mu             sync.RWMutex
	clients        map[string]map[*Client]struct{}
	upgrader       websocket.Upgrader
	allowedOrigins map[string]struct{}
}

func NewHub(allowedOrigin string) *Hub {
	origins := map[string]struct{}{}
	for _, origin := range strings.Split(allowedOrigin, ",") {
		if value := strings.TrimSpace(origin); value != "" {
			origins[value] = struct{}{}
		}
	}
	hub := &Hub{clients: make(map[string]map[*Client]struct{}), allowedOrigins: origins}
	hub.upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 4096, CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := hub.allowedOrigins[origin]
		return ok
	}}
	return hub
}
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, pollID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 16), pollID: pollID}
	h.add(client)
	go h.writePump(client)
	h.readPump(client)
}
func (h *Hub) add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.pollID] == nil {
		h.clients[c.pollID] = make(map[*Client]struct{})
	}
	h.clients[c.pollID][c] = struct{}{}
}
func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if group := h.clients[c.pollID]; group != nil {
		delete(group, c)
		if len(group) == 0 {
			delete(h.clients, c.pollID)
		}
	}
	close(c.send)
	_ = c.conn.Close()
}
func (h *Hub) Broadcast(pollID string, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[pollID] {
		select {
		case c.send <- payload:
		default:
			go h.remove(c)
		}
	}
}
func (h *Hub) readPump(c *Client) {
	defer h.remove(c)
	c.conn.SetReadLimit(2048)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
func (h *Hub) writePump(c *Client) {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func StartRedisSubscriber(rdb *redis.Client, hub *Hub) {
	ctx := context.Background()
	sub := rdb.PSubscribe(ctx, "poll:*:updates")
	go func() {
		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("redis subscriber error: %v", err)
				time.Sleep(time.Second)
				continue
			}
			var event Event
			if json.Unmarshal([]byte(msg.Payload), &event) == nil {
				hub.Broadcast(event.PollID, event)
			}
		}
	}()
}
