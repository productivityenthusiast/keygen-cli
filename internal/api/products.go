package api

import (
	"encoding/json"
	"fmt"
)

type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListProducts() ([]Product, error) {
	data, err := c.doRequest("GET", "/products", nil)
	if err != nil {
		return nil, err
	}

	var doc JSONAPIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var resources []JSONAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, fmt.Errorf("parsing products: %w", err)
	}

	products := make([]Product, len(resources))
	for i, res := range resources {
		products[i] = Product{
			ID:   res.ID,
			Name: strVal(res.Attributes, "name"),
		}
	}

	return products, nil
}
