package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	RunE:  runServer,
}

var (
	connStringServer string
)

func init() {
	serverCmd.Flags().StringVarP(&connStringServer, "conn", "c", "postgres://karman:secret@localhost:5432/karman?sslmode=disable", "Connection String to PostgreSQL database")
	rootCmd.AddCommand(serverCmd)
}

type Config struct {
	Address string
	Prefix  string
}

var defaultConfig = &Config{
	Address: ":8080",
	Prefix:  "/api",
}

func runServer(_ *cobra.Command, _ []string) error {
	// TODO: Config management, maybe with Viper
	// TODO: Proper error handling on startup
	dbConfig, err := pgxpool.ParseConfig(connStringServer)
	if err != nil {
		return err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return err
	}
	defer pool.Close()

	// TODO: Check DB Connection before startup

	songRepo := song.NewDBRepository(pool)
	songSvc := song.NewService()
	uploadRepo := upload.NewDBRepository(pool)
	uploadStore, err := upload.NewFileStore("tmp/uploads")
	if err != nil {
		return err
	}
	mediaStore, err := media.NewFileStore("tmp/media")
	if err != nil {
		return err
	}
	mediaService := media.NewService(media.NewDBRepository(pool), mediaStore)

	apiController := api.NewController(songRepo, songSvc, mediaService, mediaStore, uploadRepo, uploadStore)

	r := chi.NewRouter()
	r.Route(defaultConfig.Prefix+"/", apiController.Router)
	server := &http.Server{
		Addr:              defaultConfig.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           r,
	}
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(server.ListenAndServe())
	return nil
}
