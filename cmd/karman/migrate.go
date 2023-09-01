package main

import (
	"log"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Karaoke-Manager/karman/migrations"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migrations",
	Long:  "Run Karman schema migrations against your database.",
	Run:   runMigrate,
}

var (
	connString string
)

func init() {
	migrateCmd.Flags().StringVarP(&connString, "conn", "c", "postgres://karman:secret@localhost:5432/karman?sslmode=disable", "Connection String to PostgreSQL database")
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	// TODO: build proper CLI
	goose.SetBaseFS(migrations.FS)
	db, err := goose.OpenDBWithDriver("pgx", connString)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %s", err)
	}
	defer db.Close()

	if err := goose.Up(db, "."); err != nil {
		log.Printf("%#v", err)
		log.Fatalf("goose up: %v", err)
	}
}
