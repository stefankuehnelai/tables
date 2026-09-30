package cli

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "nextcloud-tables"

var errCredentialNotFound = errors.New("credential not found")

type credentialStore interface {
	Set(hostname, username, kind, secret string) error
	Get(hostname, username, kind string) (string, error)
	Delete(hostname, username, kind string) error
}

type keyringCredentialStore struct{}

func (keyringCredentialStore) Set(hostname, username, kind, secret string) error {
	if err := keyring.Set(keyringService, credentialAccount(hostname, username, kind), secret); err != nil {
		return fmt.Errorf("store credential in OS keyring: %w", err)
	}
	return nil
}

func (keyringCredentialStore) Get(hostname, username, kind string) (string, error) {
	secret, err := keyring.Get(keyringService, credentialAccount(hostname, username, kind))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", errCredentialNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read credential from OS keyring: %w", err)
	}
	return secret, nil
}

func (keyringCredentialStore) Delete(hostname, username, kind string) error {
	err := keyring.Delete(keyringService, credentialAccount(hostname, username, kind))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete credential from OS keyring: %w", err)
	}
	return nil
}

func credentialAccount(hostname, username, kind string) string {
	return hostname + "/" + username + "/" + kind
}
