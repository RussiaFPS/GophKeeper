package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encryptor handles encryption and decryption of data using AES-256-GCM
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new Encryptor with the provided key
// The key should be 32 bytes for AES-256
func NewEncryptor(key string) (*Encryptor, error) {
	if key == "" {
		return nil, fmt.Errorf("encryption key cannot be empty")
	}

	// Decode the key from base64 or use it directly
	keyBytes := []byte(key)

	// If the key is not 32 bytes, we need to derive it properly
	if len(keyBytes) != 32 {
		// For simplicity, we'll pad or truncate to 32 bytes
		// In production, use a proper key derivation function (KDF)
		derivedKey := make([]byte, 32)
		copy(derivedKey, keyBytes)
		keyBytes = derivedKey
	}

	return &Encryptor{key: keyBytes}, nil
}

// Encrypt encrypts plaintext data using AES-256-GCM
// Returns the encrypted data with the nonce prepended
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext data using AES-256-GCM
// Expects the nonce to be prepended to the ciphertext
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext cannot be empty")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
