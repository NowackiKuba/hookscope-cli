package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	Token  string
}

func (c *Client) Post(path string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	client := &http.Client{}
	return client.Do(req)
}
