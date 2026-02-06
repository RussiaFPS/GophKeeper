package main

import (
	"gopthkeeper/internal/auth"
	config "gopthkeeper/internal/config/server"
	delivery "gopthkeeper/internal/delivery/server"
	storage "gopthkeeper/internal/repo"
	"log"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func run() error {
	log.Println("Starting server...")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log.Printf("Configuration loaded: storage_type=%s, server_address=%s", cfg.StorageType, cfg.ServerAddress)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	store, err := setupStorage(cfg)
	if err != nil {
		return err
	}

	apiHandler := delivery.New(store, jwtManager)
	router := delivery.NewRouter(apiHandler, jwtManager)

	return startServer(cfg, router)
}

func setupStorage(cfg *config.Config) (storage.Store, error) {
	var store storage.Store
	var err error

	if cfg.IsMemoryStorage() {
		log.Println("Using in-memory storage")
		store = storage.NewMemStore()
	} else if cfg.IsPostgresStorage() {
		log.Printf("Connecting to PostgreSQL database")
		pgStore, err := storage.NewPostgresStore(cfg.GetDatabaseDSN())
		if err != nil {
			return nil, err
		}
		store = pgStore
		log.Println("Successfully connected to PostgreSQL database")
	}

	if cfg.EncryptionKey != "" {
		log.Println("Encryption enabled for secret data")
		store, err = storage.NewEncryptedStore(store, cfg.EncryptionKey)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("WARNING: Encryption is disabled. Secrets will be stored in plaintext.")
	}

	return store, nil
}

func startServer(cfg *config.Config, router http.Handler) error {
	if cfg.EnableTLS {
		log.Printf("Server is listening on %s (mTLS enabled)", cfg.ServerAddress)
		log.Printf("Using TLS certificate: %s", cfg.TLSCertFile)
		return http.ListenAndServeTLS(cfg.ServerAddress, cfg.TLSCertFile, cfg.TLSKeyFile, router)
	}

	log.Printf("Server is listening on %s (mTLS mode - consider enabling TLS for production)", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, router)
}
