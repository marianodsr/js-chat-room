package server

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

const RETRY_POLICY = 3

type Client struct {
	id   string
	Conn *websocket.Conn
	hub  *hub
	room *room
}

func NewClient(conn *websocket.Conn, hub *hub) *Client {
	return &Client{
		id:   conn.RemoteAddr().String(),
		Conn: conn,
		hub:  hub,
		room: nil,
	}
}

//Listen for client messages
func (c *Client) ListenAndProcess() {
	defer func() {
		if c.room != nil {
			c.room.leavers <- c
		}
		c.hub.leavers <- c
		c.Conn.Close()
	}()
	c.hub.mu.Lock()
	c.hub.connections[c.Conn.RemoteAddr().String()] = *c
	c.hub.mu.Unlock()
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.Sender = c
		c.dispatchMessage(msg)
	}
}

func (c *Client) dispatchMessage(msg Message) {
	switch msg.Header {
	case JOIN:
		c.hub.joiners <- msg
	case LEAVE:
		c.room.leavers <- c
	case SET_USERNAME:
		c.id = msg.Payload
	case STOCK:
		c.hub.stocks <- msg
	default:
		if c.room != nil {
			c.room.messages <- msg
			break
		}
		c.sendMessage(ERROR, "in order to send messages, please first join a room", c.id)
	}
}

func (c *Client) sendMessage(header Event, payload string, sender string) {
	retries := RETRY_POLICY

	if header == ERROR {
		sender = "ERROR"
	}

	if header == INFO {
		sender = "ERROR"
	}
	msg := map[string]string{
		"sender":  sender,
		"header":  string(header),
		"payload": payload,
	}

	encoded, _ := json.Marshal(msg)

	for i := 0; i < retries; i++ {
		err := c.Conn.WriteMessage(websocket.TextMessage, encoded)
		if err != nil {
			break
		}
	}
}
