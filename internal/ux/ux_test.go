package ux

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsuccessff(t *testing.T) {
	t.Run("writes success prefix and message", func(t *testing.T) {
		var buf bytes.Buffer
		Fsuccessf(&buf, "operation completed")

		output := buf.String()
		assert.Contains(t, output, "success:")
		assert.Contains(t, output, "operation completed")
	})

	t.Run("formats arguments", func(t *testing.T) {
		var buf bytes.Buffer
		Fsuccessf(&buf, "created %s with id %d", "user", 123)

		output := buf.String()
		assert.Contains(t, output, "created user with id 123")
	})
}

func TestFinfoff(t *testing.T) {
	t.Run("writes info prefix and message", func(t *testing.T) {
		var buf bytes.Buffer
		Finfof(&buf, "processing request")

		output := buf.String()
		assert.Contains(t, output, "info:")
		assert.Contains(t, output, "processing request")
	})

	t.Run("formats arguments", func(t *testing.T) {
		var buf bytes.Buffer
		Finfof(&buf, "found %d items", 42)

		output := buf.String()
		assert.Contains(t, output, "found 42 items")
	})
}

func TestFwarningff(t *testing.T) {
	t.Run("writes warning prefix and message", func(t *testing.T) {
		var buf bytes.Buffer
		Fwarningf(&buf, "deprecated feature")

		output := buf.String()
		assert.Contains(t, output, "warning:")
		assert.Contains(t, output, "deprecated feature")
	})

	t.Run("formats arguments", func(t *testing.T) {
		var buf bytes.Buffer
		Fwarningf(&buf, "rate limit at %d%%", 80)

		output := buf.String()
		assert.Contains(t, output, "rate limit at 80%")
	})
}

func TestFerrorff(t *testing.T) {
	t.Run("writes error prefix and message", func(t *testing.T) {
		var buf bytes.Buffer
		Ferrorf(&buf, "connection failed")

		output := buf.String()
		assert.Contains(t, output, "error:")
		assert.Contains(t, output, "connection failed")
	})

	t.Run("formats arguments", func(t *testing.T) {
		var buf bytes.Buffer
		Ferrorf(&buf, "failed to connect to %s:%d", "localhost", 5432)

		output := buf.String()
		assert.Contains(t, output, "failed to connect to localhost:5432")
	})
}

func TestFprintff(t *testing.T) {
	t.Run("writes message without prefix", func(t *testing.T) {
		var buf bytes.Buffer
		Fprintf(&buf, "plain message")

		output := buf.String()
		assert.Equal(t, "plain message", output)
		assert.NotContains(t, output, "success:")
		assert.NotContains(t, output, "info:")
		assert.NotContains(t, output, "warning:")
		assert.NotContains(t, output, "error:")
	})

	t.Run("formats arguments", func(t *testing.T) {
		var buf bytes.Buffer
		Fprintf(&buf, "Hello, %s! You have %d messages.", "Alice", 5)

		output := buf.String()
		assert.Equal(t, "Hello, Alice! You have 5 messages.", output)
	})

	t.Run("handles newlines", func(t *testing.T) {
		var buf bytes.Buffer
		Fprintf(&buf, "line1\nline2\n")

		output := buf.String()
		assert.Equal(t, "line1\nline2\n", output)
	})
}

func TestStyles(t *testing.T) {
	// Just verify styles are defined and don't panic when used
	t.Run("StyleSuccess is defined", func(t *testing.T) {
		result := StyleSuccess.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleWarning is defined", func(t *testing.T) {
		result := StyleWarning.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleError is defined", func(t *testing.T) {
		result := StyleError.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleInfo is defined", func(t *testing.T) {
		result := StyleInfo.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleSuccessLabel is defined", func(t *testing.T) {
		result := StyleSuccessLabel.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleWarningLabel is defined", func(t *testing.T) {
		result := StyleWarningLabel.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleErrorLabel is defined", func(t *testing.T) {
		result := StyleErrorLabel.Render("test")
		assert.NotEmpty(t, result)
	})

	t.Run("StyleInfoLabel is defined", func(t *testing.T) {
		result := StyleInfoLabel.Render("test")
		assert.NotEmpty(t, result)
	})
}
