package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgxutil"
	"log"
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
	"github.com/lmittmann/tint"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Karaoke-Manager/karman/api"
	"github.com/Karaoke-Manager/karman/cmd/karman/health"
	"github.com/Karaoke-Manager/karman/cmd/karman/internal"
	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/core/upload"
	"github.com/Karaoke-Manager/karman/task"
)

// The coreServices struct holds an instance of each core service.
// This is mainly used to pass around multiple coreServices more conveniently.
type coreServices struct {
	songService   song.Service
	songRepo      song.Repository
	uploadService upload.Service
	uploadRepo    upload.Repository
	uploadStore   upload.Store
	mediaService  media.Service
	mediaRepo     media.Repository
	mediaStore    media.Store
}

// migrate indicates whether the --migrate flag was specified.
var migrate bool

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
		return goose.Up(db, ".")
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		goose.SetLogger(goose.NopLogger())
		cleanups := make([]func(), 0)
		cleanup := func(close func()) {
			cleanups = append(cleanups, close)
		}
		defer func() {
			for _, cleanup := range cleanups {
				//goland:noinspection GoDeferInLoop
				defer cleanup()
			}
		}()

		// Setup application parts
		db, err := setupDatabase(cleanup)
		if err != nil {
			return err
		}
		services, err := setupServices(db)
		if err != nil {
			return err
		}
		redisConn, err := setupRedis(cleanup)
		if err != nil {
			return err
		}
		_ = setupAsynqClient(redisConn, cleanup)
		_ = setupTaskInspector(redisConn, cleanup)
		if _, err := setupTaskRunner(redisConn, services, cleanup); err != nil {
			return err
		}
		if _, err := setupTaskScheduler(redisConn, cleanup); err != nil {
			return err
		}
		healthService := setupHealthCheck(redisConn, db, cleanup)

		// Run a healthcheck to log potential connection problems directly.
		healthService.HealthCheck(context.Background())

		mainLogger.Info(fmt.Sprintf("Running HTTP server on %s.", config.API.Address))
		server := &http.Server{
			Addr:              config.API.Address,
			ReadHeaderTimeout: 3 * time.Second,
			Handler: api.NewHandler(
				logger.With("log", "api"),
				logger.With("log", "request"),
				healthService,
				services.songRepo,
				services.songService,
				services.mediaService,
				services.mediaStore,
				services.uploadRepo,
				services.uploadStore,
				config.Debug,
			),
			ErrorLog: slog.NewLogLogger(logger.With("log", "http").Handler(), config.Log.Level),
		}

		go waitForSignal(server)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			mainLogger.Error("Could not start HTTP server.", tint.Err(err))
			return fmt.Errorf("starting HTTP server: %w", err)
		}
		// there is no way the server terminates by itself without some kind of error.
		// some errors are only logged, so we return a catch-all error here.
		return errors.New("stopped")
	},
}

// setupServices initializes the core application coreServices.
func setupServices(db pgxutil.DB) (*coreServices, error) {
	mainLogger.Info("Setting up application coreServices.")
	songService := song.NewService()
	uploadStore, err := upload.NewFileStore(logger.With("log", "upload.store"), config.Uploads.Dir)
	if err != nil {
		mainLogger.Error("Could not initialize upload storage.", tint.Err(err))
		return nil, fmt.Errorf("initializing upload storage: %w", err)
	}
	mainLogger.Debug(fmt.Sprintf("Upload storage initialized at %s.", uploadStore.Root()))
	mediaStore, err := media.NewFileStore(logger.With("log", "song.store"), config.Media.Dir)
	if err != nil {
		mainLogger.Error("Could not initialize media store.", tint.Err(err))
		return nil, fmt.Errorf("initializing media storage: %w", err)
	}
	mainLogger.Debug(fmt.Sprintf("Media storage initialized at %s.", mediaStore.Root()))
	mediaRepo := media.NewDBRepository(logger.With("log", "media.repo"), db)
	songRepo := song.NewDBRepository(logger.With("log", "song.repo"), db)
	uploadRepo := upload.NewDBRepository(logger.With("log", "upload.repo"), db)
	return &coreServices{
		songService,
		songRepo,
		upload.NewService(logger.With("log", "upload.service"), uploadRepo, uploadStore, songRepo, songService),
		uploadRepo,
		uploadStore,
		media.NewService(logger.With("log", "song.service"), mediaRepo, mediaStore),
		mediaRepo,
		mediaStore,
	}, nil
}

// setupDatabase create a database connection pool.
func setupDatabase(cleanup func(func())) (*pgxpool.Pool, error) {
	mainLogger.Info("Setting up database connection pool.")
	db, err := pgxpool.New(context.Background(), config.DBConnection)
	if err != nil {
		mainLogger.Error("Could not setup database connection pool.", tint.Err(err))
		return nil, fmt.Errorf("creating database connection: %w", err)
	}
	cleanup(func() {
		mainLogger.Info("Closing database connections.")
		db.Close()
	})
	conf := db.Config().ConnConfig
	mainLogger.Debug(fmt.Sprintf("Using database connection postgres://%s:******@%s:%d/%s.", conf.User, conf.Host, conf.Port, conf.Database))
	return db, nil
}

// setupRedis either starts the embedded redis instance or prepares a connection to an external redis.
func setupRedis(cleanup func(func())) (asynq.RedisConnOpt, error) {
	if config.RedisConnection == "" {
		mainLogger.Warn("You are using the embedded redis server. This server is not intended for production use.")
		mainLogger.Info("Starting embedded redis server.")
		miniRedis := miniredis.NewMiniRedis()
		if err := miniRedis.Start(); err != nil {
			mainLogger.Error("Could not start embedded redis server.", tint.Err(err))
			return nil, fmt.Errorf("starting embedded redis server: %w", err)
		}
		cleanup(func() {
			mainLogger.Info("Stopping embedded redis server.")
			miniRedis.Close()
		})
		return asynq.RedisClientOpt{Addr: miniRedis.Addr()}, nil
	}
	conn, err := asynq.ParseRedisURI(config.RedisConnection)
	if err != nil {
		mainLogger.Error("Could not parse Redis URL.", tint.Err(err))
		return nil, fmt.Errorf("could not parse redis url: %w", err)
	}
	clientConn, ok := conn.(asynq.RedisClientOpt)
	if !ok {
		mainLogger.Error("Currently only direct redis connections are supported.")
		return nil, fmt.Errorf("redis connection must not be clustered or sentinel")
	}
	// FIXME: This should be handled by Viper
	if clientConn.Username == "" {
		clientConn.Username = os.Getenv("REDIS_USERNAME")
	}
	if clientConn.Password == "" {
		clientConn.Password = os.Getenv("REDIS_PASSWORD")
	}
	userpass := clientConn.Username
	if clientConn.Password != "" {
		userpass += ":" + clientConn.Password
	}
	if userpass != "" {
		userpass += "@"
	}
	logger.Debug(fmt.Sprintf("Using redis connection redis://%s%s/%d", userpass, clientConn.Addr, clientConn.DB))
	return clientConn, nil
}

// setupAsynqClient sets up the asynq.Client for enqueuing tasks.
func setupAsynqClient(redisConn asynq.RedisConnOpt, cleanup func(func())) *asynq.Client {
	mainLogger.Info("Setting up task queue.")
	client := asynq.NewClient(redisConn)
	cleanup(func() {
		mainLogger.Info("Closing task queue.")
		if err := client.Close(); err != nil {
			mainLogger.Error("Could not close task queue.", tint.Err(err))
		}
	})
	return client
}

// setupTaskInspector initializes an asynq.Inspector instance.
func setupTaskInspector(redisConn asynq.RedisConnOpt, cleanup func(func())) *asynq.Inspector {
	mainLogger.Info("Starting task inspector.")
	taskInspector := asynq.NewInspector(redisConn)
	cleanup(func() {
		mainLogger.Info("Stopping task inspector.")
		if err := taskInspector.Close(); err != nil {
			mainLogger.Error("Could not stop task inspector.", tint.Err(err))
		}
	})
	return taskInspector
}

// setupTaskRunner sets up the task runner that executes background tasks.
func setupTaskRunner(redisConn asynq.RedisConnOpt, services *coreServices, cleanup func(func())) (*asynq.Server, error) {
	mainLogger.Info(fmt.Sprintf("Starting task runner with %d workers.", config.TaskRunner.Workers))
	taskRunner := asynq.NewServer(redisConn, asynq.Config{
		Concurrency: config.TaskRunner.Workers,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			if retried >= maxRetry {
				logger.ErrorContext(ctx, "Retry Exhausted for task.", "log", "asynq.server", "task", task.Type(), tint.Err(err))
			} else {
				logger.WarnContext(ctx, "Task failed. Will be retried shortly.", "log", "asynq.server", "task", task.Type(), "retried", retried, "maxRetry", maxRetry, tint.Err(err))
			}
		}),
		Logger:   (*internal.AsynqLogger)(logger.With("log", "asynq.server")),
		LogLevel: internal.AsynqLogLevel(config.Log.Level),
		// We perform a health check on redis explicitly, so we do not need to use the health check of the task runner.
	})
	h := task.NewHandler(logger.With("log", "task"), services.mediaRepo, services.mediaService, services.uploadService, services.uploadRepo, services.uploadStore)
	if err := taskRunner.Start(h); err != nil {
		mainLogger.Error("Could not start task runner.", tint.Err(err))
		return nil, fmt.Errorf("starting task server: %w", err)
	}
	cleanup(func() {
		mainLogger.Info("Stopping task runner.")
		taskRunner.Shutdown()
	})
	return taskRunner, nil
}

// setupTaskScheduler sets up the task scheduler that creates specific task instances for scheduled tasks.
func setupTaskScheduler(redis asynq.RedisConnOpt, cleanup func(func())) (*asynq.Scheduler, error) {
	mainLogger.Info("Starting task scheduler.")
	schedulerLogger := logger.With("log", "asynq.scheduler")
	scheduler := asynq.NewScheduler(redis, &asynq.SchedulerOpts{
		Logger:   (*internal.AsynqLogger)(schedulerLogger),
		LogLevel: internal.AsynqLogLevel(config.Log.Level),
		Location: time.Local,
		PreEnqueueFunc: func(task *asynq.Task, opts []asynq.Option) {
			schedulerLogger.Info("Enqueuing scheduled task.", "task", task.Type())
		},
		PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
			if err != nil {
				schedulerLogger.Error("Could not enqueue scheduled task.", tint.Err(err))
			}
		},
	})
	if err := scheduler.Start(); err != nil {
		mainLogger.Error("Could not start task scheduler.", tint.Err(err))
		return nil, fmt.Errorf("starting task scheduler: %w", err)
	}
	cleanup(func() {
		mainLogger.Info("Stopping task scheduler.")
		scheduler.Shutdown()
	})
	return scheduler, nil
}

// setupHealthCheck initializes a health.Service.
func setupHealthCheck(redisConn asynq.RedisConnOpt, db *pgxpool.Pool, cleanup func(func())) *health.Service {
	mainLogger.Info("Starting health check service.")
	redisClient := redisConn.MakeRedisClient().(redis.UniversalClient)
	healthService := &health.Service{
		Logger:       logger.With("log", "health"),
		DBConnection: config.DBConnection,
		DB:           db,
		RedisClient:  redisClient,
	}
	cleanup(func() {
		mainLogger.Info("Stopping health check service.")
		if err := redisClient.Close(); err != nil {
			mainLogger.Error("Could not close health check redis connection.", tint.Err(err))
		}
	})
	return healthService
}

// waitForSignal blocks until the program receives a SIGINT or SIGTERM signal.
// When a signal is received, the server is closed.
func waitForSignal(server *http.Server) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	mainLogger.Warn(fmt.Sprintf("Stop signal %q received. Shutting down...", sig))
	mainLogger.Info("Stopping HTTP server with 30 second timeout.")
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	if err := server.Shutdown(ctx); errors.Is(err, context.DeadlineExceeded) {
		mainLogger.Error("HTTP server did not shut down for 30 seconds. Terminating forcefully.")
	} else if err != nil {
		mainLogger.Error("HTTP server shutdown caused an error.", tint.Err(err))
	}
	if err := server.Close(); err != nil {
		mainLogger.Error("Could not close HTTP server.", tint.Err(err))
	}
	cancel()
}
