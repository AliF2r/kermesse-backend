package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kermesse-backend/internal/notifications"
	"net/http"
)

type WebSocketHandler struct {
	upgrader websocket.Upgrader
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	organizerId := r.URL.Query().Get("organizerId")
	if organizerId == "" {
		http.Error(w, "Organizer ID is required", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	notifications.RegisterOrganizer(organizerId, conn)
	defer notifications.UnregisterOrganizer(organizerId)

	fmt.Printf("Organizer %s connected\n", organizerId)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}
		fmt.Printf("Received from %s: %s\n", organizerId, message)
	}
}
