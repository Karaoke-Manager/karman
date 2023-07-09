package main

import (
	"github.com/pressly/goose/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	_ "modernc.org/sqlite"
	// TODO: Add other drivers.

	_ "github.com/Karaoke-Manager/karman/migrations"
	gormdb "github.com/Karaoke-Manager/karman/migrations/db"
)

func main() {
	// TODO: Build proper CLI
	db, err := gorm.Open(sqlite.Open("test.db"))
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	defer sqlDB.Close()

	gormdb.Set(db)

	command := "up"
	if err := goose.Run(command, sqlDB, ""); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
