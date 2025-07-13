package model

import "time"

type URL struct {
	ID        int       `db:"id"         json:"id"`
	LongURL   string    `db:"long_url"   json:"long_url"`
	ShortCode string    `db:"short_code" json:"short_code"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type URLWithClicksCount struct {
	URL

	ClicksCount int `db:"clicks_count" json:"clicks_count"`
}
