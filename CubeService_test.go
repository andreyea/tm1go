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
			name:    "Valid cube - with rules",
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
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.GetLastDataUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if timestamp == "" {
				t.Errorf("CubeService.GetLastDataUpdate() error = %v, wantErr %v", "Empty timestamp", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetAll(t *testing.T) {
	t.Run("Get all cubes", func(t *testing.T) {
		cubes, err := tm1ServiceT.CubeService.GetAll()
		if err != nil {
			t.Errorf("CubeService.GetAll() error = %v", err)
		}
		if len(cubes) == 0 {
			t.Errorf("CubeService.GetAll() error = %v", "No cubes returned")
		}
	})
}

func TestCubeService_GetModelCubes(t *testing.T) {
	t.Run("Get all cubes", func(t *testing.T) {
		cubes, err := tm1ServiceT.CubeService.GetModelCubes()
		if err != nil {
			t.Errorf("CubeService.GetModelCubes() error = %v", err)
		}
		if len(cubes) == 0 {
			t.Errorf("CubeService.GetModelCubes() error = %v", "No cubes returned")
		}
	})
}

func TestCubeService_GetControlCubes(t *testing.T) {
	t.Run("Get all cubes", func(t *testing.T) {
		cubes, err := tm1ServiceT.CubeService.GetControlCubes()
		if err != nil {
			t.Errorf("CubeService.GetControlCubes() error = %v", err)
		}
		if len(cubes) == 0 {
			t.Errorf("CubeService.GetControlCubes() error = %v", "No cubes returned")
		}
	})
}

func TestCubeService_GetNumberOfCubes(t *testing.T) {
	tests := []struct {
		name             string
		skipControlCubes bool
		wantErr          bool
	}{
		{
			name:             "Get cubes count excluding control cubes",
			skipControlCubes: true,
			wantErr:          false,
		},
		{
			name:             "Get cubes count including control cubes",
			skipControlCubes: false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.CubeService.GetNumberOfCubes(tt.skipControlCubes)
			if err != nil {
				t.Errorf("CubeService.GetNumberOfCubes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if count == 0 {
				t.Errorf("CubeService.GetNumberOfCubes() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}
func TestCubeService_GetMeasureDimension(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Get measure dimension for a valid cube",
			cubeName: "Balance Sheet",
			wantErr:  false,
		},
		{
			name:     "Get measure dimension for an invalid cube",
			cubeName: "NonExistentCube",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dimension, err := tm1ServiceT.CubeService.GetMeasureDimension(tt.cubeName)
			if err != nil && !tt.wantErr {
				t.Errorf("CubeService.GetNumberOfCubes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if dimension == nil && !tt.wantErr {
				t.Errorf("CubeService.GetNumberOfCubes() error = %v, wantErr %v", "No dimension returned", tt.wantErr)
			}
			if tt.wantErr && err == nil {
				t.Errorf("CubeService.GetNumberOfCubes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCubeService_Update(t *testing.T) {
	// Create dimensions
	err := tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD1"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}
	err = tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD2"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}

	// Create a cube
	err = tm1ServiceT.CubeService.Create(tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: ""})
	if err != nil {
		t.Errorf("Error creating cube: %v", err)
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
			name:    "Valid cube - with rules",
			cube:    tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: "#this is a rule"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CubeService.Update(tt.cube)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		// Delete cubes and dimensions
		tm1ServiceT.CubeService.Delete("TestCube1")
		tm1ServiceT.DimensionService.Delete("TestD1")
		tm1ServiceT.DimensionService.Delete("TestD2")
	})
}

func TestCubeService_UpdateOrCreate(t *testing.T) {
	// Create dimensions
	err := tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD1"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}
	err = tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD2"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}

	// Create a cube
	err = tm1ServiceT.CubeService.Create(tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: ""})
	if err != nil {
		t.Errorf("Error creating cube: %v", err)
	}

	tests := []struct {
		name    string
		cube    tm1go.Cube
		wantErr bool
	}{
		{
			name:    "Valid cube - with rules",
			cube:    tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: "#this is a rule"},
			wantErr: false,
		},
		{
			name:    "Valid cube - non existent cube",
			cube:    tm1go.Cube{Name: "TestCube2", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}}, Rules: "#this is a rule"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CubeService.UpdateOrCreate(tt.cube)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.UpdateOrCreate() error = %v, wantErr %v", err, tt.wantErr)
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
