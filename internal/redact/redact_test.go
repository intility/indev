package redact

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// customRedactor implements the redactor interface for testing
type customRedactor struct {
	sensitive string
	safe      string
}

func (c *customRedactor) Error() string  { return c.sensitive }
func (c *customRedactor) Redact() string { return c.safe }

func TestError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := Error(nil)
		assert.Nil(t, result)
	})

	t.Run("simple error is redacted with placeholder", func(t *testing.T) {
		err := errors.New("sensitive data")
		redacted := Error(err)

		assert.NotNil(t, redacted)
		assert.Contains(t, redacted.Error(), "<redacted")
		assert.NotContains(t, redacted.Error(), "sensitive data")
	})

	t.Run("already redacted error is not double-redacted", func(t *testing.T) {
		original := errors.New("sensitive")
		firstRedact := Error(original)
		secondRedact := Error(firstRedact)

		// Should be the same error, not re-wrapped
		assert.Equal(t, firstRedact, secondRedact)
	})

	t.Run("error with Redact method uses custom redaction", func(t *testing.T) {
		err := &customRedactor{
			sensitive: "user password123",
			safe:      "user [REDACTED]",
		}
		redacted := Error(err)

		assert.Equal(t, "user [REDACTED]", redacted.Error())
	})

	t.Run("wrapped error chain is redacted", func(t *testing.T) {
		inner := errors.New("inner secret")
		outer := fmt.Errorf("outer error: %w", inner)
		redacted := Error(outer)

		assert.NotContains(t, redacted.Error(), "inner secret")
		assert.NotContains(t, redacted.Error(), "outer error")
		assert.Contains(t, redacted.Error(), "<redacted")
	})

	t.Run("wrapped redactor stops unwrapping", func(t *testing.T) {
		inner := &customRedactor{
			sensitive: "secret inner",
			safe:      "safe inner message",
		}
		outer := fmt.Errorf("outer wrapper: %w", inner)
		redacted := Error(outer)

		// Should contain the safe inner message
		assert.Contains(t, redacted.Error(), "safe inner message")
	})

	t.Run("redacted error preserves Unwrap behavior", func(t *testing.T) {
		inner := errors.New("inner")
		outer := fmt.Errorf("outer: %w", inner)
		redacted := Error(outer)

		// Should be able to unwrap to original
		unwrapped := errors.Unwrap(redacted)
		assert.Equal(t, outer, unwrapped)
	})
}

func TestErrorf(t *testing.T) {
	t.Run("creates error with formatted message", func(t *testing.T) {
		err := Errorf("user %s not found", "alice")

		assert.Equal(t, "user alice not found", err.Error())
	})

	t.Run("redacts arguments in Redact output", func(t *testing.T) {
		err := Errorf("user %s with id %d", "alice", 123)

		redactor, ok := err.(redactor)
		require.True(t, ok, "error should implement redactor interface")

		redacted := redactor.Redact()
		assert.NotContains(t, redacted, "alice")
		assert.NotContains(t, redacted, "123")
		assert.Contains(t, redacted, "<redacted")
	})

	t.Run("Safe arguments are not redacted", func(t *testing.T) {
		err := Errorf("user %s with id %d", "alice", Safe(123))

		redactor, ok := err.(redactor)
		require.True(t, ok)

		redacted := redactor.Redact()
		assert.NotContains(t, redacted, "alice")
		assert.Contains(t, redacted, "123") // Safe value preserved
	})

	t.Run("multiple arguments redacted correctly", func(t *testing.T) {
		err := Errorf("connecting to %s:%d as %s", "localhost", 5432, "admin")

		redactor, ok := err.(redactor)
		require.True(t, ok)

		redacted := redactor.Redact()
		assert.NotContains(t, redacted, "localhost")
		assert.NotContains(t, redacted, "5432")
		assert.NotContains(t, redacted, "admin")
	})

	t.Run("wrapped error is redacted", func(t *testing.T) {
		inner := errors.New("connection refused")
		err := Errorf("database error: %w", inner)

		assert.Contains(t, err.Error(), "connection refused")

		redactor, ok := err.(redactor)
		require.True(t, ok)

		redacted := redactor.Redact()
		assert.NotContains(t, redacted, "connection refused")
	})

	t.Run("error can be unwrapped", func(t *testing.T) {
		err := Errorf("outer: %s", "data")

		unwrapped := errors.Unwrap(err)
		assert.NotNil(t, unwrapped)
	})

	t.Run("custom redactor argument uses Redact method", func(t *testing.T) {
		custom := &customRedactor{
			sensitive: "secret-token-abc123",
			safe:      "token-[HIDDEN]",
		}
		err := Errorf("auth failed: %s", custom)

		assert.Contains(t, err.Error(), "secret-token-abc123")

		redactor, ok := err.(redactor)
		require.True(t, ok)

		redacted := redactor.Redact()
		assert.Contains(t, redacted, "token-[HIDDEN]")
		assert.NotContains(t, redacted, "secret-token-abc123")
	})

	t.Run("captures stack trace", func(t *testing.T) {
		err := Errorf("test error")

		safeErr, ok := err.(*safeError)
		require.True(t, ok)

		frames := safeErr.StackTrace()
		assert.NotEmpty(t, frames)

		// First frame should be from this test file
		found := false
		for _, frame := range frames {
			if strings.Contains(frame.File, "redact_test.go") {
				found = true
				break
			}
		}
		assert.True(t, found, "stack trace should include test file")
	})
}

func TestSafe(t *testing.T) {
	t.Run("wraps string value", func(t *testing.T) {
		result := Safe("test")
		s, ok := result.(safe)
		assert.True(t, ok)
		assert.Equal(t, "test", s.a)
	})

	t.Run("wraps integer value", func(t *testing.T) {
		result := Safe(42)
		s, ok := result.(safe)
		assert.True(t, ok)
		assert.Equal(t, 42, s.a)
	})

	t.Run("wraps nil value", func(t *testing.T) {
		result := Safe(nil)
		s, ok := result.(safe)
		assert.True(t, ok)
		assert.Nil(t, s.a)
	})
}

func TestPlaceholder(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "string type",
			input: "test",
			want:  "<redacted string>",
		},
		{
			name:  "int type",
			input: 42,
			want:  "<redacted int>",
		},
		{
			name:  "error type",
			input: errors.New("test"),
			want:  "<redacted *errors.errorString>",
		},
		{
			name:  "custom struct pointer",
			input: &customRedactor{},
			want:  "<redacted *redact.customRedactor>",
		},
		{
			name:  "slice type",
			input: []string{"a", "b"},
			want:  "<redacted []string>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := placeholder(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSafeErrorMethods(t *testing.T) {
	err := Errorf("test %s", "value")
	safeErr, ok := err.(*safeError)
	require.True(t, ok)

	t.Run("Error returns unredacted message", func(t *testing.T) {
		assert.Equal(t, "test value", safeErr.Error())
	})

	t.Run("Redact returns redacted message", func(t *testing.T) {
		redacted := safeErr.Redact()
		assert.NotContains(t, redacted, "value")
		assert.Contains(t, redacted, "<redacted")
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		unwrapped := safeErr.Unwrap()
		assert.NotNil(t, unwrapped)
	})

	t.Run("StackTrace returns frames", func(t *testing.T) {
		frames := safeErr.StackTrace()
		assert.NotEmpty(t, frames)
	})
}

func TestSafeErrorFormat(t *testing.T) {
	err := Errorf("test error")

	t.Run("format with %%s", func(t *testing.T) {
		result := fmt.Sprintf("%s", err)
		assert.Equal(t, "test error", result)
	})

	t.Run("format with %%v", func(t *testing.T) {
		result := fmt.Sprintf("%v", err)
		assert.Equal(t, "test error", result)
	})

	t.Run("format with %%q", func(t *testing.T) {
		result := fmt.Sprintf("%q", err)
		assert.Equal(t, `"test error"`, result)
	})

	t.Run("format with %%+v includes stack trace", func(t *testing.T) {
		result := fmt.Sprintf("%+v", err)
		assert.Contains(t, result, "test error")
		// Should contain file path from stack trace
		assert.Contains(t, result, ".go")
	})
}

func TestRedactedErrorMethods(t *testing.T) {
	original := errors.New("original error")
	redacted := Error(original)
	redactedErr, ok := redacted.(*redactedError)
	require.True(t, ok)

	t.Run("Error returns redacted message", func(t *testing.T) {
		msg := redactedErr.Error()
		assert.Contains(t, msg, "<redacted")
		assert.NotContains(t, msg, "original error")
	})

	t.Run("Unwrap returns original error", func(t *testing.T) {
		unwrapped := redactedErr.Unwrap()
		assert.Equal(t, original, unwrapped)
	})
}

func TestErrorIntegration(t *testing.T) {
	t.Run("Error on Errorf result", func(t *testing.T) {
		// Create a safe error with Errorf
		safeErr := Errorf("user %s has balance %d", "alice", 1000)

		// Redact it with Error()
		redacted := Error(safeErr)

		// The result should use the Redact() output
		assert.NotContains(t, redacted.Error(), "alice")
		assert.NotContains(t, redacted.Error(), "1000")
	})

	t.Run("nested Errorf calls", func(t *testing.T) {
		inner := Errorf("inner secret: %s", "password123")
		outer := Errorf("outer error: %w", inner)

		// Outer error should contain inner message
		assert.Contains(t, outer.Error(), "password123")

		// Redacting outer should redact everything
		redactor, ok := outer.(redactor)
		require.True(t, ok)
		redacted := redactor.Redact()
		assert.NotContains(t, redacted, "password123")
	})

	t.Run("errors.Is works with redacted errors", func(t *testing.T) {
		sentinel := errors.New("not found")
		wrapped := fmt.Errorf("user lookup: %w", sentinel)
		redacted := Error(wrapped)

		assert.True(t, errors.Is(redacted, sentinel))
	})

	t.Run("errors.As works with redacted errors", func(t *testing.T) {
		custom := &customRedactor{sensitive: "secret", safe: "safe"}
		wrapped := fmt.Errorf("wrapped: %w", custom)
		redacted := Error(wrapped)

		var target *customRedactor
		assert.True(t, errors.As(redacted, &target))
		assert.Equal(t, "secret", target.sensitive)
	})
}
