package model

import "time"

type Click struct {
	Id        int       `db:"id" json:"id"`
	UrlId     int       `db:"url_id" json:"url_id"`
	Ip        string    `db:"ip" json:"ip"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
