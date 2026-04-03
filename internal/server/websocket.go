package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// WSHub gerencia conexões WebSocket e broadcast.
type WSHub struct {
	mu    sync.RWMutex
	conns map[*websocket.Conn]context.CancelFunc
}

// NewWSHub cria um novo hub.
func NewWSHub() *WSHub {
	return &WSHub{
		conns: make(map[*websocket.Conn]context.CancelFunc),
	}
}

// HandleWS é o handler HTTP para upgrade de WebSocket.
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS local
	})
	if err != nil {
		log.Printf("[ws] erro no accept: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	h.mu.Lock()
	h.conns[conn] = cancel
	h.mu.Unlock()

	log.Printf("[ws] nova conexão (%d total)", h.count())

	// Manter conexão viva com pings
	go func() {
		defer func() {
			h.remove(conn)
			conn.Close(websocket.StatusNormalClosure, "")
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				if err := conn.Ping(ctx); err != nil {
					return
				}
			}
		}
	}()

	// Ler mensagens do cliente (descartadas, mas necessário para detectar disconnect)
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			return
		}
	}
}

// Broadcast envia uma mensagem para todos os clientes conectados.
func (h *WSHub) Broadcast(resource string, data interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"type":     "update",
		"resource": resource,
		"data":     data,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.conns {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := conn.Write(ctx, websocket.MessageText, msg)
		cancel()
		if err != nil {
			go h.remove(conn)
		}
	}
}

func (h *WSHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cancel, ok := h.conns[conn]; ok {
		cancel()
		delete(h.conns, conn)
	}
}

func (h *WSHub) count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

// Close fecha todas as conexões.
func (h *WSHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn, cancel := range h.conns {
		cancel()
		conn.Close(websocket.StatusGoingAway, "server shutdown")
		delete(h.conns, conn)
	}
}
