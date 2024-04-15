package credentialstore

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
	"golang.org/x/term"
)

const (
	appName = "minctl"
)

type CredentialStore struct {
	keyringConfig keyring.Config
	keyring       keyring.Keyring
}

func New() *CredentialStore {
	return &CredentialStore{
		keyringConfig: keyring.Config{ //nolint:exhaustruct
			ServiceName: appName,
			FileDir:     filepath.Join(xdg.DataHome, appName),
			FilePasswordFunc: func(prompt string) (string, error) {
				fmt.Print(prompt + ": ")

				bytePassword, err := term.ReadPassword(syscall.Stdin)
				fmt.Println()

				if err != nil {
					return "", fmt.Errorf("could not read password: %w", err)
				}

				return string(bytePassword), nil
			},
		},
		keyring: nil,
	}
}

func (c *CredentialStore) Get(key string) ([]byte, error) {
	err := c.ensureKeyringOpen()
	if err != nil {
		return nil, err
	}

	item, err := c.keyring.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from keyring: %w", err)
	}

	return item.Data, nil
}

func (c *CredentialStore) Set(key string, value []byte) error {
	err := c.ensureKeyringOpen()
	if err != nil {
		return err
	}

	err = c.keyring.Set(keyring.Item{ //nolint:exhaustruct
		Key:  key,
		Data: value,
	})

	if err != nil {
		return fmt.Errorf("failed to set item in keyring: %w", err)
	}

	return nil
}

func (c *CredentialStore) ensureKeyringOpen() error {
	if c.keyring == nil {
		ring, err := keyring.Open(c.keyringConfig)
		if err != nil {
			return fmt.Errorf("failed to open keyring: %w", err)
		}

		c.keyring = ring
	}

	return nil
}
