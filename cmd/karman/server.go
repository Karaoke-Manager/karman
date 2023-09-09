package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/cmd/karman/internal"
	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/core/upload"
	"github.com/Karaoke-Manager/karman/migrations"
)

// serverCmd implements the "server" command.
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Karman server",
	Long:  "The karman server runs the Karman backend API.",
	Args:  cobra.NoArgs,
	PreRunE: func(_ *cobra.Command, _ []string) (rErr error) {
		if !migrate {
			return nil
		}
		// TODO: Use PGX Driver directly
		// Requires: https://github.com/jackc/pgx/pull/1718
		mainLogger.Info("Running database migrations.")
		goose.SetLogger(log.Default())
		db, err := goose.OpenDBWithDriver("pgx", config.DBConnection)
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
		return goose.Up(db, ".")
	},
	RunE: runServer,
}

// init sets up the command line flags for the server command.
func init() {
	serverCmd.Flags().BoolVarP(&migrate, "migrate", "m", false, "Run database migrations before starting the server.")

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

var (
	// migrate indicates whether the --migrate flag was specified.
	migrate bool

	// db is the database connection pool.
	db *pgxpool.Pool
	// redisConn defines how to connect to redis.
	redisConn asynq.RedisConnOpt
	// taskQueue is the background task service.
	taskQueue *asynq.Client
	// taskRunner is the background task processing sever.
	taskRunner *asynq.Server
	// taskServerHealth is the last result of the taskRunner healthcheck.
	taskServerHealth error
)

// runServer starts the server and all of its components.
func runServer(_ *cobra.Command, _ []string) (rErr error) {
	goose.SetLogger(goose.NopLogger())

	closeFn, err := setupDatabase()
	if err != nil {
		return err
	}
	defer closeFn()

	mainLogger.Info("Setting up application services.")
	uploadStore, err := upload.NewFileStore(logger.With("log", "upload.store"), config.Uploads.Dir)
	if err != nil {
		mainLogger.Error("Could not initialize upload storage.", tint.Err(err))
		return fmt.Errorf("initializing upload storage: %w", err)
	}
	mainLogger.Debug(fmt.Sprintf("Upload storage initialized at %s.", uploadStore.Root()))
	mediaStore, err := media.NewFileStore(logger.With("log", "song.store"), config.Media.Dir)
	if err != nil {
		mainLogger.Error("Could not initialize media store.", tint.Err(err))
		return fmt.Errorf("initializing media storage: %w", err)
	}
	mainLogger.Debug(fmt.Sprintf("Media storage initialized at %s.", mediaStore.Root()))
	songRepo := song.NewDBRepository(logger.With("log", "song.repo"), db)
	songSvc := song.NewService()
	uploadRepo := upload.NewDBRepository(logger.With("log", "upload.repo"), db)
	mediaRepo := media.NewDBRepository(logger.With("log", "media.repo"), db)
	mediaService := media.NewService(logger.With("log", "song.service"), mediaRepo, mediaStore)

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

	// Run a healthcheck to log potential connection problems.
	_ = healthcheck(context.Background())

	mainLogger.Info(fmt.Sprintf("Running HTTP server on %s.", config.API.Address))
	server := &http.Server{
		Addr:              config.API.Address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           api.NewHandler(logger, api.HealthCheckFunc(healthcheck), songRepo, songSvc, mediaService, mediaStore, uploadRepo, uploadStore, config.Debug),
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-sigs:
			mainLogger.Warn(fmt.Sprintf("Stop signal %q received. Shutting down...", sig))
		case <-serverCtx.Done():
			mainLogger.Warn("Fatal error in task server. Shutting down...", tint.Err(serverCtx.Err()))
		case <-schedulerCtx.Done():
			mainLogger.Warn("Fatal error in task scheduler. Shutting down...", tint.Err(serverCtx.Err()))
		}
		mainLogger.Info("Stopping HTTP server with 30 second timeout.")
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		err := server.Shutdown(ctx)
		if errors.Is(err, context.DeadlineExceeded) {
			mainLogger.Error("HTTP server did not shut down for 30 seconds. Terminating forcefully.")
		} else if err != nil {
			mainLogger.Error("HTTP server shutdown caused an error.", tint.Err(err))
		}
		err = server.Close()
		if err != nil {
			mainLogger.Error("Could not close HTTP server.", tint.Err(err))
		}
		cancel()
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		mainLogger.Error("Could not start HTTP server.", tint.Err(err))
		return fmt.Errorf("starting HTTP server: %w", err)
	}
	// there is no way the server terminates by itself without some kind of error.
	// some errors are only logged, so we return a catch-all error here.
	return errors.New("stopped")
}

// healthcheck performs a health check on all system components.
func healthcheck(ctx context.Context) (ret bool) {
	redisClient := redisConn.MakeRedisClient().(redis.UniversalClient)
	defer func() {
		if err := redisClient.Close(); err != nil {
			mainLogger.Error("Could not close redis ping connection.", tint.Err(err))
			ret = false
		}
	}()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		mainLogger.Error("Redis health check failed.", tint.Err(err))
		return false
	}
	if taskServerHealth != nil {
		mainLogger.Error("Task server health check failed.", tint.Err(taskServerHealth))
		return false
	}
	if err := db.Ping(ctx); err != nil {
		mainLogger.Error("Database health check failed.", tint.Err(err))
		return false
	}

	// TODO: Rewrite migration checks using https://github.com/jackc/pgx/pull/1718
	goose.SetBaseFS(migrations.FS)
	gooseDB, err := goose.OpenDBWithDriver("pgx", config.DBConnection)
	if err != nil {
		logger.ErrorContext(ctx, "Could not create database migration connection.", "log", "health", tint.Err(err))
		return false
	}
	defer func() {
		if err := gooseDB.Close(); err != nil {
			logger.ErrorContext(ctx, "Could not close database migration connection.", "log", "health", tint.Err(err))
			ret = false
		}
	}()
	current, err := goose.GetDBVersionContext(ctx, gooseDB)
	if err != nil {
		logger.ErrorContext(ctx, "Could not fetch current migration version.", "log", "health", tint.Err(err))
		return false
	}
	ms, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil {
		logger.ErrorContext(ctx, "Could not collect pending migrations.", "log", "health", tint.Err(err))
		return false
	}
	last, _ := ms.Last()
	if last.Version != current {
		logger.WarnContext(ctx, fmt.Sprintf("The database schema is at version %d. The server expects version %d. Please migrate.", current, last.Version), "log", "health")
		// FIXME: Maybe return false if the two versions are too far apart
	}
	return true
}

// setupDatabase create a database connection pool.
func setupDatabase() (func(), error) {
	mainLogger.Info("Setting up database connection pool.")
	var err error
	db, err = pgxpool.New(context.Background(), config.DBConnection)
	if err != nil {
		mainLogger.Error("Could not setup database connection pool.", tint.Err(err))
		return nil, fmt.Errorf("creating database connection: %w", err)
	}
	conf := db.Config().ConnConfig
	mainLogger.Debug(fmt.Sprintf("Using database connection postgres://%s:******@%s:%d/%s.", conf.User, conf.Host, conf.Port, conf.Database))
	return func() {
		mainLogger.Info("Closing database connections.")
		db.Close()
	}, nil
}

// setupRedis either starts the embedded redis instance or prepares a connection to an external redis.
func setupRedis() (func(), error) {
	if config.RedisConnection == "" {
		mainLogger.Warn("You are using the embedded redis server. This server is not intended for production use.")
		mainLogger.Info("Starting embedded redis server.")
		miniRedis := miniredis.NewMiniRedis()
		if err := miniRedis.Start(); err != nil {
			mainLogger.Error("Could not start embedded redis server.", tint.Err(err))
			return nil, fmt.Errorf("starting embedded redis server: %w", err)
		}
		redisConn = asynq.RedisClientOpt{Addr: miniRedis.Addr()}
		return func() {
			mainLogger.Info("Stopping embedded redis server.")
			miniRedis.Close()
		}, nil
	}
	conn, err := asynq.ParseRedisURI(config.RedisConnection)
	if err != nil {
		mainLogger.Error("Could not parse Redis URL.", tint.Err(err))
		return nil, fmt.Errorf("could not parse redis url: %w", err)
	}
	clientConn, ok := conn.(asynq.RedisClientOpt)
	if !ok {
		mainLogger.Error("Currently only direct redis connections are supported.")
		return func() {}, fmt.Errorf("redis connection must not be clustered or sentinel")
	}
	if clientConn.Username == "" {
		clientConn.Username = os.Getenv("REDIS_USERNAME")
	}
	if clientConn.Password == "" {
		clientConn.Password = os.Getenv("REDIS_PASSWORD")
	}
	redisConn = clientConn
	return func() {}, nil
}

// setupTaskQueue sets up the asynq.Client for enqueuing tasks.
func setupTaskQueue(redis asynq.RedisConnOpt) func() {
	mainLogger.Info("Setting up task queue.")
	taskQueue = asynq.NewClient(redis)
	return func() {
		mainLogger.Info("Closing task queue.")
		if err := taskQueue.Close(); err != nil {
			mainLogger.Error("Could not close task queue.", tint.Err(err))
		}
	}
}

// setupTaskRunner sets up the task runner that executes background tasks.
func setupTaskRunner(mediaRepo media.Repository, mediaStore media.Store) (context.Context, func(), error) {
	mainLogger.Info(fmt.Sprintf("Starting task runner with %d workers.", config.TaskRunner.Workers))
	taskRunnerLogger := internal.NewAsynqLogger(logger, "asynq.server")
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
				logger.ErrorContext(ctx, "Retry Exhausted for task.", "log", "asynq.server", "task", task.Type(), tint.Err(err))
			} else {
				logger.WarnContext(ctx, "Task failed. Will be retried shortly.", "log", "asynq.server", "task", task.Type(), "retried", retried, "maxRetry", maxRetry, tint.Err(err))
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
		mainLogger.Error("Could not start task runner.", tint.Err(err))
		return nil, nil, fmt.Errorf("starting task server: %w", err)
	}
	return taskRunnerLogger.Context(), func() {
		mainLogger.Info("Stopping task runner.")
		taskRunner.Shutdown()
	}, nil
}

// setupTaskScheduler sets up the task scheduler that creates specific task instances for scheduled tasks.
func setupTaskScheduler() (context.Context, func(), error) {
	mainLogger.Info("Starting task scheduler.")
	taskSchedulerLogger := internal.NewAsynqLogger(logger, "asynq.scheduler")
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
			PreEnqueueFunc: func(task *asynq.Task, opts []asynq.Option) {
				logger.Info("Enqueuing scheduled task.", "log", "asynq.scheduler", "task", task.Type())
			},
			PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
				if err != nil {
					logger.Error("Could not enqueue scheduled task.", "log", "asynq.scheduler", "task", info.Type, tint.Err(err))
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
		mainLogger.Error("Could not start task scheduler.", tint.Err(err))
		return nil, nil, fmt.Errorf("starting cron manager: %w", err)
	}
	return taskSchedulerLogger.Context(), func() {
		mainLogger.Info("Stopping task scheduler.")
		scheduler.Shutdown()
	}, nil
}
