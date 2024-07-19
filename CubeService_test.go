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

func TestCubeService_CheckRules(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Valid cube name",
			cubeName: "Planning Assumptions",
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
			ruleErrors, err := tm1ServiceT.CubeService.CheckRules(tt.cubeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.CheckRules() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(ruleErrors) != 0 {
				t.Errorf("CubeService.CheckRules() error = %v, wantErr %v", "No rule errors returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_Delete(t *testing.T) {

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
		cube    string
		wantErr bool
	}{
		{
			name:    "Valid cube name",
			cube:    "TestCube1",
			wantErr: false,
		},
		{
			name:    "Invalid cube name",
			cube:    "NonExistentCube",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CubeService.Delete(tt.cube)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		// Delete cubes and dimensions
		tm1ServiceT.DimensionService.Delete("TestD1")
		tm1ServiceT.DimensionService.Delete("TestD2")
	})
}

func TestCubeService_Exists(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			exists, err := tm1ServiceT.CubeService.Exists(tt.cubeName)
			if err != nil {
				t.Errorf("CubeService.Exists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if exists != !tt.wantErr {
				t.Errorf("CubeService.Exists() error = %v, wantErr %v", "Cube exists", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetAllNames(t *testing.T) {
	tests := []struct {
		name            string
		skipControlCube bool
		wantErr         bool
	}{
		{
			name:            "Get all cube names",
			skipControlCube: false,
			wantErr:         false,
		},
		{
			name:            "Get only model cubes",
			skipControlCube: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.GetAllNames(tt.skipControlCube)
			if err != nil {
				t.Errorf("CubeService.GetAllNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(names) == 0 {
				t.Errorf("CubeService.GetAllNames() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetAllNamesWithRules(t *testing.T) {
	tests := []struct {
		name            string
		skipControlCube bool
		wantErr         bool
	}{
		{
			name:            "Get all cube names",
			skipControlCube: false,
			wantErr:         false,
		},
		{
			name:            "Get only model cubes",
			skipControlCube: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.GetAllNamesWithRules(tt.skipControlCube)
			if err != nil {
				t.Errorf("CubeService.GetAllNamesWithRules() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(names) == 0 {
				t.Errorf("CubeService.GetAllNamesWithRules() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetAllNamesWithoutRules(t *testing.T) {
	tests := []struct {
		name            string
		skipControlCube bool
		wantErr         bool
	}{
		{
			name:            "Get all cube names",
			skipControlCube: false,
			wantErr:         false,
		},
		{
			name:            "Get only model cubes",
			skipControlCube: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.GetAllNamesWithoutRules(tt.skipControlCube)
			if err != nil {
				t.Errorf("CubeService.GetAllNamesWithoutRules() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(names) == 0 {
				t.Errorf("CubeService.GetAllNamesWithoutRules() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetDimensionNames(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			names, err := tm1ServiceT.CubeService.GetDimensionNames(tt.cubeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.GetDimensionNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (len(names) == 0) == !tt.wantErr {
				t.Errorf("CubeService.GetDimensionNames() error = %v, wantErr %v", "No dimensions returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_SearchForDimension(t *testing.T) {
	tests := []struct {
		name            string
		dimName         string
		skipControlCube bool
		wantErr         bool
	}{
		{
			name:            "Account dimension exists. All cubes.",
			dimName:         "Account",
			skipControlCube: false,
			wantErr:         false,
		},
		{
			name:            "Account dimension exists. Only model cubes.",
			dimName:         "Account",
			skipControlCube: true,
			wantErr:         false,
		},
		{
			name:            "Invalid dimension name",
			dimName:         "NonExistentDimension",
			skipControlCube: false,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.SearchForDimension(tt.dimName, tt.skipControlCube)
			if err != nil {
				t.Errorf("CubeService.SearchForDimension() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (len(names) == 0) == !tt.wantErr {
				t.Errorf("CubeService.SearchForDimension() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_SearchForDimensionSubstring(t *testing.T) {
	tests := []struct {
		name            string
		dimName         string
		skipControlCube bool
		wantErr         bool
	}{
		{
			name:            "Acc* dimensions exists. All cubes.",
			dimName:         "Acc",
			skipControlCube: false,
			wantErr:         false,
		},
		{
			name:            "Acc* dimensions exists. Only model cubes.",
			dimName:         "Acc",
			skipControlCube: true,
			wantErr:         false,
		},
		{
			name:            "Invalid dimension name",
			dimName:         "NonExistentDimension",
			skipControlCube: false,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.SearchForDimensionSubstring(tt.dimName, tt.skipControlCube)
			if err != nil {
				t.Errorf("CubeService.SearchForDimensionSubstring() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (len(names) == 0) == !tt.wantErr {
				t.Errorf("CubeService.SearchForDimensionSubstring() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_SearchForRuleSubstring(t *testing.T) {
	tests := []struct {
		name             string
		substring        string
		skipControlCube  bool
		caseInsensitive  bool
		spaceInsensitive bool
		wantErr          bool
	}{
		{
			name:             "*skipcheck* rule exists. All cubes.",
			substring:        "skipcheck",
			skipControlCube:  false,
			caseInsensitive:  true,
			spaceInsensitive: false,
			wantErr:          false,
		},
		{
			name:             "*skipcheck* rule exists. Only model cubes.",
			substring:        "skipcheck",
			skipControlCube:  true,
			caseInsensitive:  true,
			spaceInsensitive: false,
			wantErr:          false,
		},
		{
			name:             "Non existent rule text",
			substring:        "NonExistentTextInsideRule",
			skipControlCube:  false,
			caseInsensitive:  true,
			spaceInsensitive: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.CubeService.SearchForRuleSubstring(tt.substring, tt.skipControlCube, tt.caseInsensitive, tt.spaceInsensitive)
			if err != nil {
				t.Errorf("CubeService.SearchForRuleSubstring() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (len(names) == 0) == !tt.wantErr {
				t.Errorf("CubeService.SearchForRuleSubstring() error = %v, wantErr %v", "No cubes returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_GetStorageDimensionOrder(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			names, err := tm1ServiceT.CubeService.GetStorageDimensionOrder(tt.cubeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CubeService.GetStorageDimensionOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (len(names) == 0) == !tt.wantErr {
				t.Errorf("CubeService.GetStorageDimensionOrder() error = %v, wantErr %v", "No dimensions returned", tt.wantErr)
			}
		})
	}
}

func TestCubeService_UpdateStorageDimensionOrder(t *testing.T) {
	// Create dimensions
	err := tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD1"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}
	err = tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD2"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}
	err = tm1ServiceT.DimensionService.Create(&tm1go.Dimension{Name: "TestD3"})
	if err != nil {
		t.Errorf("Error creating dimension: %v", err)
	}

	// Create a cube
	err = tm1ServiceT.CubeService.Create(tm1go.Cube{Name: "TestCube1", Dimensions: []tm1go.Dimension{{Name: "TestD1"}, {Name: "TestD2"}, {Name: "TestD3"}}, Rules: ""})
	if err != nil {
		t.Errorf("Error creating cube: %v", err)
	}

	tests := []struct {
		name       string
		cubeName   string
		dimensions []string
		wantErr    bool
	}{
		{
			name:       "Reorder dimensions",
			cubeName:   "TestCube1",
			dimensions: []string{"TestD2", "TestD3", "TestD1"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.CubeService.UpdateStorageDimensionOrder(tt.cubeName, tt.dimensions)
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
		tm1ServiceT.DimensionService.Delete("TestD3")
	})
}

func TestCubeService_Load(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			err := tm1ServiceT.CubeService.Load(tt.cubeName)
			if err != nil && !tt.wantErr {
				t.Errorf("CubeService.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCubeService_Unload(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			err := tm1ServiceT.CubeService.Unload(tt.cubeName)
			if err != nil && !tt.wantErr {
				t.Errorf("CubeService.Unload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCubeService_Lock(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			err := tm1ServiceT.CubeService.Lock(tt.cubeName)
			if err != nil && !tt.wantErr {
				t.Errorf("CubeService.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCubeService_Unlock(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		wantErr  bool
	}{
		{
			name:     "Balance Sheet cube exists",
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
			err := tm1ServiceT.CubeService.Unlock(tt.cubeName)
			if err != nil && !tt.wantErr {
				t.Errorf("CubeService.Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
