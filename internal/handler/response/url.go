package response

import "url_shortener/internal/model"

type Url struct {
	ShortCode string `json:"short_code"`
	LongUrl   string `json:"long_url"`
}

func NewUrl(u *model.Url) *Url {
	return &Url{
		ShortCode: u.ShortCode,
		LongUrl:   u.LongUrl,
	}
}

type UrlStats struct {
	Id          int    `json:"id"`
	ShortCode   string `json:"short_code"`
	LongUrl     string `json:"long_url"`
	CreatedAt   int    `json:"created_at"`
	UpdatedAt   int    `json:"updated_at"`
	ClicksCount int    `json:"clicks_count"`
}

func NewUrlStats(u *model.UrlWithClicksCount) *UrlStats {
	return &UrlStats{
		Id:          u.Id,
		ShortCode:   u.ShortCode,
		LongUrl:     u.LongUrl,
		CreatedAt:   int(u.CreatedAt.Unix()),
		UpdatedAt:   int(u.UpdatedAt.Unix()),
		ClicksCount: u.ClicksCount,
	}
}
