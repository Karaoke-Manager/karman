package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/service/media"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	RunE:  runServer,
}

type Config struct {
	Address string
	Prefix  string
}

var defaultConfig = &Config{
	Address: ":8080",
	Prefix:  "/api",
}

func runServer(cmd *cobra.Command, args []string) error {
	// TODO: Config management, maybe with Viper
	// TODO: Proper error handling on startup
	dbConfig, err := pgxpool.ParseConfig("postgres://karman:secret@localhost:5432/karman?sslmode=disable")
	if err != nil {
		return err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return err
	}
	defer pool.Close()

	mediaStore, err := media.NewFileStore("tmp/media")
	if err != nil {
		return err
	}
	mediaSvc := media.NewService(media.NewDB(pool), mediaStore)

	// uploadFS := rwfs.DirFS("tmp/uploads")
	// uploadSvc := upload.NewService(db, uploadFS)

	apiController := api.NewController(songSvc, mediaSvc, nil)

	r := chi.NewRouter()
	r.Route(defaultConfig.Prefix+"/", apiController.Router)
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(http.ListenAndServe(defaultConfig.Address, r))
	return nil
}
