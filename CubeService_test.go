package tm1go_test

import (
	"testing"

	"github.com/andreyea/tm1go"
)

func TestCubeService_Create(t *testing.T) {
	// Create dimensions
	err := tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD1"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}
	err = tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD2"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}

	tests := []struct {
		name    string
		cube    tm1go.Cube
		wantErr bool
	}{
		{
			name:    "Valid cube - no rules",
			cube:    tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: ""},
			wantErr: false,
		},
		{
			name:    "Valid cube - no rules",
			cube:    tm1go.Cube{Name: "TestCube2", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: "#this is a rule"},
			wantErr: false,
		},
		{
			name:    "Invalid cube - only name provided",
			cube:    tm1go.Cube{Name: "TestCube3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CubeService.Create(tt.cube)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		// Delete cubes and dimensions
		tm1ServiceT.CubeService.Delete("TestCube1")
		tm1ServiceT.CubeService.Delete("TestCube2")
		tm1ServiceT.DimensionService.Delete("TestD1")
		tm1ServiceT.DimensionService.Delete("TestD2")
	})

}

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

func TestCubeService_GetLastDataUpdate(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := tm1ServiceT.CubeService.GetLastDataUpdate(tt.cubeName)
			t.Logf("TIMESTAMP:%v", timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.GetLastDataUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if timestamp == "" {
				t.Errorf("CubeService.GetLastDataUpdate() error = %v, wantErr %v", "Empty timestamp", tt.wantErr)
			}
		})
	}
}
