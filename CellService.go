package tm1go

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
)

// Service for reading and writing TM1 cube cells
type CellService struct {
	rest    *RestService
	cube    *CubeService
	file    *FileService
	process *ProcessService
}

// NewCellService creates a new cell service
func NewCellService(rest *RestService, cube *CubeService, file *FileService, process *ProcessService) *CellService {
	return &CellService{rest: rest, cube: cube, file: file, process: process}
}

// CreateCellSet creates a cellset
func (cs *CellService) CreateCellSet(mdx string, sandboxName string) (string, error) {
	url := "/ExecuteMDX"
	url, err := AddURLParameters(url, map[string]string{"!sandbox": sandboxName})
	if err != nil {
		return "", err
	}
	data := map[string]string{"MDX": mdx}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	response, err := cs.rest.POST(url, string(jsonData), nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	result := Cellset{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// DeleteCellSet deletes a cellset
func (cs *CellService) DeleteCellSet(cellsetID string) error {
	url := "/Cellsets('" + cellsetID + "')"
	_, err := cs.rest.DELETE(url, nil, 0, nil)
	return err
}

// ExtractCellSetCount extracts the number of cells in a cellset
func (cs *CellService) ExtractCellSetCount(cellsetID string, sandboxName string) (int, error) {
	url := "/Cellsets('" + cellsetID + "')/Cells/$count"
	url, err := AddURLParameters(url, map[string]string{"!sandbox": sandboxName})
	if err != nil {
		return 0, err
	}
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ExtractCellsetCells extracts cells from a cellset
func (cs *CellService) ExtractCellsetCellsRaw(cellsetID string, cellProperties []string, top, skip int, skipZeros, skipConsolidatedCells, skipRuleDerivedCells bool, sandboxName string) ([]Cell, error) {
	if cellProperties == nil {
		cellProperties = []string{"Value"}
	}

	if skipRuleDerivedCells {
		cellProperties = append(cellProperties, "RuleDerived", "Updateable")
	}

	if skipConsolidatedCells {
		cellProperties = append(cellProperties, "Consolidated")
	}

	if skip > 0 || skipZeros || skipRuleDerivedCells || skipConsolidatedCells {
		if !SliceContains(cellProperties, "Ordinal") {
			cellProperties = append(cellProperties, "Ordinal")
		}
	}

	var filters []string
	if skipZeros || skipConsolidatedCells || skipRuleDerivedCells {
		if skipZeros {
			filters = append(filters, "Value ne 0 and Value ne null and Value ne ''")
		}
		if skipConsolidatedCells {
			filters = append(filters, "Consolidated eq false")
		}
		if skipRuleDerivedCells {
			filters = append(filters, "RuleDerived eq false")
		}
	}

	filterCells := strings.Join(filters, " and ")
	cellPropertiesStr := strings.Join(cellProperties, ",")
	topCells := ""
	if top > 0 {
		topCells = fmt.Sprintf(";$top=%d", top)
	}
	skipCells := ""
	if skip > 0 {
		skipCells = fmt.Sprintf(";$skip=%d", skip)
	}
	filterCellsParam := ""
	if filterCells != "" {
		filterCellsParam = fmt.Sprintf(";$filter=%s", filterCells)
	}

	url := fmt.Sprintf("/Cellsets('%s')?$expand=Cells($select=%s%s%s%s)", cellsetID, cellPropertiesStr, topCells, skipCells, filterCellsParam)

	if sandboxName != "" {
		err := error(nil)
		url, err = AddURLParameters(url, map[string]string{"!sandbox": sandboxName})
		if err != nil {
			return nil, err
		}
	}

	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := Cellset{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Cells, nil
}

// ExtractCellsetCellsAsync extracts cells from a cellset asynchronously
func (cs *CellService) ExtractCellsetCellsAsync(cellsetID string, cellProperties []string, sandboxName string, maxWorkers int) ([]Cell, error) {

	cellCount, err := cs.ExtractCellSetCount(cellsetID, sandboxName)
	if err != nil {
		return nil, err
	}
	if cellCount == 0 {
		return make([]Cell, 0), nil
	}
	if maxWorkers > cellCount {
		maxWorkers = 1
	}
	partionSize := int(math.Ceil(float64(cellCount) / float64(maxWorkers)))
	top := partionSize
	result := make([]Cell, cellCount)

	wg := sync.WaitGroup{}
	errChan := make(chan error, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		skip := i * partionSize
		wg.Add(1)
		go func(skip int) {
			defer wg.Done()
			cells, err := cs.ExtractCellsetCellsRaw(cellsetID, cellProperties, top, skip, false, false, false, sandboxName)
			if err != nil {
				errChan <- err
				return
			}
			copy(result[skip:skip+len(cells)], cells)
		}(skip)
	}
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return nil, <-errChan
	}
	return result, nil
}

// UpdateCellset updates values inside a cellset
func (cs *CellService) UpdateCellset(cellsetID string, values []interface{}, valuesOrdinalOffset int, sandboxName string) error {

	url := fmt.Sprintf("/Cellsets('%s')/Cells", cellsetID)
	url, err := AddURLParameters(url, map[string]string{"!sandbox": sandboxName})
	if err != nil {
		return err
	}
	data := []map[string]interface{}{}
	for o, value := range values {
		data = append(data, map[string]interface{}{
			"Ordinal": o + valuesOrdinalOffset,
			"Value":   value,
		})
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = cs.rest.PATCH(url, string(jsonData), nil, 0, nil)
	return err
}

// UpdateCellsetMDX updates values inside a cellset using MDX
func (cs *CellService) UpdateCellsetMDX(mdx string, values []interface{}, sandboxName string) error {
	cellsetID, err := cs.CreateCellSet(mdx, sandboxName)
	if err != nil {
		return err
	}
	defer cs.DeleteCellSet(cellsetID)
	return cs.UpdateCellset(cellsetID, values, 0, sandboxName)
}

// UpdateCellsetAsync updates values inside a cellset asynchronously
/* func (cs *CellService) UpdateCellsetAsync(cellsetID string, values []interface{}, sandboxName string, maxWorkers int) error {
	cellCount, err := cs.ExtractCellSetCount(cellsetID, sandboxName)
	if err != nil {
		return err
	}
	if cellCount == 0 {
		return nil
	}
	if maxWorkers > cellCount {
		maxWorkers = 1
	}
	partionSize := int(math.Ceil(float64(cellCount) / float64(maxWorkers)))

	wg := sync.WaitGroup{}
	errChan := make(chan error, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		skip := i * partionSize
		wg.Add(1)
		go func(skip int) {
			defer wg.Done()
			end := skip + partionSize
			if end > len(values) {
				end = len(values)
			}
			err := cs.UpdateCellset(cellsetID, values[skip:end], skip, sandboxName)
			if err != nil {
				errChan <- err
				return
			}
		}(skip)
	}
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
} */

// CellGet retrieves a cell
// First parameter is the cube name followed by the element names
func (cs *CellService) CellGet(params ...string) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("at least 3 parameters are required")
	}
	cubeName := params[0]

	// Get cube dimensions
	dimensions, err := cs.cube.GetDimensionNames(cubeName)
	if err != nil {
		return nil, err
	}

	if len(dimensions) != len(params)-1 {
		return nil, fmt.Errorf("number of dimensions does not match number of elements")
	}

	// Build MDX
	mdxBuilder := NewMDXBuilder(cubeName)
	for i, dimension := range dimensions {
		if i == 0 {
			mdxBuilder.AddMemberOnColumns(dimension, dimension, params[i+1])
		} else {
			mdxBuilder.AddMemberOnWhere(dimension, dimension, params[i+1])
		}
	}

	mdx, err := mdxBuilder.ToString()
	if err != nil {
		return nil, err
	}

	// Execute MDX
	cellsetID, err := cs.CreateCellSet(mdx, "")
	if err != nil {
		return nil, err
	}
	defer cs.DeleteCellSet(cellsetID)
	cell, err := cs.ExtractCellsetCellsRaw(cellsetID, []string{"Value"}, 1, 0, false, false, false, "")
	if err != nil {
		return nil, err
	}
	return cell[0].Value, nil
}

// CellPut updates a cell
// First parameter is the value, second parameter is the cube name followed by the element names
func (cs *CellService) CellPut(params ...interface{}) error {
	if len(params) < 4 {
		return fmt.Errorf("at least 4 parameters are required")
	}
	cubeName, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("cubeName must be of type string")
	}

	var elements = make([]string, 0, len(params)-2)
	for _, element := range params[2:] {
		elementStr, ok := element.(string)
		if !ok {
			return fmt.Errorf("element must be of type string")
		}
		elements = append(elements, elementStr)
	}

	// Get cube dimensions
	dimensions, err := cs.cube.GetDimensionNames(cubeName)
	if err != nil {
		return err
	}

	if len(dimensions) != len(elements) {
		return fmt.Errorf("number of dimensions does not match number of elements")
	}

	url := fmt.Sprintf("/Cubes('%s')/tm1.Update", cubeName)

	type CellUpdateStruct struct {
		Value interface{} `json:"Value"`
		Cells []struct {
			Tuple []string `json:"Tuple@odata.bind"`
		} `json:"Cells"`
	}

	data := CellUpdateStruct{
		Value: params[0],
		Cells: []struct {
			Tuple []string `json:"Tuple@odata.bind"`
		}{{
			Tuple: make([]string, len(elements)),
		}},
	}

	for i, element := range elements {
		data.Cells[0].Tuple[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", dimensions[i], dimensions[i], element)
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Execute request
	_, err = cs.rest.POST(url, string(jsonData), nil, 0, nil)
	return err
}

// ExtractCellsetCells extracts cells from a cellset
func (cs *CellService) ExtractCellsetAxesAndCube(cellsetID string, sandboxName string) (*Cellset, error) {
	url := fmt.Sprintf("/Cellsets('%s')?$expand=Cube($select=Name),Axes($select=Ordinal,Cardinality;$expand=Hierarchies($select=Name,UniqueName;$expand=Dimension($select=Name)),Tuples($expand=Members($select=Name,UniqueName,Attributes,DisplayInfoAbove,DisplayInfo,Type)))", cellsetID)

	if sandboxName != "" {
		err := error(nil)
		url, err = AddURLParameters(url, map[string]string{"!sandbox": sandboxName})
		if err != nil {
			return nil, err
		}
	}

	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := &Cellset{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateCellsetFromDataframe updates values inside a cellset using dataframe
func (cs *CellService) UpdateCellsetFromDataframe(cubeName string, df *DataFrame, sandboxName string) error {

	// Build mdx
	mdx, err := df.ToMDX(cubeName)
	if err != nil {
		return err
	}

	cellsetID, err := cs.CreateCellSet(mdx, sandboxName)
	if err != nil {
		return err
	}
	defer cs.DeleteCellSet(cellsetID)

	return cs.UpdateCellset(cellsetID, df.Columns[df.Headers[len(df.Headers)-1]], 0, sandboxName)
}

// ExecuteMdxToDataframe executes an MDX query and returns the result as a dataframe
func (cs *CellService) ExecuteMdxToDataframe(mdx string, sandboxName string) (*DataFrame, error) {
	cellsetID, err := cs.CreateCellSet(mdx, sandboxName)
	if err != nil {
		return nil, err
	}
	defer cs.DeleteCellSet(cellsetID)

	cellsetMetadata, err := cs.ExtractCellsetAxesAndCube(cellsetID, sandboxName)
	if err != nil {
		return nil, err
	}

	cellsetCells, err := cs.ExtractCellsetCellsRaw(cellsetID, []string{"Value"}, 0, 0, false, false, false, sandboxName)

	if err != nil {
		return nil, err
	}

	cellset := &Cellset{
		ID:    cellsetID,
		Cube:  cellsetMetadata.Cube,
		Cells: cellsetCells,
		Axes:  cellsetMetadata.Axes,
	}

	df, err := CellSetToDataFrame(cellset)
	if err != nil {
		return nil, err
	}

	return df, nil
}

// Read via blob
func (cs *CellService) ExecuteMDXViaBlob(mdx string, sandboxName string) (*DataFrame, error) {
	// create temp mdx view

	// create unbound process and asciioutput data from the view into a file (csv)

	// get the file using file service

	// process csv file into dataframe

	// cleanup temp view and file

	// return dataframe

	return &DataFrame{}, nil
}

// UpdateCellsetFromDataframeViaBlob Writes data to a cube via blob. Numeric valus only.
func (cs *CellService) UpdateCellsetFromDataframeViaBlob(cubeName string, df *DataFrame, sandboxName string) error {
	// Convert dataframe into csv file
	dataBytes, err := df.ConvertToCSVBytes()
	if err != nil {
		return err
	}

	// Upload file to tm1 using file service
	randomNumber := 10000 + rand.Intn(90000)
	randomName := fmt.Sprintf("tm1go_dataload_temp_%d.csv", randomNumber)
	err = cs.file.CreateCompressed(randomName, nil, dataBytes)
	if err != nil {
		return err
	}

	// Check TM1 version
	// v11 requires .blb extension for source files
	if !IsV1GreaterOrEqualToV2(cs.rest.version, "12.0.0") {
		randomName = randomName + ".blb"
	}

	// Create unboud process and use the file as a source to load data into cube
	process := NewProcess(randomName)

	process.DataSource.Type = "ASCII"
	process.DataSource.AsciiDecimalSeparator = "."
	process.DataSource.AsciiDelimiterChar = ","
	process.DataSource.AsciiDelimiterType = "Character"
	process.DataSource.AsciiHeaderRecords = 1
	process.DataSource.AsciiQuoteCharacter = "\""
	process.DataSource.AsciiThousandSeparator = ","
	process.DataSource.DataSourceNameForClient = randomName
	process.DataSource.DataSourceNameForServer = randomName

	process.Variables = make([]ProcessVariable, len(df.Headers))

	// Add dimension variables
	for i := 0; i < len(df.Headers)-1; i++ {
		process.Variables[i] = ProcessVariable{
			Name:      fmt.Sprintf("v%d", i+1),
			Type:      "String",
			StartByte: 0,
			EndByte:   0,
			Position:  i + 1,
		}
	}

	// Add last value variable
	valueVariable := fmt.Sprintf("v%d", len(df.Headers))
	process.Variables[len(df.Headers)-1] = ProcessVariable{
		Name:      valueVariable,
		Type:      "Numeric",
		StartByte: 0,
		EndByte:   0,
		Position:  len(df.Headers),
	}

	// Add process script
	scriptString := "CellPutN(" + valueVariable + ",'" + cubeName + "',"
	for i := 0; i < len(df.Headers)-1; i++ {
		scriptString += "v" + fmt.Sprintf("%d", i+1) + ","
	}
	scriptString = scriptString[:len(scriptString)-1]
	scriptString += ");"

	process.DataProcedure = scriptString
	executeProcessResult, err := cs.process.ExecuteProcessWithReturn(process, nil, 0)
	if err != nil {
		return err
	}

	if executeProcessResult.ProcessExecuteStatusCode != 0 {
		return fmt.Errorf("error executing process: %s", executeProcessResult.ProcessExecuteStatusCode.ToString())
	}

	// Cleanup: delete file
	err = cs.file.Delete(strings.TrimSuffix(randomName, ".blb"), nil)
	if err != nil {
		return err
	}

	return nil
}
