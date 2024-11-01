package tokencache

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/adrg/xdg"

	"github.com/intility/indev/pkg/credentialstore"
)

type TokenCache struct {
	store credentialstore.CredentialStore
}

type Option func(*TokenCache)

func New(options ...Option) *TokenCache {
	cacheFilePath := filepath.Join(xdg.DataHome, "indev", "msal.cache")

	tc := &TokenCache{
		store: credentialstore.NewFilesystemCredentialStore(cacheFilePath),
	}

	for _, option := range options {
		option(tc)
	}

	return tc
}

func WithCredentialStore(store credentialstore.CredentialStore) Option {
	return func(tc *TokenCache) {
		tc.store = store
	}
}

func (c *TokenCache) Replace(ctx context.Context, cache cache.Unmarshaler, hints cache.ReplaceHints) error {
	cacheData, err := c.store.Get(hints.PartitionKey)
	if err != nil {
		return fmt.Errorf("could not read cache file: %w", err)
	}

	err = cache.Unmarshal(cacheData)
	if err != nil {
		return fmt.Errorf("could not unmarshal cache: %w", err)
	}

	return nil
}

func (c *TokenCache) Export(ctx context.Context, cache cache.Marshaler, hints cache.ExportHints) error {
	cacheData, err := cache.Marshal()
	if err != nil {
		return fmt.Errorf("could not marshal cache: %w", err)
	}

	err = c.store.Set(cacheData, hints.PartitionKey)
	if err != nil {
		return fmt.Errorf("could not write cache file: %w", err)
	}

	return nil
}

func (c *TokenCache) Clear() error {
	err := c.store.Clear()
	if err != nil {
		return fmt.Errorf("could not clear cache: %w", err)
	}

	return nil
}
