package response

import (
	"time"
	"url_shortener/internal/model"
)

type Url struct {
	Id        int    `json:"id"`
	ShortCode string `json:"short_code"`
	LongUrl   string `json:"long_url"`
	Clicks    int    `json:"clicks"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func NewUrl(u *model.Url) *Url {
	return &Url{
		Id:        u.Id,
		ShortCode: u.ShortCode,
		LongUrl:   u.LongUrl,
		Clicks:    u.Clicks,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}

// UrlStats represents detailed URL statistics
type UrlStats struct {
	Id               int     `json:"id"`
	ShortCode        string  `json:"short_code"`
	LongUrl          string  `json:"long_url"`
	Clicks           int     `json:"clicks"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
	ClicksToday      int     `json:"clicks_today"`
	ClicksThisWeek   int     `json:"clicks_this_week"`
	ClicksThisMonth  int     `json:"clicks_this_month"`
	AverageClicksDay float64 `json:"average_clicks_per_day"`
	DaysActive       int     `json:"days_active"`
}

func NewUrlStats(u *model.Url) *UrlStats {
	now := time.Now()
	createdAt := u.CreatedAt

	// Calculate days active
	daysActive := int(now.Sub(createdAt).Hours() / 24)
	if daysActive == 0 {
		daysActive = 1 // At least 1 day to avoid division by zero
	}

	// Calculate average clicks per day
	avgClicksDay := float64(u.Clicks) / float64(daysActive)

	return &UrlStats{
		Id:               u.Id,
		ShortCode:        u.ShortCode,
		LongUrl:          u.LongUrl,
		Clicks:           u.Clicks,
		CreatedAt:        u.CreatedAt.Unix(),
		UpdatedAt:        u.UpdatedAt.Unix(),
		ClicksToday:      0, // Would need additional data to calculate
		ClicksThisWeek:   0, // Would need additional data to calculate
		ClicksThisMonth:  0, // Would need additional data to calculate
		AverageClicksDay: avgClicksDay,
		DaysActive:       daysActive,
	}
}

// UrlList represents a list of URLs with pagination
type UrlList struct {
	Data       []Url `json:"data"`
	Total      int   `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

func NewUrlList(urls []*model.Url, total, page, perPage int) *UrlList {
	data := make([]Url, 0, len(urls))
	for _, u := range urls {
		data = append(data, *NewUrl(u))
	}

	totalPages := (total + perPage - 1) / perPage

	return &UrlList{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}
