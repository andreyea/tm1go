package tm1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// CubeService handles operations for TM1 Cubes
type CubeService struct {
	rest    *RestService
	process *ProcessService
}

// NewCubeService creates a new CubeService instance
func NewCubeService(rest *RestService) *CubeService {
	return &CubeService{
		rest:    rest,
		process: NewProcessService(rest),
	}
}

// Create creates a new cube in TM1
func (cs *CubeService) Create(ctx context.Context, cube *models.Cube) error {
	body, err := cube.Body()
	if err != nil {
		return fmt.Errorf("failed to build cube body: %w", err)
	}

	resp, err := cs.rest.Post(ctx, "/Cubes", strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

// Get retrieves a cube by name
func (cs *CubeService) Get(ctx context.Context, cubeName string) (*models.Cube, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')?$expand=Dimensions($select=Name)", url.PathEscape(cubeName))

	var cube models.Cube
	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &cube)
	if err != nil {
		return nil, err
	}

	// Cater for potential EnableSandboxDimension=T setup
	if len(cube.Dimensions) > 0 && strings.EqualFold(cube.Dimensions[0].Name, "Sandboxes") {
		cube.Dimensions = cube.Dimensions[1:]
	}

	return &cube, nil
}

// GetLastDataUpdate retrieves the cube's last data update timestamp as a string
func (cs *CubeService) GetLastDataUpdate(ctx context.Context, cubeName string) (string, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/LastDataUpdate/$value", url.PathEscape(cubeName))

	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetAll retrieves all cubes
func (cs *CubeService) GetAll(ctx context.Context) ([]models.Cube, error) {
	endpoint := "/Cubes?$expand=Dimensions($select=Name)"

	var response struct {
		Value []models.Cube `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetModelCubes retrieves all model cubes (without } prefix)
func (cs *CubeService) GetModelCubes(ctx context.Context) ([]models.Cube, error) {
	endpoint := "/ModelCubes()?$expand=Dimensions($select=Name)"

	var response struct {
		Value []models.Cube `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetControlCubes retrieves all control cubes (with } prefix)
func (cs *CubeService) GetControlCubes(ctx context.Context) ([]models.Cube, error) {
	endpoint := "/ControlCubes()?$expand=Dimensions($select=Name)"

	var response struct {
		Value []models.Cube `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetNumberOfCubes retrieves the number of cubes in TM1
func (cs *CubeService) GetNumberOfCubes(ctx context.Context, skipControlCube bool) (int, error) {
	if skipControlCube {
		endpoint := "/ModelCubes()?$select=Name&$top=0&$count=true"

		var response struct {
			Count int           `json:"@odata.count"`
			Value []models.Cube `json:"value"`
		}

		err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
		if err != nil {
			return 0, err
		}

		if response.Count == 0 {
			return len(response.Value), nil
		}

		return response.Count, nil
	}

	resp, err := cs.rest.Get(ctx, "/Cubes/$count")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse cube count: %w", err)
	}

	return count, nil
}

// GetMeasureDimension retrieves the last dimension (measure) for a cube
func (cs *CubeService) GetMeasureDimension(ctx context.Context, cubeName string) (*models.Dimension, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/Dimensions?$select=Name", url.PathEscape(cubeName))

	var response struct {
		Value []models.Dimension `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Value) == 0 {
		return nil, fmt.Errorf("cube '%s' has no dimensions", cubeName)
	}

	return &response.Value[len(response.Value)-1], nil
}

// Update updates an existing cube
func (cs *CubeService) Update(ctx context.Context, cube *models.Cube) error {
	body, err := cube.Body()
	if err != nil {
		return fmt.Errorf("failed to build cube body: %w", err)
	}

	endpoint := fmt.Sprintf("/Cubes('%s')", url.PathEscape(cube.Name))
	resp, err := cs.rest.Patch(ctx, endpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

// UpdateOrCreate updates a cube if it exists, otherwise creates it
func (cs *CubeService) UpdateOrCreate(ctx context.Context, cube *models.Cube) error {
	exists, err := cs.Exists(ctx, cube.Name)
	if err != nil {
		return err
	}

	if exists {
		return cs.Update(ctx, cube)
	}

	return cs.Create(ctx, cube)
}

// CheckRules checks rules syntax for a cube
func (cs *CubeService) CheckRules(ctx context.Context, cubeName string) ([]models.RuleSyntaxError, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.CheckRules", url.PathEscape(cubeName))

	var response struct {
		Value []models.RuleSyntaxError `json:"value"`
	}

	err := cs.rest.JSON(ctx, "POST", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// Delete deletes a cube
func (cs *CubeService) Delete(ctx context.Context, cubeName string) error {
	isDataAdmin, err := cs.isDataAdmin(ctx)
	if err != nil {
		return err
	}
	if !isDataAdmin {
		return fmt.Errorf("Delete requires Data Admin privilege")
	}

	endpoint := fmt.Sprintf("/Cubes('%s')", url.PathEscape(cubeName))
	resp, err := cs.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Exists checks if a cube exists in TM1
func (cs *CubeService) Exists(ctx context.Context, cubeName string) (bool, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')", url.PathEscape(cubeName))
	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// GetAllNames retrieves all cube names
func (cs *CubeService) GetAllNames(ctx context.Context, skipControlCube bool) ([]string, error) {
	endpoint := "/Cubes?$select=Name"
	if skipControlCube {
		endpoint = "/ModelCubes()?$select=Name"
	}

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, cube := range response.Value {
		names[i] = cube.Name
	}

	return names, nil
}

// GetAllNamesWithRules retrieves all cube names that have rules
func (cs *CubeService) GetAllNamesWithRules(ctx context.Context, skipControlCube bool) ([]string, error) {
	endpoint := "/Cubes?$select=Name,Rules&$filter=Rules ne null"
	if skipControlCube {
		endpoint = "/ModelCubes()?$select=Name,Rules&$filter=Rules ne null"
	}

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, cube := range response.Value {
		names[i] = cube.Name
	}

	return names, nil
}

// GetAllNamesWithoutRules retrieves all cube names without rules
func (cs *CubeService) GetAllNamesWithoutRules(ctx context.Context, skipControlCube bool) ([]string, error) {
	endpoint := "/Cubes?$select=Name,Rules&$filter=Rules eq null"
	if skipControlCube {
		endpoint = "/ModelCubes()?$select=Name,Rules&$filter=Rules eq null"
	}

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, cube := range response.Value {
		names[i] = cube.Name
	}

	return names, nil
}

// GetDimensionNames retrieves the dimension names for a cube
func (cs *CubeService) GetDimensionNames(ctx context.Context, cubeName string) ([]string, error) {
	endpoint := fmt.Sprintf("/Cubes('%s')/Dimensions?$select=Name", url.PathEscape(cubeName))

	var response struct {
		Value []models.Dimension `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, dim := range response.Value {
		names[i] = dim.Name
	}

	return names, nil
}

// SearchForDimension finds cubes that contain a given dimension
func (cs *CubeService) SearchForDimension(ctx context.Context, dimensionName string, skipControlCube bool) ([]string, error) {
	cleaned := strings.ToLower(strings.ReplaceAll(dimensionName, " ", ""))
	endpoint := fmt.Sprintf("/Cubes?$select=Name&$filter=Dimensions/any(d: replace(tolower(d/Name), ' ', '') eq '%s')", cleaned)
	if skipControlCube {
		endpoint = fmt.Sprintf("/ModelCubes()?$select=Name&$filter=Dimensions/any(d: replace(tolower(d/Name), ' ', '') eq '%s')", cleaned)
	}

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, cube := range response.Value {
		names[i] = cube.Name
	}

	return names, nil
}

// SearchForDimensionSubstring finds cubes with dimensions matching a substring
func (cs *CubeService) SearchForDimensionSubstring(ctx context.Context, substring string, skipControlCube bool) (map[string][]string, error) {
	cleaned := strings.ToLower(strings.ReplaceAll(substring, " ", ""))
	endpoint := fmt.Sprintf("/Cubes?$select=Name&$filter=Dimensions/any(d: contains(replace(tolower(d/Name), ' ', ''),'%s'))"+
		"&$expand=Dimensions($select=Name;$filter=contains(replace(tolower(Name), ' ', ''), '%s'))", cleaned, cleaned)
	if skipControlCube {
		endpoint = fmt.Sprintf("/ModelCubes()?$select=Name&$filter=Dimensions/any(d: contains(replace(tolower(d/Name), ' ', ''),'%s'))"+
			"&$expand=Dimensions($select=Name;$filter=contains(replace(tolower(Name), ' ', ''), '%s'))", cleaned, cleaned)
	}

	var response struct {
		Value []models.Cube `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}
	for _, cube := range response.Value {
		result[cube.Name] = make([]string, 0, len(cube.Dimensions))
		for _, dim := range cube.Dimensions {
			result[cube.Name] = append(result[cube.Name], dim.Name)
		}
	}

	return result, nil
}

// SearchForRuleSubstring finds cubes whose rules contain a substring
func (cs *CubeService) SearchForRuleSubstring(ctx context.Context, substring string, skipControlCube bool, caseInsensitive bool, spaceInsensitive bool) ([]models.Cube, error) {
	cleaned := strings.ToLower(strings.ReplaceAll(substring, " ", ""))
	endpoint := "/Cubes"
	if skipControlCube {
		endpoint = "/ModelCubes()"
	}

	filter := "Rules ne null and contains("
	if caseInsensitive && spaceInsensitive {
		filter += fmt.Sprintf("tolower(replace(Rules, ' ', '')),'%s')", cleaned)
	} else if caseInsensitive {
		filter += fmt.Sprintf("tolower(Rules),'%s')", cleaned)
	} else if spaceInsensitive {
		filter += fmt.Sprintf("replace(Rules, ' ', ''),'%s')", cleaned)
	} else {
		filter += fmt.Sprintf("Rules,'%s')", substring)
	}

	endpoint = fmt.Sprintf("%s?$filter=%s&$expand=Dimensions($select=Name)", endpoint, filter)

	var response struct {
		Value []models.Cube `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetStorageDimensionOrder retrieves the storage order of dimensions for a cube
func (cs *CubeService) GetStorageDimensionOrder(ctx context.Context, cubeName string) ([]string, error) {
	if IsV1GreaterOrEqualToV2("11.4.0", cs.rest.version) {
		return nil, fmt.Errorf("GetStorageDimensionOrder requires TM1 v11.4 or greater")
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.DimensionsStorageOrder()?$select=Name", url.PathEscape(cubeName))

	var response struct {
		Value []models.Dimension `json:"value"`
	}

	err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, dim := range response.Value {
		names[i] = dim.Name
	}

	return names, nil
}

// UpdateStorageDimensionOrder updates the storage order of cube dimensions
func (cs *CubeService) UpdateStorageDimensionOrder(ctx context.Context, cubeName string, dimensions []string) (float64, error) {
	if IsV1GreaterOrEqualToV2("11.4.0", cs.rest.version) {
		return 0, fmt.Errorf("UpdateStorageDimensionOrder requires TM1 v11.4 or greater")
	}

	isDataAdmin, err := cs.isDataAdmin(ctx)
	if err != nil {
		return 0, err
	}
	if !isDataAdmin {
		return 0, fmt.Errorf("UpdateStorageDimensionOrder requires Data Admin privilege")
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.ReorderDimensions", url.PathEscape(cubeName))

	payload := map[string][]string{}
	payload["Dimensions@odata.bind"] = make([]string, 0, len(dimensions))
	for _, name := range dimensions {
		payload["Dimensions@odata.bind"] = append(payload["Dimensions@odata.bind"], fmt.Sprintf("Dimensions('%s')", name))
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	resp, err := cs.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response struct {
		Value float64 `json:"value"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	return response.Value, nil
}

// Load loads a cube into memory
func (cs *CubeService) Load(ctx context.Context, cubeName string) error {
	if IsV1GreaterOrEqualToV2("11.6.0", cs.rest.version) {
		return fmt.Errorf("Load requires TM1 v11.6 or greater")
	}

	isDataAdmin, err := cs.isDataAdmin(ctx)
	if err != nil {
		return err
	}
	if !isDataAdmin {
		return fmt.Errorf("Load requires Data Admin privilege")
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Load", url.PathEscape(cubeName))
	resp, err := cs.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Unload unloads a cube from memory
func (cs *CubeService) Unload(ctx context.Context, cubeName string) error {
	if IsV1GreaterOrEqualToV2("11.6.0", cs.rest.version) {
		return fmt.Errorf("Unload requires TM1 v11.6 or greater")
	}

	isDataAdmin, err := cs.isDataAdmin(ctx)
	if err != nil {
		return err
	}
	if !isDataAdmin {
		return fmt.Errorf("Unload requires Data Admin privilege")
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Unload", url.PathEscape(cubeName))
	resp, err := cs.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Lock locks a cube to prevent modifications by users
func (cs *CubeService) Lock(ctx context.Context, cubeName string) error {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Lock", url.PathEscape(cubeName))
	resp, err := cs.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Unlock unlocks a cube to allow modifications by users
func (cs *CubeService) Unlock(ctx context.Context, cubeName string) error {
	endpoint := fmt.Sprintf("/Cubes('%s')/tm1.Unlock", url.PathEscape(cubeName))
	resp, err := cs.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CubeSaveData serializes a cube by saving data updates
func (cs *CubeService) CubeSaveData(ctx context.Context, cubeName string) error {
	process := models.NewProcess("")
	process.EpilogProcedure = fmt.Sprintf("CubeSaveData('%s');", cubeName)

	success, status, _, err := cs.process.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Process did not complete successfully. Result: %v", status)
	}

	return nil
}

func (cs *CubeService) isDataAdmin(ctx context.Context) (bool, error) {
	var user map[string]interface{}
	err := cs.rest.JSON(ctx, "GET", "/ActiveUser", nil, &user)
	if err != nil {
		return false, err
	}

	if isDataAdmin, ok := user["IsDataAdmin"].(bool); ok {
		return isDataAdmin, nil
	}

	return false, nil
}
