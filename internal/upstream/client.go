package upstream

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL string, apiKey string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL不得为空")
	}

	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("apiKey不得为空")
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("超时必须大于零")
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Do(ctx context.Context, method string, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("请求创建失败|Method: %s, Path: %s, Error: %w", method, path, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求目标失败: %w", err)
	}

	return resp, nil
}

func (c *Client) ChatCompletion(ctx context.Context, body []byte) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, "/chat/completions", body)
}
