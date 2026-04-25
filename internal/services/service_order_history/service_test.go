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

var (
	soBaseTime = time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	soReceived = domain.ServiceOrderHistory{
		ID:             "soh-1",
		ServiceOrderID: "so-1",
		PreviousStatus: "",
		NewStatus:      domain.SERVICE_ORDER_STATUS_RECEIVED,
		CreatedAt:      soBaseTime,
	}
	soInProgress = domain.ServiceOrderHistory{
		ID:             "soh-2",
		ServiceOrderID: "so-1",
		PreviousStatus: domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL,
		NewStatus:      domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		CreatedAt:      soBaseTime.Add(2 * time.Hour),
	}
	soCompleted = domain.ServiceOrderHistory{
		ID:             "soh-3",
		ServiceOrderID: "so-1",
		PreviousStatus: domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		NewStatus:      domain.SERVICE_ORDER_STATUS_COMPLETED,
		CreatedAt:      soBaseTime.Add(4 * time.Hour),
	}
)

func TestService_GetHistoryByID(t *testing.T) {
	tests := []struct {
		name      string
		params    *domain.SearchServiceOrderHistoryParams
		setupRepo func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams)
		assert    func(t *testing.T, items []domain.ServiceOrderHistory, err error)
	}{
		{
			name:   "search error short-circuits",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				repo.EXPECT().Search(mock.Anything, params).Return(nil, errors.New("db search failure"))
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.Error(t, err)
				assert.Nil(t, items)
			},
		},
		{
			name:   "empty items returns immediately",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				repo.EXPECT().Search(mock.Anything, params).Return([]domain.ServiceOrderHistory{}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.NoError(t, err)
				assert.Empty(t, items)
			},
		},
		{
			name:   "no IN_PROGRESS entry skips work transition lookup",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				repo.EXPECT().Search(mock.Anything, params).Return([]domain.ServiceOrderHistory{soReceived}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.NoError(t, err)
				assert.Len(t, items, 1)
				assert.Empty(t, items[0].WorkTransitions)
			},
		},
		{
			name:   "work transitions grouped by work_id and attached to IN_PROGRESS entry only",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				items := []domain.ServiceOrderHistory{soReceived, soInProgress, soCompleted}
				w1t1 := domain.WorkServiceOrderHistory{
					ID: "wsh-1", WorkID: "w-1", ServiceOrderID: "so-1",
					PreviousStatus: "", NewStatus: domain.SERVICE_ORDER_STATUS_NEW,
					CreatedAt: soBaseTime.Add(1 * time.Hour),
				}
				w1t2 := domain.WorkServiceOrderHistory{
					ID: "wsh-2", WorkID: "w-1", ServiceOrderID: "so-1",
					PreviousStatus: domain.SERVICE_ORDER_STATUS_NEW, NewStatus: domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
					CreatedAt: soBaseTime.Add(2 * time.Hour),
				}
				w2t1 := domain.WorkServiceOrderHistory{
					ID: "wsh-3", WorkID: "w-2", ServiceOrderID: "so-1",
					PreviousStatus: "", NewStatus: domain.SERVICE_ORDER_STATUS_NEW,
					CreatedAt: soBaseTime.Add(1 * time.Hour),
				}
				repo.EXPECT().Search(mock.Anything, params).Return(items, nil)
				repo.EXPECT().SearchWorkTransitions(mock.Anything, "so-1").
					Return([]domain.WorkServiceOrderHistory{w1t1, w1t2, w2t1}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.NoError(t, err)
				assert.Len(t, items, 3)

				assert.Empty(t, items[0].WorkTransitions)
				assert.Empty(t, items[2].WorkTransitions)

				groups := items[1].WorkTransitions
				assert.Len(t, groups, 2)

				assert.Equal(t, "w-1", groups[0].WorkID)
				assert.Len(t, groups[0].Status, 2)
				assert.Equal(t, "wsh-1", groups[0].Status[0].ID)
				assert.Equal(t, "wsh-2", groups[0].Status[1].ID)

				assert.Equal(t, "w-2", groups[1].WorkID)
				assert.Len(t, groups[1].Status, 1)
				assert.Equal(t, "wsh-3", groups[1].Status[0].ID)
			},
		},
		{
			name:   "no work transitions yields nil WorkTransitions on IN_PROGRESS entry",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				items := []domain.ServiceOrderHistory{soReceived, soInProgress}
				repo.EXPECT().Search(mock.Anything, params).Return(items, nil)
				repo.EXPECT().SearchWorkTransitions(mock.Anything, "so-1").
					Return([]domain.WorkServiceOrderHistory{}, nil)
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.NoError(t, err)
				assert.Len(t, items, 2)
				assert.Empty(t, items[0].WorkTransitions)
				assert.Empty(t, items[1].WorkTransitions)
			},
		},
		{
			name:   "SearchWorkTransitions error propagates",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-1"},
			setupRepo: func(repo *domainmocks.ServiceOrderHistoryRepository, params *domain.SearchServiceOrderHistoryParams) {
				items := []domain.ServiceOrderHistory{soInProgress}
				repo.EXPECT().Search(mock.Anything, params).Return(items, nil)
				repo.EXPECT().SearchWorkTransitions(mock.Anything, "so-1").
					Return(nil, errors.New("work tx lookup failed"))
			},
			assert: func(t *testing.T, items []domain.ServiceOrderHistory, err error) {
				assert.Error(t, err)
				assert.Nil(t, items)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executor := &uow.UnitOfWork{}
			repo := domainmocks.NewServiceOrderHistoryRepository(t)
			tt.setupRepo(repo, tt.params)

			s := Service(executor, repo)
			items, err := s.GetHistoryByID(context.Background(), tt.params)
			tt.assert(t, items, err)
		})
	}
}

func TestGroupByWorkID(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := groupByWorkID(nil)
		assert.Nil(t, result)
	})

	t.Run("preserves insertion order and groups correctly", func(t *testing.T) {
		transitions := []domain.WorkServiceOrderHistory{
			{ID: "1", WorkID: "w-a"},
			{ID: "2", WorkID: "w-b"},
			{ID: "3", WorkID: "w-a"},
		}
		groups := groupByWorkID(transitions)
		assert.Len(t, groups, 2)
		assert.Equal(t, "w-a", groups[0].WorkID)
		assert.Len(t, groups[0].Status, 2)
		assert.Equal(t, "w-b", groups[1].WorkID)
		assert.Len(t, groups[1].Status, 1)
	})
}
