package credentialstore

import (
	"fmt"

	"github.com/99designs/keyring"
)

type KeyringCredentialStore struct {
	serviceName string
}

func NewKeyringCredentialStore() *KeyringCredentialStore {
	return &KeyringCredentialStore{
		serviceName: "minctl",
	}
}

func (c *KeyringCredentialStore) Get(key string) ([]byte, error) {
	cfg := keyring.Config{ //nolint:exhaustruct
		ServiceName: c.serviceName,
	}

	ring, err := keyring.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	item, err := ring.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from keyring: %w", err)
	}

	return item.Data, nil
}

func (c *KeyringCredentialStore) Set(key string, value []byte) error {
	ring, err := keyring.Open(keyring.Config{ //nolint:exhaustruct
		ServiceName: c.serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to open keyring: %w", err)
	}

	err = ring.Set(keyring.Item{ //nolint:exhaustruct
		Key:  key,
		Data: value,
	})

	if err != nil {
		return fmt.Errorf("failed to set item in keyring: %w", err)
	}

	return nil
}
