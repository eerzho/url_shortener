package model

import "time"

type Url struct {
	Id        int       `db:"id" json:"id"`
	ShortCode string    `db:"short_code" json:"short_code"`
	LongUrl   string    `db:"long_url" json:"long_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UrlWithClicksCount struct {
	Url
	ClicksCount int `db:"clicks_count" json:"clicks_count"`
}
