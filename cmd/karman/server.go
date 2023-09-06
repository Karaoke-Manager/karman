package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/cmd/karman/internal"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// serverCmd implements the "server" command.
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	Args:  cobra.NoArgs,
	RunE:  runServer,
}

// init sets up the command line flags for the server command.
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

// runServer starts the server and all of its components.
func runServer(_ *cobra.Command, _ []string) (rErr error) {
	closeFn, err := setupDatabase()
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

	closeFn, err = setupRedis()
	if err != nil {
		return err
	}
	defer closeFn()

	closeFn = setupTaskQueue(redisConn)
	defer closeFn()

	serverCtx, closeFn, err := setupTaskRunner(mediaRepo, mediaStore)
	if err != nil {
		return err
	}
	defer closeFn()

	schedulerCtx, closeFn, err := setupTaskScheduler()
	if err != nil {
		return err
	}
	defer closeFn()

	logger.Info("Starting HTTP server", "address", config.API.Address)
	server := &http.Server{
		Addr:              config.API.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           api.NewHandler(logger, api.HealthCheckFunc(healthcheck), songRepo, songSvc, mediaService, mediaStore, uploadRepo, uploadStore),
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
	// there is no way the server terminates by itself without some kind of error.
	// some errors are only logged so we return a catch-all error here.
	return errors.New("")
}

var (
	// db is the database connection pool.
	db *pgxpool.Pool
	// redisConn defines how to connect to redis
	redisConn asynq.RedisConnOpt
	// taskQueue is the background task service
	taskQueue *asynq.Client
	// taskRunner is the background task processing sever.
	taskRunner *asynq.Server
	// taskServerHealth is the last result of the taskRunner healthcheck.
	taskServerHealth error
)

// healthcheck performs a health check on all system components
func healthcheck(ctx context.Context) bool {
	// TODO: Check migrations in database
	redisClient := redisConn.MakeRedisClient().(redis.UniversalClient)
	redisErr := redisClient.Ping(ctx).Err()
	if err := redisClient.Close(); err != nil {
		// The connection will probably be closed by the asynq client before this is called.
		logger.Error("Could not close redis connection", tint.Err(err))
		if redisErr == nil {
			redisErr = err
		}
	}
	if redisErr != nil {
		logger.Error("Redis health check failed", tint.Err(redisErr))
		return false
	}
	if err := db.Ping(ctx); err != nil {
		logger.Error("Database health check failed", tint.Err(redisErr))
		return false
	}
	if taskServerHealth != nil {
		logger.Error("Task server health check failed", tint.Err(redisErr))
		return false
	}
	return true
}

// setupDatabase create a database connection pool.
func setupDatabase() (func(), error) {
	logger.Info("Setting up database connection pool")
	var err error
	db, err = pgxpool.New(context.Background(), config.DBConnection)
	if err != nil {
		logger.Error("Could not setup database connection pool", tint.Err(err))
		return nil, fmt.Errorf("creating database connection: %w", err)
	}
	conf := db.Config().ConnConfig
	logger.Debug("Database connection", "host", conf.Host, "port", conf.Port, "user", conf.User, "database", conf.Database)
	return func() {
		logger.Info("Closing database connections")
		db.Close()
	}, nil
}

// setupRedis either starts the embedded redis instance or prepares a connection to an external redis.
func setupRedis() (func(), error) {
	// TODO: Support external redis
	logger.Info("Starting embedded redis server")
	miniRedis := miniredis.NewMiniRedis()
	if err := miniRedis.Start(); err != nil {
		logger.Error("Could not start embedded redis server", tint.Err(err))
		return nil, fmt.Errorf("starting embedded redis server: %w", err)
	}

	redisConn = asynq.RedisClientOpt{Addr: miniRedis.Addr()}
	return func() {
		logger.Info("Stopping embedded redis server")
		miniRedis.Close()
	}, nil
}

// setupTaskQueue sets up the asynq.Client for enqueuing tasks.
func setupTaskQueue(redis asynq.RedisConnOpt) func() {
	logger.Info("Setting up task queue")
	taskQueue = asynq.NewClient(redis)
	return func() {
		logger.Info("Closing task queue")
		if err := taskQueue.Close(); err != nil {
			logger.Error("Could not close task queue", tint.Err(err))
		}
	}
}

// setupTaskRunner sets up the task runner that executes background tasks.
func setupTaskRunner(mediaRepo media.Repository, mediaStore media.Store) (context.Context, func(), error) {
	logger.Info("Starting task runner", "workers", config.TaskRunner.Workers)
	taskRunnerLogger := internal.NewAsynqLogger(logger, "runner")
	taskRunner = asynq.NewServer(redisConn, asynq.Config{
		Concurrency: config.TaskRunner.Workers,
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
		Logger:   taskRunnerLogger,
		LogLevel: internal.AsynqLogLevel(config.Log.Level),
		HealthCheckFunc: func(err error) {
			taskServerHealth = err
		},
	})
	mux := asynq.NewServeMux()
	mux.Handle(media.NewTaskHandler(mediaRepo, mediaStore))
	if err := taskRunner.Start(mux); err != nil {
		logger.Error("Could not start task runner", tint.Err(err))
		return nil, nil, fmt.Errorf("starting task server: %w", err)
	}
	return taskRunnerLogger.Context(), func() {
		logger.Info("Stopping task runner")
		taskRunner.Shutdown()
	}, nil
}

// setupTaskScheduler sets up the task scheduler that creates specific task instances for scheduled tasks.
func setupTaskScheduler() (context.Context, func(), error) {
	logger.Info("Starting task scheduler")
	taskSchedulerLogger := internal.NewAsynqLogger(logger, "scheduler")
	schedulerConfig := internal.MergePeriodicTaskConfigProviders(
		media.NewPeriodicTaskConfigProvider(),
	)
	scheduler, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		PeriodicTaskConfigProvider: schedulerConfig,
		RedisConnOpt:               redisConn,
		SchedulerOpts: &asynq.SchedulerOpts{
			Logger:   taskSchedulerLogger,
			LogLevel: internal.AsynqLogLevel(config.Log.Level),
			Location: time.Local,
			PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
				if err != nil {
					logger.Error("Could not enqueue scheduled task", "component", "scheduler", "task", info.Type, tint.Err(err))
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
		logger.Error("Could not start task scheduler", tint.Err(err))
		return nil, nil, fmt.Errorf("starting cron manager: %w", err)
	}
	return taskSchedulerLogger.Context(), func() {
		logger.Info("Stopping task scheduler")
		scheduler.Shutdown()
	}, nil
}
