package wsHub

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type WsHandler func(http.ResponseWriter, *http.Request, WsHub)

type Client struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	//Response
	response []byte
}

//Need to rewrite this, to only create clients either a ws conn, or a http.ResponseWrite / http.Request
func NewClient(w http.ResponseWriter, r *http.Request) (*Client, error) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return nil, err
	} else if err != nil {
		return nil, err
	}

	c := &Client{send: make(chan []byte, 256), ws: ws}
	return c, nil
}

//Write a message back to the client
func (c *Client) Write(msg []byte) {
	c.send <- msg
}

//Write a string message to the client
func (c *Client) WriteString(msg string) {
	c.send <- []byte(msg)
}

//Read a message from the websocket connection, wait untill you get a message
func (c *Client) Read() []byte {
	_, msg, _ := c.ws.ReadMessage()
	return msg
}

//Read a message, but give up after a given timeout (ms)
//returns nil if the timeout is hit
func (c *Client) ReadTimeout(timeout time.Duration) []byte {
	t := time.After(timeout)

	//Spin off a goroutine that reads a message from the ws, and send it down a channel
	ch := make(chan []byte)
	go func() {
		m := c.Read()
		ch <- m
	}()

	select {
	case <-t: //We hit a timeout before getting a message from the websocket conn
		return nil
	case msg := <-ch: //We got a message from the websocket before the timeout
		return msg
	}
}

//Writes to the ws conection via broadcast
func (c *Client) Start() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

//Remove maybe (Create a default wsHandler to use easily)
/*func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}
	c := &Client{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	c.writer()
}*/
