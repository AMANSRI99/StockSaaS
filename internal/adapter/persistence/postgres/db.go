package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib" // Import driver for side-effect (registers driver)
)

// NewConnection creates and returns a new database connection pool.
func NewConnection(cfg config.DBConfig) (*sql.DB, error) {

	encodedUser := url.QueryEscape(cfg.User)
	encodedPassword := url.QueryEscape(cfg.Password)

	// Construct Data Source Name (DSN)
	// Example: "postgres://user:password@host:port/dbname?sslmode=disable"
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		encodedUser, encodedPassword, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	log.Println("Attempting to connect to database...")
	// Open connection pool
	db, err := sql.Open("pgx", dsn) // Use "pgx" as the driver name
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool (optional but recommended)
	db.SetMaxOpenConns(25)                 // Max number of open connections
	db.SetMaxIdleConns(25)                 // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max lifetime of a connection

	// Verify connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close() // Close pool if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully!")
	return db, nil
}
