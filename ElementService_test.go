package tm1go_test

import (
	"strings"
	"testing"
)

func TestElementService_Get(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		elementName   string
		wantErr       bool
	}{
		{
			name:          "Get existing numeric element",
			dimensionName: "Line",
			hierarchyName: "Line",
			elementName:   "E1",
			wantErr:       false,
		},

		{
			name:          "Get non-existent element",
			dimensionName: "Time",
			hierarchyName: "Time",
			elementName:   "NonExistentElement",
			wantErr:       true,
		},
		{
			name:          "Get element from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "Time",
			elementName:   "2023",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element, err := tm1ServiceT.ElementService.Get(tt.dimensionName, tt.hierarchyName, tt.elementName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if element == nil {
					t.Errorf("ElementService.Get() returned nil element for existing element")
					return
				}

				if !strings.EqualFold(element.Name, tt.elementName) {
					t.Errorf("ElementService.Get() returned element with name = %v, want %v", element.Name, tt.elementName)
				}
			}
		})
	}
}
