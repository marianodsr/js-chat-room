package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/marianodsr/jobsity-chat-room/server"
)

func HandleRoutes(s *server.Server) {
	pgDB := NewPostgresDB()
	service := NewUserService(pgDB)

	s.Router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var user User
		if err := decoder.Decode(&user); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		_, err := mail.ParseAddress(user.Email)
		if err != nil {
			http.Error(w, "invalid email format", http.StatusBadRequest)
			return
		}

		created, err := service.createUser(user)
		if err != nil {
			fmt.Println(err)
			http.Error(w, fmt.Sprintf("error processing request: %s", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(created)
	})
	s.Router.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		var user User
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		loggedInUser, err := service.login(user.Email, user.Password)
		if err != nil {
			fmt.Println(err)
			http.Error(w, fmt.Sprintf("error processing request: %s", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(loggedInUser)
	})
}
