package main

import (
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"gorm.io/gorm"

	_ "github.com/Karaoke-Manager/karman/migrations"
	migratedb "github.com/Karaoke-Manager/karman/migrations/db"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migrations",
	Long:  "Run Karman schema migrations against your database.",
	Run:   runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) {
	// TODO: build proper CLI
	db, err := gorm.Open(sqlite.Open("test.db?_pragma=foreign_keys(1)"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migratedb.FS)
	migratedb.Set(db)

	if err := goose.SetDialect(db.Dialector.Name()); err != nil {
		log.Fatalf("goose: failed to set dialect")
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		log.Printf("%#v", err)
		log.Fatalf("goose up: %v", err)
	}
}
