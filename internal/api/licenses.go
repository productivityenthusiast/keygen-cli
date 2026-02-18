package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) ListLicenses(params map[string]string) ([]License, error) {
	query := url.Values{}
	for k, v := range params {
		if v != "" {
			query.Set(k, v)
		}
	}

	path := "/licenses"
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
		return nil, fmt.Errorf("parsing license list: %w", err)
	}

	licenses := make([]License, len(resources))
	for i, res := range resources {
		licenses[i] = parseLicense(res)
	}

	return licenses, nil
}

func (c *Client) GetLicense(id string) (*License, error) {
	data, err := c.doRequest("GET", "/licenses/"+id, nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing license: %w", err)
	}

	license := parseLicense(res)
	return &license, nil
}

func (c *Client) ValidateLicense(id string) (*LicenseValidation, *License, error) {
	data, err := c.doRequest("POST", "/licenses/"+id+"/actions/validate", nil)
	if err != nil {
		return nil, nil, err
	}

	var resp struct {
		Data json.RawMessage        `json:"data"`
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, nil, fmt.Errorf("parsing response: %w", err)
	}

	validation := &LicenseValidation{}
	if resp.Meta != nil {
		if v, ok := resp.Meta["valid"].(bool); ok {
			validation.Valid = v
		}
		if d, ok := resp.Meta["detail"].(string); ok {
			validation.Detail = d
		}
		if code, ok := resp.Meta["code"].(string); ok {
			validation.Code = code
		}
	}

	var res JSONAPIResource
	if err := json.Unmarshal(resp.Data, &res); err != nil {
		return validation, nil, nil
	}

	license := parseLicense(res)
	return validation, &license, nil
}

func (c *Client) RenewLicense(id string) (*License, error) {
	data, err := c.doRequest("POST", "/licenses/"+id+"/actions/renew", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing license: %w", err)
	}

	license := parseLicense(res)
	return &license, nil
}

func (c *Client) GetLicenseMachines(licenseID string) ([]Machine, error) {
	data, err := c.doRequest("GET", "/licenses/"+licenseID+"/machines?include=components", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var resources []JSONAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, fmt.Errorf("parsing machines: %w", err)
	}

	// Parse included components
	componentMap := make(map[string][]Component)
	for _, inc := range doc.Included {
		if inc.Type == "components" {
			comp := parseComponent(inc)
			componentMap[comp.MachineID] = append(componentMap[comp.MachineID], comp)
		}
	}

	machines := make([]Machine, len(resources))
	for i, res := range resources {
		machines[i] = parseMachine(res)
		if comps, ok := componentMap[machines[i].ID]; ok {
			machines[i].Components = comps
		}
	}

	return machines, nil
}

func (c *Client) SuspendLicense(id string) (*License, error) {
	data, err := c.doRequest("POST", "/licenses/"+id+"/actions/suspend", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing license: %w", err)
	}

	license := parseLicense(res)
	return &license, nil
}

func (c *Client) ReinstateLicense(id string) (*License, error) {
	data, err := c.doRequest("POST", "/licenses/"+id+"/actions/reinstate", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing license: %w", err)
	}

	license := parseLicense(res)
	return &license, nil
}
