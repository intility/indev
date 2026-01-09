package teams

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

func TestPrintTeamsList(t *testing.T) {
	sampleTeams := []client.Team{
		{
			ID:          "team-1",
			Name:        "Alpha Team",
			Description: "First team",
			Role:        []string{},
		},
		{
			ID:          "team-2",
			Name:        "Beta Team",
			Description: "Second team",
			Role:        []string{"owner"},
		},
		{
			ID:          "team-3",
			Name:        "Gamma Team",
			Description: "Third team",
			Role:        []string{"member"},
		},
	}

	t.Run("default format shows name, description, role", func(t *testing.T) {
		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format(""), sampleTeams)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Alpha Team")
		assert.Contains(t, output, "Beta Team")
		assert.Contains(t, output, "First team")
		// Should NOT contain ID in default format
		assert.NotContains(t, output, "team-1")
	})

	t.Run("wide format includes ID", func(t *testing.T) {
		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format("wide"), sampleTeams)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "team-1")
		assert.Contains(t, output, "team-2")
		assert.Contains(t, output, "Alpha Team")
	})

	t.Run("json format outputs valid JSON", func(t *testing.T) {
		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format("json"), sampleTeams)

		assert.NoError(t, err)

		var decoded []client.Team
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 3)
	})

	t.Run("yaml format outputs valid YAML", func(t *testing.T) {
		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format("yaml"), sampleTeams)

		assert.NoError(t, err)

		var decoded []client.Team
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 3)
	})

	t.Run("sorts teams with roles first", func(t *testing.T) {
		// Teams without roles should be sorted after teams with roles
		teamsUnsorted := []client.Team{
			{ID: "1", Name: "No Role", Role: []string{}},
			{ID: "2", Name: "Has Role", Role: []string{"member"}},
		}

		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format("json"), teamsUnsorted)

		assert.NoError(t, err)

		var decoded []client.Team
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)

		// Team with role should come first after sorting
		assert.Equal(t, "Has Role", decoded[0].Name)
		assert.Equal(t, "No Role", decoded[1].Name)
	})

	t.Run("handles empty teams list", func(t *testing.T) {
		var buf bytes.Buffer
		err := printTeamsList(&buf, outputformat.Format("json"), []client.Team{})

		assert.NoError(t, err)

		var decoded []client.Team
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 0)
	})
}
