package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL   string
	AccountID string
	Token     string
	HTTP      *http.Client
}

func NewClient(baseURL, accountID, token string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		AccountID: accountID,
		Token:     token,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s/v1/accounts/%s%s", c.BaseURL, c.AccountID, path)
}

func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.url(path), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.api+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseAPIError(resp.StatusCode, data)
	}

	return data, nil
}

func (c *Client) doRequestBasicAuth(method, path string, email, password string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.url(path), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(email, password)
	req.Header.Set("Accept", "application/vnd.api+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseAPIError(resp.StatusCode, data)
	}

	return data, nil
}

func (c *Client) parseAPIError(statusCode int, body []byte) error {
	var errResp struct {
		Errors []struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`
			Code   string `json:"code"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 {
		e := errResp.Errors[0]
		return fmt.Errorf("API error %d: %s - %s (code: %s)", statusCode, e.Title, e.Detail, e.Code)
	}

	return fmt.Errorf("API error %d: %s", statusCode, string(body))
}
