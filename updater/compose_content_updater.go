package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

func UpdateComposeTags(data []byte, updates []ServiceUpdate) ([]byte, error) {
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		logger.Error("failed to unmarshal YAML: %v", err)
		return nil, err
	}
	s, ok := doc["services"].(map[string]interface{})
	if !ok {
		logger.Error("services key missing in YAML document")
		return nil, errors.New("services key missing")
	}
	for _, u := range updates {
		svc, ok := s[u.ServiceName].(map[string]interface{})
		if !ok {
			logger.Error("service %s missing in YAML document", u.ServiceName)
			return nil, fmt.Errorf("service %s missing", u.ServiceName)
		}
		img, ok := svc["image"].(string)
		if !ok {
			logger.Error("image missing in service %s", u.ServiceName)
			return nil, fmt.Errorf("image missing in service %s", u.ServiceName)
		}
		repo := strings.Split(img, ":")[0]
		svc["image"] = fmt.Sprintf("%s:%s", repo, u.NewTag)
	}
	return yaml.Marshal(doc)
}
