package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Request represents a request to Ollama API
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response represents a response from Ollama API
type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// GenerateResponse generates a response using Ollama
func GenerateResponse(prompt string) (string, error) {
	req := Request{
		Model: "llama3.1:8b", //Using llama3-groq-tool-use:8b as it's available and good for conversational AI
		// Model: "llama3-groq-tool-use:8b", //Using llama3-groq-tool-use:8b as it's available and good for conversational AI
		// Model:  "phi3:mini", // Using phi3:mini as it's available and good for conversational AI
		Prompt: prompt,
		Stream: false,
	}

	start := time.Now()
	log.Printf("ollama: generate request started model=%s at=%s", req.Model, start.Format(time.RFC3339Nano))

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("ollama: generate request failed model=%s err=%v elapsed=%s", req.Model, err, time.Since(start))
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ollama: generate request failed model=%s err=%v elapsed=%s", req.Model, err, time.Since(start))
		return "", fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
		log.Printf("ollama: generate request failed model=%s err=%v elapsed=%s", req.Model, err, time.Since(start))
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ollama: generate request failed model=%s err=%v elapsed=%s", req.Model, err, time.Since(start))
		return "", fmt.Errorf("reading response body: %w", err)
	}

	var ollamaResp Response
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("ollama: generate request failed model=%s err=%v elapsed=%s", req.Model, err, time.Since(start))
		return "", fmt.Errorf("unmarshaling response: %w", err)
	}

	log.Printf("ollama: generate request completed model=%s elapsed=%s", req.Model, time.Since(start))
	return ollamaResp.Response, nil
}

// IsOllamaRunning checks if Ollama is running and accessible
func IsOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
