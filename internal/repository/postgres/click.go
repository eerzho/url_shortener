package postgres

import (
	"context"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
)

type Click struct {
	db *sqlx.DB
}

func NewClick(db *sqlx.DB) *Click {
	return &Click{db: db}
}

func (c *Click) Create(ctx context.Context, urlId int, ip string, userAgent string) (*model.Click, error) {
	var click model.Click
	err := c.db.GetContext(ctx, &click,
		`
			insert into clicks (url_id, ip, user_agent)
			values ($1, $2, $3)
			returning *
		`,
		urlId, ip, userAgent,
	)
	if err != nil {
		return nil, err
	}
	return &click, nil
}
