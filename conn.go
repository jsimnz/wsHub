package wsHub

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var (
	TimeoutErr = errors.New("Timeout hit before messaged recieved") //Is it best to use an error for the timeout
)

type Client struct {
	// The websocket connection.
	ws *websocket.Conn
	// Leader flag
	isLeader bool
	// Buffered channel of outbound messages.
	send chan []byte
	//Response
	response []byte
}

// Creates a new ws client to broadcast to
// TODO: Read only option
func NewClient(w http.ResponseWriter, r *http.Request, bufSize ...int) (*Client, error) {
	readBuf := 1024
	writeBuf := 1024
	if len(bufSize) == 1 {
		readBuf = bufSize[0]
	} else if len(bufSize) >= 2 {
		readBuf = bufSize[0]
		writeBuf = bufSize[1]
	}

	return newConnection(w, r, false, readBuf, writeBuf)
}

func NewLeader(w http.ResponseWriter, r *http.Request, bufSize ...int) (*Client, error) {
	readBuf := 1024
	writeBuf := 1024
	if len(bufSize) == 1 {
		readBuf = bufSize[0]
	} else if len(bufSize) >= 2 {
		readBuf = bufSize[0]
		writeBuf = bufSize[1]
	}

	return newConnection(w, r, true, readBuf, writeBuf)
}

func newConnection(w http.ResponseWriter, r *http.Request, leader bool, readbufsize, writebufsize int) (*Client, error) {

	ws, err := websocket.Upgrade(w, r, nil, readBufsize, writeBufsize)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return nil, err
	} else if err != nil {
		return nil, err
	}

	c := &Client{send: make(chan []byte, 256), ws: ws, isLeader: leader}
	return c, nil
}

// Write a message back to the client
func (c *Client) Write(msg []byte) {
	c.send <- msg
}

// Write a string message to the client
func (c *Client) WriteString(msg string) {
	c.send <- []byte(msg)
}

// Write a JSON message to the client
func (c *Client) WriteJSON(msg interface{}) {
	msgJSON, _ := json.Marshal(msg)
	c.send <- msgJSON
}

// Read a message from the websocket connection, wait untill you get a message
func (c *Client) Read() ([]byte, error) {
	_, msg, err := c.ws.ReadMessage()
	return msg, err
}

func (c *Client) ReadString() (string, error) {
	_, msg, err := c.ws.ReadMessage()
	return string(msg), err
}

func (c *Client) ReadJSON(bean interface{}) error {
	err := c.ws.ReadJSON(bean)
	return err
}

// Read a message, but give up after a given timeout
// returns nil if the timeout is hit
func (c *Client) ReadTimeout(timeout time.Duration) ([]byte, error) {
	t := time.After(timeout)

	// Spin off a goroutine that reads a message from the ws, and send it down a channel
	ch := make(chan []byte)
	errchan := make(chan error)
	go func() {
		m, err := c.Read()
		if err != nil {
			errchan <- err
		}
		ch <- m
	}()

	select {
	case <-t: // We hit a timeout before getting a message from the websocket conn
		return nil, TimeoutErr
	case msg := <-ch: // We got a message from the websocket before the timeout
		return msg, nil
	case err := <-errchan: // an error occured
		return nil, err
	}
}

// Starts the client listening handler to write messages to
func (c *Client) Start() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}
