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

type Gemini struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGemini(apiKey, model string) *Gemini {
	return &Gemini{
		apiKey: apiKey,
		model:  model,
		client: http.DefaultClient,
	}
}

type geminiRequest struct {
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	Contents          []geminiContent  `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (g *Gemini) Chat(ctx context.Context, messages []Message) (string, error) {
	var systemParts []string
	var contents []geminiContent

	for _, m := range messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, m.Content)
		case "assistant":
			contents = append(contents, geminiContent{
				Role:  "model",
				Parts: []geminiPart{{Text: m.Content}},
			})
		case "user":
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		default:
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	reqBody := geminiRequest{Contents: contents}
	if len(systemParts) > 0 {
		reqBody.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: strings.Join(systemParts, "\n\n")}},
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
		g.model,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-goog-api-key", g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("gemini: parse response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(respBody))
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
