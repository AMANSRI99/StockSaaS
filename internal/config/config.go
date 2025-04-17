package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv" // Import godotenv
)

// DBConfig holds database configuration parameters.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AppConfig holds the overall application configuration.
type AppConfig struct {
	ServerPort string
	Database   DBConfig
}

// Load loads configuration from environment variables,
// potentially loading from a .env file first.
func Load() (*AppConfig, error) {
	// --- Load .env file (if it exists) ---
	// This will load variables into the process environment.
	// It WON'T override variables already set in the environment.
	err := godotenv.Load() // Loads .env file from current directory or parent dirs
	if err != nil {
		// It's common to not have a .env file (e.g., in production),
		// so we only log a warning, not exit fatally.
		log.Printf("Warning: Could not load .env file: %v. Using existing env vars.", err)
	}
	// --- End of .env loading ---

	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Printf("Warning: Invalid DB_PORT '%s', using default 5432. Error: %v", dbPortStr, err)
		dbPort = 5432
	}

	cfg := &AppConfig{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		Database: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""), // Read password from env
			DBName:   getEnv("DB_NAME", "stocksaas_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if cfg.Database.User == "" || cfg.Database.DBName == "" {
		log.Fatal("DB_USER and DB_NAME environment variables must be set (either in environment or .env file)")
	}
	if cfg.Database.Password == "" {
		log.Println("Warning: DB_PASSWORD is not set.")
	}

	return cfg, nil
}

// Helper to get env var or default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Env variable %s not set, using default: %s", key, fallback)
	return fallback
}
