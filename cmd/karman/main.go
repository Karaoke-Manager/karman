package main

import (
	"fmt"
	"github.com/Karaoke-Manager/karman/internal/api"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Address string
	Prefix  string
}

var defaultConfig = &Config{
	Address: ":8080",
	Prefix:  "/api",
}

func main() {
	// TODO: Config management, maybe with Viper
	// TODO: Proper error handling on startup
	db, err := gorm.Open(sqlite.Open("test.db"))
	if err != nil {
		log.Fatalln(err)
	}

	uploadFS := os.DirFS("tmp/uploads")

	apiServer := api.NewServer(db, uploadFS)

	r := chi.NewRouter()
	r.Route(defaultConfig.Prefix+"/", apiServer.Router)
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(http.ListenAndServe(defaultConfig.Address, r))
}
