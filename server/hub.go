package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/streadway/amqp"
)

const DEFAULT_CAPACITY = 5

type hub struct {
	mu             *sync.Mutex
	rooms          map[string]*room
	connections    map[string]Client //map key should be remote address of client
	joiners        chan Message
	leavers        chan *Client
	stocks         chan Message
	receivedStocks <-chan amqp.Delivery
}

//meant to be a singleton
func NewHub(brokerChann <-chan amqp.Delivery) *hub {
	hub := &hub{
		mu:             &sync.Mutex{},
		rooms:          make(map[string]*room),
		connections:    make(map[string]Client),
		joiners:        make(chan Message),
		leavers:        make(chan *Client),
		stocks:         make(chan Message),
		receivedStocks: brokerChann,
	}

	go hub.Run()
	return hub
}

//meant to be run as a goroutine and should dynamically create rooms
func (h *hub) Run() {

	go h.runStockBot()

	for {
		select {
		case m := <-h.joiners:
			h.mu.Lock()
			//check rooms to see if requested room already exists
			if room, ok := h.rooms[m.Payload]; ok {
				room.joiners <- m.Sender
				h.mu.Unlock()
				continue
			}
			room := h.NewRoom(DEFAULT_CAPACITY, m.Payload)

			//Start room worker
			go room.handleJoinersAndLeavers()
			//Start room message persister worker
			go room.PersistMessages()

			h.rooms[m.Payload] = room
			h.mu.Unlock()
			room.joiners <- m.Sender
		case c := <-h.leavers:
			h.mu.Lock()
			delete(h.connections, c.Conn.RemoteAddr().String())
			fmt.Printf("Client %s has left the hub\n", c.id)
			h.mu.Unlock()
		case m := <-h.stocks:
			h.getStockPrice(m)
		}

	}
}

func (h *hub) NewRoom(capacity uint, id string) *room {
	return &room{
		id:                 id,
		members:            make(map[string]*Client),
		joiners:            make(chan *Client),
		leavers:            make(chan *Client),
		capacity:           capacity,
		messages:           make(chan Message),
		toBeStoredMessages: make([]Message, 0),
		repo:               NewPostgresDB(),
		mu:                 &sync.Mutex{},
	}
}

func (h *hub) getStockPrice(msg Message) {
	symbol := msg.Payload
	sender := msg.Sender.Conn.RemoteAddr().String()
	http.Get(fmt.Sprintf("http://localhost:9000/%s?sender=%s", symbol, sender))
}

func (h *hub) runStockBot() {
	var rabbitMQMessage RabbitMQMessage
	for msg := range h.receivedStocks {
		buffer := bytes.NewReader(msg.Body)
		err := json.NewDecoder(buffer).Decode(&rabbitMQMessage)
		if err != nil {
			fmt.Println("error decoding msg")
			continue
		}

		h.mu.Lock()
		client, ok := h.connections[rabbitMQMessage.MsgFor]
		if !ok {
			h.mu.Unlock()
			continue
		}
		h.mu.Unlock()
		client.sendMessage(STOCK, rabbitMQMessage.Payload, "INFO")

	}
}
