package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) ListUsers(params map[string]string) ([]User, error) {
	query := url.Values{}
	for k, v := range params {
		if v != "" {
			query.Set(k, v)
		}
	}

	path := "/users"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var resources []JSONAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, fmt.Errorf("parsing users: %w", err)
	}

	users := make([]User, len(resources))
	for i, res := range resources {
		users[i] = parseUser(res)
	}

	return users, nil
}

func (c *Client) GetUser(id string) (*User, error) {
	data, err := c.doRequest("GET", "/users/"+id, nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}

	user := parseUser(res)
	return &user, nil
}

func (c *Client) FindUserByEmail(email string) (*User, error) {
	users, err := c.ListUsers(map[string]string{"email": email})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found: %s", email)
	}
	return &users[0], nil
}

func (c *Client) GetUserLicenses(userID string) ([]License, error) {
	data, err := c.doRequest("GET", "/users/"+userID+"/licenses", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var resources []JSONAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, fmt.Errorf("parsing licenses: %w", err)
	}

	licenses := make([]License, len(resources))
	for i, res := range resources {
		licenses[i] = parseLicense(res)
	}

	return licenses, nil
}

func (c *Client) UpdateUser(id string, attrs map[string]interface{}) (*User, error) {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "users",
			"id":         id,
			"attributes": attrs,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	data, err := c.doRequest("PATCH", "/users/"+id, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}

	user := parseUser(res)
	return &user, nil
}
