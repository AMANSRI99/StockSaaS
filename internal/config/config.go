package config

import (
	"log"
	"os"
	"strconv"
	"time" // Import time

	"github.com/joho/godotenv"
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

// JWTConfig holds JWT configuration parameters.
type JWTConfig struct {
	SecretKey      string        // Secret key for signing tokens (MUST be kept secret)
	ExpiryDuration time.Duration // How long the access token is valid
}

// AppConfig holds the overall application configuration.
type AppConfig struct {
	ServerPort string
	Database   DBConfig
	JWT        JWTConfig // Add JWT config
}

// Load loads configuration from environment variables,
// potentially loading from a .env file first.
func Load() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v. Using existing env vars.", err)
	}

	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Printf("Warning: Invalid DB_PORT '%s', using default 5432. Error: %v", dbPortStr, err)
		dbPort = 5432
	}

	// Load JWT Expiry (in minutes from env var for simplicity)
	jwtExpiryMinutesStr := getEnv("JWT_EXPIRY_MINUTES", "60") // Default to 60 minutes
	jwtExpiryMinutes, err := strconv.Atoi(jwtExpiryMinutesStr)
	if err != nil {
		log.Printf("Warning: Invalid JWT_EXPIRY_MINUTES '%s', using default 60. Error: %v", jwtExpiryMinutesStr, err)
		jwtExpiryMinutes = 60
	}
	jwtExpiryDuration := time.Duration(jwtExpiryMinutes) * time.Minute

	// Load JWT Secret - CRITICAL: Must be set in production!
	jwtSecret := getEnv("JWT_SECRET", "") // No sensible default!
	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set!")
	}

	cfg := &AppConfig{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		Database: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "stocksaas_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{ // Populate JWT config
			SecretKey:      jwtSecret,
			ExpiryDuration: jwtExpiryDuration,
		},
	}

	if cfg.Database.User == "" || cfg.Database.DBName == "" {
		log.Fatal("DB_USER and DB_NAME environment variables must be set")
	}
	if cfg.Database.Password == "" {
		log.Println("Warning: DB_PASSWORD is not set.") // Might be ok for local dev with trusted connection
	}

	return cfg, nil
}

// Helper to get env var or default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	// Only log fallback usage if fallback is not empty, avoid logging for secrets
	if fallback != "" {
		log.Printf("Env variable %s not set, using default: %s", key, fallback)
	} else {
		log.Printf("Env variable %s not set.", key)
	}
	return fallback
}
