package wsHub

import (
	"encoding/json"
	"fmt"
)

//Central communitaion struct
type WsHub struct {
	// Registered connections.
	connections map[*Client]bool
	// Leader connections
	leaders map[*Client]bool
	// Inbound messages from the connections.
	broadcast chan message
	// Register requests from the connections.
	register chan *Client
	// Unregister requests from connections.
	unregister chan *Client
	// kill signal
	kill chan bool
}

//Create new hub
func NewHub() *WsHub {
	return &WsHub{
		broadcast:   make(chan message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		connections: make(map[*Client]bool),
		leaders:     make(map[*Client]bool),
	}
}

// Run the hub (most likely in its own goroutine)
// Handles all communitaion between connected clients
func (h *WsHub) Run() {
	for {
		select {
		case c := <-h.register:
			if c.isLeader {
				h.leaders[c] = true
			} else {
				h.connections[c] = true
			}
		case c := <-h.unregister:
			if c.isLeader {
				delete(h.leaders, c)
			} else {
				delete(h.connections, c)
			}
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				if m.ref != c {
					select {
					case c.send <- m:
						fmt.Println("Broadcasting...")
					default:
						delete(h.connections, c)
						close(c.send)
						go c.
							ws.Close()
					}
				}
			}
		case <-h.kill:
			break
		}
	}
}

// Kill the running hub
func (h *WsHub) Stop() {
	h.kill <- true
}

//Register a given client object
func (h *WsHub) RegisterClient(c *Client) {
	h.register <- c
}

//Unregister a given client
func (h *WsHub) UnregisterClient(c *Client) {
	h.unregister <- c
}

//Broadcast a message to all connected clients
func (h *WsHub) Broadcast(c *Client, msg []byte) {
	h.broadcast <- message{c, msg}
}

func (h *WsHub) BroadcastJSON(c *Client, msg interface{}) {
	msgJSON, _ := json.Marshal(msg)
	h.broadcast <- message{c, msgJSON}
}
