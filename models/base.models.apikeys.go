package models

type APIKeyInfo struct {
	APIKey     string         `json:"api_key"`
	APIKeyHash string         `json:"api_key_hash"`
	APIKeyHint string         `json:"api_key_hint"`
	UserID     string         `json:"user_id"`
	OrgID      string         `json:"org_id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Roles      []string       `json:"roles"`
	Rights     []string       `json:"rights"`
	Metadata   map[string]any `json:"metadata"`
}
