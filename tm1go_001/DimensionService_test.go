package tm1go_test

import (
	"fmt"
	"testing"

	"github.com/andreyea/tm1go"
)

func TestDimensionService_Create(t *testing.T) {
	tests := []struct {
		name      string
		dimension *tm1go.Dimension
		wantErr   bool
	}{
		{
			name: "Create new dimension",
			dimension: &tm1go.Dimension{
				Name: "TestDimension",
				Hierarchies: []tm1go.Hierarchy{
					{
						Name: "TestDimension",
						Elements: []tm1go.Element{
							{Name: "Element1", Type: "Numeric"},
							{Name: "Element2", Type: "Numeric"},
							{Name: "Total", Type: "Consolidated"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create dimension with existing name",
			dimension: &tm1go.Dimension{
				Name: "Account",
				Hierarchies: []tm1go.Hierarchy{
					{Name: "Account"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.DimensionService.Create(tt.dimension)
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the dimension was created
				createdDim, err := tm1ServiceT.DimensionService.Get(tt.dimension.Name)
				if err != nil {
					t.Errorf("Failed to get created dimension: %v", err)
					return
				}

				if createdDim == nil {
					t.Errorf("Created dimension not found")
					return
				}

				if createdDim.Name != tt.dimension.Name {
					t.Errorf("Created dimension name = %v, want %v", createdDim.Name, tt.dimension.Name)
				}

				// Clean up: delete the created dimension
				err = tm1ServiceT.DimensionService.Delete(tt.dimension.Name)
				if err != nil {
					t.Logf("Failed to delete test dimension: %v", err)
				}
			}
		})
	}
}

func TestDimensionService_Get(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		wantErr       bool
		checkFunc     func(*tm1go.Dimension) error
	}{
		{
			name:          "Get existing dimension",
			dimensionName: "Account",
			wantErr:       false,
			checkFunc: func(dim *tm1go.Dimension) error {
				if dim == nil {
					return fmt.Errorf("dimension is nil")
				}
				if dim.Name != "Account" {
					return fmt.Errorf("expected dimension name 'Account', got '%s'", dim.Name)
				}
				if len(dim.Hierarchies) == 0 {
					return fmt.Errorf("dimension has no hierarchies")
				}
				return nil
			},
		},
		{
			name:          "Get non-existent dimension",
			dimensionName: "NonExistentDimension",
			wantErr:       true,
			checkFunc:     nil,
		},
		{
			name:          "Get dimension with empty name",
			dimensionName: "",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dimension, err := tm1ServiceT.DimensionService.Get(tt.dimensionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(dimension); err != nil {
					t.Errorf("DimensionService.Get() check failed: %v", err)
				}
			}

			if tt.wantErr && dimension != nil {
				t.Errorf("DimensionService.Get() returned dimension for error case: %v", dimension)
			}
		})
	}
}

func TestDimensionService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		setupFunc     func() error
		wantErr       bool
	}{
		{
			name:          "Delete existing dimension",
			dimensionName: "TestDeleteDimension",
			setupFunc: func() error {
				dim := &tm1go.Dimension{
					Name: "TestDeleteDimension",
					Hierarchies: []tm1go.Hierarchy{
						{Name: "TestDeleteDimension"},
					},
				}
				return tm1ServiceT.DimensionService.Create(dim)
			},
			wantErr: false,
		},
		{
			name:          "Delete non-existent dimension",
			dimensionName: "NonExistentDimension",
			setupFunc:     func() error { return nil },
			wantErr:       true,
		},
		{
			name:          "Delete dimension with empty name",
			dimensionName: "",
			setupFunc:     func() error { return nil },
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Perform deletion
			err = tm1ServiceT.DimensionService.Delete(tt.dimensionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the dimension was deleted
				_, err := tm1ServiceT.DimensionService.Get(tt.dimensionName)
				if err == nil {
					t.Errorf("DimensionService.Delete() failed, dimension still exists")
				}
			}
		})
	}
}

func TestDimensionService_Exists(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		setupFunc     func() error
		cleanupFunc   func()
		want          bool
		wantErr       bool
	}{
		{
			name:          "Existing dimension",
			dimensionName: "TestExistsDimension",
			setupFunc: func() error {
				dim := &tm1go.Dimension{
					Name: "TestExistsDimension",
					Hierarchies: []tm1go.Hierarchy{
						{Name: "TestExistsDimension"},
					},
				}
				return tm1ServiceT.DimensionService.Create(dim)
			},
			cleanupFunc: func() {
				tm1ServiceT.DimensionService.Delete("TestExistsDimension")
			},
			want:    true,
			wantErr: false,
		},
		{
			name:          "Non-existent dimension",
			dimensionName: "NonExistentDimension",
			setupFunc:     func() error { return nil },
			cleanupFunc:   func() {},
			want:          false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer tt.cleanupFunc()

			// Perform existence check
			got, err := tm1ServiceT.DimensionService.Exists(tt.dimensionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DimensionService.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDimensionService_GetAllNames(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Get all dimension names",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.DimensionService.GetAllNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.GetAllNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(names) == 0 {
					t.Errorf("DimensionService.GetAllNames() returned empty list, expected at least one dimension")
				}

				// Check for some expected dimension names
				expectedDimensions := []string{"Account", "Time", "Version"}
				for _, expected := range expectedDimensions {
					found := false
					for _, name := range names {
						if name == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("DimensionService.GetAllNames() did not return expected dimension: %s", expected)
					}
				}
			}
		})
	}
}

func TestDimensionService_GetNumberOfDimensions(t *testing.T) {
	tests := []struct {
		name            string
		skipControlDims bool
		wantErr         bool
	}{
		{
			name:            "Get number of dimensions including control dimensions",
			skipControlDims: false,
			wantErr:         false,
		},
		{
			name:            "Get number of dimensions excluding control dimensions",
			skipControlDims: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.DimensionService.GetNumberOfDimensions(tt.skipControlDims)
			if (err != nil) != tt.wantErr {
				t.Errorf("DimensionService.GetNumberOfDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if count <= 0 {
					t.Errorf("DimensionService.GetNumberOfDimensions() returned %d, expected a number greater than 0", count)
				}
			}
		})
	}
}
