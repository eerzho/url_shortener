package service

import (
	"context"
	"time"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type Click struct {
}

func NewClick() *Click {
	return &Click{}
}

func (c *Click) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error) {
	clicks := make([]model.Click, 0)
	for i := range 10 {
		clicks = append(clicks, model.Click{Id: i, CreatedAt: time.Now()})
	}
	return clicks, &dto.Pagination{}, nil
}
