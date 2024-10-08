package notifications

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

var (
	organizerConnections = make(map[string]*websocket.Conn)
	connMutex            sync.Mutex
)

func RegisterOrganizer(organizerId string, conn *websocket.Conn) {
	connMutex.Lock()
	organizerConnections[organizerId] = conn
	connMutex.Unlock()
}

func UnregisterOrganizer(organizerId string) {
	connMutex.Lock()
	delete(organizerConnections, organizerId)
	connMutex.Unlock()
}

func NotifyOrganizer(organizerId, message string) {
	fmt.Print("organizer: ", organizerId)
	connMutex.Lock()
	conn, ok := organizerConnections[organizerId]
	connMutex.Unlock()
	fmt.Print("organizerConnectionId: ", conn)
	if ok {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			fmt.Println("Write error:", err)
		}
	} else {
		fmt.Printf("Organizer %s not connected\n", organizerId)
	}
}
