package service_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"url_shortener/internal/dto"
	"url_shortener/internal/model"
	"url_shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClick_GetList(t *testing.T) {
	ctx := context.Background()
	shortCode := "abc123"

	count := 10
	clicks := createClicks(count)

	tests := []struct {
		name           string
		page           int
		size           int
		setupMock      func(mockRepo *mockClickRepository)
		wantClicks     []model.Click
		wantPagination *dto.Pagination
		wantErr        bool
	}{
		{
			name: "successful request with valid parameters",
			page: 1,
			size: 10,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 10).Return(clicks, count, nil).Once()
			},
			wantClicks:     clicks,
			wantPagination: &dto.Pagination{Page: 1, Size: 10, Total: count},
			wantErr:        false,
		},
		{
			name: "empty result",
			page: 1,
			size: 10,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 10).Return([]model.Click{}, 0, nil)
			},
			wantClicks:     []model.Click{},
			wantPagination: &dto.Pagination{Page: 1, Size: 10, Total: 0},
			wantErr:        false,
		},
		{
			name: "repository error",
			page: 1,
			size: 10,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 10).Return([]model.Click{}, 0, errors.New("test error")).Once()
			},
			wantClicks:     nil,
			wantPagination: nil,
			wantErr:        true,
		},
		{
			name: "request with invalid parameters - small page",
			page: 0,
			size: 10,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 10).Return(clicks, count, nil).Once()
			},
			wantClicks:     clicks,
			wantPagination: &dto.Pagination{Page: 1, Size: 10, Total: count},
			wantErr:        false,
		},
		{
			name: "request with invalid parameters - small size",
			page: 1,
			size: 0,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 5).Return(clicks[:5], count, nil).Once()
			},
			wantClicks:     clicks[:5],
			wantPagination: &dto.Pagination{Page: 1, Size: 5, Total: count},
			wantErr:        false,
		},
		{
			name: "request with invalid parameters - large size",
			page: 1,
			size: 201,
			setupMock: func(mockRepo *mockClickRepository) {
				mockRepo.On("GetList", ctx, shortCode, 1, 200).Return(clicks, count, nil).Once()
			},
			wantClicks:     clicks,
			wantPagination: &dto.Pagination{Page: 1, Size: 200, Total: count},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockClickRepository)
			tt.setupMock(mockRepo)
			clickService := service.NewClick(mockRepo)

			clicks, pagination, err := clickService.GetList(ctx, shortCode, tt.page, tt.size)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, clicks)
				assert.Nil(t, pagination)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantClicks, clicks)
				assert.Equal(t, tt.wantPagination, pagination)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func createClicks(count int) []model.Click {
	clicks := make([]model.Click, count)
	for i := range count {
		clicks[i] = model.Click{
			Id:        i + 1,
			UrlId:     100,
			Ip:        "127.0.0.1",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			CreatedAt: time.Now(),
		}
	}
	return clicks
}
