package bgpengine

import (
	"testing"
)

func TestGetPriority(t *testing.T) {
	e := &Engine{}
	tests := []struct {
		name     string
		priority int
	}{
		{nameRouteLeak, 3},
		{nameHardOutage, 3},
		{nameDDoSMitigation, 3},
		{nameFlap, 2},
		{nameTrafficEng, 1},
		{nameDiscovery, 0},
		{"", 0},
		{"Unknown", 0},
	}

	for _, tt := range tests {
		p := e.GetPriority(tt.name)
		if p != tt.priority {
			t.Errorf("Expected priority %d for %s, got %d", tt.priority, tt.name, p)
		}
	}
}
