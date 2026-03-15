package models

import "time"

// Link represents a short link.
type Link struct {
	ID           string     `json:"id"`
	ShortCode    string     `json:"short_code"`
	LongURL      string     `json:"long_url"`
	PasswordHash *string    `json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Click represents a click event.
type Click struct {
	ID        int64     `json:"id"`
	LinkID    string    `json:"link_id"`
	ClickedAt time.Time `json:"clicked_at"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	IsUnique  bool      `json:"is_unique"`
}

// CreateLinkInput captures API request for creating link.
type CreateLinkInput struct {
	LongURL     string     `json:"long_url"`
	CustomAlias string     `json:"custom_alias"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Password    string     `json:"password"`
}

// AnalyticsSummary is aggregated stats for a link.
type AnalyticsSummary struct {
	ShortCode    string         `json:"short_code"`
	TotalClicks  int64          `json:"total_clicks"`
	UniqueClicks int64          `json:"unique_clicks"`
	TopReferrers map[string]int `json:"top_referrers"`
	TopBrowsers  map[string]int `json:"top_browsers"`
}
