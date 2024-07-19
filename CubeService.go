package tm1go

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type CubeService struct {
	rest      *RestService
	object    *ObjectService
	dimension *DimensionService
	process   *ProcessService
}

func NewCubeService(rest *RestService, object *ObjectService, dimension *DimensionService, process *ProcessService) *CubeService {
	return &CubeService{rest: rest, object: object, dimension: dimension, process: process}
}

// Create new cube in tm1
func (cs *CubeService) Create(cube Cube) error {
	url := "/Cubes"
	cubeBody, err := cube.getBody()
	if err != nil {
		return err
	}
	_, err = cs.rest.POST(url, cubeBody, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Get cube from tm1
func (cs *CubeService) Get(cubeName string) (*Cube, error) {
	url := fmt.Sprintf("/Cubes('%v')?$expand=Dimensions($select=Name)", cubeName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	cube := Cube{}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&cube)
	if err != nil {
		return nil, err
	}
	// Cater for potential EnableSandboxDimension=T setup
	if strings.EqualFold(cube.Dimensions[0].Name, "Sandboxes") {
		cube.Dimensions = cube.Dimensions[1:]
	}
	return &cube, nil
}

func (cs *CubeService) GetLastDataUpdate(cubeName string) (string, error) {
	url := fmt.Sprintf("/Cubes('%v')/LastDataUpdate/$value", cubeName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return string(body), err
}

func (cs *CubeService) GetAll() ([]Cube, error) {
	url := "/Cubes?$expand=Dimensions($select=Name)"
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Get all Cubes without } prefix from TM1 Server as TM1py.Cube instances
func (cs *CubeService) GetModelCubes() ([]Cube, error) {
	url := "/ModelCubes()?$expand=Dimensions($select=Name)"
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Get all Cubes with } prefix from TM1 Server as TM1py.Cube instances
func (cs *CubeService) GetControlCubes() ([]Cube, error) {
	url := "/ControlCubes()?$expand=Dimensions($select=Name)"
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Get number of model cubes in TM1
func (cs *CubeService) GetNumberOfCubes(skipControlCube bool) (int, error) {
	url := "/Cubes/$count"
	if skipControlCube {
		url = "/ModelCubes()?$select=Name&$top=0&$count"
	}

	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	if skipControlCube {
		result := ValueArray[Cube]{}
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			return 0, err
		}
		count = result.Count
	} else {
		err = json.NewDecoder(response.Body).Decode(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (cs *CubeService) GetMeasureDimension(cubeName string) (*Dimension, error) {
	url := fmt.Sprintf("/Cubes('%v')/Dimensions?$select=Name", cubeName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Dimension]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	// Get last element of the array
	return &result.Value[len(result.Value)-1], nil
}

// Update existing cube in TM1
func (cs *CubeService) Update(cube Cube) error {
	url := fmt.Sprintf("/Cubes('%v')", cube.Name)
	cubeBody, err := cube.getBody()
	if err != nil {
		return err
	}
	cs.rest.PATCH(url, cubeBody, nil, 0, nil)
	return nil
}

// Update if exists else create
func (cs *CubeService) UpdateOrCreate(cube Cube) error {
	url := fmt.Sprintf("/Cubes('%v')", cube.Name)
	cubeExists, err := cs.object.Exists(url)
	if err != nil {
		return err
	}
	if cubeExists {
		return cs.Update(cube)
	}
	return cs.Create(cube)
}

// Check rules syntax for a TM1 cube
func (cs *CubeService) CheckRules(cubeName string) ([]RuleSyntaxError, error) {
	url := fmt.Sprintf("/Cubes('%v')/tm1.CheckRules", cubeName)
	response, err := cs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[RuleSyntaxError]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Delete TM1 cube
func (cs *CubeService) Delete(cubeName string) error {
	if !cs.rest.IsDataAdmin() {
		return fmt.Errorf("Delete requires Data Admin privilege")
	}
	url := fmt.Sprintf("/Cubes('%v')", cubeName)
	_, err := cs.rest.DELETE(url, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Exists checks if a cube exists in TM1
func (cs *CubeService) Exists(cubeName string) (bool, error) {
	url := fmt.Sprintf("/Cubes('%v')", cubeName)
	exists, err := cs.object.Exists(url)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Get the list of all cube names in TM1
func (cs *CubeService) GetAllNames(skipControlCube bool) ([]string, error) {
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}
	url := fmt.Sprintf("/%v?$select=Name", cubesEndPoint)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	cubeNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		cubeNames = append(cubeNames, value.Name)
	}
	return cubeNames, nil
}

// Get a list of cubes with rules
func (cs *CubeService) GetAllNamesWithRules(skipControlCube bool) ([]string, error) {
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}

	url := fmt.Sprintf("/%v?$select=Name,Rules&$filter=Rules ne null", cubesEndPoint)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	cubeNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		cubeNames = append(cubeNames, value.Name)
	}
	return cubeNames, nil
}

// Get a list of cubes without rules
func (cs *CubeService) GetAllNamesWithoutRules(skipControlCube bool) ([]string, error) {
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}

	url := fmt.Sprintf("/%v?$select=Name,Rules&$filter=Rules eq null", cubesEndPoint)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	cubeNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		cubeNames = append(cubeNames, value.Name)
	}
	return cubeNames, nil
}

// Get a list of a cube dimensions names
func (cs *CubeService) GetDimensionNames(cubeName string) ([]string, error) {
	url := fmt.Sprintf("/Cubes('%v')/Dimensions?$select=Name", cubeName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Dimension]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	dimensionNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		dimensionNames = append(dimensionNames, value.Name)
	}
	return dimensionNames, nil
}

// Get list of cubes that contain dimension name provided
func (cs *CubeService) SearchForDimension(dimensionName string, skipControlCube bool) ([]string, error) {
	dimensionName = strings.ToLower(strings.ReplaceAll(dimensionName, " ", ""))
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}
	url := fmt.Sprintf("/%v?$select=Name&$filter=Dimensions/any(d: replace(tolower(d/Name), ' ', '') eq '%v')", cubesEndPoint, dimensionName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	cubeNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		cubeNames = append(cubeNames, value.Name)
	}
	return cubeNames, nil
}

// Get a list of cubes which dimensions names match the substring
func (cs *CubeService) SearchForDimensionSubstring(substring string, skipControlCube bool) (map[string][]string, error) {
	substring = strings.ToLower(strings.ReplaceAll(substring, " ", ""))
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}
	url := fmt.Sprintf("/%v?$select=Name&$filter=Dimensions/any(d: contains(replace(tolower(d/Name), ' ', ''),'%v'))"+
		"&$expand=Dimensions($select=Name;$filter=contains(replace(tolower(Name), ' ', ''), '%v'))", cubesEndPoint, substring, substring)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	cubeMap := map[string][]string{}

	for _, cube := range result.Value {
		cubeMap[cube.Name] = make([]string, 0, len(cube.Dimensions))
		for _, dim := range cube.Dimensions {
			cubeMap[cube.Name] = append(cubeMap[cube.Name], dim.Name)
		}
	}
	return cubeMap, nil
}

// Get a list of cubes which rules contain the substring
func (cs *CubeService) SearchForRuleSubstring(substring string, skipControlCube bool, caseInsensitive bool, spaceInsensitive bool) ([]Cube, error) {
	substring = strings.ToLower(strings.ReplaceAll(substring, " ", ""))
	cubesEndPoint := "Cubes"
	if skipControlCube {
		cubesEndPoint = "ModelCubes()"
	}

	urlFilter := "Rules ne null and contains("
	if caseInsensitive && spaceInsensitive {
		urlFilter += fmt.Sprintf("tolower(replace(Rules, ' ', '')),'%v')", substring)
	} else if caseInsensitive {
		urlFilter += fmt.Sprintf("tolower(Rules),'%v')", substring)
	} else if spaceInsensitive {
		urlFilter += fmt.Sprintf("replace(Rules, ' ', ''),'%v')", substring)
	} else {
		urlFilter += fmt.Sprintf("Rules,'%v')", substring)
	}
	url := fmt.Sprintf("/%v?$filter=%v&$expand=Dimensions($select=Name)", cubesEndPoint, urlFilter)

	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Cube]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

// Get list of dimensions names in storage order
func (cs *CubeService) GetStorageDimensionOrder(cubeName string) ([]string, error) {
	if isV1GreaterOrEqualToV2("11.4.0", cs.rest.version) {
		err := fmt.Errorf("GetStorageDimensionOrder requires TM1 v11.4 or greater")
		return nil, err
	}
	url := fmt.Sprintf("/Cubes('%v')/tm1.DimensionsStorageOrder()?$select=Name", cubeName)
	response, err := cs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Dimension]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	dimensionNames := make([]string, 0, len(result.Value))
	for _, value := range result.Value {
		dimensionNames = append(dimensionNames, value.Name)
	}
	return dimensionNames, nil
}

// Update internal dimension storage order. The function returns percentage of memory saved
func (cs *CubeService) UpdateStorageDimensionOrder(cubeName string, dimensions []string) (float64, error) {
	if isV1GreaterOrEqualToV2("11.4.0", cs.rest.version) {
		err := fmt.Errorf("UpdateStorageDimensionOrder requires TM1 v11.4 or greater")
		return 0, err
	}
	if !cs.rest.IsDataAdmin() {
		return 0, fmt.Errorf("UpdateStorageDimensionOrder requires Data Admin privilege")
	}

	url := fmt.Sprintf("/Cubes('%v')/tm1.ReorderDimensions", cubeName)

	payload := map[string][]string{}
	payload["Dimensions@odata.bind"] = make([]string, 0, len(dimensions))
	for _, value := range dimensions {
		dimUrl := fmt.Sprintf("Dimensions('%v')", value)
		payload["Dimensions@odata.bind"] = append(payload["Dimensions@odata.bind"], dimUrl)
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	response, err := cs.rest.POST(url, string(jsonData), nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	var result Value[float64]
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Value, nil
}

// Load cube into memory
func (cs *CubeService) Load(cubeName string) error {
	if isV1GreaterOrEqualToV2("11.6.0", cs.rest.version) {
		err := fmt.Errorf("Load requires TM1 v11.6 or greater")
		return err
	}
	if !cs.rest.IsDataAdmin() {
		return fmt.Errorf("Load requires Data Admin privilege")
	}
	url := fmt.Sprintf("/Cubes('%v')/tm1.Load", cubeName)
	_, err := cs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Unload cube into memory
func (cs *CubeService) Unload(cubeName string) error {
	if isV1GreaterOrEqualToV2("11.6.0", cs.rest.version) {
		err := fmt.Errorf("Unload requires TM1 v11.6 or greater")
		return err
	}
	if !cs.rest.IsDataAdmin() {
		return fmt.Errorf("Unload requires Data Admin privilege")
	}
	url := fmt.Sprintf("/Cubes('%v')/tm1.Unload", cubeName)
	_, err := cs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Lock cube to prevent any modification by the users
func (cs *CubeService) Lock(cubeName string) error {
	url := fmt.Sprintf("/Cubes('%v')/tm1.Lock", cubeName)
	_, err := cs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Unlock cube to prevent any modification by the users
func (cs *CubeService) Unlock(cubeName string) error {
	url := fmt.Sprintf("/Cubes('%v')/tm1.Unlock", cubeName)
	_, err := cs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Serializes a cube by saving data updates
func (cs *CubeService) CubeSaveData(cubeName string) error {

	epilog := fmt.Sprintf("CubeSaveData('%v');", cubeName)
	result, err := cs.process.ExecuteTICode("", epilog)
	if err != nil {
		return err
	}

	if result.ProcessExecuteStatusCode != CompletedSuccessfully {
		return fmt.Errorf("Process did not complete successfully. Result: %v", result.ProcessExecuteStatusCode)
	}

	return nil
}
