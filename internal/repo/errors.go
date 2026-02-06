package storage

import "fmt"

// ErrUserExists is returned when trying to create a user that already exists.
type ErrUserExists struct {
	Login string
}

func (e ErrUserExists) Error() string {
	return fmt.Sprintf("user with login '%s' already exists", e.Login)
}

func NewErrUserExists(login string) ErrUserExists {
	return ErrUserExists{Login: login}
}

// ErrUserNotFound is returned when a user is not found.
type ErrUserNotFound struct {
	Login string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with login '%s' not found", e.Login)
}

func NewErrUserNotFound(login string) ErrUserNotFound {
	return ErrUserNotFound{Login: login}
}

// ErrSecretNotFound is returned when a secret is not found.
type ErrSecretNotFound struct {
	SecretID int
}

func (e ErrSecretNotFound) Error() string {
	return fmt.Sprintf("secret with ID '%d' not found", e.SecretID)
}

func NewErrSecretNotFound(secretID int) ErrSecretNotFound {
	return ErrSecretNotFound{SecretID: secretID}
}
