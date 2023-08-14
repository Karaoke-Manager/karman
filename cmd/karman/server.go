package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
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
	db, err := gorm.Open(sqlite.Open("test.db?_pragma=foreign_keys(1)"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}()

	songSvc := song.NewService(db)

	mediaStore, err := media.NewFileStore("tmp/media")
	if err != nil {
		log.Fatalln(err)
	}
	mediaSvc := media.NewService(db, mediaStore)

	// uploadFS := rwfs.DirFS("tmp/uploads")
	// uploadSvc := upload.NewService(db, uploadFS)

	apiController := api.NewController(songSvc, mediaSvc, nil)

	r := chi.NewRouter()
	r.Route(defaultConfig.Prefix+"/", apiController.Router)
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(http.ListenAndServe(defaultConfig.Address, r))
}
