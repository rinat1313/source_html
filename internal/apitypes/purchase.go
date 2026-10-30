package apitypes

import "time"

type FileLink struct {
	URL  string `json:"url,omitempty"`
	Name string `json:"name,omitempty"`
}

type PurchaseCard struct {
	Law           string         `json:"law"`
	EntityKind    string         `json:"entity_kind"`
	RegNumber     string         `json:"reg_number,omitempty"`
	GUID          string         `json:"guid,omitempty"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	CustomerName  string         `json:"customer_name"`
	CustomerINN   string         `json:"customer_inn,omitempty"`
	CustomerExtra map[string]any `json:"customer_extra,omitempty"`
	FileLinks     []FileLink     `json:"file_links,omitempty"`
	EffectiveDate time.Time      `json:"effective_date"`
	RawMeta       map[string]any `json:"raw_meta,omitempty"`
}

type UpsertResult struct {
	PurchaseID string `json:"purchase_id"`
	Action     string `json:"action"`
	NewSlot    bool   `json:"new_slot"`
}
