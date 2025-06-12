package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestIsSecondVersionNewerThanFirstVersion(t *testing.T) {
	cases := []struct {
		first   string
		second  string
		newer   bool
		wantErr bool
	}{
		{"1.2.3", "1.3.0", true, false},
		{"1.2.3", "1.2.3", false, false},
		{"1.2.3", "1.2.2", false, false},
		{"v1.2.3", "v1.2.4", true, false},
		{"v1.2.3", "1.2.4", false, true},
		{"1.2.3-alpine", "1.2.4-alpine", true, false},
		{"1.2.3-alpine", "1.2.4-slim", false, true},
		{"1.2", "1.2.1", true, false},
	}
	for _, c := range cases {
		newer, err := isSecondVersionNewerThanFirstVersion(c.first, c.second)
		if c.wantErr {
			assert.Error(t, err, c.first+" vs "+c.second)
		} else {
			assert.NoError(t, err, c.first+" vs "+c.second)
			assert.Equal(t, c.newer, newer, c.first+" vs "+c.second)
		}
	}
}

func TestParseVersion(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, parseVersion("v1.2.3-alpine"))
	assert.Equal(t, []int{1, 0, 0}, parseVersion("1.0.0"))
	assert.Equal(t, []int{1, 1}, parseVersion("1.1"))
}
