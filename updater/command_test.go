package main

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestDockerTagPull(t *testing.T) {
	// When changing the real DockerHub client code, remove the lines and execute the test to assert that the code is working.
	t.Skip()
	return

	realDockerHubClient := dockerHubClientReal{}
	tags, err := realDockerHubClient.listImageTags("nginx")
	assert.Nil(t, err)
	assert.Equal(t, 100, len(tags))
	println("listing tags for nginx:")
	for _, tag := range tags {
		println(tag)
	}
}

func TestDockerhubMock(t *testing.T) {
	mockDockerHubClient := NewDockerHubClientMock(t)
	mockDockerHubClient.EXPECT().listImageTags("nginx").Return([]string{"latest", "1.21", "1.22"}, nil)

	tags, err := mockDockerHubClient.listImageTags("nginx")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, tags[0], "latest")
	assert.Equal(t, tags[1], "1.21")
	assert.Equal(t, tags[2], "1.22")
	mockDockerHubClient.AssertExpectations(t)
}
