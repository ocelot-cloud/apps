package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type DockerHubClientImpl struct{}

func (d *DockerHubClientImpl) listImageTags(image string) ([]string, error) {
	var repoPath string
	if !strings.Contains(image, "/") {
		repoPath = "library/" + image
	} else {
		repoPath = image
	}
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags?page_size=100", repoPath)
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

func FilterLatestImageTag(originalTag string, tagList []string) (string, error) {
	prefix, originalTag, tagList := trimPrefix(originalTag, tagList)
	suffix, originalTag, tagList := trimSuffix(originalTag, tagList)

	originalTagNumbers, err := parse(originalTag)
	if err != nil {
		return "", err
	}
	listOfAllTagNumbers := parseTagList(tagList)
	listOfAllTagNumbers = append(listOfAllTagNumbers, originalTagNumbers)

	tagNumbersWithHighestVersion := findMaxIntSlice(len(originalTagNumbers), listOfAllTagNumbers)
	if intSlicesEqual(originalTagNumbers, tagNumbersWithHighestVersion) {
		return "", nil
	} else {
		newestTag := intSliceToString(tagNumbersWithHighestVersion)
		newestTag = prefix + newestTag
		newestTag = newestTag + suffix
		return newestTag, nil
	}
}

func parseTagList(tagList []string) [][]int {
	var listOfAllTagNumbers [][]int
	for _, tag := range tagList {
		parsed, err := parse(tag)
		if err != nil {
			continue
		}
		listOfAllTagNumbers = append(listOfAllTagNumbers, parsed)
	}
	return listOfAllTagNumbers
}

func trimSuffix(originalTag string, tagList []string) (string, string, []string) {
	suffix := ""
	suffixIdx := strings.Index(originalTag, "-")
	if suffixIdx != -1 {
		suffix = originalTag[suffixIdx:]
		originalTag = originalTag[:suffixIdx]
	}
	if suffix != "" {
		filtered := make([]string, 0, len(tagList))
		for _, tag := range tagList {
			if strings.HasSuffix(tag, suffix) {
				tagWithoutSuffix := strings.TrimSuffix(tag, suffix)
				filtered = append(filtered, tagWithoutSuffix)
			}
		}
		tagList = filtered
	}
	return suffix, originalTag, tagList
}

func trimPrefix(originalTag string, tagList []string) (string, string, []string) {
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
	return prefix, originalTag, tagList
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
			logger.Debug("integer conversion failed for '%s': %v", tag, err)
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
