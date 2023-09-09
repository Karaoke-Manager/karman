package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // required to run migrations
	"github.com/jackc/pgxutil"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Karaoke-Manager/karman/migrations"
)

// init sets up default values in the environment.
func init() {
	if os.Getenv("POSTGRES_VERSION") == "" {
		_ = os.Setenv("POSTGRES_VERSION", "15")
	}
	if os.Getenv("POSTGRES_IMAGE") == "" {
		_ = os.Setenv("POSTGRES_IMAGE", fmt.Sprintf("postgres:%s-alpine", strings.TrimPrefix(os.Getenv("POSTGRES_VERSION"), "v")))
	}
	if os.Getenv("PGUSER") == "" {
		_ = os.Setenv("PGUSER", "postgres")
	}
	if os.Getenv("PGPASSWORD") == "" {
		_ = os.Setenv("PGPASSWORD", "postgres")
	}
}

// NewDB creates a new database connection for a single test.
// Depending on the environment this method behaves in one of two ways:
//
// If the PGHOST environment variable is set, a database connection to an existing database is created.
// The connection uses the usual environment variables for postgres connections (PGPORT, PGUSER, etc.).
// If PGUSER or PGPASSWORD is empty, a default value of "postgres" will be used for either.
//
// If PGHOST is not present in the environment, this functions uses testcontainers to create a testing database.
// You can control the postgres version using the POSTGRES_VERSION variable or specify a custom image using POSTGRES_IMAGE.
// Only a single database container is created for multiple tests (for each version/image).
// The database container will not be terminated automatically, so that it can be reused by multiple tests.
// You should rely on Reaper/Ryuk to automatically remove these containers.
//
// In both cases this function creates a new database using CREATE DATABASE.
// NewDB runs all migrations on the database and then returns a connection.
// When the test is over the database is dropped automatically.
// The -pg-user must have the appropriate permissions on the database.
// When using testcontainers this is the case by default.
func NewDB(t *testing.T) pgxutil.DB {
	if os.Getenv("PGHOST") == "" {
		if err := runPostgresContainer(); err != nil {
			t.Fatalf("NewDB() could not create postgres container: %s", err)
		}
	}
	database, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("NewDB() could not create new database for tests: %s", err)
	}
	t.Cleanup(func() {
		if err := dropTestingDatabase(database); err != nil {
			t.Fatalf("NewDB() could not drop testing database: %s", err)
		}
	})

	if err = migrate(database); err != nil {
		t.Fatalf("NewDB() could not migrate testing database: %s", err)
	}

	pool, err := pgxpool.New(context.TODO(), fmt.Sprintf("dbname=%s", database))
	if err != nil {
		t.Fatalf("Could not connect to testing database: %s", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// runPostgresContainer starts a new testcontainers instance of PostgreSQL.
// This function sets up pgHost, pgPort, etc. to point to this container.
func runPostgresContainer() error {
	container, err := postgres.RunContainer(context.TODO(),
		testcontainers.WithImage(os.Getenv("POSTGRES_IMAGE")),
		testcontainers.WithWaitStrategy(wait.ForExposedPort()),
		postgres.WithUsername(os.Getenv("PGUSER")),
		postgres.WithPassword(os.Getenv("PGPASSWORD")),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) {
			req.Name = "karman-tests-" + strings.ReplaceAll(strings.ReplaceAll(os.Getenv("POSTGRES_IMAGE"), ":", "_"), "/", "-")
			req.Reuse = true
			// container is terminated by Reaper/Ryuk
		}),
	)
	if err != nil {
		return err
	}
	host, err := container.Host(context.TODO())
	if err != nil {
		return err
	}
	_ = os.Setenv("PGHOST", host)
	rawPort, err := container.MappedPort(context.TODO(), "5432/tcp")
	if err != nil {
		return err
	}
	_ = os.Setenv("PGPORT", rawPort.Port())
	return nil
}

// createTestingDatabase creates a new, empty database.
// This function does not run migrations.
func createTestingDatabase() (string, error) {
	database := uuid.New().String()
	// Initially we connect to the postgres DB to create the database for this test
	db, err := pgx.Connect(context.TODO(), "dbname=postgres")
	if err != nil {
		return database, err
	}
	// $1 does not work for identifiers so we use concatenation
	if _, err = db.Exec(context.TODO(), "CREATE DATABASE"+" "+pgx.Identifier{database}.Sanitize()); err != nil {
		return database, err
	}
	return database, db.Close(context.TODO())
}

// dropTestingDatabase tries to drop the specified database.
func dropTestingDatabase(database string) error {
	db, err := pgx.Connect(context.TODO(), "dbname=postgres")
	if err != nil {
		return err
	}
	// $1 does not work for identifiers so we use concatenation
	if _, err = db.Exec(context.TODO(), "DROP DATABASE"+" "+pgx.Identifier{database}.Sanitize()); err != nil {
		return err
	}
	return db.Close(context.TODO())
}

// migrate applies all known migrations to the specified database.
func migrate(database string) error {
	if err := goose.SetDialect("pgx"); err != nil {
		return err
	}
	db, err := sql.Open("pgx", fmt.Sprintf("dbname=%s", database))
	if err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err = goose.Up(db, "."); err != nil {
		return err
	}
	return db.Close()
}
