package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

//go:generate mockery
type DockerHubClient interface {
	listImageTags(image string) ([]string, error)
}

type dockerHubClientReal struct{}

func (d *dockerHubClientReal) listImageTags(image string) ([]string, error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags?page_size=100", image)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	tags := make([]string, len(out.Results))
	for i, r := range out.Results {
		tags[i] = r.Name
	}
	return tags, nil
}

func filterLatestImageTag(originalTag string, tagList []string) (string, bool, error) {
	prefix := ""
	if strings.HasPrefix(originalTag, "v") {
		prefix = "v"
		originalTag = strings.TrimPrefix(originalTag, "v")
	}
	if prefix != "" {
		filtered := make([]string, 0, len(tagList))
		for _, tag := range tagList {
			if strings.HasPrefix(tag, "v") {
				tagWithoutPrefix := strings.TrimPrefix(tag, "v")
				filtered = append(filtered, tagWithoutPrefix)
			}
		}
		tagList = filtered
	}

	originalTagNumbers, err := parse(originalTag)
	if err != nil {
		return "", false, err
	}

	var listOfAllTagNumbers [][]int
	for _, tag := range tagList {
		parsed, err := parse(tag)
		if err != nil {
			continue
		}
		listOfAllTagNumbers = append(listOfAllTagNumbers, parsed)
	}
	listOfAllTagNumbers = append(listOfAllTagNumbers, originalTagNumbers)

	tagNumbersWithHighestVersion := findMaxIntSlice(len(originalTagNumbers), listOfAllTagNumbers)
	if intSlicesEqual(originalTagNumbers, tagNumbersWithHighestVersion) {
		return "", false, nil
	} else {
		newestTag := intSliceToString(tagNumbersWithHighestVersion)
		newestTag = prefix + newestTag
		return newestTag, true, nil
	}
}

func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func intSliceToString(slice []int) string {
	strs := make([]string, len(slice))
	for i, n := range slice {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, ".")
}

func parse(tag string) ([]int, error) {
	parts := strings.Split(tag, ".")
	ints := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			logger.Info("integer conversion failed for '%s': %v", tag, err)
			return nil, fmt.Errorf("integer conversion failed")
		}
		ints[i] = n
	}
	return ints, nil
}

func findMaxIntSlice(desiredLength int, slices [][]int) []int {
	if len(slices) == 0 {
		return nil
	}
	var maxSlice []int
	for _, s := range slices {
		if len(s) != desiredLength {
			continue
		}
		if maxSlice == nil {
			maxSlice = s
			continue
		}

		for i := 0; i < desiredLength; i++ {
			if s[i] > maxSlice[i] {
				maxSlice = s
				break
			} else if s[i] < maxSlice[i] {
				break
			}
		}
	}
	return maxSlice
}
