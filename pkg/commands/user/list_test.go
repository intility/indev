package user

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/intility/indev/pkg/client"
	"github.com/intility/indev/pkg/outputformat"
)

func TestSortUsersByOwnerThenName(t *testing.T) {
	tests := []struct {
		name  string
		users []client.User
		want  []string // Expected order of names after sorting
	}{
		{
			name:  "empty list",
			users: []client.User{},
			want:  []string{},
		},
		{
			name: "single user",
			users: []client.User{
				{Name: "Alice", Roles: []string{"member"}},
			},
			want: []string{"Alice"},
		},
		{
			name: "owners come first",
			users: []client.User{
				{Name: "Bob", Roles: []string{"member"}},
				{Name: "Alice", Roles: []string{"owner"}},
			},
			want: []string{"Alice", "Bob"},
		},
		{
			name: "multiple owners sorted by name",
			users: []client.User{
				{Name: "Charlie", Roles: []string{"owner"}},
				{Name: "Alice", Roles: []string{"owner"}},
				{Name: "Bob", Roles: []string{"owner"}},
			},
			want: []string{"Alice", "Bob", "Charlie"},
		},
		{
			name: "non-owners sorted by name",
			users: []client.User{
				{Name: "Charlie", Roles: []string{"member"}},
				{Name: "Alice", Roles: []string{"member"}},
				{Name: "Bob", Roles: []string{"member"}},
			},
			want: []string{"Alice", "Bob", "Charlie"},
		},
		{
			name: "mixed owners and non-owners",
			users: []client.User{
				{Name: "Dave", Roles: []string{"member"}},
				{Name: "Alice", Roles: []string{"owner"}},
				{Name: "Charlie", Roles: []string{"member"}},
				{Name: "Bob", Roles: []string{"owner"}},
			},
			want: []string{"Alice", "Bob", "Charlie", "Dave"},
		},
		{
			name: "owner with multiple roles",
			users: []client.User{
				{Name: "Bob", Roles: []string{"member"}},
				{Name: "Alice", Roles: []string{"admin", "owner", "member"}},
			},
			want: []string{"Alice", "Bob"},
		},
		{
			name: "user with no roles treated as non-owner",
			users: []client.User{
				{Name: "Alice", Roles: []string{}},
				{Name: "Bob", Roles: []string{"owner"}},
			},
			want: []string{"Bob", "Alice"},
		},
		{
			name: "user with nil roles treated as non-owner",
			users: []client.User{
				{Name: "Alice", Roles: nil},
				{Name: "Bob", Roles: []string{"owner"}},
			},
			want: []string{"Bob", "Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			users := make([]client.User, len(tt.users))
			copy(users, tt.users)

			sortUsersByOwnerThenName(users)

			got := make([]string, len(users))
			for i, u := range users {
				got[i] = u.Name
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintUsersList(t *testing.T) {
	sampleUsers := []client.User{
		{
			ID:    "user-1",
			Name:  "Alice Admin",
			UPN:   "alice@example.com",
			Roles: []string{"owner"},
		},
		{
			ID:    "user-2",
			Name:  "Bob Builder",
			UPN:   "bob@example.com",
			Roles: []string{"member"},
		},
		{
			ID:    "user-3",
			Name:  "Charlie Coder",
			UPN:   "charlie@example.com",
			Roles: []string{"member", "developer"},
		},
	}

	t.Run("default format shows name, UPN, role", func(t *testing.T) {
		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format(""), sampleUsers)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Alice Admin")
		assert.Contains(t, output, "alice@example.com")
		assert.Contains(t, output, "owner")
		// Default format should NOT contain ID
		assert.NotContains(t, output, "user-1")
	})

	t.Run("wide format includes ID", func(t *testing.T) {
		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format("wide"), sampleUsers)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "user-1")
		assert.Contains(t, output, "user-2")
		assert.Contains(t, output, "Alice Admin")
		assert.Contains(t, output, "alice@example.com")
	})

	t.Run("json format outputs valid JSON", func(t *testing.T) {
		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format("json"), sampleUsers)

		assert.NoError(t, err)

		var decoded []client.User
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 3)
	})

	t.Run("yaml format outputs valid YAML", func(t *testing.T) {
		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format("yaml"), sampleUsers)

		assert.NoError(t, err)

		var decoded []client.User
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 3)
	})

	t.Run("sorts users with owners first", func(t *testing.T) {
		// Users in non-sorted order
		unsortedUsers := []client.User{
			{ID: "1", Name: "Zach", UPN: "zach@example.com", Roles: []string{"member"}},
			{ID: "2", Name: "Alice", UPN: "alice@example.com", Roles: []string{"owner"}},
			{ID: "3", Name: "Bob", UPN: "bob@example.com", Roles: []string{"member"}},
		}

		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format("json"), unsortedUsers)

		assert.NoError(t, err)

		var decoded []client.User
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)

		// Owner (Alice) should be first, then alphabetically (Bob, Zach)
		assert.Equal(t, "Alice", decoded[0].Name)
		assert.Equal(t, "Bob", decoded[1].Name)
		assert.Equal(t, "Zach", decoded[2].Name)
	})

	t.Run("handles empty user list", func(t *testing.T) {
		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format("json"), []client.User{})

		assert.NoError(t, err)

		var decoded []client.User
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 0)
	})

	t.Run("handles multiple roles in output", func(t *testing.T) {
		users := []client.User{
			{Name: "Multi", UPN: "multi@example.com", Roles: []string{"admin", "developer", "owner"}},
		}

		var buf bytes.Buffer
		err := printUsersList(&buf, outputformat.Format(""), users)

		assert.NoError(t, err)
		output := buf.String()
		// Roles should be comma-separated
		assert.Contains(t, output, "admin,developer,owner")
	})
}
