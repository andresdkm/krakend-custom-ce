package main

import (
	"encoding/json"
	"kraken-builder-plugins/pkg/hotels/common/models"
	"os"
	"testing"
)

func TestCanTransform(t *testing.T) {
	transformer := ExpediaTransformerImpl()
	data := loadExpediaResponse(t)
	if !transformer.CanTransform(data) {
		t.Errorf("Expected CanTransform to return true")
	}
	data = map[string]interface{}{"OtherKey": "value"}
	if transformer.CanTransform(data) {
		t.Errorf("Expected CanTransform to return false")
	}
}

func TestTransform_EmptyResults(t *testing.T) {
	transformer := ExpediaTransformerImpl()
	data := map[string]interface{}{}
	config := models.TransformationConfig{}
	resp, err := transformer.Transform(data, config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(resp.Hotels) != 0 {
		t.Errorf("Expected 0 hotels, got %d", len(resp.Hotels))
	}
}

func TestTransform_WithResults(t *testing.T) {
	transformer := ExpediaTransformerImpl()
	data := loadExpediaResponse(t)
	config := models.TransformationConfig{}
	resp, err := transformer.Transform(data, config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(resp.Hotels) != 11 {
		t.Errorf("Expected 11 hotel, got %d", len(resp.Hotels))
	}
	if resp.Hotels[0].ProductID != "118938708" {
		t.Errorf("Expected ProductID '118938708', got '%s'", resp.Hotels[0].ProductID)
	}
}

func loadExpediaResponse(t *testing.T) map[string]interface{} {
	file, err := os.Open("expedia_response.json")
	if err != nil {
		t.Fatalf("Failed to open expedia_response.json: %v", err)
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		t.Fatalf("Failed to decode expedia_response.json: %v", err)
	}
	return data
}
