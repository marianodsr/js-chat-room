package server

import (
	"fmt"

	"github.com/marianodsr/jobsity-chat-room/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBMessage struct {
	gorm.Model
	Sender  string `json:"sender" db:"sender"`
	Payload string `json:"payload" db:"payload"`
	Room    string `json:"room" db:"room"`
}

func (m DBMessage) TableName() string {
	return "messages"
}

type MessageRepository interface {
	PersistMessages(messages []DBMessage)
	GetLatestMessagesForRoom(room string) ([]DBMessage, error)
}

type PostgresDB struct {
	*gorm.DB
}

func NewPostgresDB() *PostgresDB {
	return &PostgresDB{
		db.GetDB(),
	}
}

func (pg *PostgresDB) PersistMessages(messages []DBMessage) {
	if tx := pg.Create(&messages); tx.Error != nil {
		fmt.Println("error inserting messages to db")
		return
	}
}

func (pg *PostgresDB) GetLatestMessagesForRoom(room string) ([]DBMessage, error) {
	var messages []DBMessage
	if tx := pg.Where(&DBMessage{Room: room}).Order(clause.OrderByColumn{
		Column: clause.Column{
			Name: "room",
		},
		Desc: false,
	}).Find(&messages); tx.Error != nil {
		return nil, fmt.Errorf("error retrieving messages")
	}

	return messages, nil

}
