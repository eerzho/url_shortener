package service_test

import (
	"context"
	"url_shortener/internal/model"

	"github.com/stretchr/testify/mock"
)

type mockClickRepository struct {
	mock.Mock
}

func (m *mockClickRepository) GetList(ctx context.Context, shortCode string, page, size int) ([]model.Click, int, error) {
	args := m.Called(ctx, shortCode, page, size)
	return args.Get(0).([]model.Click), args.Int(1), args.Error(2)
}

func (m *mockClickRepository) Create(ctx context.Context, urlId int, ip string, userAgent string) (*model.Click, error) {
	args := m.Called(ctx, urlId, ip, userAgent)
	return args.Get(0).(*model.Click), args.Error(1)
}

type mockUrlRepository struct {
	mock.Mock
}

func (m *mockUrlRepository) Create(ctx context.Context, longUrl, shortCode string) (*model.Url, error) {
	args := m.Called(ctx, longUrl, shortCode)
	return args.Get(0).(*model.Url), args.Error(1)
}

func (m *mockUrlRepository) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

func (m *mockUrlRepository) GetByShortCode(ctx context.Context, shortCode string) (*model.Url, error) {
	args := m.Called(ctx, shortCode)
	return args.Get(0).(*model.Url), args.Error(1)
}

func (m *mockUrlRepository) GetWithClicksCountByShortCode(ctx context.Context, shortCode string) (*model.UrlWithClicksCount, error) {
	args := m.Called(ctx, shortCode)
	return args.Get(0).(*model.UrlWithClicksCount), args.Error(1)
}
