package api

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetMachine(id string) (*Machine, error) {
	data, err := c.doRequest("GET", "/machines/"+id+"?include=components", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var res JSONAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, fmt.Errorf("parsing machine: %w", err)
	}

	machine := parseMachine(res)

	// Parse included components
	for _, inc := range doc.Included {
		if inc.Type == "components" {
			comp := parseComponent(inc)
			machine.Components = append(machine.Components, comp)
		}
	}

	return &machine, nil
}
