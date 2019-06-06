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
