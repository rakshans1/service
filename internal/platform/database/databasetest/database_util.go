package databasetest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"

	// imported to register the postgres migration driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// imported to register the "file" source migration driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ory/dockertest/v3"
	"github.com/rakshans1/service/internal/platform/database"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP
	Port string // Port
}

// NewTestDatabaseWithConfig creates a new database suitable for use in testing.
// This should not be used outside of testing, but it is exposed in the main
// package so it can be shared with other packages.
//
// All database tests can be skipped by running `go test -short`
func NewTestDatabaseWithConfig(t *testing.T) (*sqlx.DB, func(), *Container) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping database tests (short!")
	}

	// Context
	ctx := context.Background()

	// Create the pool (docker instance).
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("failed to create Docker pool: %s", err)
	}

	// Start the container.
	dbname, username, password := "sales", "username", "abc123"
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12-alpine",
		Env: []string{
			"LANG=C",
			"POSTGRES_DB=" + dbname,
			"POSTGRES_USER=" + username,
			"POSTGRES_PASSWORD=" + password,
			"TZ=" + "Asia/Kolkata",
		},
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}

	// Get the host. On Mac, Docker runs in a VM.
	host := container.Container.NetworkSettings.IPAddress
	if runtime.GOOS == "darwin" {
		host = net.JoinHostPort(container.GetBoundIP("5432/tcp"), container.GetPort("5432/tcp"))
	}

	// Build the connection Config.
	connConfig := database.Config{
		User:       username,
		Password:   password,
		Host:       host,
		Name:       dbname,
		DisableTLS: true,
	}

	// Build the connection URL.
	connURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(username, password),
		Host:   host,
		Path:   dbname,
	}
	q := connURL.Query()
	q.Add("sslmode", "disable")
	connURL.RawQuery = q.Encode()

	// Wait for the container to start - we'll retry connections in a loop below,
	// but there's no point in trying immediately.
	time.Sleep(1 * time.Second)

	// Establish a connection to the database. Use a Fibonacci backoff instead of
	// exponential so wait times scale appropriately.
	var db *sqlx.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = database.Open(connConfig)
		if err != nil {
			return err
		}
		return database.StatusCheck(ctx, db)
	}); err != nil {
		t.Fatalf("failed to start postgres: %s", err)
	}

	// Run the migrations.
	if err := dbMigrate(connURL.String()); err != nil {
		t.Fatalf("failed to migrate database: %s", err)
	}

	// Run the seed.
	if err := dbSeed(db); err != nil {
		t.Fatalf("failed to seed database: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(container); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}

	}
	return db, cleanup, &Container{
		Host: container.GetBoundIP("5432/tcp"),
		Port: container.GetPort("5432/tcp"),
		ID:   container.Container.ID,
	}

}

func NewTestDatabase(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	db, cleanup, _ := NewTestDatabaseWithConfig(t)
	return db, cleanup
}

// dbMigrate runs the migrations. u is the connection URL string (e.g.
// postgres://...).
func dbMigrate(u string) error {
	// Run the migrations
	migrationsDir := fmt.Sprintf("file://%s", dbMigrationsDir())
	m, err := migrate.New(migrationsDir, u)
	if err != nil {
		return fmt.Errorf("failed create migrate: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed run migrate: %w", err)
	}
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("migrate source error: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migrate database error: %w", dbErr)
	}
	return nil
}

// dbMigrationsDir returns the path on disk to the migrations. It uses
// runtime.Caller() to get the path to the caller, since this package is
// imported by multiple others at different levels.
func dbMigrationsDir() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "../../../../migrations")
}

func dbSeed(db *sqlx.DB) error {
	_, filename, _, _ := runtime.Caller(1)
	seedFile := filepath.Join(filepath.Dir(filename), "../../../../testdata/seed.sql")
	dat, err := ioutil.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("seed database error: %w", err)
	}

	seeds := string(dat)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
