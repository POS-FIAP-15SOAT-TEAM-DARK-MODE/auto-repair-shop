package work

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestMapResponseDTOFromDomain(t *testing.T) {
	price := decimal.NewFromFloat(199.90)
	svc := &domain.Work{
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

	expectedPrice := price.String()
	if resp.Price != expectedPrice {
		t.Errorf("expected Price %q, got %q", expectedPrice, resp.Price)
	}
	if resp.Status != svc.Status.String() {
		t.Errorf("expected Status %q, got %q", svc.Status.String(), resp.Status)
	}
}

func TestMapListResponseDTOFromDomain(t *testing.T) {
	price := decimal.RequireFromString("249.99")
	work1 := domain.Work{
		ID:          "w1",
		Name:        "Brake Inspection",
		Description: "Front and rear brake inspection",
		Price:       price,
		Status:      domain.ACTIVE,
	}
	work2 := domain.Work{
		ID:          "w2",
		Name:        "Tire Rotation",
		Description: "All four tires rotation",
		Price:       decimal.RequireFromString("49.50"),
		Status:      domain.INACTIVE,
	}

	domainPaginator := &domain.PaginatorResponse[domain.Work]{
		Items:      []domain.Work{work1, work2},
		TotalItems: 2,
		TotalPages: 1,
		PageSize:   10,
		Page:       1,
	}

	dto := mapListResponseDTOFromDomain(domainPaginator)

	if got, want := dto.TotalItems, domainPaginator.TotalItems; got != want {
		t.Errorf("TotalItems: got %d, want %d", got, want)
	}
	if got, want := dto.TotalPages, domainPaginator.TotalPages; got != want {
		t.Errorf("TotalPages: got %d, want %d", got, want)
	}
	if got, want := dto.PageSize, domainPaginator.PageSize; got != want {
		t.Errorf("PageSize: got %d, want %d", got, want)
	}
	if got, want := dto.Page, domainPaginator.Page; got != want {
		t.Errorf("Page: got %d, want %d", got, want)
	}
	if len(dto.Items) != len(domainPaginator.Items) {
		t.Fatalf("expected %d items, got %d", len(domainPaginator.Items), len(dto.Items))
	}

	for i, item := range dto.Items {
		expected := domainPaginator.Items[i]
		if item.ID != expected.ID {
			t.Errorf("Items[%d].ID: got %q, want %q", i, item.ID, expected.ID)
		}
		if item.Name != expected.Name {
			t.Errorf("Items[%d].Name: got %q, want %q", i, item.Name, expected.Name)
		}
		if item.Description != expected.Description {
			t.Errorf("Items[%d].Description: got %q, want %q", i, item.Description, expected.Description)
		}
		expectedPrice := expected.Price.String()
		if item.Price != expectedPrice {
			t.Errorf("Items[%d].Price: got %q, want %q", i, item.Price, expectedPrice)
		}
		if item.Status != expected.Status.String() {
			t.Errorf("Items[%d].Status: got %q, want %q", i, item.Status, expected.Status.String())
		}
	}
}
