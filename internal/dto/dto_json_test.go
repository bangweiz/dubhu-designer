package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func assertNoSnakeCaseKeys(t *testing.T, dtoName string, data []byte) {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("%s failed to unmarshal json: %v", dtoName, err)
	}

	for k := range raw {
		if strings.Contains(k, "_") {
			t.Errorf("%s contains snake_case key '%s': %s", dtoName, k, string(data))
		}
	}
}

func TestDTOJSONSerialization_CamelCase(t *testing.T) {
	now := time.Now().UTC()

	t.Run("ToolResponseDTO", func(t *testing.T) {
		dto := ToolResponseDTO{
			ID:          "651da3b57f8641a27e36c579",
			Name:        "tool",
			Description: "desc",
			Inputs:      []ToolInputResponseDTO{},
			Outputs:     []ToolOutputResponseDTO{},
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "ToolResponseDTO", data)
		str := string(data)
		if !strings.Contains(str, `"createdAt"`) || !strings.Contains(str, `"updatedAt"`) {
			t.Errorf("expected createdAt and updatedAt in json: %s", str)
		}
	})

	t.Run("InstructionSummaryResponseDTO", func(t *testing.T) {
		dto := InstructionSummaryResponseDTO{
			ID:        "651da3b57f8641a27e36c579",
			Name:      "inst",
			Content:   "content",
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "InstructionSummaryResponseDTO", data)
	})

	t.Run("InstructionResponseDTO", func(t *testing.T) {
		dto := InstructionResponseDTO{
			ID:        "651da3b57f8641a27e36c579",
			Name:      "inst",
			Content:   "content",
			Tools:     []ToolResponseDTO{},
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "InstructionResponseDTO", data)
	})

	t.Run("ConciergeSummaryResponseDTO", func(t *testing.T) {
		dto := ConciergeSummaryResponseDTO{
			ID:          "651da3b57f8641a27e36c579",
			Name:        "concierge",
			Description: "desc",
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "ConciergeSummaryResponseDTO", data)
	})

	t.Run("ConciergeResponseDTO", func(t *testing.T) {
		dto := ConciergeResponseDTO{
			ID:          "651da3b57f8641a27e36c579",
			Name:        "concierge",
			Description: "desc",
			Agents:      []AgentResponseDTO{},
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "ConciergeResponseDTO", data)
	})

	t.Run("AgentResponseDTO", func(t *testing.T) {
		dto := AgentResponseDTO{
			ID:          "651da3b57f8641a27e36c582",
			ConciergeID: "651da3b57f8641a27e36c579",
			Name:        "agent",
			Description: "desc",
			Model:       "gemini-3.5-flash",
			Version:     1,
		}
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		assertNoSnakeCaseKeys(t, "AgentResponseDTO", data)
		str := string(data)
		if !strings.Contains(str, `"conciergeId"`) {
			t.Errorf("expected conciergeId in json: %s", str)
		}
	})
}
