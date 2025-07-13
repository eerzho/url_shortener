package service_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
	"url_shortener/internal/constant"
	"url_shortener/internal/model"
	"url_shortener/internal/service"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	code := m.Run()
	os.Exit(code)
}

func TestUrl_Create(t *testing.T) {
	ctx := context.Background()
	longUrl := "https://example.com"
	ip := "127.0.0.1"
	userAgent := "Mozilla/5.0"

	tests := []struct {
		name        string
		longUrl     string
		ip          string
		userAgent   string
		setupMock   func(urlRepo *mockUrlRepository)
		wantUrl     *model.Url
		wantErr     bool
		wantErrType error
	}{
		{
			name:      "successful creation",
			longUrl:   longUrl,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(false, nil).Once()
				urlRepo.On("Create", ctx, longUrl, mock.AnythingOfType("string")).Return(&model.Url{
					Id:        1,
					ShortCode: "abc123",
					LongUrl:   longUrl,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil).Once()
			},
			wantUrl: &model.Url{
				Id:        1,
				ShortCode: "abc123",
				LongUrl:   longUrl,
			},
			wantErr: false,
		},
		{
			name:      "short code already exists",
			longUrl:   longUrl,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(true, nil).Once()
			},
			wantUrl:     nil,
			wantErr:     true,
			wantErrType: constant.ErrAlreadyExists,
		},
		{
			name:      "repository error on exists check",
			longUrl:   longUrl,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(false, errors.New("db error")).Once()
			},
			wantUrl: nil,
			wantErr: true,
		},
		{
			name:      "repository error on create",
			longUrl:   longUrl,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(false, nil).Once()
				urlRepo.On("Create", ctx, longUrl, mock.AnythingOfType("string")).Return((*model.Url)(nil), errors.New("create error")).Once()
			},
			wantUrl: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlRepo := new(mockUrlRepository)
			clickRepo := new(mockClickRepository)
			tt.setupMock(urlRepo)

			urlService := service.NewUrl(urlRepo, clickRepo)
			defer urlService.Close()

			url, err := urlService.Create(ctx, tt.longUrl, tt.ip, tt.userAgent)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, url)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, url)
				assert.Equal(t, tt.wantUrl.LongUrl, url.LongUrl)
				assert.Equal(t, tt.wantUrl.Id, url.Id)
				assert.NotEmpty(t, url.ShortCode)
			}

			urlRepo.AssertExpectations(t)
		})
	}
}

func TestUrl_Click(t *testing.T) {
	ctx := context.Background()
	shortCode := "abc123"
	ip := "127.0.0.1"
	userAgent := "Mozilla/5.0"

	existingUrl := &model.Url{
		Id:        1,
		ShortCode: shortCode,
		LongUrl:   "https://example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		shortCode string
		ip        string
		userAgent string
		setupMock func(urlRepo *mockUrlRepository, clickRepo *mockClickRepository)
		wantUrl   *model.Url
		wantErr   bool
	}{
		{
			name:      "successful click",
			shortCode: shortCode,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository, clickRepo *mockClickRepository) {
				urlRepo.On("GetByShortCode", ctx, shortCode).Return(existingUrl, nil).Once()
				clickRepo.On("Create", mock.AnythingOfType("*context.cancelCtx"), existingUrl.Id, ip, userAgent).Return(&model.Click{
					Id:        1,
					UrlId:     existingUrl.Id,
					Ip:        ip,
					UserAgent: userAgent,
					CreatedAt: time.Now(),
				}, nil).Maybe()
			},
			wantUrl: existingUrl,
			wantErr: false,
		},
		{
			name:      "url not found",
			shortCode: "nonexistent",
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository, clickRepo *mockClickRepository) {
				urlRepo.On("GetByShortCode", ctx, "nonexistent").Return((*model.Url)(nil), errors.New("not found")).Once()
			},
			wantUrl: nil,
			wantErr: true,
		},
		{
			name:      "repository error",
			shortCode: shortCode,
			ip:        ip,
			userAgent: userAgent,
			setupMock: func(urlRepo *mockUrlRepository, clickRepo *mockClickRepository) {
				urlRepo.On("GetByShortCode", ctx, shortCode).Return((*model.Url)(nil), errors.New("db error")).Once()
			},
			wantUrl: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlRepo := new(mockUrlRepository)
			clickRepo := new(mockClickRepository)
			tt.setupMock(urlRepo, clickRepo)

			urlService := service.NewUrl(urlRepo, clickRepo)
			defer urlService.Close()

			url, err := urlService.Click(ctx, tt.shortCode, tt.ip, tt.userAgent)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, url)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUrl, url)
			}

			urlRepo.AssertExpectations(t)
		})
	}
}

func TestUrl_GetStats(t *testing.T) {
	ctx := context.Background()
	shortCode := "abc123"

	urlWithStats := &model.UrlWithClicksCount{
		Url: model.Url{
			Id:        1,
			ShortCode: shortCode,
			LongUrl:   "https://example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ClicksCount: 42,
	}

	tests := []struct {
		name      string
		shortCode string
		setupMock func(urlRepo *mockUrlRepository)
		wantStats *model.UrlWithClicksCount
		wantErr   bool
	}{
		{
			name:      "successful stats retrieval",
			shortCode: shortCode,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("GetWithClicksCountByShortCode", ctx, shortCode).Return(urlWithStats, nil).Once()
			},
			wantStats: urlWithStats,
			wantErr:   false,
		},
		{
			name:      "url not found",
			shortCode: "nonexistent",
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("GetWithClicksCountByShortCode", ctx, "nonexistent").Return((*model.UrlWithClicksCount)(nil), errors.New("not found")).Once()
			},
			wantStats: nil,
			wantErr:   true,
		},
		{
			name:      "repository error",
			shortCode: shortCode,
			setupMock: func(urlRepo *mockUrlRepository) {
				urlRepo.On("GetWithClicksCountByShortCode", ctx, shortCode).Return((*model.UrlWithClicksCount)(nil), errors.New("db error")).Once()
			},
			wantStats: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlRepo := new(mockUrlRepository)
			clickRepo := new(mockClickRepository)
			tt.setupMock(urlRepo)

			urlService := service.NewUrl(urlRepo, clickRepo)
			defer urlService.Close()

			stats, err := urlService.GetStats(ctx, tt.shortCode)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, stats)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStats, stats)
			}

			urlRepo.AssertExpectations(t)
		})
	}
}

func BenchmarkUrl_Create(b *testing.B) {
	urlRepo := new(mockUrlRepository)
	clickRepo := new(mockClickRepository)
	ctx := context.Background()
	longUrl := "https://example.com"
	ip := "127.0.0.1"
	userAgent := "Mozilla/5.0"

	// Используем Maybe() для неопределенного количества вызовов в бенчмарке
	urlRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(false, nil).Maybe()
	urlRepo.On("Create", ctx, longUrl, mock.AnythingOfType("string")).Return(&model.Url{
		Id:        1,
		ShortCode: "abc123",
		LongUrl:   longUrl,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil).Maybe()

	urlService := service.NewUrl(urlRepo, clickRepo)
	defer urlService.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := urlService.Create(ctx, longUrl, ip, userAgent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUrl_Click(b *testing.B) {
	urlRepo := new(mockUrlRepository)
	clickRepo := new(mockClickRepository)
	ctx := context.Background()
	shortCode := "abc123"
	ip := "127.0.0.1"
	userAgent := "Mozilla/5.0"

	url := &model.Url{
		Id:        1,
		ShortCode: shortCode,
		LongUrl:   "https://example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Используем Maybe() для неопределенного количества вызовов
	urlRepo.On("GetByShortCode", ctx, shortCode).Return(url, nil).Maybe()
	clickRepo.On("Create", mock.AnythingOfType("*context.cancelCtx"), url.Id, ip, userAgent).Return(&model.Click{
		Id:        1,
		UrlId:     url.Id,
		Ip:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}, nil).Maybe()

	urlService := service.NewUrl(urlRepo, clickRepo)
	defer urlService.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := urlService.Click(ctx, shortCode, ip, userAgent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUrl_GetStats(b *testing.B) {
	urlRepo := new(mockUrlRepository)
	clickRepo := new(mockClickRepository)
	ctx := context.Background()
	shortCode := "abc123"

	stats := &model.UrlWithClicksCount{
		Url: model.Url{
			Id:        1,
			ShortCode: shortCode,
			LongUrl:   "https://example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ClicksCount: 42,
	}

	// Используем Maybe() для неопределенного количества вызовов
	urlRepo.On("GetWithClicksCountByShortCode", ctx, shortCode).Return(stats, nil).Maybe()

	urlService := service.NewUrl(urlRepo, clickRepo)
	defer urlService.Close()

	b.ResetTimer()
	for b.Loop() {
		_, err := urlService.GetStats(ctx, shortCode)
		if err != nil {
			b.Fatal(err)
		}
	}
}
