package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitializeDB() {
	dsn := "host=postgresDB-container user=postgres password=123 dbname=chat-room port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	pgDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db = pgDB
}

func GetDB() *gorm.DB {
	return db
}
