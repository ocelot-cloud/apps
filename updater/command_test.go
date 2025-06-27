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
	tests := []struct {
		name        string
		originalTag string
		tagList     []string
		newTag      string
		errorMsg    string
	}{
		{"same tag", "1.22", []string{"1.22"}, "", ""},
		{"same and lower tag", "1.22", []string{"1.21", "1.22"}, "", ""},
		{"ignore non-numeric tags", "1.22", []string{"latest", "1.21", "stable", ""}, "", ""},
		{"find simple higher tag", "1.22", []string{"1.21", "1.23"}, "1.23", ""},
		{"ignore tags with different number count", "1.22", []string{"1.21", "1.23", "2", "2.23.2"}, "1.23", ""},

		{"empty original tag causes error", "", []string{"1.21", "1.23", "1.23.2"}, "", "integer conversion failed"},
		{"word original tag causes error", "latest", []string{"1.21", "1.23", "1.23.2"}, "", "integer conversion failed"},

		{"ignore different prefixes and suffixes", "1.22", []string{"1.21", "1.23", "v1.24", "1.24-alpine", "v1.24-alpine"}, "1.23", ""},
		{"find newest tag with same prefix", "v1.22", []string{"1.21", "1.24", "v1.21", "v1.23"}, "v1.23", ""},
		{"find newest tag with same suffix", "1.22-alpine", []string{"1.21", "1.24", "1.21-alpine", "1.23-alpine"}, "1.23-alpine", ""},
		{"find newest tag with same prefix and suffix", "v1.22-alpine", []string{"1.21", "1.24", "v1.21", "v1.24", "1.21-alpine", "1.24-alpine", "v1.21-alpine", "v1.23-alpine"}, "v1.23-alpine", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newTag, err := FilterLatestImageTag(tc.originalTag, tc.tagList)
			if tc.errorMsg != "" {
				assert.Equal(t, tc.errorMsg, err.Error())
			}
			assert.Equal(t, tc.newTag, newTag)
		})
	}
}

func TestIntSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []int
		expected bool
	}{
		{"both empty", []int{}, []int{}, true},
		{"equal slices", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different lengths", []int{1, 2}, []int{1, 2, 3}, false},
		{"different content", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"one empty", []int{}, []int{1}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := intSlicesEqual(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("intSlicesEqual(%v, %v) = %v, expected %v", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		want      []int
		errSubstr string
	}{
		{"valid single number", "1", []int{1}, ""},
		{"valid two numbers with prefix", "1.22", []int{1, 22}, ""},
		{"valid three numbers with prefix", "1.2.3", []int{1, 2, 3}, ""},
		{"invalid non-numeric", "latest", nil, "integer conversion failed"},
		{"invalid mixed", "1.2.latest", nil, "integer conversion failed"},
		{"empty string", "", nil, "integer conversion failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.tag)
			if tt.errSubstr != "" {
				assert.Equal(t, tt.errSubstr, err.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaxIntSlice(t *testing.T) {
	tests := []struct {
		name          string
		desiredLength int
		inputSlices   [][]int
		expected      []int
	}{
		{"empty input", 3, [][]int{}, nil},
		{"slices must have same length", 3, [][]int{{1, 2}, {1, 2, 3}}, []int{1, 2, 3}},
		{"single slice", 3, [][]int{{1, 2, 3}}, []int{1, 2, 3}},

		{"multiple slices, max at start", 3, [][]int{{2, 2, 3}, {1, 2, 3}}, []int{2, 2, 3}},
		{"multiple slices, max at end", 3, [][]int{{1, 2, 3}, {1, 2, 4}}, []int{1, 2, 4}},
		{"multiple slices, max in middle", 3, [][]int{{1, 2, 3}, {1, 3, 2}, {1, 2, 2}}, []int{1, 3, 2}},
		{"equal slices", 3, [][]int{{1, 2, 3}, {1, 2, 3}}, []int{1, 2, 3}},

		{"take two element slice", 2, [][]int{{1, 2}, {1, 2, 3}, {1, 2, 3, 4}}, []int{1, 2}},
		{"take three element slice", 3, [][]int{{1, 2}, {1, 2, 3}, {1, 2, 3, 4}}, []int{1, 2, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := findMaxIntSlice(tc.desiredLength, tc.inputSlices)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIntSliceToString(t *testing.T) {
	tests := []struct {
		input    []int
		expected string
	}{
		{[]int{1}, "1"},
		{[]int{1, 2}, "1.2"},
		{[]int{1, 2, 3}, "1.2.3"},
		{[]int{}, ""},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, intSliceToString(tc.input))
	}
}

// TODO add ci pipeline
// TODO skip when tag is same as original tag, should not be returned as new tag -> hasNewerVersion = false
// sampleTagList := []string{"latest", "1.21", "1.22", "1.23", "v1.24", "1.25-alpine", "v1.26-alpine"}
// TODO case: mixes tag schemas, like 1.2 and 1.2.3 -> stick to the original tag schema
// TODO also test with custom 1) prefix and 2) suffix
// TODO maybe publish this at the end as CLI tool so others can use it
// TODO do I need and app store GUI at all, if I can simply interact with the server via CLI? Maybe even smarter, since it can automate stuff like updating, zipping, signing and uploading the app, etc.
// TODO add mocking to cloud, app store and shared module; improve test coverage? e.g. request logic in shared module -> separate requests from processing logic
// TODO app store: make mock that ignores the .env file in non-prod profiles
