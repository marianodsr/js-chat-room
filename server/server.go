package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type Server struct {
	Router      Router
	hub         *hub
	MessageRepo MessageRepository
}

type Router interface {
	Get(path string, handlerFunc http.HandlerFunc)
	Post(path string, handlerFunc http.HandlerFunc)
}

func NewServer(router Router, receivedStocks <-chan amqp.Delivery) *Server {
	return &Server{
		Router:      router,
		hub:         NewHub(receivedStocks),
		MessageRepo: NewPostgresDB(),
	}
}

//HandleRoutes manages the endpoints for the application.
func (s *Server) ServeWebsocket() {
	s.Router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("err upgrading websocket connection: ", err)
			w.Write([]byte("server was not able to establish a websocket connection"))
			return
		}

		client := NewClient(conn, s.hub)
		fmt.Printf("Client: %s connected\n", client.id)

		//Spawns a goroutine for each client
		go client.ListenAndProcess()
	})

	s.Router.Get("/messages", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		room := params.Get("room")
		if room == "" {
			http.Error(w, "invalid param, should be in the form of ?room={room}", http.StatusBadRequest)
			return
		}

		messages, err := s.MessageRepo.GetLatestMessagesForRoom(room)
		if err != nil {
			http.Error(w, "error retrieving messages", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(messages)
	})
}
