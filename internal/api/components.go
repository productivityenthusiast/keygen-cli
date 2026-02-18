package api

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListComponents(machineID string, page, limit int) ([]Component, error) {
	path := fmt.Sprintf("/machines/%s/components?page[size]=%d&page[number]=%d", machineID, limit, page)

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
		return nil, fmt.Errorf("parsing components: %w", err)
	}

	components := make([]Component, len(resources))
	for i, res := range resources {
		components[i] = parseComponent(res)
		components[i].MachineID = machineID
	}

	return components, nil
}

func (c *Client) GetComponent(id string) (*Component, error) {
	data, err := c.doRequest("GET", "/components/"+id, nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing component: %w", err)
	}

	comp := parseComponent(res)
	return &comp, nil
}

func (c *Client) DeleteComponent(id string) error {
	_, err := c.doRequest("DELETE", "/components/"+id, nil)
	return err
}

// FindComponentByFingerprint paginates through all machines and components to find a match
func (c *Client) FindComponentByFingerprint(fingerprint string) (*Component, error) {
	page := 1
	for {
		path := fmt.Sprintf("/machines?page[size]=100&page[number]=%d", page)
		data, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var doc JSONAPIDocument
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		var machines []JSONAPIResource
		if err := json.Unmarshal(doc.Data, &machines); err != nil {
			return nil, fmt.Errorf("parsing machines: %w", err)
		}

		if len(machines) == 0 {
			break
		}

		for _, machRes := range machines {
			machine := parseMachine(machRes)
			comps, err := c.ListComponents(machine.ID, 1, 100)
			if err != nil {
				continue
			}
			for _, comp := range comps {
				if comp.Fingerprint == fingerprint {
					return &comp, nil
				}
			}
		}

		page++
	}

	return nil, nil
}
