package util

import (
	"math"
	"testing"
)

func TestBackgroundLuminance(t *testing.T) {
	tests := []struct {
		name  string
		bg    string
		want  float64
		wantK bool
	}{
		{"white", "rgba(255,255,255,0.98)", 255, true},
		{"black", "rgba(0,0,0,1.00)", 0, true},
		{"spaces", "rgba( 255, 255, 255, 1.00 )", 255, true},
		{"no alpha", "rgb(255,255,255)", 255, true},
		{"empty", "", 0, false},
		{"malformed", "rgba(255,255", 0, false},
		{"too few components", "rgba(255,255)", 0, false},
		{"non-numeric", "rgba(a,b,c,1.0)", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := backgroundLuminance(tt.bg)
			if ok != tt.wantK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantK)
			}
			if ok && math.Abs(got-tt.want) > 1e-6 {
				t.Errorf("luminance = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIconVariantSelection(t *testing.T) {
	tests := []struct {
		name string
		bg   string
		want string
	}{
		{"light theme", "rgba(255,255,255,0.98)", iconVariantLight},
		{"dark theme", "rgba(30,30,30,1.00)", iconVariantDark},
		{"unset falls back to dark", "", iconVariantDark},
		{"malformed falls back to dark", "rgba(oops)", iconVariantDark},
		{"just above midpoint", "rgba(128,128,128,1.00)", iconVariantLight},
		{"just below midpoint", "rgba(127,127,127,1.00)", iconVariantDark},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("alfred_theme_background", tt.bg)
			if got := iconVariant(); got != tt.want {
				t.Errorf("variant = %q, want %q", got, tt.want)
			}
		})
	}
}
