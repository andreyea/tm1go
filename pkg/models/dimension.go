package models

import (
	"encoding/json"
	"strings"
)

// Dimension represents a TM1 Dimension with its hierarchies
type Dimension struct {
	Name        string      `json:"Name"`
	Hierarchies []Hierarchy `json:"Hierarchies,omitempty"`
}

// NewDimension creates a new Dimension instance
func NewDimension(name string) *Dimension {
	return &Dimension{
		Name:        name,
		Hierarchies: make([]Hierarchy, 0),
	}
}

// AddHierarchy adds a hierarchy to the dimension
func (d *Dimension) AddHierarchy(hierarchy Hierarchy) {
	hierarchy.DimensionName = d.Name
	d.Hierarchies = append(d.Hierarchies, hierarchy)
}

// GetHierarchy returns a hierarchy by name (case-insensitive)
func (d *Dimension) GetHierarchy(name string) *Hierarchy {
	for i := range d.Hierarchies {
		if strings.EqualFold(d.Hierarchies[i].Name, name) {
			return &d.Hierarchies[i]
		}
	}
	return nil
}

// HierarchyNames returns the names of all hierarchies
func (d *Dimension) HierarchyNames() []string {
	names := make([]string, len(d.Hierarchies))
	for i, h := range d.Hierarchies {
		names[i] = h.Name
	}
	return names
}

// HasHierarchy checks if a hierarchy exists (case-insensitive)
func (d *Dimension) HasHierarchy(name string) bool {
	return d.GetHierarchy(name) != nil
}

// Body returns the JSON representation for API requests
func (d *Dimension) Body() string {
	data, _ := json.Marshal(d)
	return string(data)
}

// FromJSON creates a Dimension from JSON string
func DimensionFromJSON(jsonStr string) (*Dimension, error) {
	var dim Dimension
	err := json.Unmarshal([]byte(jsonStr), &dim)
	if err != nil {
		return nil, err
	}

	// Set DimensionName for each hierarchy
	for i := range dim.Hierarchies {
		dim.Hierarchies[i].DimensionName = dim.Name
	}

	return &dim, nil
}

// Hierarchy represents a hierarchy within a dimension
type Hierarchy struct {
	Name              string             `json:"Name"`
	DimensionName     string             `json:"-"` // Not part of JSON, used internally
	Elements          []Element          `json:"Elements,omitempty"`
	Edges             []Edge             `json:"Edges,omitempty"`
	ElementAttributes []ElementAttribute `json:"ElementAttributes,omitempty"`
	Subsets           []Subset           `json:"Subsets,omitempty"`
}

// NewHierarchy creates a new Hierarchy instance
func NewHierarchy(name, dimensionName string) *Hierarchy {
	return &Hierarchy{
		Name:              name,
		DimensionName:     dimensionName,
		Elements:          make([]Element, 0),
		Edges:             make([]Edge, 0),
		ElementAttributes: make([]ElementAttribute, 0),
		Subsets:           make([]Subset, 0),
	}
}

// AddElement adds an element to the hierarchy
func (h *Hierarchy) AddElement(element Element) {
	h.Elements = append(h.Elements, element)
}

// AddEdge adds a parent-child relationship
func (h *Hierarchy) AddEdge(parentName, childName string, weight float64) {
	h.Edges = append(h.Edges, Edge{
		ParentName: parentName,
		ChildName:  childName,
		Weight:     weight,
	})
}

// AddElementAttribute adds an element attribute
func (h *Hierarchy) AddElementAttribute(attr ElementAttribute) {
	h.ElementAttributes = append(h.ElementAttributes, attr)
}

// Element represents an element in a hierarchy
type Element struct {
	Name  string      `json:"Name"`
	Type  ElementType `json:"Type"`
	Index int         `json:"Index,omitempty"`
}

// ElementType represents the type of an element
type ElementType string

const (
	ElementTypeNumeric      ElementType = "Numeric"
	ElementTypeString       ElementType = "String"
	ElementTypeConsolidated ElementType = "Consolidated"
)

// Edge represents a parent-child relationship between elements
type Edge struct {
	ParentName string  `json:"ParentName"`
	ChildName  string  `json:"ComponentName"`
	Weight     float64 `json:"Weight,omitempty"`
}

// ElementAttribute represents an attribute of elements
type ElementAttribute struct {
	Name          string               `json:"Name"`
	AttributeType ElementAttributeType `json:"Type"`
}

// ElementAttributeType represents the type of an element attribute
type ElementAttributeType string

const (
	AttributeTypeNumeric ElementAttributeType = "Numeric"
	AttributeTypeString  ElementAttributeType = "String"
	AttributeTypeAlias   ElementAttributeType = "Alias"
)

// Subset represents a subset within a hierarchy
type Subset struct {
	Name          string `json:"Name"`
	Expression    string `json:"Expression,omitempty"`
	HierarchyName string `json:"-"`
	DimensionName string `json:"-"`
}
