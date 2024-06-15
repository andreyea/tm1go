package tm1go_test

import (
	"testing"
)

func TestCubeService_Get(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Valid cube name",
			cubeName: "Balance Sheet",
			wantErr:  false,
		},
		{
			name:     "Invalid cube name",
			cubeName: "NonExistentCube",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.CubeService.Get(tt.cubeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
