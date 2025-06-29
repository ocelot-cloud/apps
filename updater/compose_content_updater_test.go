package main

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func equalYAML(t *testing.T, a, b []byte) {
	var v1, v2 interface{}
	yaml.Unmarshal(a, &v1)
	yaml.Unmarshal(b, &v2)
	assert.Equal(t, v2, v1)
}

func TestUpdateComposeTags(t *testing.T) {
	in := `
services:
  web:
    image: nginx:1.24
`
	want := `
services:
  web:
    image: nginx:1.25
`
	out, err := UpdateComposeTags([]byte(in), []ServiceUpdate{{ServiceName: "web", OldTag: "1.24", NewTag: "1.25"}})
	assert.NoError(t, err)
	equalYAML(t, out, []byte(want))
}

func TestUpdateTwoServices(t *testing.T) {
	in := `
services:
  db:
    image: postgres:4.5
  web:
    image: nginx:1.2
  cache:
    image: redis:6.0
`
	want := `
services:
  db:
    image: postgres:4.6
  web:
    image: nginx:1.3
  cache:
    image: redis:6.0
`
	out, err := UpdateComposeTags([]byte(in), []ServiceUpdate{
		{ServiceName: "web", OldTag: "1.2", NewTag: "1.3"},
		{ServiceName: "db", OldTag: "4.5", NewTag: "4.6"},
	})
	assert.NoError(t, err)
	equalYAML(t, out, []byte(want))
}

func TestMissingServices(t *testing.T) {
	_, err := UpdateComposeTags([]byte(`version: "3"`), nil)
	assert.Error(t, err)
}

func TestMissingService(t *testing.T) {
	in := `
services:
  db:
    image: postgres:15
`
	_, err := UpdateComposeTags([]byte(in), []ServiceUpdate{{ServiceName: "web", NewTag: "1.25"}})
	assert.Error(t, err)
}

func TestMissingImage(t *testing.T) {
	in := `
services:
  web:
    build: .
`
	_, err := UpdateComposeTags([]byte(in), []ServiceUpdate{{ServiceName: "web", NewTag: "1.25"}})
	assert.Error(t, err)
}
