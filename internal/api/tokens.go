package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) CreateToken(email, password string) (*Token, error) {
	data, err := c.doRequestBasicAuth("POST", "/tokens", email, password, nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	token := parseToken(res)
	return &token, nil
}

func (c *Client) ValidateToken() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.url("/me"), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseAPIError(resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result, nil
}
