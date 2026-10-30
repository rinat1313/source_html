package apitypes

import (
	"encoding/json"
	"strings"
)

type Period struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type EIS44Params struct {
	OrgRegion      string   `json:"org_region,omitempty"`
	DocumentTypes  []string `json:"document_types,omitempty"`
	Period         Period   `json:"period,omitempty"`
	IsAllOrg44     bool     `json:"is_all_organizations_44,omitempty"`
	IsAllOrg223    bool     `json:"is_all_organizations_223,omitempty"`
}

type EIS223Params struct {
	EntityTypes []string `json:"entity_types,omitempty"`
	Period      Period   `json:"period,omitempty"`
	Region      string   `json:"region,omitempty"`
}

type SearchConfig struct {
	Law      string       `json:"law"`
	Keywords []string     `json:"keywords"`
	Limit    int          `json:"limit"`
	EIS44    EIS44Params  `json:"eis44,omitempty"`
	EIS223   EIS223Params `json:"eis223,omitempty"`
}

var secretFieldNames = []string{
	"password", "passwd", "login", "secret", "token", "certificate", "cert_path",
	"private_key", "api_key", "system_identifier",
}

func SanitizeConfigJSON(raw json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	stripSecretsMap(m)
	out, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func stripSecretsMap(m map[string]json.RawMessage) {
	for k := range m {
		if isSecretKey(k) {
			delete(m, k)
			continue
		}
		var nested map[string]json.RawMessage
		if json.Unmarshal(m[k], &nested) == nil {
			stripSecretsMap(nested)
			b, _ := json.Marshal(nested)
			m[k] = b
		}
	}
}

func isSecretKey(k string) bool {
	lk := strings.ToLower(k)
	for _, s := range secretFieldNames {
		if strings.Contains(lk, s) {
			return true
		}
	}
	return false
}

func (c SearchConfig) Validate() string {
	switch c.Law {
	case "44", "223", "both":
	default:
		return "law must be 44, 223, or both"
	}
	if c.Limit <= 0 {
		return "limit must be positive"
	}
	return ""
}

func NormalizeKeywords(words []string) []string {
	var out []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		out = append(out, strings.ToLower(w))
	}
	return out
}
