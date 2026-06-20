package db

import (
	"testing"

	"go-api/internal/model"
)

func TestSanitizeOrderForRoleForUser(t *testing.T) {
	order := model.Order{
		UID:        1001,
		HID:        2002,
		DockStatus: "1",
		YID:        "UP-123",
		SupplierPT: "pup",
	}

	sanitizeOrderForRole(&order, false)

	if order.UID != 0 {
		t.Fatalf("expected uid to be hidden, got %d", order.UID)
	}
	if order.HID != 0 {
		t.Fatalf("expected hid to be hidden, got %d", order.HID)
	}
	if order.DockStatus != "" {
		t.Fatalf("expected dockstatus to be hidden, got %q", order.DockStatus)
	}
	if order.YID != "" {
		t.Fatalf("expected yid to be hidden, got %q", order.YID)
	}
	if order.SupplierPT != "" {
		t.Fatalf("expected supplier_pt to be hidden, got %q", order.SupplierPT)
	}
	if !order.CanPupLogin {
		t.Fatal("expected can_pup_login to stay available for pup orders")
	}
}

func TestSanitizeOrderForRoleForAdmin(t *testing.T) {
	order := model.Order{
		UID:        1001,
		HID:        2002,
		DockStatus: "1",
		YID:        "UP-123",
		SupplierPT: "pup",
	}

	sanitizeOrderForRole(&order, true)

	if order.UID != 1001 {
		t.Fatalf("expected uid to be kept, got %d", order.UID)
	}
	if order.HID != 2002 {
		t.Fatalf("expected hid to be kept, got %d", order.HID)
	}
	if order.DockStatus != "1" {
		t.Fatalf("expected dockstatus to be kept, got %q", order.DockStatus)
	}
	if order.YID != "UP-123" {
		t.Fatalf("expected yid to be kept, got %q", order.YID)
	}
	if order.SupplierPT != "pup" {
		t.Fatalf("expected supplier_pt to be kept, got %q", order.SupplierPT)
	}
	if !order.CanPupLogin {
		t.Fatal("expected can_pup_login to be available for admin pup orders")
	}
}
