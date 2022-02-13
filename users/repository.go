package users

import (
	"errors"
	"fmt"

	"github.com/marianodsr/jobsity-chat-room/db"
	"gorm.io/gorm"
)

type UserRepository interface {
	createUser(user User) (*User, error)
	getUserByEmail(email string) (*User, error)
}

type PostgresDB struct {
	*gorm.DB
}

func NewPostgresDB() *PostgresDB {
	return &PostgresDB{
		db.GetDB(),
	}
}

func (pg *PostgresDB) createUser(user User) (*User, error) {
	tx := pg.Create(&user)
	if tx.Error != nil {
		return nil, errors.New("error creating user")
	}

	return &user, nil
}

func (pg *PostgresDB) getUserByEmail(email string) (*User, error) {
	var user *User
	tx := pg.Where(&User{Email: email}).First(&user)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return nil, tx.Error
	}
	return user, nil
}
