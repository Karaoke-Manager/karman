package test

import (
	"context"
	"fmt"
	"net"
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

var (
	pgVersion = os.Getenv("PGVERSION") // postgres version, used with testcontainers
	pgImage   = os.Getenv("PGIMAGE")   // full postgres image, used with testcontainers, overrides pgVersion

	pgHost = os.Getenv("PGHOST") // host of the database
	pgPort = os.Getenv("PGPORT") // port of the database
	pgUser = os.Getenv("PGUSER") // user of the database, must have CREATE/DROP DATABASE permission
	pgPass = os.Getenv("PGPASS") // password of pgUser
)

// NewDB creates a new database connection for a single test.
// Depending on how the go test command was invoked this method behaves in one of two ways:
//
// If the -pg-host flag has been specified, a database connection to an existing database is created.
// The connection uses the values from the -pg-port, -pg-user, and -pg-pass for the connection.
//
// If -pg-host is not specified, this functions uses testcontainers to create a testing database.
// You can control the postgres version using the -pg-version flag or specify a custom image using -pg-image.
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
	if pgHost == "" {
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

	pool, err := pgxpool.New(context.TODO(), connectionString(database))
	if err != nil {
		t.Fatalf("Could not connect to testing database: %s", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// runPostgresContainer starts a new testcontainers instance of PostgreSQL.
// This function sets up pgHost, pgPort, etc. to point to this container.
func runPostgresContainer() error {
	if pgVersion == "" {
		pgVersion = "15"
	}
	if pgImage == "" {
		pgImage = fmt.Sprintf("postgres:%s-alpine", strings.TrimPrefix(pgVersion, "v"))
	}
	if pgUser == "" {
		pgUser = "postgres"
	}
	if pgPass == "" {
		pgPass = "postgres"
	}
	container, err := postgres.RunContainer(context.TODO(),
		testcontainers.WithImage(pgImage),
		testcontainers.WithWaitStrategy(wait.ForExposedPort()),
		postgres.WithUsername(pgUser),
		postgres.WithPassword(pgPass),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) {
			req.Name = "karman-tests-" + strings.ReplaceAll(strings.ReplaceAll(pgImage, ":", "_"), "/", "-")
			req.Reuse = true
			// container is terminated by Reaper/Ryuk
		}),
	)
	if err != nil {
		return err
	}
	if pgHost, err = container.Host(context.TODO()); err != nil {
		return err
	}
	rawPort, err := container.MappedPort(context.TODO(), "5432/tcp")
	if err != nil {
		return err
	}
	pgPort = rawPort.Port()
	return nil
}

// createTestingDatabase creates a new, empty database.
// This function does not run migrations.
func createTestingDatabase() (string, error) {
	database := uuid.New().String()
	// Initially we connect to the postgres DB to create the database for this test
	db, err := pgx.Connect(context.TODO(), connectionString("postgres"))
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
	db, err := pgx.Connect(context.TODO(), connectionString("postgres"))
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
	db, err := goose.OpenDBWithDriver("pgx", connectionString(database))
	if err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err = goose.Up(db, "."); err != nil {
		return err
	}
	return db.Close()
}

// connectionString is a helper method to construct a connection string to the specified database.
func connectionString(database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", pgUser, pgPass, net.JoinHostPort(pgHost, pgPort), database)
}
