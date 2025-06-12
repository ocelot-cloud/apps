package main

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestDockerTagPull(t *testing.T) {
	realDockerHubClient := dockerHubClientReal{}
	tags, err := realDockerHubClient.listImageTags("nginx")
	assert.Nil(t, err)
	assert.Equal(t, 100, len(tags))
	println("listing tags for nginx:")
	for _, tag := range tags {
		println(tag)
	}
}
