package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Cube represents a TM1 cube
type Cube struct {
	OdataContext      string      `json:"@odata.context,omitempty"`
	OdataEtag         string      `json:"@odata.etag,omitempty"`
	Name              string      `json:"Name"`
	Rules             string      `json:"Rules,omitempty"`
	DrillthroughRules string      `json:"DrillthroughRules,omitempty"`
	LastSchemaUpdate  time.Time   `json:"LastSchemaUpdate,omitempty"`
	LastDataUpdate    time.Time   `json:"LastDataUpdate,omitempty"`
	Dimensions        []Dimension `json:"Dimensions,omitempty"`
	Views             []View      `json:"Views,omitempty"`
	PrivateViews      []View      `json:"PrivateViews,omitempty"`

	// DimensionNames can be used to build cube creation payload without requiring Dimension objects
	DimensionNames []string `json:"-"`
}

// View represents a cube view
type View struct {
	Name string `json:"Name"`
}

// RuleSyntaxError represents a cube rule syntax error
type RuleSyntaxError struct {
	LineNumber int    `json:"LineNumber"`
	Message    string `json:"Message"`
}

// NewCube creates a new Cube instance
func NewCube(name string, dimensionNames ...string) *Cube {
	return &Cube{
		Name:           name,
		Dimensions:     make([]Dimension, 0),
		Views:          make([]View, 0),
		PrivateViews:   make([]View, 0),
		DimensionNames: append([]string{}, dimensionNames...),
	}
}

// AddDimension adds a Dimension object to the cube
func (c *Cube) AddDimension(dimension Dimension) {
	c.Dimensions = append(c.Dimensions, dimension)
}

// AddDimensionName adds a dimension name to the cube
func (c *Cube) AddDimensionName(name string) {
	c.DimensionNames = append(c.DimensionNames, name)
}

// DimensionNamesResolved returns the list of dimension names for the cube
func (c *Cube) DimensionNamesResolved() []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(c.DimensionNames)+len(c.Dimensions))

	for _, name := range c.DimensionNames {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}

	for _, dim := range c.Dimensions {
		if dim.Name == "" || seen[dim.Name] {
			continue
		}
		seen[dim.Name] = true
		result = append(result, dim.Name)
	}

	return result
}

// Body returns the JSON representation for cube create/update requests
func (c *Cube) Body() (string, error) {
	payload := map[string]interface{}{
		"Name": c.Name,
	}

	dimensionNames := c.DimensionNamesResolved()
	if len(dimensionNames) > 0 {
		bindings := make([]string, 0, len(dimensionNames))
		for _, name := range dimensionNames {
			bindings = append(bindings, fmt.Sprintf("Dimensions('%s')", name))
		}
		payload["Dimensions@odata.bind"] = bindings
	}

	if c.Rules != "" {
		payload["Rules"] = c.Rules
	}

	if c.DrillthroughRules != "" {
		payload["DrillthroughRules"] = c.DrillthroughRules
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
