package class

import "testing"

func TestApplyMiJiaModes(t *testing.T) {
	tests := []struct {
		name      string
		basePrice float64
		addprice  float64
		yunsuan   string
		mode      int
		secret    float64
		want      float64
		applied   bool
	}{
		{name: "subtract from final", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeSubtractFromFinal, secret: 3, want: 17, applied: true},
		{name: "subtract before rate", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeSubtractBeforeRate, secret: 3, want: 14, applied: true},
		{name: "direct price", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeDirectPrice, secret: 12, want: 12, applied: true},
		{name: "multiplier", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeMultiplier, secret: 1.5, want: 15, applied: true},
		{name: "cap at original price", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeMultiplier, secret: 3, want: 20, applied: true},
		{name: "plus formula", basePrice: 10, addprice: 2, yunsuan: "+", mode: MiJiaModeSubtractFromFinal, secret: 1, want: 11, applied: true},
		{name: "mode one still uses multiplier logic", basePrice: 10, addprice: 2, yunsuan: "+", mode: MiJiaModeSubtractBeforeRate, secret: 1, want: 12, applied: true},
		{name: "unsupported mode", basePrice: 10, addprice: 2, yunsuan: "*", mode: 3, secret: 5, want: 20, applied: false},
		{name: "floor at zero", basePrice: 10, addprice: 2, yunsuan: "*", mode: MiJiaModeSubtractFromFinal, secret: 30, want: 0, applied: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, original, applied := ApplyMiJia(tt.basePrice, tt.addprice, tt.yunsuan, tt.mode, tt.secret, 4)
			if original != 20 && tt.yunsuan == "*" {
				t.Fatalf("unexpected original price: got %.4f", original)
			}
			if got != tt.want {
				t.Fatalf("unexpected price: got %.4f want %.4f", got, tt.want)
			}
			if applied != tt.applied {
				t.Fatalf("unexpected applied flag: got %v want %v", applied, tt.applied)
			}
		})
	}
}
