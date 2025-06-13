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
	_, err := parse(originalTag)
	if err != nil {
		return "", false, err
	}
	return "", false, nil
}

func parse(tag string) ([]int, error) {
	parts := strings.Split(tag, ".")
	ints := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			logger.Error("integer conversion failed for '%s': %v", tag, err)
			return nil, fmt.Errorf("integer conversion failed")
		}
		ints[i] = n
	}
	return ints, nil
}

func maxIntSlice(slices [][]int) ([]int, error) {
	if len(slices) == 0 {
		return nil, fmt.Errorf("no slices passed")
	}
	length := len(slices[0])
	for i, s := range slices {
		if len(s) != length {
			logger.Error("slice at index %d has a length %d but it must have same length as first slice with length %d", i, len(s), length)
			return nil, fmt.Errorf("slices must have the same length")
		}
	}
	maxSlice := slices[0]
	for _, s := range slices[1:] {
		for i := 0; i < length; i++ {
			if s[i] > maxSlice[i] {
				maxSlice = s
				break
			} else if s[i] < maxSlice[i] {
				break
			}
		}
	}
	return maxSlice, nil
}
