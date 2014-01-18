package wsHub

import ()

//Central communitaion struct
type WsHub struct {
	// Registered connections.
	connections map[*Client]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *Client

	// Unregister requests from connections.
	unregister chan *Client

	kill chan bool
}

//Create new hub
func NewHub() WsHub {
	h := WsHub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		connections: make(map[*Client]bool),
	}

	return h
}

// Run the hub (most likely in its own goroutine)
// Handles all communitaion between connected clients
func (h *WsHub) Run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
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
	defer func() { h.unregister <- c }()
}
