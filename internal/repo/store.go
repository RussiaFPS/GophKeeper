package storage

import (
	"context"
	"gopthkeeper/internal/models"
)

type Store interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)

	CreateSecret(ctx context.Context, secret models.Secret) (models.Secret, error)
	GetSecrets(ctx context.Context, userID int) ([]models.Secret, error)
	GetSecretByID(ctx context.Context, userID, secretID int) (models.Secret, error)
	UpdateSecret(ctx context.Context, secret models.Secret) (models.Secret, error)
	DeleteSecret(ctx context.Context, userID, secretID int) error
}
