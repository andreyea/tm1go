package tm1go_test

import (
	"testing"

	"github.com/andreyea/tm1go"
)

func TestHierarchyService_Create(t *testing.T) {
	tests := []struct {
		name      string
		hierarchy *tm1go.Hierarchy
		wantErr   bool
	}{
		{
			name: "Create new hierarchy",
			hierarchy: &tm1go.Hierarchy{
				Name: "TestHierarchy",
				Dimension: tm1go.Dimension{
					Name: "TestDimension",
				},
				Elements: []tm1go.Element{
					{Name: "Element1", Type: "Numeric"},
					{Name: "Element2", Type: "Numeric"},
					{Name: "Total", Type: "Consolidated"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create dimension if it doesn't exist
			if tt.hierarchy.Dimension.Name != "Account" {
				dim := &tm1go.Dimension{
					Name: tt.hierarchy.Dimension.Name,
					Hierarchies: []tm1go.Hierarchy{
						{Name: tt.hierarchy.Dimension.Name},
					},
				}
				err := tm1ServiceT.DimensionService.Create(dim)
				if err != nil {
					t.Fatalf("Failed to create test dimension: %v", err)
				}
			}

			err := tm1ServiceT.HierarchyService.Create(tt.hierarchy)
			if (err != nil) != tt.wantErr {
				t.Errorf("HierarchyService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the hierarchy was created
				createdHierarchy, err := tm1ServiceT.HierarchyService.Get(tt.hierarchy.Dimension.Name, tt.hierarchy.Name)
				if err != nil {
					t.Errorf("Failed to get created hierarchy: %v", err)
					return
				}

				if createdHierarchy == nil {
					t.Errorf("Created hierarchy not found")
					return
				}

				if createdHierarchy.Name != tt.hierarchy.Name {
					t.Errorf("Created hierarchy name = %v, want %v", createdHierarchy.Name, tt.hierarchy.Name)
				}

				// Clean up: delete the created hierarchy
				err = tm1ServiceT.HierarchyService.Delete(tt.hierarchy.Dimension.Name, tt.hierarchy.Name)
				if err != nil {
					t.Logf("Failed to delete test hierarchy: %v", err)
				}
			}
		})
	}
}
