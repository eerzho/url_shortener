package response

import "url_shortener/internal/model"

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
