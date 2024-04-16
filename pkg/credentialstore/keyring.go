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

type KeyringCredentialStore struct {
	keyringConfig keyring.Config
	keyring       keyring.Keyring
}

func NewKeyringCredentialStore() *KeyringCredentialStore {
	return &KeyringCredentialStore{
		keyringConfig: keyring.Config{ //nolint:exhaustruct
			ServiceName:                    appName,
			FileDir:                        filepath.Join(xdg.DataHome, appName),
			KeychainTrustApplication:       true,
			KeychainAccessibleWhenUnlocked: true,
			FilePasswordFunc:               passwdPromptFunc,
		},
		keyring: nil,
	}
}

func (c *KeyringCredentialStore) Get(partitionKey string) ([]byte, error) {
	err := c.ensureKeyringOpen()
	if err != nil {
		return nil, err
	}

	item, err := c.keyring.Get("minctl-msal-cache")
	if err != nil {
		if err.Error() == "The specified item could not be found in the keyring" {
			return []byte{}, nil
		}

		return nil, fmt.Errorf("failed to get item from keyring: %w", err)
	}

	return item.Data, nil
}

func (c *KeyringCredentialStore) Set(data []byte, partitionKey string) error {
	err := c.ensureKeyringOpen()
	if err != nil {
		return err
	}

	err = c.keyring.Set(keyring.Item{ //nolint:exhaustruct
		Key:  "minctl-msal-cache",
		Data: data,
	})

	if err != nil {
		return fmt.Errorf("failed to set item in keyring: %w", err)
	}

	return nil
}

func (c *KeyringCredentialStore) Clear() error {
	err := c.ensureKeyringOpen()
	if err != nil {
		return err
	}

	err = c.keyring.Remove("minctl-msal-cache")
	if err != nil {
		return fmt.Errorf("failed to remove item from keyring: %w", err)
	}

	return nil
}

func (c *KeyringCredentialStore) ensureKeyringOpen() error {
	if c.keyring == nil {
		ring, err := keyring.Open(c.keyringConfig)
		if err != nil {
			return fmt.Errorf("failed to open keyring: %w", err)
		}

		c.keyring = ring
	}

	return nil
}

func passwdPromptFunc(prompt string) (string, error) {
	fmt.Print(prompt + ": ")

	bytePassword, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	if err != nil {
		return "", fmt.Errorf("could not read password: %w", err)
	}

	return string(bytePassword), nil
}
