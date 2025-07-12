package service

import (
	"context"
	"fmt"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type Click struct {
	clickRepository ClickRepository
}

func NewClick(clickRepository ClickRepository) *Click {
	return &Click{
		clickRepository: clickRepository,
	}
}

func (c *Click) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error) {
	const op = "service.Click.GetList"
	page = max(page, 1)
	size = min(max(size, 5), 200)
	list, total, err := c.clickRepository.GetList(ctx, shortCode, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	return list, &dto.Pagination{Page: page, Size: size, Total: total}, nil
}
