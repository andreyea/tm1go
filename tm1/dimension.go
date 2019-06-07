package tm1

import (
	"encoding/json"
	"fmt"
)

//Hierarchy describes a hierarchy of a dimension
type Hierarchy struct {
	OdataEtag   string    `json:"@odata.etag"`
	Name        string    `json:"Name"`
	UniqueName  string    `json:"UniqueName"`
	Cardinality int       `json:"Cardinality"`
	Elements    []Element `json:"Elements"`
	Structure   int       `json:"Structure"`
	Visible     bool      `json:"Visible"`
}

//DimensionsResponse
type DimensionsResponse struct {
	OdataContext string      `json:"@odata.context"`
	Value        []Dimension `json:"value"`
}

//Dimension is describing a single dimension
type Dimension struct {
	OdataEtag              string      `json:"@odata.etag"`
	Name                   string      `json:"Name"`
	Hierarchies            []Hierarchy `json:"Hierarchies"`
	UniqueName             string      `json:"UniqueName"`
	AllLeavesHierarchyName string      `json:"AllLeavesHierarchyName"`
}

//Element of a hierarchy
type Element struct {
	OdataContext string            `json:"@odata.context"`
	OdataEtag    string            `json:"@odata.etag"`
	Name         string            `json:"Name"`
	UniqueName   string            `json:"UniqueName"`
	Type         string            `json:"Type"`
	Level        int               `json:"Level"`
	Index        int               `json:"Index"`
	Attributes   map[string]string `json:"Attributes"`
}

//DimensionCreate creates new dimension
func (s Tm1Session) DimensionCreate(dim Dimension) error {

	p1, _ := json.Marshal(dim)
	payload := string(p1)
	fmt.Println(payload)
	_, err := s.Tm1SendHttpRequest("POST", "/Dimensions", payload)

	if err != nil {
		return err
	}
	return nil
}

//GetDimensions gets all dimensions from tm1
func (s Tm1Session) GetDimensions() ([]Dimension, error) {

	dims := DimensionsResponse{}
	res, err := s.Tm1SendHttpRequest("GET", "/Dimensions", nil)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(res, &dims)
	return dims.Value, nil
}

//GetDimension brings a dimension from tm1
func (s Tm1Session) GetDimension(dimName string) (Dimension, error) {

	dim := Dimension{}
	res, err := s.Tm1SendHttpRequest("GET", "/Dimensions('"+dimName+"')", nil)

	if err != nil {
		return dim, err
	}

	json.Unmarshal(res, &dim)
	return dim, nil
}

//DimensionExists check if a cube exists in tm1
func (s Tm1Session) DimensionExists(dimName string) (bool, error) {
	_, err := s.GetDimension(dimName)
	if err != nil {
		return false, err
	}
	return true, nil
}

//DimensionCreate creates local dimension struct
func DimensionCreate(dName string) (Dimension, error) {
	dim := Dimension{
		Name: dName,
		Hierarchies: []Hierarchy{
			Hierarchy{
				Name:     dName,
				Elements: []Element{},
			},
		},
	}
	return dim, nil
}
