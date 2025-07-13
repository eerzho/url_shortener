package postgres

import (
	"context"
	"fmt"
	"url_shortener/internal/model"

	"github.com/jmoiron/sqlx"
)

type Click struct {
	db *sqlx.DB
}

func NewClick(db *sqlx.DB) *Click {
	return &Click{db: db}
}

func (c *Click) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error) {
	const op = "repository.postgres.Click.GetList"
	var total int
	err := c.db.GetContext(ctx, &total,
		`
			select count(*) from clicks c
			join urls u on c.url_id = u.id
			where u.short_code = $1
		`,
		shortCode,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	if total == 0 {
		return []model.Click{}, 0, nil
	}

	offset := (page - 1) * size
	clicks := []model.Click{}
	err = c.db.SelectContext(ctx, &clicks,
		`
			select c.id, c.url_id, c.ip, c.user_agent, c.created_at
			from clicks c
			join urls u on c.url_id = u.id
			where u.short_code = $1
			order by c.created_at desc
			limit $2 offset $3
		`,
		shortCode, size, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	return clicks, total, nil
}

func (c *Click) Create(ctx context.Context, urlID int, ip string, userAgent string) (*model.Click, error) {
	const op = "repository.postgres.Click.Create"
	var click model.Click
	err := c.db.GetContext(ctx, &click,
		`
			insert into clicks (url_id, ip, user_agent)
			values ($1, $2, $3)
			returning *
		`,
		urlID, ip, userAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &click, nil
}
