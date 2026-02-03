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
	log.Println("Starting server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: storage_type=%s, server_address=%s", cfg.StorageType, cfg.ServerAddress)

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize storage based on configuration
	var store storage.Store

	if cfg.IsMemoryStorage() {
		log.Println("Using in-memory storage")
		store = storage.NewMemStore()
	} else if cfg.IsPostgresStorage() {
		log.Printf("Connecting to PostgreSQL database")

		pgStore, err := storage.NewPostgresStore(cfg.GetDatabaseDSN())
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
		}
		defer pgStore.Close()

		store = pgStore
		log.Println("Successfully connected to PostgreSQL database")
	}

	// Wrap store with encryption if encryption key is provided
	if cfg.EncryptionKey != "" {
		log.Println("Encryption enabled for secret data")
		encryptedStore, err := storage.NewEncryptedStore(store, cfg.EncryptionKey)
		if err != nil {
			log.Fatalf("Failed to initialize encryption: %v", err)
		}
		store = encryptedStore
	} else {
		log.Println("WARNING: Encryption is disabled. Secrets will be stored in plaintext.")
	}

	// Initialize API handlers
	apiHandler := delivery.New(store, jwtManager)

	// Initialize router
	router := delivery.NewRouter(apiHandler, jwtManager)

	// Start server with or without TLS
	if cfg.EnableTLS {
		log.Printf("Server is listening on %s (HTTPS enabled)", cfg.ServerAddress)
		log.Printf("Using TLS certificate: %s", cfg.TLSCertFile)
		log.Fatal(http.ListenAndServeTLS(cfg.ServerAddress, cfg.TLSCertFile, cfg.TLSKeyFile, router))
	} else {
		log.Printf("Server is listening on %s (HTTP mode - consider enabling TLS for production)", cfg.ServerAddress)
		log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
	}
}
