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
