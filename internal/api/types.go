package api

import (
	"encoding/json"
	"time"
)

// JSON:API envelope types

type JSONAPIDocument struct {
	Data     json.RawMessage        `json:"data"`
	Included []JSONAPIResource      `json:"included,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	Links    map[string]interface{} `json:"links,omitempty"`
}

type JSONAPIResource struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Attributes    map[string]interface{}  `json:"attributes"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         map[string]interface{}  `json:"links,omitempty"`
}

type Relationship struct {
	Data  json.RawMessage        `json:"data,omitempty"`
	Links map[string]interface{} `json:"links,omitempty"`
}

type RelationshipData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Domain types

type License struct {
	ID        string                 `json:"id"`
	Key       string                 `json:"key"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Expiry    string                 `json:"expiry"`
	Created   string                 `json:"created"`
	Updated   string                 `json:"updated"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	PolicyID  string                 `json:"policy_id,omitempty"`
	ProductID string                 `json:"product_id,omitempty"`
	OwnerID   string                 `json:"owner_id,omitempty"`
}

type Machine struct {
	ID          string      `json:"id"`
	Fingerprint string      `json:"fingerprint"`
	Name        string      `json:"name"`
	Hostname    string      `json:"hostname"`
	Platform    string      `json:"platform"`
	IP          string      `json:"ip"`
	Cores       int         `json:"cores"`
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	LicenseID   string      `json:"license_id,omitempty"`
	Components  []Component `json:"components,omitempty"`
}

type Component struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
	Name        string `json:"name"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	MachineID   string `json:"machine_id,omitempty"`
}

type User struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Role      string                 `json:"role"`
	Status    string                 `json:"status"`
	Created   string                 `json:"created"`
	Updated   string                 `json:"updated"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Token struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Token      string `json:"token"`
	Expiry     string `json:"expiry"`
	Created    string `json:"created"`
	BearerID   string `json:"bearer_id,omitempty"`
	BearerType string `json:"bearer_type,omitempty"`
}

type LicenseValidation struct {
	Valid    bool                   `json:"valid"`
	Detail   string                 `json:"detail"`
	Code     string                 `json:"code"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Parse functions

func parseLicense(res JSONAPIResource) License {
	attr := res.Attributes
	l := License{
		ID:      res.ID,
		Key:     strVal(attr, "key"),
		Name:    strVal(attr, "name"),
		Status:  strVal(attr, "status"),
		Expiry:  strVal(attr, "expiry"),
		Created: strVal(attr, "created"),
		Updated: strVal(attr, "updated"),
	}

	if md, ok := attr["metadata"].(map[string]interface{}); ok {
		l.Metadata = md
	}

	// Extract relationship IDs
	if rel, ok := res.Relationships["policy"]; ok {
		l.PolicyID = extractRelID(rel)
	}
	if rel, ok := res.Relationships["product"]; ok {
		l.ProductID = extractRelID(rel)
	}
	if rel, ok := res.Relationships["owner"]; ok {
		l.OwnerID = extractRelID(rel)
	}

	return l
}

func parseMachine(res JSONAPIResource) Machine {
	attr := res.Attributes
	m := Machine{
		ID:          res.ID,
		Fingerprint: strVal(attr, "fingerprint"),
		Name:        strVal(attr, "name"),
		Hostname:    strVal(attr, "hostname"),
		Platform:    strVal(attr, "platform"),
		IP:          strVal(attr, "ip"),
		Created:     strVal(attr, "created"),
		Updated:     strVal(attr, "updated"),
	}

	if cores, ok := attr["cores"].(float64); ok {
		m.Cores = int(cores)
	}

	if rel, ok := res.Relationships["license"]; ok {
		m.LicenseID = extractRelID(rel)
	}

	return m
}

func parseComponent(res JSONAPIResource) Component {
	attr := res.Attributes
	c := Component{
		ID:          res.ID,
		Fingerprint: strVal(attr, "fingerprint"),
		Name:        strVal(attr, "name"),
		Created:     strVal(attr, "created"),
		Updated:     strVal(attr, "updated"),
	}

	if rel, ok := res.Relationships["machine"]; ok {
		c.MachineID = extractRelID(rel)
	}

	return c
}

func parseUser(res JSONAPIResource) User {
	attr := res.Attributes
	u := User{
		ID:        res.ID,
		Email:     strVal(attr, "email"),
		FirstName: strVal(attr, "firstName"),
		LastName:  strVal(attr, "lastName"),
		Role:      strVal(attr, "role"),
		Status:    strVal(attr, "status"),
		Created:   strVal(attr, "created"),
		Updated:   strVal(attr, "updated"),
	}

	if md, ok := attr["metadata"].(map[string]interface{}); ok {
		u.Metadata = md
	}

	return u
}

func parseToken(res JSONAPIResource) Token {
	attr := res.Attributes
	t := Token{
		ID:      res.ID,
		Kind:    strVal(attr, "kind"),
		Token:   strVal(attr, "token"),
		Expiry:  strVal(attr, "expiry"),
		Created: strVal(attr, "created"),
	}

	if rel, ok := res.Relationships["bearer"]; ok {
		t.BearerID = extractRelID(rel)
	}

	return t
}

// Helper functions

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func extractRelID(rel Relationship) string {
	var rd RelationshipData
	if err := json.Unmarshal(rel.Data, &rd); err == nil {
		return rd.ID
	}
	return ""
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
