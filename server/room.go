package server

import (
	"fmt"
	"sync"
	"time"
)

type room struct {
	id                 string
	members            map[string]*Client
	joiners            chan *Client
	leavers            chan *Client
	capacity           uint
	messages           chan Message
	toBeStoredMessages []Message
	repo               MessageRepository
	mu                 *sync.Mutex
}

//listens to joiners and leavers channels and handles joining and leaving a room
func (r *room) handleJoinersAndLeavers() {
	for {
		select {
		case c := <-r.joiners:
			_, ok := r.members[c.id]
			if ok {
				c.sendMessage(ERROR, fmt.Sprintf("you are already on room %s\n", r.id), c.id)
				continue
			}
			if len(r.members) >= int(r.capacity) {
				c.sendMessage(ERROR, fmt.Sprintf("room %s is full, please try again later\n", r.id), c.id)
				continue
			}

			if c.room != nil {
				c.sendMessage(ERROR, "in order to join a new room you first need to leave your current one", c.id)
				continue
			}
			r.members[c.id] = c
			c.room = r
			c.sendMessage(JOIN, r.id, "INFO")
			r.broadcast(Message{
				Sender:  c,
				Header:  INFO,
				Payload: fmt.Sprintf("client %s has joined the room", c.id),
			}, "INFO")
		case c := <-r.leavers:
			fmt.Println("READING FROM LEAVERS CHANNEL IN ROOM")
			_, ok := r.members[c.id]
			if !ok {
				//user is not in the room
				continue
			}
			delete(r.members, c.id)
			c.room = nil
			c.sendMessage(LEAVE, r.id, "INFO")
			r.broadcast(Message{
				Sender:  c,
				Header:  INFO,
				Payload: fmt.Sprintf("client %s has left the room", c.id),
			}, "INFO")
		case m := <-r.messages:
			r.mu.Lock()
			r.toBeStoredMessages = append(r.toBeStoredMessages, m)
			r.mu.Unlock()
			r.broadcast(m, m.Sender.id)
		}
	}
}

func (r *room) broadcast(m Message, sender string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, client := range r.members {
		client.sendMessage(MESSAGE, m.Payload, sender)
	}
}

//Meant to be run as a goroutine for each room. Should check periodically for buffered messages and store them
func (r *room) PersistMessages() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		<-ticker.C
		r.mu.Lock()
		if len(r.toBeStoredMessages) > 0 {
			var dbMessages []DBMessage
			for _, msg := range r.toBeStoredMessages {
				dbMessages = append(dbMessages, DBMessage{
					Sender:  msg.Sender.id,
					Payload: msg.Payload,
					Room:    r.id,
				})
			}
			r.repo.PersistMessages(dbMessages)
			r.toBeStoredMessages = nil
		}
		r.mu.Unlock()
	}
}
