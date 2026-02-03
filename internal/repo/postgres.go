package storage

import (
	"context"
	"errors"
	"fmt"
	"gopthkeeper/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore is a PostgreSQL-based data store.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates and returns a new PostgresStore.
func NewPostgresStore(connString string) (*PostgresStore, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	store := &PostgresStore{pool: pool}

	// Initialize database schema
	if err := store.initSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}

// initSchema creates the necessary tables if they don't exist.
func (s *PostgresStore) initSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS secrets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type INTEGER NOT NULL,
			data BYTEA NOT NULL,
			metadata TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id)`,
	}

	for _, query := range queries {
		if _, err := s.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// CreateUser adds a new user to the store.
func (s *PostgresStore) CreateUser(ctx context.Context, user models.User) (models.User, error) {

	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	err := s.pool.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID)
	if err != nil {
		// Check for unique constraint violation (PostgreSQL error code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, NewErrUserExists(user.Login)
		}
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByLogin retrieves a user by their login.
func (s *PostgresStore) GetUserByLogin(ctx context.Context, login string) (models.User, error) {

	query := `SELECT id, login, password FROM users WHERE login = $1`

	var user models.User
	err := s.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, NewErrUserNotFound(login)
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateSecret adds a new secret for a user.
func (s *PostgresStore) CreateSecret(ctx context.Context, secret models.Secret) (models.Secret, error) {

	query := `INSERT INTO secrets (user_id, type, data, metadata) VALUES ($1, $2, $3, $4) RETURNING id`

	err := s.pool.QueryRow(ctx, query, secret.UserID, secret.Type, secret.Data, secret.Metadata).Scan(&secret.ID)
	if err != nil {
		return models.Secret{}, fmt.Errorf("failed to create secret: %w", err)
	}

	return secret, nil
}

// GetSecrets retrieves all secrets for a specific user.
func (s *PostgresStore) GetSecrets(ctx context.Context, userID int) ([]models.Secret, error) {

	query := `SELECT id, user_id, type, data, metadata FROM secrets WHERE user_id = $1`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}
	defer rows.Close()

	var secrets []models.Secret
	for rows.Next() {
		var secret models.Secret
		err := rows.Scan(&secret.ID, &secret.UserID, &secret.Type, &secret.Data, &secret.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}
		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secrets: %w", err)
	}

	if secrets == nil {
		return []models.Secret{}, nil
	}

	return secrets, nil
}

// GetSecretByID retrieves a specific secret for a user by its ID.
func (s *PostgresStore) GetSecretByID(ctx context.Context, userID, secretID int) (models.Secret, error) {

	query := `SELECT id, user_id, type, data, metadata FROM secrets WHERE id = $1 AND user_id = $2`

	var secret models.Secret
	err := s.pool.QueryRow(ctx, query, secretID, userID).Scan(
		&secret.ID, &secret.UserID, &secret.Type, &secret.Data, &secret.Metadata,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Secret{}, NewErrSecretNotFound(secretID)
		}
		return models.Secret{}, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// UpdateSecret updates an existing secret for a user.
func (s *PostgresStore) UpdateSecret(ctx context.Context, secret models.Secret) (models.Secret, error) {

	query := `UPDATE secrets SET type = $1, data = $2, metadata = $3 WHERE id = $4 AND user_id = $5`

	result, err := s.pool.Exec(ctx, query, secret.Type, secret.Data, secret.Metadata, secret.ID, secret.UserID)
	if err != nil {
		return models.Secret{}, fmt.Errorf("failed to update secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.Secret{}, NewErrSecretNotFound(secret.ID)
	}

	return secret, nil
}

// DeleteSecret deletes a secret for a user by its ID.
func (s *PostgresStore) DeleteSecret(ctx context.Context, userID, secretID int) error {

	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`

	result, err := s.pool.Exec(ctx, query, secretID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return NewErrSecretNotFound(secretID)
	}

	return nil
}
