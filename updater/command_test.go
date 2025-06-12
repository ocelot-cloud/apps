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

func TestFilterLatestImageTag(t *testing.T) {
	originalTag := "1.22"
	sampleTagList := []string{"1.22"}

	newTag, wasNewerVersionFound, err := filterLatestImageTag(originalTag, sampleTagList)
	assert.Nil(t, err)
	assert.False(t, wasNewerVersionFound)
	assert.Equal(t, "", newTag)
}

// TODO skip when tag is same as original tag, should not be returned as new tag -> hasNewerVersion = false
// sampleTagList := []string{"latest", "1.21", "1.22", "1.23", "v1.24", "1.25-alpine", "v1.26-alpine"}
// TODO case: mixes tag schemas, like 1.2 and 1.2.3 -> stick to the original tag schema
// TODO also test with custom 1) prefix and 2) suffix
