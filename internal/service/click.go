package service

import (
	"context"
	"fmt"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type Click struct {
	minPage         int
	minSize         int
	maxSize         int
	clickRepository ClickRepository
}

func NewClick(
	minPage int,
	minSize int,
	maxSize int,
	clickReader ClickRepository,
) *Click {
	return &Click{
		minPage:         minPage,
		minSize:         minSize,
		maxSize:         maxSize,
		clickRepository: clickReader,
	}
}

func (c *Click) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error) {
	const op = "service.Click.GetList"
	page = max(page, c.minPage)
	size = min(max(size, c.minSize), c.maxSize)
	list, total, err := c.clickRepository.GetList(ctx, shortCode, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	return list, &dto.Pagination{Page: page, Size: size, Total: total}, nil
}
