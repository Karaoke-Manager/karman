package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgxutil"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/cmd/karman/internal"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	Args:  cobra.NoArgs,
	RunE:  runServer,
}

func init() {
	serverCmd.Flags().StringP("address", "l", ":8080", "The address on which the server listens for HTTP requests.")
	viper.SetDefault("api.address", ":8080")
	_ = viper.BindPFlag("api.address", serverCmd.Flag("address"))

	serverCmd.Flags().IntP("workers", "w", 2*runtime.NumCPU(), "Number of workers for processing background tasks.")
	viper.SetDefault("task-server.workers", 2*runtime.NumCPU())
	_ = viper.BindPFlag("task-server.workers", serverCmd.Flag("workers"))

	serverCmd.Flags().String("uploads-dir", "/usr/local/share/karman/uploads", "Directory in which uploads will be stored.")
	viper.SetDefault("uploads.dir", "/usr/local/share/karman/uploads")
	_ = viper.BindPFlag("uploads.dir", serverCmd.Flag("uploads-dir"))

	serverCmd.Flags().String("media-dir", "/usr/local/share/karman/media", "Directory in which media files will be stored.")
	viper.SetDefault("media.dir", "/usr/local/share/karman/media")
	_ = viper.BindPFlag("media.dir", serverCmd.Flag("media-dir"))

	rootCmd.AddCommand(serverCmd)
}

func runServer(_ *cobra.Command, _ []string) (rErr error) {
	db, closeFn, err := setupDatabase()
	if err != nil {
		return err
	}
	defer closeFn()

	logger.Info("Setting up application services")
	uploadStore, err := upload.NewFileStore(config.Uploads.Dir)
	if err != nil {
		logger.Error("Could not initialize upload storage", tint.Err(err))
		return fmt.Errorf("initializing upload storage: %w", err)
	}
	logger.Debug("Set up upload storage", "dir", uploadStore.Root())
	mediaStore, err := media.NewFileStore(config.Media.Dir)
	if err != nil {
		logger.Error("Could not initialize media storage", tint.Err(err))
		return fmt.Errorf("initializing media storage: %w", err)
	}
	logger.Debug("Set up media storage", "dir", mediaStore.Root())
	songRepo := song.NewDBRepository(db)
	songSvc := song.NewService()
	uploadRepo := upload.NewDBRepository(db)
	mediaRepo := media.NewDBRepository(db)
	mediaService := media.NewService(media.NewDBRepository(db), mediaStore)

	redis, closeFn, err := setupRedis()
	if err != nil {
		return err
	}
	defer closeFn()

	_, closeFn = setupAsynqClient(redis)
	defer closeFn()

	serverCtx, closeFn, err := setupTaskServer(redis, mediaRepo, mediaStore)
	if err != nil {
		return err
	}
	defer closeFn()

	schedulerCtx, closeFn, err := setupTaskScheduler(redis)
	if err != nil {
		return err
	}
	defer closeFn()

	logger.Info("Starting HTTP server", "address", config.API.Address)
	server := &http.Server{
		Addr:              config.API.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           api.NewHandler(songRepo, songSvc, mediaService, mediaStore, uploadRepo, uploadStore),
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-sigs:
			logger.Warn("Received stop signal. Shutting down", "signal", sig)
		case <-serverCtx.Done():
			logger.Warn("Fatal error in task server. Shutting down")
		case <-schedulerCtx.Done():
			logger.Warn("Fatal error in task scheduler. Shutting down")
		}
		logger.Info("Stopping HTTP server")
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		err := server.Shutdown(ctx)
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("HTTP server did not shut down for 30 seconds. Terminating forcefully.")
		} else if err != nil {
			logger.Error("Error during HTTP server shutdown", tint.Err(err))
		}
		err = server.Close()
		if err != nil {
			logger.Error("Error closing HTTP server", tint.Err(err))
		}
		cancel()
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Could not start HTTP server", tint.Err(err))
		return fmt.Errorf("starting HTTP server: %w", err)
	}
	return nil
}

func setupDatabase() (pgxutil.DB, func(), error) {
	logger.Info("Setting up database connection pool")
	db, err := pgxpool.New(context.Background(), config.DBConnection)
	if err != nil {
		logger.Error("Could not setup database connection pool", tint.Err(err))
		return nil, nil, fmt.Errorf("creating database connection: %w", err)
	}
	conf := db.Config().ConnConfig
	logger.Debug("Database connection", "host", conf.Host, "port", conf.Port, "user", conf.User, "database", conf.Database)
	return db, func() {
		logger.Info("Closing database connections")
		db.Close()
	}, nil
}

func setupRedis() (asynq.RedisConnOpt, func(), error) {
	logger.Info("Starting embedded redis server")
	redis := miniredis.NewMiniRedis()
	if err := redis.Start(); err != nil {
		logger.Error("Could not start embedded redis server", tint.Err(err))
		return nil, nil, fmt.Errorf("starting embedded redis server: %w", err)
	}
	return asynq.RedisClientOpt{Addr: redis.Addr()}, func() {
		logger.Info("Stopping embedded redis server")
		redis.Close()
	}, nil
}

func setupAsynqClient(redis asynq.RedisConnOpt) (*asynq.Client, func()) {
	logger.Info("Setting up asynq client")
	client := asynq.NewClient(redis)
	return client, func() {
		logger.Info("Closing asynq client")
		if err := client.Close(); err != nil {
			logger.Error("Could not close asynq client", tint.Err(err))
		}
	}
}

func setupTaskServer(redis asynq.RedisConnOpt, mediaRepo media.Repository, mediaStore media.Store) (context.Context, func(), error) {
	// we expect tasks to not be primarily CPU bound but mostly IO bound by the database
	logger.Info("Starting task server", "workers", config.TaskServer.Workers)
	serverLogger := internal.NewAsynqLogger(logger, "task-server")
	var logLevel asynq.LogLevel
	if config.Log.Level >= slog.LevelWarn {
		logLevel = asynq.ErrorLevel
	} else if config.Log.Level >= slog.LevelInfo {
		logLevel = asynq.WarnLevel
	} else if config.Log.Level >= slog.LevelDebug {
		logLevel = asynq.InfoLevel
	} else {
		logLevel = asynq.DebugLevel
	}
	srv := asynq.NewServer(redis, asynq.Config{
		Concurrency: config.TaskServer.Workers,
		Queues: map[string]int{
			"default":       1,
			media.TaskQueue: 1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			if retried >= maxRetry {
				logger.ErrorContext(ctx, "Retry Exhausted for task", "task", task.Type(), tint.Err(err))
			} else {
				logger.WarnContext(ctx, "Task failed. Will be retried shortly", "task", task.Type(), "retried", retried, "maxRetry", maxRetry, tint.Err(err))
			}
		}),
		Logger:          serverLogger,
		LogLevel:        logLevel,
		HealthCheckFunc: nil, // TODO: record errors and provide then to a /healthz endpoint.
	})
	mux := asynq.NewServeMux()
	mux.Handle(media.NewTaskHandler(mediaRepo, mediaStore))
	if err := srv.Start(mux); err != nil {
		logger.Error("Could not start task server", tint.Err(err))
		return nil, nil, fmt.Errorf("starting task server: %w", err)
	}
	return serverLogger.Context(), func() {
		logger.Info("Stopping task server")
		srv.Shutdown()
	}, nil
}

func setupTaskScheduler(redis asynq.RedisConnOpt) (context.Context, func(), error) {
	logger.Info("Starting task scheduler")
	schedulerLogger := internal.NewAsynqLogger(logger, "cron")
	var logLevel asynq.LogLevel
	if config.Log.Level >= slog.LevelWarn {
		logLevel = asynq.ErrorLevel
	} else if config.Log.Level >= slog.LevelInfo {
		logLevel = asynq.WarnLevel
	} else if config.Log.Level >= slog.LevelDebug {
		logLevel = asynq.InfoLevel
	} else {
		logLevel = asynq.DebugLevel
	}
	periodicTaskConfigProvider := media.NewPeriodicTaskConfigProvider()
	scheduler, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		PeriodicTaskConfigProvider: periodicTaskConfigProvider,
		RedisConnOpt:               redis,
		SchedulerOpts: &asynq.SchedulerOpts{
			Logger:   schedulerLogger,
			LogLevel: logLevel,
			Location: time.Local,
			PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
				if err != nil {
					logger.Error("Could not enqueue task", "component", "cron", "task", info.Type, tint.Err(err))
				}
			},
		},
	})
	if err != nil {
		// This error only occurs if the task manager is not set up correctly.
		// This is a programmer error!
		panic(err)
	}
	if err = scheduler.Start(); err != nil {
		logger.Error("Could not start cron manager", tint.Err(err))
		return nil, nil, fmt.Errorf("starting cron manager: %w", err)
	}
	return schedulerLogger.Context(), func() {
		logger.Info("Stopping cron manager")
		scheduler.Shutdown()
	}, nil
}
