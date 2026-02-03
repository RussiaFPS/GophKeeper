package storage

import (
	"context"
	"fmt"
	"gopthkeeper/internal/crypto"
	"gopthkeeper/internal/models"
)

// EncryptedStore wraps a Store and provides transparent encryption/decryption of secret data
type EncryptedStore struct {
	store     Store
	encryptor *crypto.Encryptor
}

// NewEncryptedStore creates a new EncryptedStore that wraps the provided store
// If encryptionKey is empty, encryption is disabled and data passes through unchanged
func NewEncryptedStore(store Store, encryptionKey string) (*EncryptedStore, error) {
	var encryptor *crypto.Encryptor
	var err error

	if encryptionKey != "" {
		encryptor, err = crypto.NewEncryptor(encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create encryptor: %w", err)
		}
	}

	return &EncryptedStore{
		store:     store,
		encryptor: encryptor,
	}, nil
}

// CreateUser delegates to the underlying store (no encryption needed for users)
func (es *EncryptedStore) CreateUser(ctx context.Context, user models.User) (models.User, error) {

	return es.store.CreateUser(ctx, user)

}

// GetUserByLogin delegates to the underlying store (no encryption needed for users)
func (es *EncryptedStore) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return es.store.GetUserByLogin(ctx, login)
}

// CreateSecret encrypts the secret data before storing
func (es *EncryptedStore) CreateSecret(ctx context.Context, secret models.Secret) (models.Secret, error) {
	if es.encryptor != nil && len(secret.Data) > 0 {
		encryptedData, err := es.encryptor.Encrypt(secret.Data)
		if err != nil {
			return models.Secret{}, fmt.Errorf("failed to encrypt secret data: %w", err)
		}
		secret.Data = encryptedData
	}

	return es.store.CreateSecret(ctx, secret)
}

// GetSecrets retrieves and decrypts all secrets for a user
func (es *EncryptedStore) GetSecrets(ctx context.Context, userID int) ([]models.Secret, error) {
	secrets, err := es.store.GetSecrets(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Decrypt each secret
	for i := range secrets {
		if es.encryptor != nil && len(secrets[i].Data) > 0 {
			decryptedData, err := es.encryptor.Decrypt(secrets[i].Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt secret %d: %w", secrets[i].ID, err)
			}
			secrets[i].Data = decryptedData
		}
	}

	return secrets, nil
}

// GetSecretByID retrieves and decrypts a specific secret
func (es *EncryptedStore) GetSecretByID(ctx context.Context, userID, secretID int) (models.Secret, error) {
	secret, err := es.store.GetSecretByID(ctx, userID, secretID)
	if err != nil {
		return models.Secret{}, err
	}

	// Decrypt the secret data
	if es.encryptor != nil && len(secret.Data) > 0 {
		decryptedData, err := es.encryptor.Decrypt(secret.Data)
		if err != nil {
			return models.Secret{}, fmt.Errorf("failed to decrypt secret: %w", err)
		}
		secret.Data = decryptedData
	}

	return secret, nil
}

// UpdateSecret encrypts the secret data before updating
func (es *EncryptedStore) UpdateSecret(ctx context.Context, secret models.Secret) (models.Secret, error) {
	if es.encryptor != nil && len(secret.Data) > 0 {
		encryptedData, err := es.encryptor.Encrypt(secret.Data)
		if err != nil {
			return models.Secret{}, fmt.Errorf("failed to encrypt secret data: %w", err)
		}
		secret.Data = encryptedData
	}

	return es.store.UpdateSecret(ctx, secret)
}

// DeleteSecret delegates to the underlying store
func (es *EncryptedStore) DeleteSecret(ctx context.Context, userID, secretID int) error {
	return es.store.DeleteSecret(ctx, userID, secretID)
}
