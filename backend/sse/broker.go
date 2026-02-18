package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

// Event types sent over SSE
const (
	EventOddsUpdated   = "odds_updated"
	EventEventCreated  = "event_created"
	EventEventResolved = "event_resolved"
	EventUserResolved  = "user_resolved"
	EventBingoResolved = "bingo_resolved"
	EventBingoWinner   = "bingo_winner"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type client struct {
	ch     chan []byte
	userID int // 0 = anonymous
}

type Broker struct {
	mu        sync.RWMutex
	clients   map[*client]struct{}
	jwtSecret string
}

func NewBroker(jwtSecret string) *Broker {
	return &Broker{
		clients:   make(map[*client]struct{}),
		jwtSecret: jwtSecret,
	}
}

func (b *Broker) subscribe(userID int) *client {
	c := &client{ch: make(chan []byte, 64), userID: userID}
	b.mu.Lock()
	b.clients[c] = struct{}{}
	b.mu.Unlock()
	return c
}

func (b *Broker) unsubscribe(c *client) {
	b.mu.Lock()
	delete(b.clients, c)
	b.mu.Unlock()
	close(c.ch)
}

// Broadcast sends a message to all connected clients.
func (b *Broker) Broadcast(msgType string, data interface{}) {
	msg := Message{Type: msgType, Data: data}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for c := range b.clients {
		select {
		case c.ch <- bytes:
		default:
			// Client too slow, skip
		}
	}
}

// Send sends a message only to clients matching the given userID.
func (b *Broker) Send(userID int, msgType string, data interface{}) {
	msg := Message{Type: msgType, Data: data}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for c := range b.clients {
		if c.userID != userID {
			continue
		}
		select {
		case c.ch <- bytes:
		default:
		}
	}
}

// parseUserID extracts the user ID from a JWT token string.
// Returns 0 if the token is missing or invalid (anonymous).
func (b *Broker) parseUserID(tokenStr string) int {
	if tokenStr == "" {
		return 0
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(b.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0
	}
	uid, ok := claims["user_id"].(float64)
	if !ok {
		return 0
	}
	return int(uid)
}

// ServeHTTP handles the SSE endpoint.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	userID := b.parseUserID(r.URL.Query().Get("token"))

	c := b.subscribe(userID)
	defer b.unsubscribe(c)

	// Send a ping so the client knows we're connected
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
