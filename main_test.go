package main

import (
	"testing"
)

func TestIcpctl(t *testing.T) {
	run([]string{"icpctl", "version"})
}
