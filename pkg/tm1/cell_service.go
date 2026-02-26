package tm1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
	"github.com/go-gota/gota/dataframe"
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
	Cube    *models.Cube                      `json:"Cube,omitempty"`
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

func (cs *CellStatus) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		*cs = CellStatus(intValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		switch strings.ToLower(strings.TrimSpace(stringValue)) {
		case "null":
			*cs = CellStatusNull
		case "data":
			*cs = CellStatusData
		case "error":
			*cs = CellStatusError
		default:
			return fmt.Errorf("invalid cell status: %q", stringValue)
		}
		return nil
	}

	return fmt.Errorf("invalid cell status payload: %s", string(data))
}

// MemberType represents the type of a member.
type MemberType int

const (
	MemberTypeUnknown MemberType = 0
	MemberTypeRegular MemberType = 1
	MemberTypeAll     MemberType = 2
	MemberTypeMeasure MemberType = 3
	MemberTypeFormula MemberType = 4
)

func (mt *MemberType) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		*mt = MemberType(intValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		switch strings.ToLower(strings.TrimSpace(stringValue)) {
		case "unknown":
			*mt = MemberTypeUnknown
		case "regular":
			*mt = MemberTypeRegular
		case "all":
			*mt = MemberTypeAll
		case "measure":
			*mt = MemberTypeMeasure
		case "formula":
			*mt = MemberTypeFormula
		default:
			return fmt.Errorf("invalid member type: %q", stringValue)
		}
		return nil
	}

	return fmt.Errorf("invalid member type payload: %s", string(data))
}

// ElementType represents the type of an element.
type ElementType int

const (
	ElementTypeNumeric      ElementType = 1
	ElementTypeString       ElementType = 2
	ElementTypeConsolidated ElementType = 3
)

func (et *ElementType) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		*et = ElementType(intValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		switch strings.ToLower(strings.TrimSpace(stringValue)) {
		case "numeric":
			*et = ElementTypeNumeric
		case "string":
			*et = ElementTypeString
		case "consolidated":
			*et = ElementTypeConsolidated
		default:
			return fmt.Errorf("invalid element type: %q", stringValue)
		}
		return nil
	}

	return fmt.Errorf("invalid element type payload: %s", string(data))
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

// UpdateCellsetFromDataframeViaBlob writes data to a cube via blob using a dataframe.
// The last dataframe column is treated as the value column; preceding columns are dimensions.
func (cs *CellService) UpdateCellsetFromDataframeViaBlob(ctx context.Context, cubeName string, df dataframe.DataFrame, sandboxName string) error {
	if df.Nrow() == 0 {
		return nil
	}

	headers := df.Names()
	if len(headers) < 2 {
		return fmt.Errorf("dataframe must contain at least one dimension and one value column")
	}

	var buffer bytes.Buffer
	if err := df.WriteCSV(&buffer); err != nil {
		return fmt.Errorf("write dataframe to csv: %w", err)
	}

	fileService := NewFileService(cs.rest)
	processService := NewProcessService(cs.rest)

	fileName := fmt.Sprintf("tm1go_dataload_temp_%s.csv", RandomString(8))
	if err := fileService.CreateCompressed(ctx, fileName, nil, buffer.Bytes()); err != nil {
		return err
	}

	loadFileName := fileName
	if !IsV1GreaterOrEqualToV2(cs.rest.version, "12.0.0") {
		loadFileName = fileName + ".blb"
	}

	deleteName := strings.TrimSuffix(loadFileName, ".blb")
	defer func() {
		_ = fileService.Delete(ctx, deleteName, nil)
	}()

	dataSourceType := "ASCII"
	odataType := ""
	if IsV1GreaterOrEqualToV2(cs.rest.version, "12.0.0") {
		odataType = "#ibm.tm1.api.v1.ASCIIDataSource"
	}

	process := models.NewProcess(loadFileName)
	process.DataSource = &models.ProcessDataSource{
		Type:                    dataSourceType,
		ODataType:               odataType,
		ASCIIDecimalSeparator:   ".",
		ASCIIDelimiterChar:      ",",
		ASCIIDelimiterType:      "Character",
		ASCIIHeaderRecords:      1,
		ASCIIQuoteCharacter:     "\"",
		ASCIIThousandSeparator:  ",",
		DataSourceNameForClient: loadFileName,
		DataSourceNameForServer: loadFileName,
	}

	process.Variables = make([]models.ProcessVariable, len(headers))
	for i := 0; i < len(headers)-1; i++ {
		process.Variables[i] = models.ProcessVariable{
			Name:      fmt.Sprintf("v%d", i+1),
			Type:      "String",
			StartByte: 0,
			EndByte:   0,
			Position:  i + 1,
		}
	}

	valueVariable := fmt.Sprintf("v%d", len(headers))
	process.Variables[len(headers)-1] = models.ProcessVariable{
		Name:      valueVariable,
		Type:      "Numeric",
		StartByte: 0,
		EndByte:   0,
		Position:  len(headers),
	}

	script := "CellPutN(" + valueVariable + ",'" + cubeName + "',"
	for i := 0; i < len(headers)-1; i++ {
		script += "v" + fmt.Sprintf("%d", i+1) + ","
	}
	script = strings.TrimSuffix(script, ",") + ");"
	process.DataProcedure = script

	success, status, _, err := processService.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("error executing process: %s", status)
	}

	return nil
}

// CalculationType represents the type of calculation for a cell.
type CalculationType int

const (
	CalculationTypeSimple        CalculationType = 0
	CalculationTypeConsolidation CalculationType = 1
	CalculationTypeRule          CalculationType = 2
)

func (ct *CalculationType) UnmarshalJSON(data []byte) error {
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		*ct = CalculationType(intValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		switch strings.ToLower(strings.TrimSpace(stringValue)) {
		case "simple":
			*ct = CalculationTypeSimple
		case "consolidation":
			*ct = CalculationTypeConsolidation
		case "rule":
			*ct = CalculationTypeRule
		default:
			return fmt.Errorf("invalid calculation type: %q", stringValue)
		}
		return nil
	}

	return fmt.Errorf("invalid calculation type payload: %s", string(data))
}

// CalculationComponent represents a component in a cell calculation trace.
type CalculationComponent struct {
	Cube       *models.Cube           `json:"Cube,omitempty"`
	Tuple      []Element              `json:"Tuple,omitempty"`
	Type       CalculationType        `json:"Type,omitempty"`
	Status     CellStatus             `json:"Status,omitempty"`
	Value      interface{}            `json:"Value,omitempty"`
	Statements []string               `json:"Statements,omitempty"`
	Components []CalculationComponent `json:"Components,omitempty"`
}

// FedCellDescriptor describes whether a cell is properly fed.
type FedCellDescriptor struct {
	Cube  *models.Cube `json:"Cube,omitempty"`
	Tuple []Element    `json:"Tuple,omitempty"`
	Fed   bool         `json:"Fed"`
}

// FeederTrace represents the result of a feeder trace operation.
type FeederTrace struct {
	FedCells   []FedCellDescriptor `json:"FedCells,omitempty"`
	Statements []string            `json:"Statements,omitempty"`
}

// RuleSyntaxError represents a syntax error found in cube rules.
type RuleSyntaxError struct {
	LineNumber int    `json:"LineNumber,omitempty"`
	Message    string `json:"Message,omitempty"`
}

// composeODataTupleFromElements builds an OData tuple binding from element names and dimension names.
// elements and dimensions must be in the same order and have the same length.
func (cs *CellService) composeODataTupleFromElements(cubeName string, elements []string, dimensions []string) (map[string]interface{}, error) {
	if len(dimensions) == 0 {
		var err error
		dimensions, err = cs.getDimensionNamesForCube(context.Background(), cubeName)
		if err != nil {
			return nil, fmt.Errorf("get dimensions: %w", err)
		}
	}

	if len(elements) != len(dimensions) {
		return nil, fmt.Errorf("elements count (%d) must match dimensions count (%d)", len(elements), len(dimensions))
	}

	tupleBindings := make([]string, 0, len(elements))
	for i, elem := range elements {
		dim := dimensions[i]
		hier := dim // Default hierarchy has same name as dimension
		tupleBindings = append(tupleBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
			url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
	}

	return map[string]interface{}{
		"Tuple@odata.bind": tupleBindings,
	}, nil
}

// TraceCellCalculation traces the calculation of a single cell.
// Returns the calculation components including rule statements, consolidation components and their values.
// depth controls how many levels of component recursion to return.
func (cs *CellService) TraceCellCalculation(ctx context.Context, cubeName string, elements []string, dimensions []string, sandboxName string, depth int) (*CalculationComponent, error) {
	if depth <= 0 {
		depth = 1
	}

	// Build $expand and $select for component depth
	expandQuery := ""
	selectQuery := ""
	for x := 1; x <= depth; x++ {
		componentDepth := ""
		for j := 0; j < x; j++ {
			if j > 0 {
				componentDepth += "/"
			}
			componentDepth += "Components"
		}
		componentsTupleCube := fmt.Sprintf("%s/Tuple($select=Name,UniqueName,Type),%s/Cube($select=Name)", componentDepth, componentDepth)
		if expandQuery != "" {
			expandQuery += ","
		}
		expandQuery += componentsTupleCube

		componentFields := fmt.Sprintf("%s/Type,%s/Value,%s/Statements", componentDepth, componentDepth, componentDepth)
		if selectQuery != "" {
			selectQuery += ","
		}
		selectQuery += componentFields
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.TraceCellCalculation?$select=Type,Value,Statements,%s&$expand=Tuple($select=Name,UniqueName,Type),%s",
		url.PathEscape(cubeName), selectQuery, expandQuery)

	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	body, err := cs.composeODataTupleFromElements(cubeName, elements, dimensions)
	if err != nil {
		return nil, fmt.Errorf("compose tuple: %w", err)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("trace cell calculation: %w", err)
	}
	defer resp.Body.Close()

	var result CalculationComponent
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode trace result: %w", err)
	}

	return &result, nil
}

// TraceCellFeeders traces the feeders from a cell.
// Returns the feeder statements and the collection of fed cells.
func (cs *CellService) TraceCellFeeders(ctx context.Context, cubeName string, elements []string, dimensions []string, sandboxName string) (*FeederTrace, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.TraceFeeders?$select=Statements,FedCells&$expand=FedCells/Tuple($select=Name,UniqueName,Type),FedCells/Cube($select=Name)",
		url.PathEscape(cubeName))

	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	body, err := cs.composeODataTupleFromElements(cubeName, elements, dimensions)
	if err != nil {
		return nil, fmt.Errorf("compose tuple: %w", err)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("trace cell feeders: %w", err)
	}
	defer resp.Body.Close()

	var result FeederTrace
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode feeder trace: %w", err)
	}

	return &result, nil
}

// CheckCellFeeders checks whether the components of a consolidated cell are properly fed.
// Returns a list of fed cell descriptors indicating which components are not properly fed.
func (cs *CellService) CheckCellFeeders(ctx context.Context, cubeName string, elements []string, dimensions []string, sandboxName string) ([]FedCellDescriptor, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.CheckFeeders?$select=Fed&$expand=Tuple($select=Name,UniqueName,Type),Cube($select=Name)",
		url.PathEscape(cubeName))

	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	body, err := cs.composeODataTupleFromElements(cubeName, elements, dimensions)
	if err != nil {
		return nil, fmt.Errorf("compose tuple: %w", err)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("check cell feeders: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []FedCellDescriptor `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode check feeders result: %w", err)
	}

	return result.Value, nil
}

// CheckRules checks cube rules for syntax errors.
// If rules is empty, the cube's existing rules are checked.
// Returns a list of RuleSyntaxError; an empty list means the rules are valid.
func (cs *CellService) CheckRules(ctx context.Context, cubeName string, rules string) ([]RuleSyntaxError, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.CheckRules", url.PathEscape(cubeName))

	body := map[string]interface{}{}
	if rules != "" {
		body["Rules"] = rules
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("check rules: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []RuleSyntaxError `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode check rules result: %w", err)
	}

	return result.Value, nil
}

// postAgainstCellset executes a POST request against a cellset's tm1.Update endpoint.
// Used for spreading operations.
func (cs *CellService) postAgainstCellset(ctx context.Context, cellsetID string, payload map[string]interface{}, sandboxName string) error {
	endpoint := fmt.Sprintf("/Cellsets('%s')/tm1.Update", cellsetID)
	if sandboxName != "" {
		endpoint = addSandboxParam(endpoint, sandboxName)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := cs.rest.Post(ctx, endpoint, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("post against cellset: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// parseDimensionHierarchyElementFromUniqueName parses a unique element name like "[dimension].[element]"
// or "[dimension].[hierarchy].[element]" and returns (dimension, hierarchy, element).
func parseDimensionHierarchyElementFromUniqueName(uniqueName string) (string, string, string) {
	// Remove leading/trailing whitespace
	uniqueName = strings.TrimSpace(uniqueName)

	// Extract parts between brackets
	parts := make([]string, 0, 3)
	for _, segment := range strings.Split(uniqueName, "]") {
		segment = strings.TrimLeft(segment, ".[")
		segment = strings.TrimSpace(segment)
		if segment != "" {
			parts = append(parts, segment)
		}
	}

	switch len(parts) {
	case 3:
		return parts[0], parts[1], parts[2]
	case 2:
		return parts[0], parts[0], parts[1]
	default:
		return uniqueName, uniqueName, uniqueName
	}
}

// RelativeProportionalSpread executes a relative proportional spread on a cube.
// value is the value to spread.
// uniqueElementNames are the target cell coordinates as unique element names (e.g. "[dim1].[elem1]").
// referenceUniqueElementNames are the reference cell coordinates as unique element names.
// referenceCube is the name of the reference cube (uses the same cube if empty).
func (cs *CellService) RelativeProportionalSpread(ctx context.Context, value float64, cubeName string, uniqueElementNames []string, referenceUniqueElementNames []string, referenceCube string, sandboxName string) error {
	// Build MDX to create a cellset targeting the cell
	mdxParts := make([]string, 0, len(uniqueElementNames))
	for _, uen := range uniqueElementNames {
		mdxParts = append(mdxParts, "{"+uen+"}")
	}
	mdx := fmt.Sprintf("SELECT %s ON 0 FROM [%s]", strings.Join(mdxParts, "*"), cubeName)

	cellsetID, err := cs.CreateCellset(ctx, mdx, sandboxName)
	if err != nil {
		return fmt.Errorf("create cellset for spread: %w", err)
	}

	// Build reference cell bindings
	refCellBindings := make([]string, 0, len(referenceUniqueElementNames))
	for _, refUEN := range referenceUniqueElementNames {
		dim, hier, elem := parseDimensionHierarchyElementFromUniqueName(refUEN)
		refCellBindings = append(refCellBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
			url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
	}

	if referenceCube == "" {
		referenceCube = cubeName
	}

	payload := map[string]interface{}{
		"BeginOrdinal":             0,
		"Value":                    fmt.Sprintf("RP%g", value),
		"ReferenceCell@odata.bind": refCellBindings,
		"ReferenceCube@odata.bind": fmt.Sprintf("Cubes('%s')", url.PathEscape(referenceCube)),
	}

	err = cs.postAgainstCellset(ctx, cellsetID, payload, sandboxName)
	// Always clean up the cellset
	_ = cs.DeleteCellset(ctx, cellsetID, sandboxName)

	return err
}

// ClearSpread executes a clear spread on a cube, zeroing out cells at the specified coordinates.
// uniqueElementNames are the target cell coordinates as unique element names (e.g. "[dim1].[elem1]").
func (cs *CellService) ClearSpread(ctx context.Context, cubeName string, uniqueElementNames []string, sandboxName string) error {
	// Build MDX to create a cellset
	mdxParts := make([]string, 0, len(uniqueElementNames))
	for _, uen := range uniqueElementNames {
		mdxParts = append(mdxParts, "{"+uen+"}")
	}
	mdx := fmt.Sprintf("SELECT %s ON 0 FROM [%s]", strings.Join(mdxParts, "*"), cubeName)

	cellsetID, err := cs.CreateCellset(ctx, mdx, sandboxName)
	if err != nil {
		return fmt.Errorf("create cellset for clear spread: %w", err)
	}

	// Build reference cell bindings (same as target for clear)
	refCellBindings := make([]string, 0, len(uniqueElementNames))
	for _, uen := range uniqueElementNames {
		dim, hier, elem := parseDimensionHierarchyElementFromUniqueName(uen)
		refCellBindings = append(refCellBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
			url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
	}

	payload := map[string]interface{}{
		"BeginOrdinal":             0,
		"Value":                    "C",
		"ReferenceCell@odata.bind": refCellBindings,
	}

	err = cs.postAgainstCellset(ctx, cellsetID, payload, sandboxName)
	// Always clean up the cellset
	_ = cs.DeleteCellset(ctx, cellsetID, sandboxName)

	return err
}

// EqualSpread executes an equal spread on a cube.
// value is the value to spread equally across leaf cells.
// uniqueElementNames are the target cell coordinates as unique element names.
func (cs *CellService) EqualSpread(ctx context.Context, value float64, cubeName string, uniqueElementNames []string, sandboxName string) error {
	// Build MDX to create a cellset
	mdxParts := make([]string, 0, len(uniqueElementNames))
	for _, uen := range uniqueElementNames {
		mdxParts = append(mdxParts, "{"+uen+"}")
	}
	mdx := fmt.Sprintf("SELECT %s ON 0 FROM [%s]", strings.Join(mdxParts, "*"), cubeName)

	cellsetID, err := cs.CreateCellset(ctx, mdx, sandboxName)
	if err != nil {
		return fmt.Errorf("create cellset for equal spread: %w", err)
	}

	// Build reference cell bindings
	refCellBindings := make([]string, 0, len(uniqueElementNames))
	for _, uen := range uniqueElementNames {
		dim, hier, elem := parseDimensionHierarchyElementFromUniqueName(uen)
		refCellBindings = append(refCellBindings, fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
			url.PathEscape(dim), url.PathEscape(hier), url.PathEscape(elem)))
	}

	payload := map[string]interface{}{
		"BeginOrdinal":             0,
		"Value":                    fmt.Sprintf("S%g", value),
		"ReferenceCell@odata.bind": refCellBindings,
	}

	err = cs.postAgainstCellset(ctx, cellsetID, payload, sandboxName)
	_ = cs.DeleteCellset(ctx, cellsetID, sandboxName)

	return err
}

// ClearWithMDX clears (zeros out) a slice of a cube based on an MDX query.
// This creates a temporary MDX view and uses a TI process with ViewZeroOut to clear the data.
// Requires admin permissions.
func (cs *CellService) ClearWithMDX(ctx context.Context, cubeName string, mdx string, sandboxName string) error {
	processService := NewProcessService(cs.rest)

	viewName := fmt.Sprintf("}tm1go_%s", RandomString(16))

	// Create the MDX view
	viewBody := map[string]interface{}{
		"@odata.type": "#ibm.tm1.api.v1.MDXView",
		"Name":        viewName,
		"MDX":         mdx,
	}

	viewPayload, err := json.Marshal(viewBody)
	if err != nil {
		return fmt.Errorf("marshal view body: %w", err)
	}

	viewEndpoint := fmt.Sprintf("/Cubes('%s')/Views", url.PathEscape(cubeName))
	resp, err := cs.rest.Post(ctx, viewEndpoint, strings.NewReader(string(viewPayload)))
	if err != nil {
		return fmt.Errorf("create view: %w", err)
	}
	resp.Body.Close()

	// Ensure view cleanup
	defer func() {
		deleteEndpoint := fmt.Sprintf("/Cubes('%s')/Views('%s')", url.PathEscape(cubeName), url.PathEscape(viewName))
		if delResp, delErr := cs.rest.Delete(ctx, deleteEndpoint); delErr == nil {
			delResp.Body.Close()
		}
	}()

	// Build and execute TI process with ViewZeroOut
	enableSandbox := ""
	if sandboxName != "" {
		enableSandbox = fmt.Sprintf("ServerActiveSandbox('%s');", sandboxName)
	}

	process := models.NewProcess("")
	process.PrologProcedure = enableSandbox
	process.EpilogProcedure = fmt.Sprintf("ViewZeroOut('%s','%s');", cubeName, viewName)

	success, status, _, err := processService.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return fmt.Errorf("execute clear process: %w", err)
	}
	if !success {
		return fmt.Errorf("clear failed for cube '%s': %s", cubeName, status)
	}

	return nil
}
