package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Karaoke-Manager/karman/migrations"
)

// migrateCmd represents the "migrate" subcommand.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migrations",
	Long:  "Run Karman schema migrations against your database.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runMigrate,
}

// init sets up command line flags for the "migrate" command.
func init() {
	migrateCmd.Flags().BoolVarP(&status, "status", "s", false, "Show current migration status.")
	migrateCmd.Flags().BoolVar(&allowMissing, "allow-missing", false, "Applies missing (out-of-order) migrations.")
	rootCmd.AddCommand(migrateCmd)

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("pgx"); err != nil {
		// This error only occurs for an unsupported dialect.
		// This is a programmer error!
		panic(err)
	}
}

var (
	// allowMissing indicates whether migrations can be applied out of order.
	allowMissing bool
	// status indicates whether the --status flag was set.
	status bool
)

// runMigrate executes the "migrate" command.
func runMigrate(_ *cobra.Command, args []string) (rErr error) {
	// TODO: The CLI could probably be more consistent...
	goose.SetLogger(log.New(os.Stdout, "", 0))
	db, err := sql.Open("pgx", config.DBConnection)
	if err != nil {
		// This error indicates an unsupported or invalid driver.
		// This is a programmer error!
		panic(err)
	}
	defer func() {
		if cErr := db.Close(); rErr == nil {
			rErr = cErr
		}
	}()

	if status {
		return goose.Status(db, ".")
	}

	var opts []goose.OptionsFunc
	if allowMissing {
		opts = []goose.OptionsFunc{goose.WithAllowMissing()}
	}
	if len(args) == 0 {
		return goose.Up(db, ".", opts...)
	}
	targetStr := args[0]
	target, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid argument. %s is not a valid integer", targetStr)
	}
	current, err := goose.GetDBVersion(db)
	if errors.Is(err, goose.ErrNoCurrentVersion) {
		current = 0
	} else if err != nil {
		return fmt.Errorf("could not fetch current migration state: %w", err)
	}
	if strings.HasPrefix(targetStr, "+") || strings.HasPrefix(targetStr, "-") {
		target = target + current
	}
	if target >= current {
		return goose.UpTo(db, ".", target, opts...)
	}
	return goose.DownTo(db, ".", target, opts...)
}
