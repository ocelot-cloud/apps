package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestIsSecondVersionNewerThanFirstVersion(t *testing.T) {
	cases := []struct {
		first  string
		second string
		expect bool
	}{
		{"1.2.3", "1.3.0", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.3", "1.2.2", false},
		{"v1.2.3", "1.2.4-alpine", true},
		{"1.2", "1.2.1", true},
		{"1.2.3-alpine", "v1.2.4", true},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, isSecondVersionNewerThanFirstVersion(c.first, c.second), c.first+" vs "+c.second)
	}
}

func TestParseVersion(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, parseVersion("v1.2.3-alpine"))
	assert.Equal(t, []int{1, 0, 0}, parseVersion("1.0.0"))
}
