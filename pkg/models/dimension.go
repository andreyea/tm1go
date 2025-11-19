package models

import (
	"encoding/json"
	"fmt"
	"net/url"
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
		ParentName:    parentName,
		ComponentName: childName,
		Weight:        weight,
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
	ParentName    string  `json:"ParentName"`
	ComponentName string  `json:"ComponentName"`
	Weight        float64 `json:"Weight,omitempty"`
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

// Subset represents a subset within a hierarchy (static or dynamic)
type Subset struct {
	Name          string   `json:"Name"`
	DimensionName string   `json:"-"` // Not part of JSON, used internally
	HierarchyName string   `json:"-"` // Not part of JSON, used internally
	Alias         string   `json:"Alias,omitempty"`
	Expression    string   `json:"Expression,omitempty"`
	Elements      []string `json:"-"` // For static subsets, handled separately
}

// NewSubset creates a new Subset instance
func NewSubset(dimensionName, hierarchyName, subsetName string) *Subset {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}
	return &Subset{
		Name:          subsetName,
		DimensionName: dimensionName,
		HierarchyName: hierarchyName,
		Elements:      make([]string, 0),
	}
}

// NewStaticSubset creates a new static subset with elements
func NewStaticSubset(dimensionName, hierarchyName, subsetName string, elements []string) *Subset {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}
	return &Subset{
		Name:          subsetName,
		DimensionName: dimensionName,
		HierarchyName: hierarchyName,
		Elements:      elements,
	}
}

// NewDynamicSubset creates a new dynamic subset with MDX expression
func NewDynamicSubset(dimensionName, hierarchyName, subsetName, expression string) *Subset {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}
	return &Subset{
		Name:          subsetName,
		DimensionName: dimensionName,
		HierarchyName: hierarchyName,
		Expression:    expression,
	}
}

// AddElements adds elements to a static subset
func (s *Subset) AddElements(elements ...string) {
	s.Elements = append(s.Elements, elements...)
	s.Expression = ""
}

// SetExpression sets the MDX expression for a dynamic subset
func (s *Subset) SetExpression(expression string) {
	s.Expression = expression
	s.Elements = nil
}

// Type returns "dynamic" or "static" based on whether the subset has an expression
func (s *Subset) Type() string {
	if s.IsDynamic() {
		return "dynamic"
	}
	return "static"
}

// IsDynamic returns true if the subset is dynamic (has an MDX expression)
func (s *Subset) IsDynamic() bool {
	return s.Expression != ""
}

// IsStatic returns true if the subset is static (no MDX expression)
func (s *Subset) IsStatic() bool {
	return !s.IsDynamic()
}

// Body returns the JSON representation for API requests
func (s *Subset) Body() (string, error) {
	bodyDict, err := s.BodyAsDict()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(bodyDict)
	if err != nil {
		return "", fmt.Errorf("failed to marshal subset: %w", err)
	}

	return string(data), nil
}

// BodyAsDict returns the body as a map for API requests
func (s *Subset) BodyAsDict() (map[string]interface{}, error) {
	if s.IsDynamic() {
		return s.constructBodyDynamic(), nil
	}
	return s.constructBodyStatic(), nil
}

// constructBodyDynamic constructs the body for a dynamic subset
func (s *Subset) constructBodyDynamic() map[string]interface{} {
	body := make(map[string]interface{})
	body["Name"] = s.Name

	if s.Alias != "" {
		body["Alias"] = s.Alias
	}

	body["Hierarchy@odata.bind"] = fmt.Sprintf(
		"Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(s.DimensionName),
		url.PathEscape(s.HierarchyName))

	body["Expression"] = s.Expression

	return body
}

// constructBodyStatic constructs the body for a static subset
func (s *Subset) constructBodyStatic() map[string]interface{} {
	body := make(map[string]interface{})
	body["Name"] = s.Name

	if s.Alias != "" {
		body["Alias"] = s.Alias
	}

	body["Hierarchy@odata.bind"] = fmt.Sprintf(
		"Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(s.DimensionName),
		url.PathEscape(s.HierarchyName))

	if len(s.Elements) > 0 {
		elementBindings := make([]string, len(s.Elements))
		for i, elem := range s.Elements {
			elementBindings[i] = fmt.Sprintf(
				"Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
				url.PathEscape(s.DimensionName),
				url.PathEscape(s.HierarchyName),
				url.PathEscape(elem))
		}
		body["Elements@odata.bind"] = elementBindings
	}

	return body
}

// SubsetFromJSON creates a Subset from JSON response
func SubsetFromJSON(data []byte) (*Subset, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subset: %w", err)
	}

	return SubsetFromDict(raw)
}

// SubsetFromDict creates a Subset from a dictionary/map
func SubsetFromDict(dict map[string]interface{}) (*Subset, error) {
	subset := &Subset{}

	// Extract name
	if name, ok := dict["Name"].(string); ok {
		subset.Name = name
	}

	// Extract alias
	if alias, ok := dict["Alias"].(string); ok {
		subset.Alias = alias
	}

	// Extract expression for dynamic subsets
	if expr, ok := dict["Expression"].(string); ok && expr != "" {
		subset.Expression = expr
	}

	// Extract dimension and hierarchy from expanded Hierarchy object or UniqueName
	if hierarchy, ok := dict["Hierarchy"].(map[string]interface{}); ok {
		if hierName, ok := hierarchy["Name"].(string); ok {
			subset.HierarchyName = hierName
		}
		if dim, ok := hierarchy["Dimension"].(map[string]interface{}); ok {
			if dimName, ok := dim["Name"].(string); ok {
				subset.DimensionName = dimName
			}
		}
	} else if uniqueName, ok := dict["UniqueName"].(string); ok {
		// Parse UniqueName: [DimensionName].[HierarchyName].[SubsetName]
		// Extract dimension name from UniqueName
		if len(uniqueName) > 0 && uniqueName[0] == '[' {
			endIdx := 1
			for endIdx < len(uniqueName) && uniqueName[endIdx] != ']' {
				endIdx++
			}
			if endIdx < len(uniqueName) {
				subset.DimensionName = uniqueName[1:endIdx]
			}
		}
		// If hierarchy name not extracted from Hierarchy object, use dimension name
		if subset.HierarchyName == "" {
			subset.HierarchyName = subset.DimensionName
		}
	}

	// Extract elements for static subsets (only if no expression)
	if subset.Expression == "" {
		if elements, ok := dict["Elements"].([]interface{}); ok {
			subset.Elements = make([]string, len(elements))
			for i, elem := range elements {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					if name, ok := elemMap["Name"].(string); ok {
						subset.Elements[i] = name
					}
				}
			}
		}
	}

	// If hierarchy name is still empty, use dimension name
	if subset.HierarchyName == "" {
		subset.HierarchyName = subset.DimensionName
	}

	return subset, nil
}
