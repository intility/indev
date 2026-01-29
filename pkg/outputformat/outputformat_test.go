package outputformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{
			name:   "empty format",
			format: Format(""),
			want:   "",
		},
		{
			name:   "wide format",
			format: Format("wide"),
			want:   "wide",
		},
		{
			name:   "json format",
			format: Format("json"),
			want:   "json",
		},
		{
			name:   "yaml format",
			format: Format("yaml"),
			want:   "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_Set(t *testing.T) {
	t.Run("valid formats", func(t *testing.T) {
		validFormats := []string{"wide", "json", "yaml"}

		for _, format := range validFormats {
			t.Run(format, func(t *testing.T) {
				var f Format
				err := f.Set(format)

				assert.NoError(t, err)
				assert.Equal(t, Format(format), f)
			})
		}
	})

	t.Run("invalid formats", func(t *testing.T) {
		invalidFormats := []string{
			"",
			"table",
			"csv",
			"xml",
			"JSON", // case sensitive
			"YAML", // case sensitive
			"Wide", // case sensitive
			"plain",
			"text",
		}

		for _, format := range invalidFormats {
			t.Run(format, func(t *testing.T) {
				var f Format
				err := f.Set(format)

				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidOutputFormat)
				assert.Equal(t, Format(""), f) // Should not be modified
			})
		}
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		f := Format("json")
		err := f.Set("yaml")

		assert.NoError(t, err)
		assert.Equal(t, Format("yaml"), f)
	})
}

func TestFormat_Type(t *testing.T) {
	var f Format
	assert.Equal(t, "outputFormat", f.Type())
}

func TestErrInvalidOutputFormat(t *testing.T) {
	assert.NotNil(t, ErrInvalidOutputFormat)
	assert.Contains(t, ErrInvalidOutputFormat.Error(), "wide")
	assert.Contains(t, ErrInvalidOutputFormat.Error(), "json")
	assert.Contains(t, ErrInvalidOutputFormat.Error(), "yaml")
}
