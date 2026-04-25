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

	expWorks := []domain.WorkStatusTimeline{
		{
			WorkID: "wrk-1",
			History: []domain.WorkStatusHistoryEntry{
				{
					PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
					NewStatus:      domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
					CreatedAt:      time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			WorkID:  "wrk-2",
			History: []domain.WorkStatusHistoryEntry{},
		},
	}

	tests := []struct {
		name         string
		serviceID    string
		countResult  int64
		countErr     error
		searchResult []domain.ServiceOrderHistory
		searchErr    error
		worksResult  []domain.WorkStatusTimeline
		worksErr     error
		expectErr    bool
		expectWorks  int
	}{
		{
			name:         "success with works",
			serviceID:    "so-1",
			countResult:  1,
			searchResult: []domain.ServiceOrderHistory{expItem},
			worksResult:  expWorks,
			expectWorks:  2,
		},
		{
			name:         "success with empty works",
			serviceID:    "so-empty",
			countResult:  1,
			searchResult: []domain.ServiceOrderHistory{expItem},
			worksResult:  []domain.WorkStatusTimeline{},
			expectWorks:  0,
		},
		{
			name:      "count error",
			serviceID: "so-2",
			countErr:  errors.New("db count failure"),
			expectErr: true,
		},
		{
			name:      "search error",
			serviceID: "so-3",
			searchErr: errors.New("db search failure"),
			expectErr: true,
		},
		{
			name:      "works error",
			serviceID: "so-4",
			worksErr:  errors.New("db works failure"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executor := &uow.UnitOfWork{}
			repo := domainmocks.NewServiceOrderHistoryRepository(t)
			searchParams := &domain.SearchServiceOrderHistoryParams{ID: tt.serviceID, Page: 1, PageSize: 10}

			repo.EXPECT().Count(mock.Anything, searchParams).Return(tt.countResult, tt.countErr).Maybe()
			repo.EXPECT().Search(mock.Anything, searchParams).Return(tt.searchResult, tt.searchErr).Maybe()
			repo.EXPECT().WorkTimelineByServiceOrderID(mock.Anything, tt.serviceID).Return(tt.worksResult, tt.worksErr).Maybe()

			s := Service(executor, repo)
			resp, err := s.GetHistoryByID(context.Background(), &domain.SearchServiceOrderHistoryParams{ID: tt.serviceID, Page: 1, PageSize: 10})

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.countResult, resp.Page.TotalItems)
			assert.Equal(t, tt.searchResult, resp.Page.Items)
			assert.Len(t, resp.Works, tt.expectWorks)
			if tt.expectWorks > 0 {
				assert.Equal(t, tt.worksResult, resp.Works)
			}
		})
	}
}
