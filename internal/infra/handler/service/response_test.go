package service

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestMapResponseDTOFromDomain(t *testing.T) {
	price := decimal.NewFromFloat(199.90)
	svc := &domain.Service{
		ID:          "svc-id",
		Name:        "Oil Change",
		Description: "Complete synthetic oil change",
		Price:       price,
		Status:      domain.ACTIVE,
	}

	resp := mapResponseDTOFromDomain(svc)

	if resp.ID != svc.ID {
		t.Errorf("expected ID %q, got %q", svc.ID, resp.ID)
	}
	if resp.Name != svc.Name {
		t.Errorf("expected Name %q, got %q", svc.Name, resp.Name)
	}
	if resp.Description != svc.Description {
		t.Errorf("expected Description %q, got %q", svc.Description, resp.Description)
	}

	priceFloat, _ := price.Float64()
	if resp.Price != priceFloat {
		t.Errorf("expected Price %f, got %f", priceFloat, resp.Price)
	}
	if resp.Status != svc.Status.String() {
		t.Errorf("expected Status %q, got %q", svc.Status.String(), resp.Status)
	}
}
