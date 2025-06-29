package main

import (
	"fmt"
	"net/http"
	"time"
)

type EndpointCheckerImpl struct{}

func (e EndpointCheckerImpl) TryAccessingIndexPageOnLocalhost(port string, path string) error {
	url := fmt.Sprintf("http://localhost:%s%s", port, path)
	timeout := 2 * time.Minute
	deadline := time.Now().Add(timeout)
	client := http.Client{Timeout: 5 * time.Second}

	for {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			logger.Error("unexpected status code %d for %s", resp.StatusCode, url)
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}

		if time.Now().After(deadline) {
			logger.Error("timeout reaching %s: %v", url, err)
			return fmt.Errorf("timeout reaching %s: %w", url, err)
		}
		time.Sleep(2 * time.Second)
	}
}
