package runtimeops

import "testing"

func TestClampInt(t *testing.T) {
	if got := clampInt(1, 3, 5); got != 3 {
		t.Fatalf("expected lower bound, got %d", got)
	}
	if got := clampInt(9, 3, 5); got != 5 {
		t.Fatalf("expected upper bound, got %d", got)
	}
	if got := clampInt(4, 3, 5); got != 4 {
		t.Fatalf("expected unchanged value, got %d", got)
	}
}

func TestCalcProfileFallbackAndModeShape(t *testing.T) {
	normal := calcProfile("normal")
	unknown := calcProfile("unknown")
	if unknown.Name != "normal" {
		t.Fatalf("expected unknown mode to fall back to normal, got %q", unknown.Name)
	}
	if normal.DBMaxOpen <= 0 || normal.DockWorkers <= 0 || normal.GOMAXPROCS <= 0 {
		t.Fatalf("expected positive tuning values: %+v", normal)
	}

	eco := calcProfile("eco")
	turbo := calcProfile("turbo")
	if eco.Name != "eco" || turbo.Name != "turbo" {
		t.Fatalf("unexpected mode names: eco=%q turbo=%q", eco.Name, turbo.Name)
	}
	if turbo.DBMaxOpen < eco.DBMaxOpen {
		t.Fatalf("expected turbo to be at least as aggressive as eco: eco=%d turbo=%d", eco.DBMaxOpen, turbo.DBMaxOpen)
	}
}
