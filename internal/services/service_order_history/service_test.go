package service_order_history

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	uow "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetHistoryByID(t *testing.T) {
	expItem := domain.ServiceOrderHistory{
		ID:             "h-1",
		PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
		NewStatus:      domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS,
		CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
	}

	tests := []struct {
		name         string
		serviceID    string
		countResult  int64
		countErr     error
		searchResult []domain.ServiceOrderHistory
		searchErr    error
		expectErr    error
	}{
		{
			name:         "success",
			serviceID:    "so-1",
			countResult:  1,
			searchResult: []domain.ServiceOrderHistory{expItem},
		},
		{
			name:      "count error",
			serviceID: "so-2",
			countErr:  errors.New("db count failure"),
			expectErr: errors.New("db count failure"),
		},
		{
			name:        "search error",
			serviceID:   "so-3",
			countResult: 0,
			searchErr:   errors.New("db search failure"),
			expectErr:   errors.New("db search failure"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executor := &uow.UnitOfWork{}
			repo := domainmocks.NewServiceOrderHistoryRepository(t)
			searchParams := &domain.SearchServiceOrderHistoryParams{ID: tt.serviceID, Page: 1, PageSize: 10}

			repo.EXPECT().Count(mock.Anything, searchParams).Return(tt.countResult, tt.countErr)
			repo.EXPECT().Search(mock.Anything, searchParams).Return(tt.searchResult, tt.searchErr)

			s := Service(executor, repo)
			resp, err := s.GetHistoryByID(context.Background(), &domain.ListServiceOrderHistoryParams{ID: tt.serviceID, Page: 1, PageSize: 10})

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.countResult, resp.TotalItems)
				assert.Equal(t, tt.searchResult, resp.Items)
			}

			repo.AssertExpectations(t)
		})
	}
}
