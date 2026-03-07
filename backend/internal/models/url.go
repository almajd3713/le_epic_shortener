package models

import "time"

type URL struct {
	ID        int64      `json:"id"`
	ShortCode string     `json:"short_code"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"` // nil = no expiration
	IsActive  bool       `json:"is_active"`
}

func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

type URLRequest struct {
	// validate:"required,url" ensures the field is present and a valid URL.
	LongURL   string     `json:"long_url"   validate:"required,url"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type URLUpdateRequest struct {
	Action string `json:"action" validate:"required,oneof=activate deactivate"`
}

type URLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	CreatedAt string `json:"created_at"`
}

// URLListItem is the response shape for GET /api/urls entries.
type URLListItem struct {
	ShortCode string  `json:"short_code"`
	LongURL   string  `json:"long_url"`
	ShortURL  string  `json:"short_url"`
	CreatedAt string  `json:"created_at"`
	ExpiresAt *string `json:"expires_at"`
}
