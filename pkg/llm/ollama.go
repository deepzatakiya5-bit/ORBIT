package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllama(baseURL, model string) *Ollama {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		client:  http.DefaultClient,
	}
}

type ollamaChatRequest struct {
	Model    string         `json:"model"`
	Messages []ollamaMsg    `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaMsg `json:"message"`
	Error   string    `json:"error,omitempty"`
}

func (o *Ollama) Chat(ctx context.Context, messages []Message) (string, error) {
	reqMessages := make([]ollamaMsg, len(messages))
	for i, m := range messages {
		reqMessages[i] = ollamaMsg{Role: m.Role, Content: m.Content}
	}

	body, err := json.Marshal(ollamaChatRequest{
		Model:    o.model,
		Messages: reqMessages,
		Stream:   false,
		Options: map[string]any{
			"temperature": 0.5,
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed ollamaChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("ollama: parse response: %w", err)
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("ollama: %s", parsed.Error)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ollama: status %d: %s", resp.StatusCode, string(respBody))
	}
	if strings.TrimSpace(parsed.Message.Content) == "" {
		return "", fmt.Errorf("ollama: empty response")
	}
	return parsed.Message.Content, nil
}
