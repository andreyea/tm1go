package tm1go

import (
	"encoding/json"
	"fmt"
)

type HierarchyService struct {
	rest   *RestService
	object *ObjectService
}

// NewHierarchyService creates a new instance of HierarchyService
func NewHierarchyService(rest *RestService, object *ObjectService) *HierarchyService {
	return &HierarchyService{rest: rest, object: object}
}

// Create creates a new hierarchy in the dimension
func (hs *HierarchyService) Create(hierarchy *Hierarchy) error {
	url := "/Dimensions('" + hierarchy.Dimension.Name + "')/Hierarchies"
	hierarchyBody, err := hierarchy.getBody(true)
	if err != nil {
		return err
	}
	hierarchyBodyJson, err := json.Marshal(hierarchyBody)
	if err != nil {
		return err
	}
	_, err = hs.rest.POST(url, string(hierarchyBodyJson), nil, 0, nil)
	return err
}

// Get retrieves a hierarchy from the dimension
func (hs *HierarchyService) Get(dimensionName string, hierarchyName string) (*Hierarchy, error) {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')?$expand=Edges,Elements,ElementAttributes,Subsets,DefaultMember", dimensionName, hierarchyName)
	response, err := hs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	hierarchy := &Hierarchy{}
	err = json.NewDecoder(response.Body).Decode(hierarchy)
	if err != nil {
		return nil, err
	}
	return hierarchy, nil
}

// GetAll retrieves all hierarchies from the dimension
func (hs *HierarchyService) GetAllNames(dimensionName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies?$select=Name", dimensionName)
	response, err := hs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Hierarchy]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(result.Value))
	for _, hierarchy := range result.Value {
		names = append(names, hierarchy.Name)
	}
	return names, nil
}

// Exists checks if the hierarchy exists in the dimension
func (hs *HierarchyService) Exists(dimensionName string, hierarchyName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')", dimensionName, hierarchyName)
	return hs.object.Exists(url)
}

// Update updates the hierarchy in the dimension
func (hs *HierarchyService) Delete(dimensionName string, hierarchyName string) error {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')", dimensionName, hierarchyName)
	_, err := hs.rest.DELETE(url, nil, 0, nil)
	return err
}

// GetDefaultMember retrieves the default member in the hierarchy in the dimension
func (hs *HierarchyService) GetDefaultMember(dimensionName string, hierarchyName string) (*Member, error) {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')/DefaultMember", dimensionName, hierarchyName)
	response, err := hs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	member := &Member{}
	err = json.NewDecoder(response.Body).Decode(member)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// UpdateDefaultMemberViaApi updates the default member in the hierarchy in the dimension
func (hs *HierarchyService) UpdateDefaultMemberViaApi(dimensionName string, hierarchyName string, memberName string) error {
	if !IsV1GreaterOrEqualToV2(hs.rest.version, "12.0.0") {
		return fmt.Errorf("UpdateDefaultMemberViaApi is only supported in TM1 12.0.0 and later")
	}
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')", dimensionName, hierarchyName)
	payload := `{"DefaultMemberName":"` + memberName + `"}`
	_, err := hs.rest.PATCH(url, payload, nil, 0, nil)
	return err
}

// UpdateDefaultMember updates the default member in the hierarchy in the dimension
func (hs *HierarchyService) UpdateDefaultMember(dimensionName string, hierarchyName string, memberName string) error {
	if IsV1GreaterOrEqualToV2(hs.rest.version, "12.0.0") {
		return hs.UpdateDefaultMemberViaApi(dimensionName, hierarchyName, memberName)
	} else {
		return fmt.Errorf("UpdateDefaultMember is not imlemented for TM1 versions earlier than 12.0.0")
	}
}

// RemoveAllEdges removes all edges from the hierarchy in the dimension
func (hs *HierarchyService) RemoveAllEdges(dimensionName string, hierarchyName string) error {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')", dimensionName, hierarchyName)
	body := `{"Edges": []}`
	_, err := hs.rest.PATCH(url, body, nil, 0, nil)
	return err
}

// IsBalanced checks if the hierarchy in the dimension is balanced
func (hs *HierarchyService) IsBalanced(dimensionName string, hierarchyName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%v')/Hierarchies('%v')/Structure/$value", dimensionName, hierarchyName)
	response, err := hs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	structure := -1
	err = json.NewDecoder(response.Body).Decode(&structure)
	if err != nil {
		return false, err
	}
	if structure == 0 {
		return true, nil
	} else if structure == 2 {
		return false, nil
	} else {
		return false, fmt.Errorf("unknown structure value: %v", structure)
	}
}

// func(hs *HierarchyService) UpdateOrCreateHierarchyFromDataframe
// func(hs *HierarchyService) GetDimensionService
// func(hs *HierarchyService) GetCellService
// func(hs *HierarchyService) AttributeTypeFromCode ​
// ValidateEdges
// ValidateAliasUniqueness
// func(hs *HierarchyService) RemoveEdgesUnderConsolidation
// func(hs *HierarchyService) AddEdges
// func(hs *HierarchyService) AddElements
// func(hs *HierarchyService) AddElementAttributes
// func(hs *HierarchyService) UpdateDefaultMemberViaPropsCube
// func(hs *HierarchyService) Update
// func(hs *HierarchyService) UpdateOrCreate
// func(hs *HierarchyService) GetHierarchySummary
// func(hs *HierarchyService) UpdateElementAttributes
