package health

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	"github.com/Karaoke-Manager/karman/migrations"
)

// Service implements a health check for various components.
type Service struct {
	Logger       *slog.Logger
	DBConnection string
	DB           interface {
		Ping(ctx context.Context) error
	}
	RedisClient redis.UniversalClient
}

// HealthCheck performs a health check on all system components.
func (s *Service) HealthCheck(ctx context.Context) (ret bool) {
	if err := s.RedisClient.Ping(ctx).Err(); err != nil {
		s.Logger.ErrorContext(ctx, "Redis health check failed.", tint.Err(err))
		return false
	}
	if err := s.DB.Ping(ctx); err != nil {
		s.Logger.ErrorContext(ctx, "Database health check failed.", tint.Err(err))
		return false
	}

	// TODO: Rewrite migration checks using https://github.com/jackc/pgx/pull/1718
	goose.SetBaseFS(migrations.FS)
	gooseDB, err := sql.Open("pgx", s.DBConnection)
	if err != nil {
		s.Logger.ErrorContext(ctx, "Could not create database migration connection.", tint.Err(err))
		return false
	}
	defer func() {
		if err := gooseDB.Close(); err != nil {
			s.Logger.ErrorContext(ctx, "Could not close database migration connection.", tint.Err(err))
			ret = false
		}
	}()
	current, err := goose.GetDBVersionContext(ctx, gooseDB)
	if err != nil {
		s.Logger.ErrorContext(ctx, "Could not fetch current migration version.", tint.Err(err))
		return false
	}
	ms, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil {
		s.Logger.ErrorContext(ctx, "Could not collect pending migrations.", tint.Err(err))
		return false
	}
	last, _ := ms.Last()
	if last.Version != current {
		s.Logger.WarnContext(ctx, fmt.Sprintf("The database schema is at version %d. The server expects version %d. Please migrate.", current, last.Version), "log", "health")
		// FIXME: Maybe return false if the two versions are too far apart
	}
	return true
}
