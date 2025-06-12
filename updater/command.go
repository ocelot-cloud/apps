package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DockerHubClient interface{}

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
