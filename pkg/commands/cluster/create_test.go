package cluster

import (
	"regexp"
	"testing"
)

func TestGenerateSuffixReturnsValidString(t *testing.T) {
	suffix := generateSuffix()
	want := regexp.MustCompile(`[a-z0-9]{6}`)
	if !want.MatchString(suffix) {
		t.Errorf("Suffix does not return expected format, expected: [a-z0-9]{6}, got: %s.", suffix)
	}
}
