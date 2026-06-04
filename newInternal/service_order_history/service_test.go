package service_order_history_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	soHistory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
	soHistoryMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces/mocks"
)

var (
	soBaseTime = time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	soReceived = domain.ServiceOrderHistoryItem{
		NewStatus: soDomain.SERVICE_ORDER_STATUS_RECEIVED,
		CreatedAt: soBaseTime,
	}
	soInProgress = domain.ServiceOrderHistoryItem{
		PreviousStatus: soDomain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL,
		NewStatus:      soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		CreatedAt:      soBaseTime.Add(2 * time.Hour),
	}
	soCompleted = domain.ServiceOrderHistoryItem{
		PreviousStatus: soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		NewStatus:      soDomain.SERVICE_ORDER_STATUS_COMPLETED,
		CreatedAt:      soBaseTime.Add(4 * time.Hour),
	}
)

func TestService_GetHistoryByID(t *testing.T) {
	tests := []struct {
		name      string
		params    *domain.SearchParams
		setupRepo func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams)
		assert    func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error)
	}{
		{
			name:   "search error short-circuits",
			params: &domain.SearchParams{ID: "so-1"},
			setupRepo: func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams) {
				repo.EXPECT().Search(mock.Anything, params).Return(nil, errors.New("db search failure"))
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error) {
				assert.Error(t, err)
				assert.Nil(t, items)
			},
		},
		{
			name:   "empty items returns immediately",
			params: &domain.SearchParams{ID: "so-1"},
			setupRepo: func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams) {
				repo.EXPECT().Search(mock.Anything, params).Return([]domain.ServiceOrderHistoryItem{}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error) {
				assert.NoError(t, err)
				assert.Empty(t, items)
			},
		},
		{
			name:   "no IN_PROGRESS entry skips work transition lookup",
			params: &domain.SearchParams{ID: "so-1"},
			setupRepo: func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams) {
				repo.EXPECT().Search(mock.Anything, params).Return([]domain.ServiceOrderHistoryItem{soReceived}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error) {
				assert.NoError(t, err)
				assert.Len(t, items, 1)
				assert.Empty(t, items[0].WorkTransitions)
			},
		},
		{
			name:   "work transitions grouped by work_id and attached to IN_PROGRESS entry only",
			params: &domain.SearchParams{ID: "so-1"},
			setupRepo: func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams) {
				items := []domain.ServiceOrderHistoryItem{soReceived, soInProgress, soCompleted}
				groups := []domain.WorkTransitionGroup{
					{
						WorkID: "w-1",
						Status: []domain.WorkStatusItem{
							{NewStatus: soDomain.SERVICE_ORDER_STATUS_NEW, CreatedAt: soBaseTime.Add(time.Hour)},
							{PreviousStatus: soDomain.SERVICE_ORDER_STATUS_NEW, NewStatus: soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS, CreatedAt: soBaseTime.Add(2 * time.Hour)},
						},
					},
					{
						WorkID: "w-2",
						Status: []domain.WorkStatusItem{
							{NewStatus: soDomain.SERVICE_ORDER_STATUS_NEW, CreatedAt: soBaseTime.Add(time.Hour)},
						},
					},
				}
				repo.EXPECT().Search(mock.Anything, params).Return(items, nil)
				repo.EXPECT().SearchWorkTransitionsByServiceOrderID(mock.Anything, "so-1").Return(groups, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error) {
				assert.NoError(t, err)
				assert.Len(t, items, 3)
				assert.Empty(t, items[0].WorkTransitions)
				assert.Empty(t, items[2].WorkTransitions)

				groups := items[1].WorkTransitions
				assert.Len(t, groups, 2)
				assert.Equal(t, "w-1", groups[0].WorkID)
				assert.Len(t, groups[0].Status, 2)
				assert.Equal(t, soDomain.SERVICE_ORDER_STATUS_NEW, groups[0].Status[0].NewStatus)
				assert.Equal(t, soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS, groups[0].Status[1].NewStatus)
				assert.Equal(t, "w-2", groups[1].WorkID)
				assert.Len(t, groups[1].Status, 1)
			},
		},
		{
			name:   "SearchWorkTransitions error propagates",
			params: &domain.SearchParams{ID: "so-1"},
			setupRepo: func(repo *soHistoryMocks.ServiceOrderHistoryRepository, params *domain.SearchParams) {
				repo.EXPECT().Search(mock.Anything, params).Return([]domain.ServiceOrderHistoryItem{soInProgress}, nil)
				repo.EXPECT().SearchWorkTransitionsByServiceOrderID(mock.Anything, "so-1").
					Return(nil, errors.New("work tx lookup failed"))
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistoryItem, err error) {
				assert.Error(t, err)
				assert.Nil(t, items)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := soHistoryMocks.NewServiceOrderHistoryRepository(t)
			tt.setupRepo(repo, tt.params)

			svc := soHistory.NewService(repo)
			items, err := svc.GetHistoryByID(context.Background(), tt.params)
			tt.assert(t, items, err)
		})
	}
}
