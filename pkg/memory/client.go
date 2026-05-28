package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	if baseURL == "" {
		return nil
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		http: &http.Client{
			Timeout: 1800 * time.Millisecond,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != ""
}

func (c *Client) BuildContext(userID, query string) (string, error) {
	if !c.Enabled() {
		return "", nil
	}

	endpoint := fmt.Sprintf("%s/v1/users/%s/memory/context?query=%s", c.baseURL, userID, url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("memory context request failed: %s", res.Status)
	}

	var payload struct {
		Context string `json:"context"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.Context, nil
}

func (c *Client) EnqueueConversationUpdated(userID, conversationID string, messageIDs []string) error {
	if !c.Enabled() {
		return nil
	}
	payload := map[string]any{
		"userId":         userID,
		"conversationId": conversationID,
		"messageIds":     messageIDs,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/events/conversation-updated", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("enqueue failed: %s", res.Status)
	}
	return nil
}
