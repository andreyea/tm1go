package tm1go_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andreyea/tm1go"
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

func TestElementService_Create(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		element       *tm1go.Element
		wantErr       bool
	}{
		{
			name:          "Create new string element",
			dimensionName: "Account",
			hierarchyName: "Account",
			element: &tm1go.Element{
				Name: "NewStrElement2",
				Type: "String",
			},
			wantErr: false,
		},
		{
			name:          "Create new numeric element",
			dimensionName: "Account",
			hierarchyName: "Account",
			element: &tm1go.Element{
				Name: "NewAccountElement2",
				Type: "Numeric",
			},
			wantErr: false,
		},
		{
			name:          "Create element in non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "TestHierarchy",
			element: &tm1go.Element{
				Name: "NewElement",
				Type: "Numeric",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ElementService.Create(tt.dimensionName, tt.hierarchyName, tt.element)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the element was created
				createdElement, err := tm1ServiceT.ElementService.Get(tt.dimensionName, tt.hierarchyName, tt.element.Name)
				if err != nil {
					t.Errorf("Failed to get created element: %v", err)
					return
				}

				if createdElement == nil {
					t.Errorf("Created element not found")
					return
				}

				if !strings.EqualFold(createdElement.Name, tt.element.Name) {
					t.Errorf("Created element name = %v, want %v", createdElement.Name, tt.element.Name)
				}

				if createdElement.Type != tt.element.Type {
					t.Errorf("Created element type = %v, want %v", createdElement.Type, tt.element.Type)
				}

				// Clean up: delete the created element
				err = tm1ServiceT.ElementService.Delete(tt.dimensionName, tt.hierarchyName, tt.element.Name)
				if err != nil {
					t.Logf("Failed to delete test element: %v", err)
				}
			}
		})
	}
}

func TestElementService_Update(t *testing.T) {
	tests := []struct {
		name            string
		dimensionName   string
		hierarchyName   string
		originalElement *tm1go.Element
		updatedElement  *tm1go.Element
		wantErr         bool
	}{
		{
			name:          "Update numeric element to string",
			dimensionName: "Account",
			hierarchyName: "Account",
			originalElement: &tm1go.Element{
				Name: "UpdateTestElement1",
				Type: "Numeric",
			},
			updatedElement: &tm1go.Element{
				Name: "UpdateTestElement1",
				Type: "String",
			},
			wantErr: false,
		},
		{
			name:            "Update non-existent element",
			dimensionName:   "TestDimension",
			hierarchyName:   "TestHierarchy",
			originalElement: nil,
			updatedElement: &tm1go.Element{
				Name: "NonExistentElement",
				Type: "String",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the original element if it exists
			if tt.originalElement != nil {
				err := tm1ServiceT.ElementService.Create(tt.dimensionName, tt.hierarchyName, tt.originalElement)
				if err != nil {
					t.Fatalf("Failed to create test element: %v", err)
				}
			}

			// Perform the update
			err := tm1ServiceT.ElementService.Update(tt.dimensionName, tt.hierarchyName, tt.updatedElement)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the element was updated
				updatedElement, err := tm1ServiceT.ElementService.Get(tt.dimensionName, tt.hierarchyName, tt.updatedElement.Name)
				if err != nil {
					t.Errorf("Failed to get updated element: %v", err)
					return
				}

				if updatedElement == nil {
					t.Errorf("Updated element not found")
					return
				}

				if updatedElement.Name != tt.updatedElement.Name {
					t.Errorf("Updated element name = %v, want %v", updatedElement.Name, tt.updatedElement.Name)
				}

				if updatedElement.Type != tt.updatedElement.Type {
					t.Errorf("Updated element type = %v, want %v", updatedElement.Type, tt.updatedElement.Type)
				}

				// Clean up: delete the updated element
				err = tm1ServiceT.ElementService.Delete(tt.dimensionName, tt.hierarchyName, tt.updatedElement.Name)
				if err != nil {
					t.Logf("Failed to delete test element: %v", err)
				}
			}
		})
	}
}

func TestElementService_Exists(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		elementName   string
		want          bool
		wantErr       bool
	}{
		{
			name:          "Existing element",
			dimensionName: "Line",
			hierarchyName: "Line",
			elementName:   "e1",
			want:          true,
			wantErr:       false,
		},
		{
			name:          "Non-existing element",
			dimensionName: "Line",
			hierarchyName: "Line",
			elementName:   "NonExistentElement",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "Element in non-existing dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "Default",
			elementName:   "SomeElement",
			want:          false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm1ServiceT.ElementService.Exists(tt.dimensionName, tt.hierarchyName, tt.elementName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ElementService.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElementService_UpdateOrCreate(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		element       *tm1go.Element
		wantCreated   bool
		wantErr       bool
	}{
		{
			name:          "Update existing element",
			dimensionName: "Line",
			hierarchyName: "Line",
			element: &tm1go.Element{
				Name: "E1",
				Type: "Numeric",
			},
			wantCreated: false,
			wantErr:     false,
		},
		{
			name:          "Create new element",
			dimensionName: "Line",
			hierarchyName: "Line",
			element: &tm1go.Element{
				Name: "NewTestElement",
				Type: "String",
			},
			wantCreated: true,
			wantErr:     false,
		},
		{
			name:          "Attempt update with invalid type",
			dimensionName: "Line",
			hierarchyName: "Line",
			element: &tm1go.Element{
				Name: "E1",
				Type: "InvalidType",
			},
			wantCreated: false,
			wantErr:     true,
		},
		{
			name:          "Attempt create in non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			element: &tm1go.Element{
				Name: "TestElement",
				Type: "Numeric",
			},
			wantCreated: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ElementService.UpdateOrCreate(tt.dimensionName, tt.hierarchyName, tt.element)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.UpdateOrCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the element exists and has the correct properties
				element, err := tm1ServiceT.ElementService.Get(tt.dimensionName, tt.hierarchyName, tt.element.Name)
				if err != nil {
					t.Errorf("Failed to get element after UpdateOrCreate: %v", err)
					return
				}

				if element == nil {
					t.Errorf("Element not found after UpdateOrCreate")
					return
				}

				if !strings.EqualFold(element.Name, tt.element.Name) {
					t.Errorf("Element name = %v, want %v", element.Name, tt.element.Name)
				}

				if element.Type != tt.element.Type {
					t.Errorf("Element type = %v, want %v", element.Type, tt.element.Type)
				}

				// Clean up: delete the newly created element if it was created in this test
				if tt.wantCreated {
					err = tm1ServiceT.ElementService.Delete(tt.dimensionName, tt.hierarchyName, tt.element.Name)
					if err != nil {
						t.Logf("Failed to delete test element: %v", err)
					}
				}
			}
		})
	}
}

func TestElementService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		elementName   string
		setupFunc     func() error
		wantErr       bool
	}{
		{
			name:          "Delete existing element",
			dimensionName: "Line",
			hierarchyName: "Line",
			elementName:   "ElementToDelete",
			setupFunc: func() error {
				return tm1ServiceT.ElementService.Create("Line", "Line", &tm1go.Element{
					Name: "ElementToDelete",
					Type: "Numeric",
				})
			},
			wantErr: false,
		},
		{
			name:          "Delete non-existent element",
			dimensionName: "Line",
			hierarchyName: "Line",
			elementName:   "NonExistentElement",
			setupFunc:     func() error { return nil },
			wantErr:       true,
		},
		{
			name:          "Delete element from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			elementName:   "TestElement",
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
			err = tm1ServiceT.ElementService.Delete(tt.dimensionName, tt.hierarchyName, tt.elementName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the element was deleted
				exists, err := tm1ServiceT.ElementService.Exists(tt.dimensionName, tt.hierarchyName, tt.elementName)
				if err != nil {
					t.Errorf("Failed to check if element exists after deletion: %v", err)
					return
				}
				if exists {
					t.Errorf("Element still exists after deletion")
				}
			}
		})
	}
}

func TestElementService_GetElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(elements []tm1go.Element) error
	}{
		{
			name:          "Get elements from Line dimension",
			dimensionName: "Line",
			hierarchyName: "Line",
			wantErr:       false,
			checkFunc: func(elements []tm1go.Element) error {
				if len(elements) == 0 {
					return fmt.Errorf("no elements returned")
				}
				// Check for some expected elements
				expectedElements := map[string]bool{
					"e1": false,
					"e2": false,
					"e3": false,
				}
				for _, elem := range elements {
					if _, exists := expectedElements[elem.Name]; exists {
						expectedElements[elem.Name] = true
					}
				}
				for name, found := range expectedElements {
					if !found {
						return fmt.Errorf("expected element %s not found", name)
					}
				}
				return nil
			},
		},
		{
			name:          "Get elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := tm1ServiceT.ElementService.GetElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(elements); err != nil {
					t.Errorf("ElementService.GetElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetEdges(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(edges []tm1go.Edge) error
	}{
		{
			name:          "Get edges from Line dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(edges []tm1go.Edge) error {
				if len(edges) == 0 {
					return fmt.Errorf("no edges returned")
				}

				// Check for expected structure
				for _, edge := range edges {
					if edge.ParentName == "" {
						return fmt.Errorf("edge has empty ParentName")
					}
					if edge.ComponentName == "" {
						return fmt.Errorf("edge has empty ComponentName")
					}
					// Weight is a float64, so we don't need to check if it's empty
				}

				// Check for at least one known edge
				foundKnownEdge := false
				for _, edge := range edges {
					if edge.ParentName == "1" && edge.ComponentName == "10" {
						foundKnownEdge = true
						break
					}
				}
				if !foundKnownEdge {
					return fmt.Errorf("expected edge '1' -> '10' not found")
				}

				return nil
			},
		},
		{
			name:          "Get edges from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges, err := tm1ServiceT.ElementService.GetEdges(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetEdges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(edges); err != nil {
					t.Errorf("ElementService.GetEdges() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetLeafElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(elements []tm1go.Element) error
	}{
		{
			name:          "Get leaf elements from Line dimension",
			dimensionName: "Line",
			hierarchyName: "Line",
			wantErr:       false,
			checkFunc: func(elements []tm1go.Element) error {
				if len(elements) == 0 {
					return fmt.Errorf("no leaf elements returned")
				}

				// Check that all returned elements are leaf elements (not consolidated)
				for _, elem := range elements {
					if elem.Type == tm1go.Consolidated.String() {
						return fmt.Errorf("consolidated element found: %s", elem.Name)
					}
				}

				// Check for expected leaf elements
				expectedLeaves := map[string]bool{
					"e1": false,
					"e2": false,
					"e3": false,
				}
				for _, elem := range elements {
					if _, exists := expectedLeaves[elem.Name]; exists {
						expectedLeaves[elem.Name] = true
					}
				}
				for name, found := range expectedLeaves {
					if !found {
						return fmt.Errorf("expected leaf element %s not found", name)
					}
				}

				return nil
			},
		},
		{
			name:          "Get leaf elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := tm1ServiceT.ElementService.GetLeafElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetLeafElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(elements); err != nil {
					t.Errorf("ElementService.GetLeafElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetLeafElementNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(names []string) error
	}{
		{
			name:          "Get leaf element names from Line dimension",
			dimensionName: "Line",
			hierarchyName: "Line",
			wantErr:       false,
			checkFunc: func(names []string) error {
				if len(names) == 0 {
					return fmt.Errorf("no leaf element names returned")
				}

				// Check for expected leaf element names
				expectedLeaves := map[string]bool{
					"e1": false,
					"e2": false,
					"e3": false,
				}
				for _, name := range names {
					if _, exists := expectedLeaves[name]; exists {
						expectedLeaves[name] = true
					}
				}
				for name, found := range expectedLeaves {
					if !found {
						return fmt.Errorf("expected leaf element name %s not found", name)
					}
				}

				// Check that no consolidated element names are present
				consolidatedElements := []string{"Total Line"}
				for _, name := range names {
					for _, consolidated := range consolidatedElements {
						if name == consolidated {
							return fmt.Errorf("consolidated element name found: %s", name)
						}
					}
				}

				return nil
			},
		},
		{
			name:          "Get leaf element names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.ElementService.GetLeafElementNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetLeafElementNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(names); err != nil {
					t.Errorf("ElementService.GetLeafElementNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetConsolidatedElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(elements []tm1go.Element) error
	}{
		{
			name:          "Get consolidated elements from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(elements []tm1go.Element) error {
				if len(elements) == 0 {
					return fmt.Errorf("no consolidated elements returned")
				}

				// Check that all returned elements are consolidated
				for _, elem := range elements {
					if elem.Type != tm1go.Consolidated.String() {
						return fmt.Errorf("non-consolidated element found: %s (Type: %s)", elem.Name, elem.Type)
					}
				}

				// Check for expected consolidated elements
				expectedConsolidated := map[string]bool{
					"1": false,
				}
				for _, elem := range elements {
					if _, exists := expectedConsolidated[elem.Name]; exists {
						expectedConsolidated[elem.Name] = true
					}
				}
				for name, found := range expectedConsolidated {
					if !found {
						return fmt.Errorf("expected consolidated element %s not found", name)
					}
				}

				return nil
			},
		},
		{
			name:          "Get consolidated elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := tm1ServiceT.ElementService.GetConsolidatedElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetConsolidatedElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(elements); err != nil {
					t.Errorf("ElementService.GetConsolidatedElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetConsolidatedElementNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(names []string) error
	}{
		{
			name:          "Get consolidated element names from account dimension",
			dimensionName: "Account",
			hierarchyName: "ACcount",
			wantErr:       false,
			checkFunc: func(names []string) error {
				if len(names) == 0 {
					return fmt.Errorf("no consolidated element names returned")
				}

				// Check for expected consolidated element names
				expectedConsolidated := map[string]bool{
					"1": false,
				}
				for _, name := range names {
					if _, exists := expectedConsolidated[name]; exists {
						expectedConsolidated[name] = true
					}
				}
				for name, found := range expectedConsolidated {
					if !found {
						return fmt.Errorf("expected consolidated element name %s not found", name)
					}
				}

				return nil
			},
		},
		{
			name:          "Get consolidated element names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.ElementService.GetConsolidatedElementNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetConsolidatedElementNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(names); err != nil {
					t.Errorf("ElementService.GetConsolidatedElementNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumericElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(elements []tm1go.Element) error
	}{
		{
			name:          "Get numeric elements from measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(elements []tm1go.Element) error {
				if len(elements) == 0 {
					return fmt.Errorf("no numeric elements returned")
				}

				// Check that all returned elements are numeric
				for _, elem := range elements {
					if elem.Type != tm1go.Numeric.String() {
						return fmt.Errorf("non-numeric element found: %s (Type: %s)", elem.Name, elem.Type)
					}
				}

				// Check for expected numeric elements
				expectedNumeric := map[string]bool{
					"Value": false,
				}
				for _, elem := range elements {
					if _, exists := expectedNumeric[elem.Name]; exists {
						expectedNumeric[elem.Name] = true
					}
				}
				for name, found := range expectedNumeric {
					if !found {
						return fmt.Errorf("expected numeric element %s not found", name)
					}
				}

				// Check that no string or consolidated elements are present
				nonNumericElements := []string{"StringElement", "String"}
				for _, elem := range elements {
					for _, nonNumeric := range nonNumericElements {
						if elem.Name == nonNumeric {
							return fmt.Errorf("non-numeric element found: %s", nonNumeric)
						}
					}
				}

				return nil
			},
		},
		{
			name:          "Get numeric elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := tm1ServiceT.ElementService.GetNumericElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumericElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(elements); err != nil {
					t.Errorf("ElementService.GetNumericElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumericElementNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(names []string) error
	}{
		{
			name:          "Get numeric element names from Measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(names []string) error {
				if len(names) == 0 {
					return fmt.Errorf("no numeric element names returned")
				}

				// Check for expected numeric element names
				expectedNumeric := map[string]bool{
					"Value": false,
				}
				for _, name := range names {
					if _, exists := expectedNumeric[name]; exists {
						expectedNumeric[name] = true
					}
				}
				for name, found := range expectedNumeric {
					if !found {
						return fmt.Errorf("expected numeric element name %s not found", name)
					}
				}

				// Check that no string or consolidated element names are present
				nonNumericElements := []string{"StringElement", "Total Line"}
				for _, name := range names {
					for _, nonNumeric := range nonNumericElements {
						if name == nonNumeric {
							return fmt.Errorf("non-numeric element name found: %s", nonNumeric)
						}
					}
				}

				return nil
			},
		},
		{
			name:          "Get numeric element names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.ElementService.GetNumericElementNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumericElementNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(names); err != nil {
					t.Errorf("ElementService.GetNumericElementNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetStringElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(elements []tm1go.Element) error
	}{
		{
			name:          "Get string elements from Measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(elements []tm1go.Element) error {
				if len(elements) == 0 {
					return fmt.Errorf("no string elements returned")
				}

				// Check that all returned elements are string type
				for _, elem := range elements {
					if elem.Type != tm1go.String.String() {
						return fmt.Errorf("non-string element found: %s (Type: %s)", elem.Name, elem.Type)
					}
				}

				// Check for expected string element
				foundString := false
				for _, elem := range elements {
					if elem.Name == "String" {
						foundString = true
						break
					}
				}
				if !foundString {
					return fmt.Errorf("expected string element 'String' not found")
				}

				// Check that no numeric or consolidated elements are present
				nonStringElements := []string{"Value", "Total Measure"}
				for _, elem := range elements {
					for _, nonString := range nonStringElements {
						if elem.Name == nonString {
							return fmt.Errorf("non-string element found: %s", nonString)
						}
					}
				}

				return nil
			},
		},
		{
			name:          "Get string elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := tm1ServiceT.ElementService.GetStringElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetStringElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(elements); err != nil {
					t.Errorf("ElementService.GetStringElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetStringElementNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(names []string) error
	}{
		{
			name:          "Get string element names from Measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(names []string) error {
				if len(names) == 0 {
					return fmt.Errorf("no string element names returned")
				}

				// Check for expected string element name
				foundString := false
				for _, name := range names {
					if name == "String" {
						foundString = true
						break
					}
				}
				if !foundString {
					return fmt.Errorf("expected string element name 'String' not found")
				}

				// Check that no numeric or consolidated element names are present
				nonStringElements := []string{"Value", "Total Measure"}
				for _, name := range names {
					for _, nonString := range nonStringElements {
						if name == nonString {
							return fmt.Errorf("non-string element name found: %s", nonString)
						}
					}
				}

				return nil
			},
		},
		{
			name:          "Get string element names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.ElementService.GetStringElementNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetStringElementNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(names); err != nil {
					t.Errorf("ElementService.GetStringElementNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetElementNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(names []string) error
	}{
		{
			name:          "Get all element names from Line dimension",
			dimensionName: "Line",
			hierarchyName: "Line",
			wantErr:       false,
			checkFunc: func(names []string) error {
				if len(names) == 0 {
					return fmt.Errorf("no element names returned")
				}

				// Check for expected element names
				expectedElements := map[string]bool{
					"e1": false,
					"e2": false,
					"e3": false,
				}
				for _, name := range names {
					if _, exists := expectedElements[name]; exists {
						expectedElements[name] = true
					}
				}
				for name, found := range expectedElements {
					if !found {
						return fmt.Errorf("expected element name %s not found", name)
					}
				}

				// Check that the number of returned names is at least the number of expected elements
				if len(names) < len(expectedElements) {
					return fmt.Errorf("fewer element names returned than expected")
				}

				return nil
			},
		},
		{
			name:          "Get element names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := tm1ServiceT.ElementService.GetElementNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetElementNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(names); err != nil {
					t.Errorf("ElementService.GetElementNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumberOfElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get number of elements from Line dimension",
			dimensionName: "Line",
			hierarchyName: "Line",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of elements to be greater than 0, got %d", count)
				}
				return nil
			},
		},
		{
			name:          "Get number of elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetNumberOfElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumberOfElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetNumberOfElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumberOfConsolidatedElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get number of consolidated elements from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of consolidated elements to be greater than 0, got %d", count)
				}
				return nil
			},
		},
		{
			name:          "Get number of consolidated elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetNumberOfConsolidatedElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumberOfConsolidatedElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetNumberOfConsolidatedElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumberOfLeafElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get number of leaf elements from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of leaf elements to be greater than 0, got %d", count)
				}
				return nil
			},
		},
		{
			name:          "Get number of leaf elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetNumberOfLeafElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumberOfLeafElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetNumberOfLeafElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumberOfNumericElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get number of numeric elements from Measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of numeric elements to be greater than 0, got %d", count)
				}

				return nil
			},
		},
		{
			name:          "Get number of numeric elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetNumberOfNumericElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumberOfNumericElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetNumberOfNumericElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetNumberOfStringElements(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get number of string elements from Measure dimension",
			dimensionName: "Measure",
			hierarchyName: "Measure",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of string elements to be greater than 0, got %d", count)
				}

				return nil
			},
		},
		{
			name:          "Get number of string elements from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetNumberOfStringElements(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetNumberOfStringElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetNumberOfStringElements() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetLevelNames(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(levelNames []string) error
	}{
		{
			name:          "Get level names from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(levelNames []string) error {
				if len(levelNames) == 0 {
					return fmt.Errorf("expected at least one level name, got none")
				}

				// Check for expected level names
				// Adjust these expected names based on your actual Account dimension structure
				expectedLevels := map[string]bool{
					"level000": false,
					"level001": false,
					"level002": false,
				}

				for _, name := range levelNames {
					if _, exists := expectedLevels[name]; exists {
						expectedLevels[name] = true
					}
				}

				for level, found := range expectedLevels {
					if !found {
						return fmt.Errorf("expected level '%s' not found in level names", level)
					}
				}

				// Check that level names are in the correct order (top to bottom)
				for i := 1; i < len(levelNames); i++ {
					if levelNames[i-1] == "Line Item" && levelNames[i] == "Total" {
						return fmt.Errorf("level names are not in the correct order: %v", levelNames)
					}
				}

				return nil
			},
		},
		{
			name:          "Get level names from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			levelNames, err := tm1ServiceT.ElementService.GetLevelNames(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetLevelNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(levelNames); err != nil {
					t.Errorf("ElementService.GetLevelNames() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_GetLevelsCount(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(count int) error
	}{
		{
			name:          "Get levels count from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(count int) error {
				if count <= 0 {
					return fmt.Errorf("expected number of levels to be greater than 0, got %d", count)
				}
				return nil
			},
		},
		{
			name:          "Get levels count from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.ElementService.GetLevelsCount(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetLevelsCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(count); err != nil {
					t.Errorf("ElementService.GetLevelsCount() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_AttributeCubeExists(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string

		want    bool
		wantErr bool
	}{
		{
			name:          "Check attribute cube exists for Account dimension",
			dimensionName: "Account",

			want:    true,
			wantErr: false,
		},
		{
			name:          "Check attribute cube for non-existent dimension",
			dimensionName: "NonExistentDimension",

			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := tm1ServiceT.ElementService.AttributeCubeExists(tt.dimensionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.AttributeCubeExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exists != tt.want {
				t.Errorf("ElementService.AttributeCubeExists() = %v, want %v", exists, tt.want)
			}
		})
	}
}

func TestElementService_GetElementAttributes(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		wantErr       bool
		checkFunc     func(attributes []tm1go.ElementAttribute) error
	}{
		{
			name:          "Get element attributes from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			wantErr:       false,
			checkFunc: func(attributes []tm1go.ElementAttribute) error {
				if len(attributes) == 0 {
					return fmt.Errorf("expected at least one attribute, got none")
				}

				// Check for expected attributes
				expectedAttributes := map[string]bool{
					"Description":  false,
					"Account Type": false,
					"Operator":     false,
				}

				for _, attr := range attributes {
					if _, exists := expectedAttributes[attr.Name]; exists {
						expectedAttributes[attr.Name] = true
					}

					// Check that each attribute has a non-empty name and a valid type
					if attr.Name == "" {
						return fmt.Errorf("found an attribute with empty name")
					}
					if attr.Type == "" {
						return fmt.Errorf("attribute %s has an empty type", attr.Name)
					}
				}

				for attrName, found := range expectedAttributes {
					if !found {
						return fmt.Errorf("expected attribute '%s' not found", attrName)
					}
				}

				return nil
			},
		},
		{
			name:          "Get element attributes from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			wantErr:       true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes, err := tm1ServiceT.ElementService.GetElementAttributes(tt.dimensionName, tt.hierarchyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.GetElementAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(attributes); err != nil {
					t.Errorf("ElementService.GetElementAttributes() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_CreateElementAttribute(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		attribute     tm1go.ElementAttribute
		wantErr       bool
	}{
		{
			name:          "Create new attribute for Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			attribute: tm1go.ElementAttribute{
				Name: "TestAttribute",
				Type: "String",
			},
			wantErr: false,
		},
		{
			name:          "Create attribute for non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			attribute: tm1go.ElementAttribute{
				Name: "TestAttribute",
				Type: "String",
			},
			wantErr: true,
		},
		{
			name:          "Create attribute with invalid type",
			dimensionName: "Account",
			hierarchyName: "Account",
			attribute: tm1go.ElementAttribute{
				Name: "InvalidAttribute",
				Type: "InvalidType",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ElementService.CreateElementAttribute(tt.dimensionName, tt.hierarchyName, &tt.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.CreateElementAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify that the attribute was created
				attributes, err := tm1ServiceT.ElementService.GetElementAttributes(tt.dimensionName, tt.hierarchyName)
				if err != nil {
					t.Errorf("Failed to get attributes after creation: %v", err)
					return
				}

				found := false
				for _, attr := range attributes {
					if attr.Name == tt.attribute.Name && attr.Type == tt.attribute.Type {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Created attribute %s not found in dimension attributes", tt.attribute.Name)
				}

				// Clean up: Delete the created attribute
				err = tm1ServiceT.ElementService.DeleteElementAttribute(tt.dimensionName, tt.hierarchyName, tt.attribute.Name)
				if err != nil {
					t.Errorf("Failed to delete test attribute: %v", err)
				}
			}
		})
	}
}

func TestElementService_DeleteElementAttribute(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		attributeName string
		wantErr       bool
	}{
		{
			name:          "Delete attribute from Account dimension",
			dimensionName: "Account",
			hierarchyName: "Account",
			attributeName: "TestDeleteAttribute",
			wantErr:       false,
		},
		{
			name:          "Delete non-existent attribute",
			dimensionName: "Account",
			hierarchyName: "Account",
			attributeName: "NonExistentAttribute",
			wantErr:       true,
		},
		{
			name:          "Delete attribute from non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			attributeName: "TestDeleteAttribute",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test attribute if we're not testing a non-existent attribute or dimension
			if tt.dimensionName == "Account" && tt.attributeName != "NonExistentAttribute" {
				testAttr := tm1go.ElementAttribute{
					Name: tt.attributeName,
					Type: "String",
				}
				err := tm1ServiceT.ElementService.CreateElementAttribute(tt.dimensionName, tt.hierarchyName, &testAttr)
				if err != nil {
					t.Fatalf("Failed to create test attribute: %v", err)
				}
			}

			// Perform the delete operation
			err := tm1ServiceT.ElementService.DeleteElementAttribute(tt.dimensionName, tt.hierarchyName, tt.attributeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.DeleteElementAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If the delete should have succeeded, verify that the attribute is gone
			if !tt.wantErr {
				attributes, err := tm1ServiceT.ElementService.GetElementAttributes(tt.dimensionName, tt.hierarchyName)
				if err != nil {
					t.Errorf("Failed to get attributes after deletion: %v", err)
					return
				}

				for _, attr := range attributes {
					if attr.Name == tt.attributeName {
						t.Errorf("Attribute %s still exists after deletion", tt.attributeName)
						return
					}
				}
			}
		})
	}
}

func TestElementService_ExecuteSetMDX(t *testing.T) {
	tests := []struct {
		name          string
		dimensionName string
		hierarchyName string
		mdx           string
		wantErr       bool
	}{
		{
			name:          "Execute MDX to get all Account members",
			dimensionName: "Account",
			hierarchyName: "Account",
			mdx:           "[Account].[Account].MEMBERS",
			wantErr:       false,
		},
		{
			name:          "Execute MDX on non-existent dimension",
			dimensionName: "NonExistentDimension",
			hierarchyName: "NonExistentHierarchy",
			mdx:           "[NonExistentDimension].[NonExistentHierarchy].MEMBERS",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tm1go.MDXExecuteParams{
				MDX: tt.mdx,
			}
			cellset, err := tm1ServiceT.ElementService.ExecuteSetMDX(params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.ExecuteSetMDX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(cellset.Tuples) == 0 {
					t.Errorf("ElementService.ExecuteSetMDX() check failed: %v", err)
				}
			}
		})
	}
}

func TestElementService_AddElements(t *testing.T) {
	// Setup: Create a test dimension
	testDimName := "TestDimension"
	testHierName := "TestDimension"

	testDimStruct := tm1go.NewDimension(testDimName)

	setupDimension := func() error {
		return tm1ServiceT.DimensionService.Create(testDimStruct)
	}

	cleanupDimension := func() {
		tm1ServiceT.DimensionService.Delete(testDimName)
	}

	if err := setupDimension(); err != nil {
		t.Fatalf("Failed to set up test dimension: %v", err)
	}
	defer cleanupDimension()

	tests := []struct {
		name     string
		elements []tm1go.Element
		wantErr  bool
	}{
		{
			name: "Add multiple elements",
			elements: []tm1go.Element{
				{Name: "Total", Type: "Consolidated"},
				{Name: "North", Type: "Consolidated"},
				{Name: "South", Type: "Consolidated"},
				{Name: "NY", Type: "Numeric"},
				{Name: "NJ", Type: "Numeric"},
				{Name: "FL", Type: "Numeric"},
				{Name: "GA", Type: "Numeric"},
			},
			wantErr: false,
		},
		{
			name: "Add element with existing name",
			elements: []tm1go.Element{
				{Name: "Total", Type: "Numeric"}, // This should cause an error as "Total" already exists
			},
			wantErr: true,
		},
		{
			name: "Add element with invalid type",
			elements: []tm1go.Element{
				{Name: "Invalid", Type: "InvalidType"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ElementService.AddElements(testDimName, testHierName, tt.elements)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementService.AddElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify that the elements were added
				elements, err := tm1ServiceT.ElementService.GetElements(testDimName, testHierName)
				if err != nil {
					t.Errorf("Failed to get elements after addition: %v", err)
					return
				}

				// Check if all added elements exist
				for _, addedElem := range tt.elements {
					found := false
					for _, elem := range elements {
						if elem.Name == addedElem.Name && elem.Type == addedElem.Type {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Added element %s of type %s not found in dimension", addedElem.Name, addedElem.Type)
					}
				}

				// Check if the number of elements matches
				if len(elements) != len(tt.elements) {
					t.Errorf("Number of elements after addition (%d) does not match number of added elements (%d)", len(elements), len(tt.elements))
				}
			}
		})
	}
}
