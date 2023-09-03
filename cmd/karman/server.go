package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
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
}

var defaultConfig = &Config{
	Address: ":8080",
}

func runServer(_ *cobra.Command, _ []string) (err error) {
	// TODO: Config management, maybe with Viper
	// TODO: Proper error handling on startup

	pool, err := pgxpool.New(context.Background(), connStringServer)
	if err != nil {
		return err
	}
	defer pool.Close()

	redis := miniredis.NewMiniRedis()
	if err = redis.Start(); err != nil {
		return err
	}
	redisConn := asynq.RedisClientOpt{Addr: redis.Addr()}
	client := asynq.NewClient(redisConn)
	defer func() {
		if cErr := client.Close(); err == nil {
			err = cErr
		}
	}()

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
	mediaRepo := media.NewDBRepository(pool)
	mediaService := media.NewService(media.NewDBRepository(pool), mediaStore)

	srv := asynq.NewServer(redisConn, asynq.Config{
		// we expect tasks to not be primarily CPU bound but mostly IO bound by the database
		Concurrency: 2 * runtime.NumCPU(),
		Queues: map[string]int{
			"default":       1,
			media.TaskQueue: 1,
		},
		ErrorHandler:    nil, // probably error logging
		Logger:          nil,
		HealthCheckFunc: nil, // record errors and provide then to a /healthz endpoint.
	})
	mux := asynq.NewServeMux()
	mux.Handle(media.NewTaskHandler(mediaRepo, mediaStore))
	// FIXME: this must run in background
	if err = srv.Start(mux); err != nil {
		log.Fatalln(err)
	}

	periodicTaskConfigProvider := media.NewPeriodicTaskConfigProvider()
	scheduler, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		PeriodicTaskConfigProvider: periodicTaskConfigProvider,
		RedisConnOpt:               redisConn,
		SchedulerOpts: &asynq.SchedulerOpts{
			Logger:          nil,
			Location:        nil, // probably use TZ environment
			PostEnqueueFunc: nil, // probably error logging
		},
	})
	// FIXME: This must run in background
	if err = scheduler.Start(); err != nil {
		log.Fatalln(err)
	}

	apiHandler := api.NewHandler(songRepo, songSvc, mediaService, mediaStore, uploadRepo, uploadStore)
	server := &http.Server{
		Addr:              defaultConfig.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           apiHandler,
	}
	// TODO: gracefully terminate server on Ctrl-C
	fmt.Printf("Running on %s\n", defaultConfig.Address)
	log.Fatalln(server.ListenAndServe())
	return nil
}
