package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSuccessfulStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		// Successful status codes (200-299)
		{
			name:   "200 OK is successful",
			status: http.StatusOK,
			want:   true,
		},
		{
			name:   "201 Created is successful",
			status: http.StatusCreated,
			want:   true,
		},
		{
			name:   "204 No Content is successful",
			status: http.StatusNoContent,
			want:   true,
		},
		{
			name:   "299 is successful (upper boundary)",
			status: 299,
			want:   true,
		},
		// Unsuccessful status codes
		{
			name:   "199 is not successful (below range)",
			status: 199,
			want:   false,
		},
		{
			name:   "300 is not successful (above range)",
			status: 300,
			want:   false,
		},
		{
			name:   "400 Bad Request is not successful",
			status: http.StatusBadRequest,
			want:   false,
		},
		{
			name:   "401 Unauthorized is not successful",
			status: http.StatusUnauthorized,
			want:   false,
		},
		{
			name:   "403 Forbidden is not successful",
			status: http.StatusForbidden,
			want:   false,
		},
		{
			name:   "404 Not Found is not successful",
			status: http.StatusNotFound,
			want:   false,
		},
		{
			name:   "500 Internal Server Error is not successful",
			status: http.StatusInternalServerError,
			want:   false,
		},
		{
			name:   "502 Bad Gateway is not successful",
			status: http.StatusBadGateway,
			want:   false,
		},
		{
			name:   "503 Service Unavailable is not successful",
			status: http.StatusServiceUnavailable,
			want:   false,
		},
		// Edge cases
		{
			name:   "0 is not successful",
			status: 0,
			want:   false,
		},
		{
			name:   "negative status is not successful",
			status: -1,
			want:   false,
		},
		{
			name:   "100 Continue is not successful",
			status: http.StatusContinue,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSuccessfulStatusCode(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestError(t *testing.T) {
	t.Run("Error returns message", func(t *testing.T) {
		err := &RequestError{Message: "something went wrong"}
		assert.Equal(t, "something went wrong", err.Error())
	})

	t.Run("Error returns empty string for empty message", func(t *testing.T) {
		err := &RequestError{Message: ""}
		assert.Equal(t, "", err.Error())
	})

	t.Run("Error returns message with status prefix", func(t *testing.T) {
		err := &RequestError{Message: "404 Not Found: resource not found"}
		assert.Equal(t, "404 Not Found: resource not found", err.Error())
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = &RequestError{Message: "test"}
		assert.NotNil(t, err)
		assert.Equal(t, "test", err.Error())
	})
}

func TestDoRequest(t *testing.T) {
	type testResponse struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("successful request with JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(testResponse{Name: "test", Value: 42})
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		assert.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("successful request with nil result", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		req, err := http.NewRequest("DELETE", server.URL, nil)
		require.NoError(t, err)

		err = doRequest[any](server.Client(), req, nil)

		assert.NoError(t, err)
	})

	t.Run("error status without body returns RequestError with status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		require.Error(t, err)
		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Equal(t, "404 Not Found", reqErr.Message)
	})

	t.Run("error status with body returns RequestError with status and body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid request parameters"))
		}))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		require.Error(t, err)
		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Contains(t, reqErr.Message, "400 Bad Request")
		assert.Contains(t, reqErr.Message, "invalid request parameters")
	})

	t.Run("401 Unauthorized returns RequestError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("authentication required"))
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		require.Error(t, err)
		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Contains(t, reqErr.Message, "401 Unauthorized")
	})

	t.Run("500 Internal Server Error returns RequestError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		require.Error(t, err)
		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Contains(t, reqErr.Message, "500 Internal Server Error")
	})

	t.Run("invalid JSON response returns decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not valid json"))
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not decode response")
	})

	t.Run("201 Created is successful", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(testResponse{Name: "created", Value: 1})
		}))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL, nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(server.Client(), req, &result)

		assert.NoError(t, err)
		assert.Equal(t, "created", result.Name)
	})

	t.Run("connection error returns wrapped error", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:99999", nil)
		require.NoError(t, err)

		var result testResponse
		err = doRequest(http.DefaultClient, req, &result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not perform request")
	})
}
