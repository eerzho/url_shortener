package service

import (
	"context"
	"fmt"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
)

type Click struct {
	minPage     int
	minSize     int
	maxSize     int
	clickReader ClickReader
}

func NewClick(
	minPage int,
	minSize int,
	maxSize int,
	clickReader ClickReader,
) *Click {
	return &Click{
		minPage:     minPage,
		minSize:     minSize,
		maxSize:     maxSize,
		clickReader: clickReader,
	}
}

func (c *Click) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, *dto.Pagination, error) {
	const op = "service.Click.GetList"
	page = max(page, c.minPage)
	size = min(max(size, c.minSize), c.maxSize)
	list, total, err := c.clickReader.GetList(ctx, shortCode, page, size)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	return list, &dto.Pagination{Page: page, Size: size, Total: total}, nil
}
