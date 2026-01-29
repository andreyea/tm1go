package tm1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CellService handles read and write operations to TM1 cubes
type CellService struct {
	rest *RestService
}

// NewCellService creates a new CellService instance
func NewCellService(rest *RestService) *CellService {
	return &CellService{
		rest: rest,
	}
}

// CellValue represents a single cell value with its properties
type CellValue struct {
	Value          interface{}            `json:"Value"`
	Ordinal        int                    `json:"Ordinal,omitempty"`
	RuleDerived    bool                   `json:"RuleDerived,omitempty"`
	Consolidated   bool                   `json:"Consolidated,omitempty"`
	Updateable     int                    `json:"Updateable,omitempty"`
	FormattedValue string                 `json:"FormattedValue,omitempty"`
	Properties     map[string]interface{} `json:"-"`
}

// Cellset represents a complete cellset response
type Cellset struct {
	ID      string                            `json:"ID,omitempty"`
	Cube    *Cube                             `json:"Cube,omitempty"`
	Axes    []Axis                            `json:"Axes,omitempty"`
	Cells   []Cell                            `json:"Cells,omitempty"`
	CellMap map[string]map[string]interface{} `json:"-"` // Coordinate tuple -> cell properties
}

// Axis represents an axis in a cellset
type Axis struct {
	Ordinal     int         `json:"Ordinal"`
	Cardinality int         `json:"Cardinality,omitempty"`
	Hierarchies []Hierarchy `json:"Hierarchies,omitempty"`
	Tuples      []Tuple     `json:"Tuples,omitempty"`
}

// Tuple represents a tuple on an axis
type Tuple struct {
	Ordinal int      `json:"Ordinal"`
	Members []Member `json:"Members,omitempty"`
}

// Member represents a member in a tuple
type Member struct {
	Name                string                `json:"Name"`
	UniqueName          string                `json:"UniqueName,omitempty"`
	Type                MemberType            `json:"Type,omitempty"`
	Ordinal             int                   `json:"Ordinal,omitempty"`
	IsPlaceholder       bool                  `json:"IsPlaceholder,omitempty"`
	Weight              float64               `json:"Weight,omitempty"`
	Attributes          Attributes            `json:"Attributes,omitempty"`
	Hierarchy           *Hierarchy            `json:"Hierarchy,omitempty"`
	Level               *Level                `json:"Level,omitempty"`
	Element             *Element              `json:"Element,omitempty"`
	Parent              *Member               `json:"Parent,omitempty"`
	Children            []Member              `json:"Children,omitempty"`
	LocalizedAttributes []LocalizedAttributes `json:"LocalizedAttributes,omitempty"`
	DimensionName       string                `json:"DimensionName,omitempty"`
	HierarchyName       string                `json:"HierarchyName,omitempty"`
}

// Cell represents a single cell in a cellset
type Cell struct {
	Ordinal             int            `json:"Ordinal"`
	Status              CellStatus     `json:"Status,omitempty"`
	Value               interface{}    `json:"Value,omitempty"`
	FormatString        string         `json:"FormatString,omitempty"`
	FormattedValue      string         `json:"FormattedValue,omitempty"`
	Updateable          int            `json:"Updateable,omitempty"`
	RuleDerived         bool           `json:"RuleDerived,omitempty"`
	Annotated           bool           `json:"Annotated,omitempty"`
	Consolidated        bool           `json:"Consolidated,omitempty"`
	NullIntersected     bool           `json:"NullIntersected,omitempty"`
	Language            int            `json:"Language,omitempty"`
	HasPicklist         bool           `json:"HasPicklist,omitempty"`
	PicklistValues      []string       `json:"PicklistValues,omitempty"`
	HasDrillthrough     bool           `json:"HasDrillthrough,omitempty"`
	DrillthroughScripts []Drillthrough `json:"DrillthroughScripts,omitempty"`
	Members             []Member       `json:"Members,omitempty"`
	Annotations         []Annotation   `json:"Annotations,omitempty"`
}

// Attributes is an open type key-value container used by TM1.
type Attributes map[string]interface{}

// CellStatus represents the status of a cell.
type CellStatus int

const (
	CellStatusNull  CellStatus = 0
	CellStatusData  CellStatus = 1
	CellStatusError CellStatus = 2
)

// MemberType represents the type of a member.
type MemberType int

const (
	MemberTypeUnknown MemberType = 0
	MemberTypeRegular MemberType = 1
	MemberTypeAll     MemberType = 2
	MemberTypeMeasure MemberType = 3
	MemberTypeFormula MemberType = 4
)

// ElementType represents the type of an element.
type ElementType int

const (
	ElementTypeNumeric      ElementType = 1
	ElementTypeString       ElementType = 2
	ElementTypeConsolidated ElementType = 3
)

// Cube represents a TM1 cube.
type Cube struct {
	Name              string     `json:"Name"`
	Rules             string     `json:"Rules,omitempty"`
	DrillthroughRules string     `json:"DrillthroughRules,omitempty"`
	LastSchemaUpdate  string     `json:"LastSchemaUpdate,omitempty"`
	LastDataUpdate    string     `json:"LastDataUpdate,omitempty"`
	Attributes        Attributes `json:"Attributes,omitempty"`
}

// Hierarchy represents a TM1 hierarchy.
type Hierarchy struct {
	Name        string     `json:"Name"`
	UniqueName  string     `json:"UniqueName,omitempty"`
	Cardinality int        `json:"Cardinality,omitempty"`
	Structure   int        `json:"Structure,omitempty"`
	Visible     bool       `json:"Visible,omitempty"`
	Attributes  Attributes `json:"Attributes,omitempty"`
}

// Level represents a TM1 level.
type Level struct {
	Number      int    `json:"Number"`
	Name        string `json:"Name"`
	UniqueName  string `json:"UniqueName,omitempty"`
	Cardinality int    `json:"Cardinality,omitempty"`
	Type        int    `json:"Type,omitempty"`
}

// Element represents a TM1 element.
type Element struct {
	Name       string      `json:"Name"`
	UniqueName string      `json:"UniqueName,omitempty"`
	Type       ElementType `json:"Type,omitempty"`
	Level      int         `json:"Level,omitempty"`
	Index      int         `json:"Index,omitempty"`
	Attributes Attributes  `json:"Attributes,omitempty"`
}

// LocalizedAttributes represents localized attributes for an object.
type LocalizedAttributes struct {
	LocaleID   string     `json:"LocaleID"`
	Attributes Attributes `json:"Attributes,omitempty"`
}

// Drillthrough represents a drillthrough script.
type Drillthrough struct {
	Name string `json:"Name"`
}

// Annotation represents a cell annotation.
type Annotation struct {
	ID            string `json:"ID"`
	Text          string `json:"Text,omitempty"`
	Creator       string `json:"Creator,omitempty"`
	Created       string `json:"Created,omitempty"`
	LastUpdatedBy string `json:"LastUpdatedBy,omitempty"`
	LastUpdated   string `json:"LastUpdated,omitempty"`
}

// GetValue returns a single cube value from specified coordinates
// elements can be a slice of element names in the correct dimension order
// dimensions should contain the dimension names in their natural order
func (cs *CellService) GetValue(ctx context.Context, cubeName string, elements []string, dimensions []string, sandboxName string) (interface{}, error) {
	if len(elements) == 0 {
		return nil, fmt.Errorf("elements cannot be empty")
	}

	if len(dimensions) == 0 {
		// If dimensions not provided, retrieve from cube
		var err error
		dimensions, err = cs.getDimensionNamesForCube(ctx, cubeName)
		if err != nil {
			return nil, fmt.Errorf("get dimensions: %w", err)
		}
	}

	if len(elements) != len(dimensions) {
		return nil, fmt.Errorf("elements count (%d) must match dimensions count (%d)", len(elements), len(dimensions))
	}

	// Build MDX query
	// SELECT {} ON ROWS, {} ON COLUMNS FROM [cube]
	// Only the last element is used as the MDX ON COLUMN statement
	mdxParts := make([]string, 0, len(elements))
	for i, elem := range elements {
		dim := dimensions[i]
		mdxParts = append(mdxParts, fmt.Sprintf("[%s].[%s].[%s]", dim, dim, elem))
	}

	var mdxRows, mdxColumns string
	if len(mdxParts) > 1 {
		mdxRows = strings.Join(mdxParts[:len(mdxParts)-1], "*")
		mdxColumns = mdxParts[len(mdxParts)-1]
	} else {
		mdxRows = "{}"
		mdxColumns = mdxParts[0]
	}

	mdx := fmt.Sprintf("SELECT %s ON ROWS, %s ON COLUMNS FROM [%s]", mdxRows, mdxColumns, cubeName)

	// Execute MDX
	cellset, err := cs.ExecuteMDX(ctx, mdx, nil, sandboxName)
	if err != nil {
		return nil, fmt.Errorf("execute mdx: %w", err)
	}

	// Extract first value
	for _, cell := range cellset.CellMap {
		if val, ok := cell["Value"]; ok {
			return val, nil
		}
	}

	return nil, fmt.Errorf("no value found in cellset")
}

// ExecuteMDX executes an MDX query and returns a cellset.
func (cs *CellService) ExecuteMDX(ctx context.Context, mdx string, cellProperties []string, sandboxName string) (*Cellset, error) {
	// Create cellset
	cellsetID, err := cs.CreateCellset(ctx, mdx, sandboxName)
	if err != nil {
		return nil, fmt.Errorf("create cellset: %w", err)
	}

	// Extract cellset and delete it
	cellset, err := cs.ExtractCellset(ctx, cellsetID, cellProperties, true, sandboxName)
	if err != nil {
		return nil, fmt.Errorf("extract cellset: %w", err)
	}

	return cellset, nil
}

// ExecuteView executes an existing cube view and returns a cellset.
func (cs *CellService) ExecuteView(ctx context.Context, cubeName, viewName string, private bool, cellProperties []string, sandboxName string) (*Cellset, error) {
	// Create cellset from view
	cellsetID, err := cs.CreateCellsetFromView(ctx, cubeName, viewName, private, sandboxName)
	if err != nil {
		return nil, fmt.Errorf("create cellset from view: %w", err)
	}

	// Extract cellset and delete it
	cellset, err := cs.ExtractCellset(ctx, cellsetID, cellProperties, true, sandboxName)
	if err != nil {
		return nil, fmt.Errorf("extract cellset: %w", err)
	}

	return cellset, nil
}

// WriteValue writes a single value to a cube at the specified coordinates
func (cs *CellService) WriteValue(ctx context.Context, cubeName string, elements []string, dimensions []string, value interface{}, sandboxName string) error {
	if len(elements) == 0 {
		return fmt.Errorf("elements cannot be empty")
	}

	if len(dimensions) == 0 {
		// If dimensions not provided, retrieve from cube
		var err error
		dimensions, err = cs.getDimensionNamesForCube(ctx, cubeName)
		if err != nil {
			return fmt.Errorf("get dimensions: %w", err)
		}
	}

	if len(elements) != len(dimensions) {
		return fmt.Errorf("elements count (%d) must match dimensions count (%d)", len(elements), len(dimensions))
	}

	// Build OData tuple
	tupleBindings := make([]string, 0, len(elements))
	for i, elem := range elements {
		dim := dimensions[i]
		hier := dim // Default hierarchy has same name as dimension
		tupleBindings = append(tupleBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
			url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
	}

	// Build request body
	body := map[string]interface{}{
		"Cells": []map[string]interface{}{
			{
				"Tuple@odata.bind": tupleBindings,
			},
		},
		"Value": value,
	}

	// Build URL
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Update", url.PathEscape(cubeName))
	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	// Execute request
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("post update: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// WriteValues writes multiple cell values to a cube using comma-separated coordinates.
// cells is a map where keys are coordinate tuples (comma-separated element names) and values are the cell values.
// dimensions should contain the dimension names in their natural order.
func (cs *CellService) WriteValues(ctx context.Context, cubeName string, cells map[string]interface{}, dimensions []string, sandboxName string) error {
	if len(cells) == 0 {
		return nil // Nothing to write
	}

	coords := make([][]string, 0, len(cells))
	values := make([]interface{}, 0, len(cells))

	for coordKey, value := range cells {
		// Parse coordinate key (comma-separated format)
		elements := strings.Split(coordKey, ",")
		for i := range elements {
			elements[i] = strings.TrimSpace(elements[i])
		}
		coords = append(coords, elements)
		values = append(values, value)
	}

	return cs.WriteValuesByCoords(ctx, cubeName, coords, values, dimensions, sandboxName)
}

// WriteValuesByCoords writes multiple cell values to a cube using explicit coordinates.
// coords is a slice of element tuples (one tuple per cell), and values is the matching list of values.
// dimensions should contain the dimension names in their natural order.
func (cs *CellService) WriteValuesByCoords(ctx context.Context, cubeName string, coords [][]string, values []interface{}, dimensions []string, sandboxName string) error {
	if len(coords) == 0 {
		return nil
	}

	if len(coords) != len(values) {
		return fmt.Errorf("coords count (%d) must match values count (%d)", len(coords), len(values))
	}

	if len(dimensions) == 0 {
		// If dimensions not provided, retrieve from cube
		var err error
		dimensions, err = cs.getDimensionNamesForCube(ctx, cubeName)
		if err != nil {
			return fmt.Errorf("get dimensions: %w", err)
		}
	}

	// Build an array of cell updates
	cellUpdates := make([]map[string]interface{}, 0, len(coords))

	for i, elements := range coords {
		if len(elements) != len(dimensions) {
			return fmt.Errorf("coordinate at index %d has %d elements but expected %d dimensions", i, len(elements), len(dimensions))
		}

		// Build tuple bindings for this cell
		tupleBindings := make([]string, 0, len(elements))
		for j, elem := range elements {
			dim := dimensions[j]
			hier := dim // Default hierarchy
			tupleBindings = append(tupleBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
				url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
		}

		cellUpdate := map[string]interface{}{
			"Cells": []map[string]interface{}{
				{
					"Tuple@odata.bind": tupleBindings,
				},
			},
			"Value": values[i],
		}
		cellUpdates = append(cellUpdates, cellUpdate)
	}

	// Build URL
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Update", url.PathEscape(cubeName))
	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	// Execute each update (could be optimized with batch requests)
	for _, cellUpdate := range cellUpdates {
		payload, err := json.Marshal(cellUpdate)
		if err != nil {
			return fmt.Errorf("marshal cell update: %w", err)
		}

		resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
		if err != nil {
			return fmt.Errorf("post cell update: %w", err)
		}
		resp.Body.Close()
	}

	return nil
}

// CreateCellset creates a cellset from an MDX query
func (cs *CellService) CreateCellset(ctx context.Context, mdx string, sandboxName string) (string, error) {
	endpoint := "/ExecuteMDX"
	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	body := map[string]interface{}{
		"MDX": mdx,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("post execute mdx: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"ID"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.ID, nil
}

// CreateCellsetFromView creates a cellset from a cube view
func (cs *CellService) CreateCellsetFromView(ctx context.Context, cubeName, viewName string, private bool, sandboxName string) (string, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s('%s')/tm1.Execute",
		url.PathEscape(cubeName), viewType, url.PathEscape(viewName))

	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	var result struct {
		ID string `json:"ID"`
	}

	err := cs.rest.JSON(ctx, http.MethodPost, endpoint, nil, &result)
	if err != nil {
		return "", fmt.Errorf("create cellset from view: %w", err)
	}

	return result.ID, nil
}

// ExtractCellset extracts cell data from a cellset and returns the cellset payload.
func (cs *CellService) ExtractCellset(ctx context.Context, cellsetID string, cellProperties []string, deleteCellset bool, sandboxName string) (*Cellset, error) {
	// Build query parameters
	selectClause := "Ordinal,Value"
	if len(cellProperties) > 0 {
		selectClause = strings.Join(cellProperties, ",")
	}

	endpoint := fmt.Sprintf("/Cellsets('%s')?$expand=Axes($expand=Tuples($expand=Members($select=Name,UniqueName,Ordinal))),Cells($select=%s;$expand=Members($select=Name,UniqueName))",
		cellsetID, selectClause)

	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	// Get cellset data
	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get cellset: %w", err)
	}
	defer resp.Body.Close()

	cellsetData := &Cellset{}
	if err := json.NewDecoder(resp.Body).Decode(cellsetData); err != nil {
		return nil, fmt.Errorf("decode cellset: %w", err)
	}

	// Delete cellset if requested
	if deleteCellset {
		defer cs.DeleteCellset(ctx, cellsetID, sandboxName)
	}

	// Build result map: coordinate tuple -> cell properties
	cellsetData.CellMap = make(map[string]map[string]interface{})

	// Get cardinality of each axis for ordinal calculation
	axisCardinalities := make([]int, len(cellsetData.Axes))
	for i, axis := range cellsetData.Axes {
		axisCardinalities[i] = len(axis.Tuples)
	}

	// Process each cell
	for _, cell := range cellsetData.Cells {
		// Calculate coordinates from ordinal
		coords := cs.ordinalToCoordinates(cell.Ordinal, axisCardinalities)

		// Build coordinate key from member names
		coordParts := make([]string, 0)
		if len(cell.Members) > 0 {
			for memberIdx, member := range cell.Members {
				coordPart := member.Name
				if coordPart == "" {
					coordPart = member.UniqueName
				}
				if coordPart == "" {
					coordPart = fmt.Sprintf("Member%d", memberIdx)
				}
				coordParts = append(coordParts, coordPart)
			}
		} else {
			for axisIdx, tupleIdx := range coords {
				if axisIdx < len(cellsetData.Axes) && tupleIdx < len(cellsetData.Axes[axisIdx].Tuples) {
					tuple := cellsetData.Axes[axisIdx].Tuples[tupleIdx]
					for memberIdx, member := range tuple.Members {
						coordPart := member.Name
						if coordPart == "" {
							coordPart = member.UniqueName
						}
						if coordPart == "" {
							coordPart = fmt.Sprintf("Axis%dTuple%dMember%d", axisIdx, tupleIdx, memberIdx)
						}
						coordParts = append(coordParts, coordPart)
					}
				}
			}
		}
		coordKey := strings.Join(coordParts, ",")

		// Build cell properties map
		value := cell.Value
		if value == nil {
			value = 0
		}
		cellProps := map[string]interface{}{
			"Value":   value,
			"Ordinal": cell.Ordinal,
		}

		if cell.FormattedValue != "" {
			cellProps["FormattedValue"] = cell.FormattedValue
		}
		if len(cellProperties) == 0 || contains(cellProperties, "RuleDerived") {
			cellProps["RuleDerived"] = cell.RuleDerived
		}
		if len(cellProperties) == 0 || contains(cellProperties, "Consolidated") {
			cellProps["Consolidated"] = cell.Consolidated
		}
		if len(cellProperties) == 0 || contains(cellProperties, "Updateable") {
			cellProps["Updateable"] = cell.Updateable
		}

		cellsetData.CellMap[coordKey] = cellProps
	}

	return cellsetData, nil
}

// DeleteCellset deletes a cellset
func (cs *CellService) DeleteCellset(ctx context.Context, cellsetID string, sandboxName string) error {
	endpoint := fmt.Sprintf("/Cellsets('%s')", cellsetID)
	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	resp, err := cs.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("delete cellset: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// getDimensionNamesForCube retrieves dimension names for a cube (excluding sandbox dimension)
func (cs *CellService) getDimensionNamesForCube(ctx context.Context, cubeName string) ([]string, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/Dimensions?$select=Name", url.PathEscape(cubeName))

	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get dimensions: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode dimensions: %w", err)
	}

	dimensions := make([]string, 0, len(result.Value))
	for _, dim := range result.Value {
		// Skip sandbox dimension
		if !strings.HasPrefix(dim.Name, "Sandboxes") {
			dimensions = append(dimensions, dim.Name)
		}
	}

	return dimensions, nil
}

// ordinalToCoordinates converts a cell ordinal to axis coordinates
func (cs *CellService) ordinalToCoordinates(ordinal int, axisCardinalities []int) []int {
	if len(axisCardinalities) == 0 {
		return []int{}
	}

	coords := make([]int, len(axisCardinalities))
	remaining := ordinal

	for i := len(axisCardinalities) - 1; i >= 0; i-- {
		if axisCardinalities[i] > 0 {
			coords[i] = remaining % axisCardinalities[i]
			remaining = remaining / axisCardinalities[i]
		}
	}

	return coords
}

// addSandboxParam adds sandbox parameter to URL
func addSandboxParam(endpoint string, sandboxName string) string {
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s!sandbox=%s", endpoint, separator, url.QueryEscape(sandboxName))
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}
