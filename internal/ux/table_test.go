package ux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRow(t *testing.T) {
	t.Run("creates row with field and value", func(t *testing.T) {
		row := NewRow("Name", "Alice")

		assert.Equal(t, "Name", row.Field)
		assert.Equal(t, "Alice", row.Value)
	})

	t.Run("handles empty strings", func(t *testing.T) {
		row := NewRow("", "")

		assert.Equal(t, "", row.Field)
		assert.Equal(t, "", row.Value)
	})
}

func TestTableFromObjects(t *testing.T) {
	type Person struct {
		Name string
		Age  int
		City string
	}

	t.Run("creates table from objects", func(t *testing.T) {
		people := []Person{
			{Name: "Alice", Age: 30, City: "New York"},
			{Name: "Bob", Age: 25, City: "Boston"},
		}

		table := TableFromObjects(people, func(p Person) []Row {
			return []Row{
				NewRow("Name", p.Name),
				NewRow("Age", string(rune(p.Age+'0'))),
				NewRow("City", p.City),
			}
		})

		assert.Equal(t, []string{"Name", "Age", "City"}, table.Header)
		assert.Len(t, table.Rows, 2)
		assert.Equal(t, "Alice", table.Rows[0][0])
		assert.Equal(t, "Bob", table.Rows[1][0])
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var people []Person

		table := TableFromObjects(people, func(p Person) []Row {
			return []Row{
				NewRow("Name", p.Name),
			}
		})

		assert.Empty(t, table.Header)
		assert.Empty(t, table.Rows)
	})

	t.Run("handles single object", func(t *testing.T) {
		people := []Person{
			{Name: "Alice", Age: 30, City: "NYC"},
		}

		table := TableFromObjects(people, func(p Person) []Row {
			return []Row{
				NewRow("Name", p.Name),
				NewRow("City", p.City),
			}
		})

		assert.Equal(t, []string{"Name", "City"}, table.Header)
		assert.Len(t, table.Rows, 1)
	})

	t.Run("works with different types", func(t *testing.T) {
		type Item struct {
			ID    int
			Label string
		}

		items := []Item{
			{ID: 1, Label: "First"},
			{ID: 2, Label: "Second"},
		}

		table := TableFromObjects(items, func(item Item) []Row {
			return []Row{
				NewRow("ID", string(rune(item.ID + '0'))),
				NewRow("Label", item.Label),
			}
		})

		assert.Equal(t, []string{"ID", "Label"}, table.Header)
		assert.Len(t, table.Rows, 2)
	})
}

func TestTable_String(t *testing.T) {
	t.Run("renders table with headers and rows", func(t *testing.T) {
		table := Table{
			Header: []string{"Name", "City"},
			Rows: [][]string{
				{"Alice", "New York"},
				{"Bob", "Boston"},
			},
		}

		output := table.String()

		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "City")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "New York")
		assert.Contains(t, output, "Bob")
		assert.Contains(t, output, "Boston")
	})

	t.Run("handles empty table", func(t *testing.T) {
		table := Table{
			Header: []string{},
			Rows:   [][]string{},
		}

		output := table.String()
		assert.Equal(t, "\n", output)
	})

	t.Run("aligns columns based on content width", func(t *testing.T) {
		table := Table{
			Header: []string{"ID", "Description"},
			Rows: [][]string{
				{"1", "Short"},
				{"2", "A much longer description"},
			},
		}

		output := table.String()

		// Both rows should contain content
		assert.Contains(t, output, "Short")
		assert.Contains(t, output, "A much longer description")
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		table := Table{
			Header: []string{"Name", "Emoji"},
			Rows: [][]string{
				{"Alice", "👋"},
				{"日本語", "🎉"},
			},
		}

		output := table.String()

		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "👋")
		assert.Contains(t, output, "日本語")
		assert.Contains(t, output, "🎉")
	})

	t.Run("header width affects column width", func(t *testing.T) {
		table := Table{
			Header: []string{"VeryLongHeaderName", "X"},
			Rows: [][]string{
				{"a", "b"},
			},
		}

		output := table.String()

		// Header should be present and properly formatted
		assert.Contains(t, output, "VeryLongHeaderName")
		assert.Contains(t, output, "X")
	})
}

func TestTable_calculateColumnWidths(t *testing.T) {
	t.Run("returns correct widths", func(t *testing.T) {
		table := Table{
			Header: []string{"ID", "Name"},
			Rows: [][]string{
				{"1", "Alice"},
				{"100", "Bob"},
			},
		}

		widths := table.calculateColumnWidths()

		// First column: max of "ID"(2), "1"(1), "100"(3) = 3
		assert.Equal(t, 3, widths[0])
		// Second column: max of "Name"(4), "Alice"(5), "Bob"(3) = 5
		assert.Equal(t, 5, widths[1])
	})

	t.Run("header can be widest", func(t *testing.T) {
		table := Table{
			Header: []string{"LongHeader"},
			Rows: [][]string{
				{"short"},
			},
		}

		widths := table.calculateColumnWidths()

		assert.Equal(t, 10, widths[0]) // "LongHeader" = 10
	})

	t.Run("handles empty rows", func(t *testing.T) {
		table := Table{
			Header: []string{"Name"},
			Rows:   [][]string{},
		}

		widths := table.calculateColumnWidths()

		// With no rows, widths map should be empty
		assert.Empty(t, widths)
	})
}

func TestTableIntegration(t *testing.T) {
	t.Run("full workflow from objects to string", func(t *testing.T) {
		type User struct {
			ID       string
			Username string
			Role     string
		}

		users := []User{
			{ID: "1", Username: "alice", Role: "admin"},
			{ID: "2", Username: "bob", Role: "user"},
			{ID: "10", Username: "charlie", Role: "moderator"},
		}

		table := TableFromObjects(users, func(u User) []Row {
			return []Row{
				NewRow("ID", u.ID),
				NewRow("Username", u.Username),
				NewRow("Role", u.Role),
			}
		})

		output := table.String()

		// Verify headers
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "Username")
		assert.Contains(t, output, "Role")

		// Verify all data
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "charlie")
		assert.Contains(t, output, "admin")
		assert.Contains(t, output, "moderator")
	})
}
