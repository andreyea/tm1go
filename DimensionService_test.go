package tm1go_test

import (
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
