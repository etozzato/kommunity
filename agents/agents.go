package agents

import (
	"encoding/json"
	"fmt"
	"os"
)

// Agent represents an AI agent in the community
type Agent struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Style    string  `json:"style"`
	Courage  float64 `json:"courage"`
	Empathy  float64 `json:"empathy"`
	Elegance float64 `json:"elegance"`
}

// LoadAgents loads agent definitions from a JSON file
func LoadAgents(filename string) ([]Agent, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening agents file: %w", err)
	}
	defer file.Close()

	var agents []Agent
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&agents); err != nil {
		return nil, fmt.Errorf("decoding agents JSON: %w", err)
	}

	return agents, nil
}

// SaveAgents saves agent definitions to a JSON file
func SaveAgents(agents []Agent, filename string) error {
	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling agents: %w", err)
	}

	// Atomic write
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
