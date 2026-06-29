package class

import "testing"

func TestApplyMiJiaCapsAtOriginalPrice(t *testing.T) {
	finalPrice, originalPrice, applied := ApplyMiJia(10, 2, "*", MiJiaModeDirectPrice, 30, 4)
	if !applied {
		t.Fatal("expected mijia rule to apply")
	}
	if originalPrice != 20 {
		t.Fatalf("original price = %v, want 20", originalPrice)
	}
	if finalPrice != 20 {
		t.Fatalf("final price = %v, want capped original price 20", finalPrice)
	}
}

func TestNormalizeMiJiaScopeRecord(t *testing.T) {
	tests := []struct {
		name      string
		scopeType string
		scopeID   int
		cid       int
		wantType  string
		wantID    int
		wantCID   int
	}{
		{name: "legacy product", scopeType: "", scopeID: 0, cid: 12, wantType: "product", wantID: 12, wantCID: 12},
		{name: "explicit product", scopeType: "product", scopeID: 21, cid: 12, wantType: "product", wantID: 21, wantCID: 21},
		{name: "category", scopeType: "category", scopeID: 7, cid: 12, wantType: "category", wantID: 7, wantCID: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotID, gotCID := normalizeMiJiaScopeRecord(tt.scopeType, tt.scopeID, tt.cid)
			if gotType != tt.wantType || gotID != tt.wantID || gotCID != tt.wantCID {
				t.Fatalf("normalizeMiJiaScopeRecord() = (%q, %d, %d), want (%q, %d, %d)", gotType, gotID, gotCID, tt.wantType, tt.wantID, tt.wantCID)
			}
		})
	}
}

func TestValidateMiJiaPrice(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		price   string
		wantErr bool
	}{
		{name: "direct price", mode: "2", price: "1.25"},
		{name: "trim spaces", mode: "2", price: " 1.25 "},
		{name: "zero direct price", mode: "2", price: "0"},
		{name: "valid multiplier", mode: "3", price: "0.8"},
		{name: "empty", mode: "2", price: "", wantErr: true},
		{name: "not numeric", mode: "2", price: "abc", wantErr: true},
		{name: "negative", mode: "2", price: "-1", wantErr: true},
		{name: "zero multiplier", mode: "3", price: "0", wantErr: true},
		{name: "multiplier over one", mode: "3", price: "1.2", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateMiJiaPrice(tt.mode, tt.price)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestResolveClassPricesWithRules(t *testing.T) {
	results := resolveClassPricesWithRules([]PricingInput{
		{CID: 10, BasePrice: 2, Yunsuan: "*"},
		{CID: 11, BasePrice: 2, Yunsuan: "+"},
		{CID: 12, BasePrice: 10, Yunsuan: "*"},
	}, 3, 4, map[int]MiJiaRule{12: {Mode: MiJiaModeDirectPrice, Price: 20}})
	if got := results[10].Price; got != 6 {
		t.Fatalf("multiplication price = %v, want 6", got)
	}
	if got := results[11].Price; got != 5 {
		t.Fatalf("addition price = %v, want 5", got)
	}
	if got := results[12].Price; got != 20 {
		t.Fatalf("mijia price = %v, want 20", got)
	}
	if !results[12].MiJiaApplied {
		t.Fatal("expected mijia rule to apply")
	}
}
