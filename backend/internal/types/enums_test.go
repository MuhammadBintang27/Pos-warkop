package types

import "testing"

func TestValidPrimary(t *testing.T) {
	for s, want := range map[string]bool{
		"amber": true, "sky": true, "#8B4513": true, "#8b4513": true,
		"": false, "#fff": false, "#GGGGGG": false, "merah": false, "amber; x": false,
	} {
		if got := ValidPrimary(s); got != want {
			t.Errorf("ValidPrimary(%q) = %v, want %v", s, got, want)
		}
	}
}
