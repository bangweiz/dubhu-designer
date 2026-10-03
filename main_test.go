package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestOpenAPISpecValid(t *testing.T) {
	data, err := os.ReadFile("swagger/openapi.yaml")
	if err != nil {
		t.Fatalf("failed to read swagger/openapi.yaml: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse swagger/openapi.yaml: %v", err)
	}

	if parsed["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %v", parsed["openapi"])
	}
}

func TestPostmanCollectionValid(t *testing.T) {
	data, err := os.ReadFile("dubhu-designer.postman_collection.json")
	if err != nil {
		t.Fatalf("failed to read dubhu-designer.postman_collection.json: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse dubhu-designer.postman_collection.json: %v", err)
	}

	info, ok := parsed["info"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected info object in postman collection")
	}

	if info["schema"] != "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" {
		t.Errorf("unexpected schema in postman collection: %v", info["schema"])
	}
}

