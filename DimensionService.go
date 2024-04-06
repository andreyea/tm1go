package tm1go

import (
	"encoding/json"
	"fmt"
)

type DimensionService struct {
	rest      *RestService
	object    *ObjectService
	hierarchy *HierarchyService
}

// NewDimensionService creates a new instance of DimensionService
func NewDimensionService(rest *RestService, object *ObjectService, hierarchy *HierarchyService) *DimensionService {
	return &DimensionService{rest: rest, object: object, hierarchy: hierarchy}
}

// Create a new dimension
func (ds *DimensionService) Create(dimension *Dimension) error {
	url := "/Dimensions"
	dimensionBody, err := dimension.getBody(false)
	if err != nil {
		return err
	}
	_, err = ds.rest.POST(url, dimensionBody, nil, 0, nil)
	return err
}

// Get retrieves a dimension
func (ds *DimensionService) Get(dimensionName string) (*Dimension, error) {
	url := fmt.Sprintf("/Dimensions('%v')?$expand=Hierarchies($expand=*)", dimensionName)
	response, err := ds.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	dimension := &Dimension{}
	err = json.NewDecoder(response.Body).Decode(dimension)
	if err != nil {
		return nil, err
	}
	return dimension, nil
}

// Delete a dimension
func (ds *DimensionService) Delete(dimensionName string) error {
	url := fmt.Sprintf("/Dimensions('%v')", dimensionName)
	_, err := ds.rest.DELETE(url, nil, 0, nil)
	return err
}

// Exists checks if a dimension exists
func (ds *DimensionService) Exists(dimensionName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%v')", dimensionName)
	return ds.object.Exists(url)
}

// GetAllNames retrieves all dimension names
func (dx *DimensionService) GetAllNames() ([]string, error) {
	url := "/Dimensions?$select=Name"
	response, err := dx.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Dimension]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(result.Value))
	for _, dimension := range result.Value {
		names = append(names, dimension.Name)
	}
	return names, nil
}

// GetNumberOfDimensions retrieves the number of dimensions
func (ds *DimensionService) GetNumberOfDimensions(skipControlDims bool) (int, error) {
	url := "/Dimensions/$count"
	if skipControlDims {
		url += "/ModelDimensions()?$select=Name&$top=0&$count"
	}
	response, err := ds.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	if skipControlDims {
		result := ValueArray[Dimension]{}
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			return 0, err
		}
	} else {
		err = json.NewDecoder(response.Body).Decode(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

//execute_mdx
//create_element_attributes_through_ti
//uses_alternate_hierarchies
//update
//update_or_create
