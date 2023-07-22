package main

import (
	"fmt"
	"github.com/Karaoke-Manager/karman/internal/api"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	Run:   runServer,
}

type Config struct {
	Address string
	Prefix  string
}

var defaultConfig = &Config{
	Address: ":8080",
	Prefix:  "/api",
}

func runServer(cmd *cobra.Command, args []string) {
	// TODO: Config management, maybe with Viper
	// TODO: Proper error handling on startup
	db, err := gorm.Open(sqlite.Open("test.db"))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}()

	uploadFS := rwfs.DirFS("tmp/uploads")
	uploadSvc := upload.NewService(db, uploadFS)

	// songFS := rwfs.DirFS("tmp/data")
	songSvc := song.NewService(db)
	apiController := api.NewController(songSvc, uploadSvc)

	r := chi.NewRouter()
	r.Route(defaultConfig.Prefix+"/", apiController.Router)
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(http.ListenAndServe(defaultConfig.Address, r))
}