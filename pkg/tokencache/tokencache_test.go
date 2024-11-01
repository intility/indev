package tokencache_test

import (
	"errors"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/intility/indev/mocks"
	"github.com/intility/indev/pkg/tokencache"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/fs"
	"os"
	"testing"
)

type MockFs struct {
	afero.Fs
}

func (m *MockFs) Open(name string) (afero.File, error) {
	if name == "no_permissions_file" {
		return nil, fs.ErrPermission
	}

	return m.Fs.Open(name)
}

func (m *MockFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if name == "no_permissions_file" {
		return nil, fs.ErrPermission
	}

	if name == "non_existent_file" {
		return nil, fs.ErrNotExist
	}

	return m.Fs.OpenFile(name, flag, perm)
}

func (m *MockFs) Remove(name string) error {
	if name == "permission_denied_file" {
		return fs.ErrPermission
	}

	return m.Fs.Remove(name)
}

func TestTokenCache_Replace(t *testing.T) {
	tests := []struct {
		name          string
		cacheFilePath string
		prepareStore  func(store *mocks.CredentialStore)
		unmarshal     func(unmarshaler *mocks.Unmarshaler)
		wantErr       string // expected error message, empty if no error
	}{
		{
			name:          "successful replacement",
			cacheFilePath: "cache_file",
			prepareStore: func(store *mocks.CredentialStore) {
				store.On("Get", "test").Return([]byte{}, nil).Once() // Expecting the GetCluster function to be called once
			},
			unmarshal: func(u *mocks.Unmarshaler) {
				u.On("Unmarshal", mock.Anything).Return(nil).Once() // Expecting the Unmarshal function to be called once
			},
			wantErr: "",
		},
		{
			name: "error reading credential store",
			prepareStore: func(store *mocks.CredentialStore) {
				store.On("Get", "test").Return(nil, errors.New("unknown err")).Once() // Expecting the GetCluster function to be called once
			},
			wantErr: "could not read cache file: unknown err",
		},
		{
			name:          "error unmarshalling cache",
			cacheFilePath: "cache_file",
			prepareStore: func(store *mocks.CredentialStore) {
				store.On("Get", "test").Return([]byte{}, nil).Once() // Expecting the GetCluster function to be called once
			},
			unmarshal: func(u *mocks.Unmarshaler) {
				u.On("Unmarshal", mock.Anything).Return(errors.New("unknown err")).Once() // Expecting the Unmarshal function to be called once
			},
			wantErr: "could not unmarshal cache: unknown err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up the file system
			mockStore := &mocks.CredentialStore{}

			if tt.prepareStore != nil {
				tt.prepareStore(mockStore)
			}

			// set up the unmarshal expectations
			unmarshaler := &mocks.Unmarshaler{}
			if tt.unmarshal != nil {
				tt.unmarshal(unmarshaler)
			}

			tokenCache := tokencache.New(tokencache.WithCredentialStore(mockStore))

			hints := cache.ReplaceHints{PartitionKey: "test"}
			err := tokenCache.Replace(nil, unmarshaler, hints)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}

			unmarshaler.AssertExpectations(t)
		})
	}
}

func TestTokenCache_Export(t *testing.T) {
	tests := []struct {
		name         string
		prepareStore func(store *mocks.CredentialStore)
		marshal      func(marshaler *mocks.Marshaler)
		wantErr      string // expected error message, empty if no error
	}{
		{
			name: "successful export",
			prepareStore: func(store *mocks.CredentialStore) {
				store.On("Set", mock.Anything, "test").Return(nil).Once() // Expecting the Set function to be called once
			},
			marshal: func(m *mocks.Marshaler) {
				m.On("Marshal").Return([]byte("foo"), nil).Once() // Expecting the Marshal function to be called once
			},
			wantErr: "",
		},
		{
			name: "error writing credential store",
			prepareStore: func(store *mocks.CredentialStore) {
				store.On("Set", mock.Anything, "test").Return(errors.New("unknown err")).Once() // Expecting the Set function to be called once
			},
			marshal: func(m *mocks.Marshaler) {
				m.On("Marshal").Return([]byte("foo"), nil).Once() // Expecting the Marshal function to be called once
			},
			wantErr: "could not write cache file: unknown err",
		},
		{
			name: "error marshalling cache",
			marshal: func(m *mocks.Marshaler) {
				m.On("Marshal").Return(nil, errors.New("unknown err")).Once() // Expecting the Marshal function to be called once
			},
			wantErr: "could not marshal cache: unknown err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up the file system
			mockStore := &mocks.CredentialStore{}
			if tt.prepareStore != nil {
				tt.prepareStore(mockStore)
			}

			// set up the marshal expectations
			marshaler := &mocks.Marshaler{}
			if tt.marshal != nil {
				tt.marshal(marshaler)
			}

			tokenCache := tokencache.New(tokencache.WithCredentialStore(mockStore))

			hints := cache.ExportHints{PartitionKey: "test"}
			err := tokenCache.Export(nil, marshaler, hints)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}

			marshaler.AssertExpectations(t)
		})
	}
}
