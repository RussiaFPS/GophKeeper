package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// StorageType represents the type of storage backend to use
type StorageType string

const (
	StorageMemory   StorageType = "memory"
	StoragePostgres StorageType = "postgres"
)

// Config holds the server configuration
type Config struct {
	ServerAddress string      `json:"server_address" env:"SERVER_ADDRESS" env-default:":8080"`
	DatabaseDSN   string      `json:"database_dsn" env:"DATABASE_DSN" env-default:""`
	JWTSecret     string      `json:"jwt_secret" env:"JWT_SECRET" env-default:"your-secret-key"`
	StorageType   StorageType `json:"storage_type" env:"STORAGE_TYPE" env-default:"memory"`
	EnableTLS     bool        `json:"enable_tls" env:"ENABLE_TLS" env-default:"false"`
	TLSCertFile   string      `json:"tls_cert_file" env:"TLS_CERT_FILE" env-default:""`
	TLSKeyFile    string      `json:"tls_key_file" env:"TLS_KEY_FILE" env-default:""`
	EncryptionKey string      `json:"encryption_key" env:"ENCRYPTION_KEY" env-default:""`
}

// Load loads configuration from environment variables, JSON file, and command-line flags
// Priority: flags > env vars > JSON file > defaults
func Load() (*Config, error) {
	cfg := &Config{}

	// Define command-line flags
	configFile := flag.String("config", "", "Path to JSON configuration file")
	serverAddr := flag.String("server-address", "", "Server address (e.g., :8080)")
	dbDSN := flag.String("database-dsn", "", "Database DSN connection string")
	jwtSecret := flag.String("jwt-secret", "", "JWT secret key")
	storageType := flag.String("storage-type", "", "Storage type: memory or postgres")
	enableTLS := flag.Bool("enable-tls", false, "Enable HTTPS/TLS")
	tlsCertFile := flag.String("tls-cert", "", "Path to TLS certificate file")
	tlsKeyFile := flag.String("tls-key", "", "Path to TLS private key file")
	encryptionKey := flag.String("encryption-key", "", "Master encryption key for secrets (32 bytes)")

	flag.Parse()

	// Load from JSON file if specified
	if *configFile != "" {
		if err := loadFromJSON(cfg, *configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from JSON: %w", err)
		}
	}

	// Load from environment variables (overrides JSON)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	// Override with command-line flags (highest priority)
	if *serverAddr != "" {
		cfg.ServerAddress = *serverAddr
	}
	if *dbDSN != "" {
		cfg.DatabaseDSN = *dbDSN
	}
	if *jwtSecret != "" {
		cfg.JWTSecret = *jwtSecret
	}
	if *storageType != "" {
		cfg.StorageType = StorageType(*storageType)
	}
	if flag.Lookup("enable-tls").Value.String() == "true" {
		cfg.EnableTLS = *enableTLS
	}
	if *tlsCertFile != "" {
		cfg.TLSCertFile = *tlsCertFile
	}
	if *tlsKeyFile != "" {
		cfg.TLSKeyFile = *tlsKeyFile
	}
	if *encryptionKey != "" {
		cfg.EncryptionKey = *encryptionKey
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromJSON loads configuration from a JSON file
func loadFromJSON(cfg *Config, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.StorageType != StorageMemory && c.StorageType != StoragePostgres {
		return fmt.Errorf("invalid storage type: %s (must be 'memory' or 'postgres')", c.StorageType)
	}

	if c.StorageType == StoragePostgres && c.DatabaseDSN == "" {
		return fmt.Errorf("database_dsn is required when storage_type is 'postgres'")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}

	if c.EnableTLS {
		if c.TLSCertFile == "" {
			return fmt.Errorf("tls_cert_file is required when enable_tls is true")
		}
		if c.TLSKeyFile == "" {
			return fmt.Errorf("tls_key_file is required when enable_tls is true")
		}
	}

	return nil
}

// GetDatabaseDSN returns the database DSN connection string
func (c *Config) GetDatabaseDSN() string {
	return c.DatabaseDSN
}

// IsMemoryStorage returns true if memory storage is configured
func (c *Config) IsMemoryStorage() bool {
	return c.StorageType == StorageMemory
}

// IsPostgresStorage returns true if PostgreSQL storage is configured
func (c *Config) IsPostgresStorage() bool {
	return c.StorageType == StoragePostgres
}
