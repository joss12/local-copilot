package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// handlelling commucation with Ollama
type OllamaClient struct {
	baseURL string
	model   string
}

// create an ollama client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	if model == "" {
		model = "deepseek-coder:1.3b-base"
	}

	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
	}
}

// GenerateRequest represents a generation request to Ollama
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResponse represents Ollama's response
type GenerateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// Using Ollama to generate code completion
func (c *OllamaClient) GenerateCompletion(language, contextBefore, contextAfter string) (string, error) {
	// Building prompt with context
	prompt := c.buildPrompt(language, contextBefore, contextAfter)

	reqBody := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Calling Ollama API
	resp, err := client.Post(
		c.baseURL+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	//Parse response
	var generateResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	//Cleaning up generated code
	completion := c.cleanCompletion(generateResp.Response)
	return completion, nil
}

func (c *OllamaClient) buildPrompt(language, contextBefore, contextAfter string) string {
	var prompt strings.Builder

	prompt.WriteString("You are a code completion assistant. Complete the code at the cursor position.\n\n")
	prompt.WriteString(fmt.Sprintf("Language: %s\n\n", language))
	prompt.WriteString("Code:\n")
	prompt.WriteString(contextBefore)
	prompt.WriteString("<CURSOR>")
	prompt.WriteString(contextAfter)
	prompt.WriteString("\n\nComplete ONLY the code at <CURSOR>. Output just the completion, no explanations:")

	return prompt.String()
}

func (c *OllamaClient) cleanCompletion(response string) string {
	//Removing markdown code blocks
	response = strings.ReplaceAll(response, "```go", "")
	response = strings.ReplaceAll(response, "```javascript", "")
	response = strings.ReplaceAll(response, "```typescript", "")
	response = strings.ReplaceAll(response, "```", "")

	//Trim whitespace
	response = strings.TrimSpace(response)

	//The frst line for inline completion
	lines := strings.Split(response, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return response
}

func (c *OllamaClient) IsAvailable() bool {
	resp, err := http.Get(c.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
