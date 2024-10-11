package utils

import (
	"testing"
)

func TestUpdateImageName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  string
	}{
		{
			name:      "Already has correct prefix",
			imageName: "harbor.power.sd.k9s.space/victory/myapp:latest",
			expected:  "harbor.power.sd.k9s.space/victory/myapp:latest",
		},
		{
			name:      "Has different prefix -1",
			imageName: "docker.io/library/victory/myapp:latest",
			expected:  "harbor.power.sd.k9s.space/victory/myapp:latest",
		},
		{
			name:      "Has different prefix -2",
			imageName: "docker.io/library/myapp:latest",
			expected:  "harbor.power.sd.k9s.space/myapp:latest",
		},
		{
			name:      "No prefix",
			imageName: "myapp:latest",
			expected:  "harbor.power.sd.k9s.space/myapp:latest",
		},
		{
			name:      "two layer",
			imageName: "victory/myapp:latest",
			expected:  "harbor.power.sd.k9s.space/victory/myapp:latest",
		},
		{
			name:      "three layer",
			imageName: "victory/x/myapp:latest",
			expected:  "harbor.power.sd.k9s.space/victory/x/myapp:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			imageName := UpdateImageName(tt.imageName, "harbor.power.sd.k9s.space")
			if imageName != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, imageName)
			}
		})
	}
}
