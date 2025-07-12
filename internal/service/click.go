package service

import (
	"context"
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
	page = max(page, 1)
	size = min(max(size, 5), 200)
	list, total, err := c.clickRepository.GetList(ctx, shortCode, page, size)
	if err != nil {
		return nil, nil, err
	}
	return list, &dto.Pagination{Page: page, Size: size, Total: total}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
