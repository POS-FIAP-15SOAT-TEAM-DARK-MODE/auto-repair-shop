package domain_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestWorkServiceOrderHistoryList_GroupByWorkID(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		var list domain.WorkServiceOrderHistoryList
		assert.Nil(t, list.ToWorkTransitionGroup())
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		list := domain.WorkServiceOrderHistoryList{}
		assert.Nil(t, list.ToWorkTransitionGroup())
	})

	t.Run("preserves insertion order and groups correctly", func(t *testing.T) {
		list := domain.WorkServiceOrderHistoryList{
			{WorkID: "w-a"},
			{WorkID: "w-b"},
			{WorkID: "w-a"},
		}
		groups := list.ToWorkTransitionGroup()
		assert.Len(t, groups, 2)
		assert.Equal(t, "w-a", groups[0].WorkID)
		assert.Len(t, groups[0].Status, 2)
		assert.Equal(t, "w-b", groups[1].WorkID)
		assert.Len(t, groups[1].Status, 1)
	})
}
