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

	if err := goose.SetDialect(db.Dialector.Name()); err != nil {
		log.Fatalf("goose: failed to set dialect")
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Printf("%#v", err)
		log.Fatalf("goose up: %v", err)
	}
}
