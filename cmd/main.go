package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/marianodsr/jobsity-chat-room/db"
	"github.com/marianodsr/jobsity-chat-room/server"
	"github.com/marianodsr/jobsity-chat-room/users"
	"github.com/streadway/amqp"
)

const QUEUE_NAME = "stock_prices"

func main() {

	db.InitializeDB()
	DB := db.GetDB()

	DB.AutoMigrate(&users.User{}, &server.DBMessage{})

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	chann, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer chann.Close()

	q, err := chann.QueueDeclare(
		QUEUE_NAME,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	stockChann, err := chann.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	server := server.NewServer(r, stockChann)

	go http.ListenAndServe("localhost:8000", r)

	server.ServeWebsocket()

	users.HandleRoutes(server)

	for {
	}

}
