package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	tokenFileName    = "token.txt"
	defaultServerURL = "http://localhost:8080"
	serverURLEnvVar  = "SERVER_URL"
)

func GetServerURL() string {
	url := os.Getenv(serverURLEnvVar)
	if url == "" {
		return defaultServerURL
	}
	return url
}

// GetConfigDir returns the appropriate configuration directory for the OS.
func GetConfigDir() (string, error) {
	var configDir string
	switch {
	case os.Getenv("APPDATA") != "": // Windows
		configDir = filepath.Join(os.Getenv("APPDATA"), "gophkeeper")
	case os.Getenv("HOME") != "": // Linux, macOS
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "gophkeeper")
	default:
		return "", fmt.Errorf("cannot determine user home directory")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return configDir, nil
}

// SaveToken saves the JWT token to a file.
func SaveToken(token string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	tokenPath := filepath.Join(configDir, tokenFileName)
	return os.WriteFile(tokenPath, []byte(token), 0600)
}

// LoadToken loads the JWT token from a file.
func LoadToken() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	tokenPath := filepath.Join(configDir, tokenFileName)
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	return string(data), nil
}
