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
	"time"

	"github.com/andreyea/tm1go/pkg/models"
)

// ProcessService handles operations for TM1 Processes (TurboIntegrator)
type ProcessService struct {
	rest *RestService
}

// NewProcessService creates a new ProcessService instance
func NewProcessService(rest *RestService) *ProcessService {
	return &ProcessService{
		rest: rest,
	}
}

// Get retrieves a process from TM1 Server
func (ps *ProcessService) Get(ctx context.Context, processName string) (*models.Process, error) {
	endpoint := fmt.Sprintf("/Processes('%s')?$select=*,UIData,VariablesUIData,"+
		"DataSource/dataSourceNameForServer,"+
		"DataSource/dataSourceNameForClient,"+
		"DataSource/asciiDecimalSeparator,"+
		"DataSource/asciiDelimiterChar,"+
		"DataSource/asciiDelimiterType,"+
		"DataSource/asciiHeaderRecords,"+
		"DataSource/asciiQuoteCharacter,"+
		"DataSource/asciiThousandSeparator,"+
		"DataSource/view,"+
		"DataSource/query,"+
		"DataSource/userName,"+
		"DataSource/password,"+
		"DataSource/usesUnicode,"+
		"DataSource/subset,"+
		"DataSource/jsonRootPointer,"+
		"DataSource/jsonVariableMapping",
		url.PathEscape(processName))

	var process models.Process
	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &process)
	if err != nil {
		return nil, err
	}

	return &process, nil
}

// GetAll retrieves all processes from TM1 Server
func (ps *ProcessService) GetAll(ctx context.Context, skipControlProcesses bool) ([]*models.Process, error) {
	query := url.Values{}
	query.Set("$select", "*,UIData,VariablesUIData,"+
		"DataSource/dataSourceNameForServer,"+
		"DataSource/dataSourceNameForClient,"+
		"DataSource/asciiDecimalSeparator,"+
		"DataSource/asciiDelimiterChar,"+
		"DataSource/asciiDelimiterType,"+
		"DataSource/asciiHeaderRecords,"+
		"DataSource/asciiQuoteCharacter,"+
		"DataSource/asciiThousandSeparator,"+
		"DataSource/view,"+
		"DataSource/query,"+
		"DataSource/userName,"+
		"DataSource/password,"+
		"DataSource/usesUnicode,"+
		"DataSource/subset,"+
		"DataSource/jsonRootPointer,"+
		"DataSource/jsonVariableMapping")
	if skipControlProcesses {
		query.Set("$filter", "startswith(Name,'}') eq false and startswith(Name,'{') eq false")
	}

	endpoint := "/Processes?" + EncodeODataQuery(query)

	var response struct {
		Value []*models.Process `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetAllNames retrieves all process names from TM1 Server
func (ps *ProcessService) GetAllNames(ctx context.Context, skipControlProcesses bool) ([]string, error) {
	query := url.Values{}
	query.Set("$select", "Name")
	if skipControlProcesses {
		query.Set("$filter", "startswith(Name,'}') eq false and startswith(Name,'{') eq false")
	}

	endpoint := "/Processes?" + EncodeODataQuery(query)

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, p := range response.Value {
		names[i] = p.Name
	}

	return names, nil
}

// SearchStringInCode searches for processes containing a string in their code
func (ps *ProcessService) SearchStringInCode(ctx context.Context, searchString string, skipControlProcesses bool) ([]string, error) {
	searchString = strings.ToLower(strings.ReplaceAll(searchString, " ", ""))

	filter := fmt.Sprintf(
		"contains(tolower(replace(PrologProcedure, ' ', '')),'%s') "+
			"or contains(tolower(replace(MetadataProcedure, ' ', '')),'%s') "+
			"or contains(tolower(replace(DataProcedure, ' ', '')),'%s') "+
			"or contains(tolower(replace(EpilogProcedure, ' ', '')),'%s')",
		searchString, searchString, searchString, searchString)

	if skipControlProcesses {
		filter += " and (startswith(Name,'}') eq false and startswith(Name,'{') eq false)"
	}

	query := url.Values{}
	query.Set("$select", "Name")
	query.Set("$filter", filter)
	endpoint := "/Processes?" + EncodeODataQuery(query)

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, p := range response.Value {
		names[i] = p.Name
	}

	return names, nil
}

// SearchStringInName searches for processes by name patterns
func (ps *ProcessService) SearchStringInName(ctx context.Context, nameStartsWith string, nameContains []string, nameContainsOperator string, skipControlProcesses bool) ([]string, error) {
	nameContainsOperator = strings.ToLower(strings.TrimSpace(nameContainsOperator))
	if nameContainsOperator != "and" && nameContainsOperator != "or" {
		return nil, fmt.Errorf("nameContainsOperator must be either 'and' or 'or'")
	}

	nameFilters := []string{}

	if nameStartsWith != "" {
		nameFilters = append(nameFilters, fmt.Sprintf("startswith(toupper(Name),toupper('%s'))", nameStartsWith))
	}

	if len(nameContains) > 0 {
		containsFilters := []string{}
		for _, wildcard := range nameContains {
			containsFilters = append(containsFilters, fmt.Sprintf("contains(toupper(Name),toupper('%s'))", wildcard))
		}
		nameFilters = append(nameFilters, fmt.Sprintf("(%s)", strings.Join(containsFilters, " "+nameContainsOperator+" ")))
	}

	endpoint := "/Processes?$select=Name"
	query := url.Values{}
	query.Set("$select", "Name")
	if len(nameFilters) > 0 {
		query.Set("$filter", strings.Join(nameFilters, " and "))
	}

	if skipControlProcesses {
		if existing := query.Get("$filter"); existing != "" {
			query.Set("$filter", existing+" and (startswith(Name,'}') eq false and startswith(Name,'{') eq false)")
		} else {
			query.Set("$filter", "(startswith(Name,'}') eq false and startswith(Name,'{') eq false)")
		}
	}

	endpoint = "/Processes?" + EncodeODataQuery(query)

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, p := range response.Value {
		names[i] = p.Name
	}

	return names, nil
}

// Create creates a new process on TM1 Server
func (ps *ProcessService) Create(ctx context.Context, process *models.Process) error {
	// Adjust process body if TM1 version is lower than 11
	version := ps.rest.version
	if len(version) >= 2 {
		majorVersion, err := strconv.Atoi(version[0:2])
		if err == nil && majorVersion < 11 {
			process.DropParameterTypes()
		}
	}

	bodyJSON, err := json.Marshal(process)
	if err != nil {
		return fmt.Errorf("failed to marshal process: %w", err)
	}

	resp, err := ps.rest.Post(ctx, "/Processes", bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create process: %s", string(body))
	}

	return nil
}

// Update updates an existing process on TM1 Server
func (ps *ProcessService) Update(ctx context.Context, process *models.Process) error {
	// Adjust process body if TM1 version is lower than 11
	version := ps.rest.version
	if len(version) >= 2 {
		majorVersion, err := strconv.Atoi(version[0:2])
		if err == nil && majorVersion < 11 {
			process.DropParameterTypes()
		}
	}

	bodyJSON, err := json.Marshal(process)
	if err != nil {
		return fmt.Errorf("failed to marshal process: %w", err)
	}

	endpoint := fmt.Sprintf("/Processes('%s')", url.PathEscape(process.Name))
	resp, err := ps.rest.Patch(ctx, endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update process: %s", string(body))
	}

	return nil
}

// UpdateOrCreate updates or creates a process on TM1 Server
func (ps *ProcessService) UpdateOrCreate(ctx context.Context, process *models.Process) error {
	exists, err := ps.Exists(ctx, process.Name)
	if err != nil {
		return err
	}

	if exists {
		return ps.Update(ctx, process)
	}

	return ps.Create(ctx, process)
}

// Delete deletes a process from TM1 Server
func (ps *ProcessService) Delete(ctx context.Context, processName string) error {
	endpoint := fmt.Sprintf("/Processes('%s')", url.PathEscape(processName))
	resp, err := ps.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Exists checks if a process exists on TM1 Server
func (ps *ProcessService) Exists(ctx context.Context, processName string) (bool, error) {
	endpoint := fmt.Sprintf("/Processes('%s')", url.PathEscape(processName))
	resp, err := ps.rest.Get(ctx, endpoint)
	if err != nil {
		// Check if it's a 404 error
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// Compile compiles a process and returns syntax errors
func (ps *ProcessService) Compile(ctx context.Context, processName string) ([]interface{}, error) {
	endpoint := fmt.Sprintf("/Processes('%s')/tm1.Compile", url.PathEscape(processName))

	var response struct {
		Value []interface{} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "POST", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// CompileProcess compiles a process object and returns syntax errors
func (ps *ProcessService) CompileProcess(ctx context.Context, process *models.Process) ([]interface{}, error) {
	endpoint := "/CompileProcess"

	payload := map[string]interface{}{
		"Process": process,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var response struct {
		Value []interface{} `json:"value"`
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Value, nil
}

// Execute executes a process on TM1 Server
func (ps *ProcessService) Execute(ctx context.Context, processName string, parameters map[string]interface{}, timeout *time.Duration, cancelAtTimeout bool) error {
	endpoint := fmt.Sprintf("/Processes('%s')/tm1.Execute", url.PathEscape(processName))

	payload := map[string]interface{}{}
	if len(parameters) > 0 {
		params := []map[string]interface{}{}
		for name, value := range parameters {
			params = append(params, map[string]interface{}{
				"Name":  name,
				"Value": value,
			})
		}
		payload["Parameters"] = params
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to execute process: %s", string(body))
	}

	return nil
}

// ExecuteWithReturn executes a process and returns execution status
func (ps *ProcessService) ExecuteWithReturn(ctx context.Context, processName string, parameters map[string]interface{}, timeout *time.Duration, cancelAtTimeout bool) (bool, string, string, error) {
	endpoint := fmt.Sprintf("/Processes('%s')/tm1.ExecuteWithReturn?$expand=*", url.PathEscape(processName))

	payload := map[string]interface{}{}
	if len(parameters) > 0 {
		params := []map[string]interface{}{}
		for name, value := range parameters {
			params = append(params, map[string]interface{}{
				"Name":  name,
				"Value": value,
			})
		}
		payload["Parameters"] = params
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to marshal parameters: %w", err)
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", err
	}

	var result models.ProcessExecutionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	success := result.ProcessExecuteStatusCode == "CompletedSuccessfully"
	status := result.ProcessExecuteStatusCode
	errorLogFile := ""
	if result.ErrorLogFile != nil {
		errorLogFile = result.ErrorLogFile.Filename
	}

	return success, status, errorLogFile, nil
}

// ExecuteProcessWithReturn executes an unbound process object and returns execution status
func (ps *ProcessService) ExecuteProcessWithReturn(ctx context.Context, process *models.Process, parameters map[string]interface{}) (bool, string, string, error) {
	endpoint := "/ExecuteProcessWithReturn?$expand=*"

	// Update parameters if provided
	if len(parameters) > 0 {
		for name, value := range parameters {
			process.RemoveParameter(name)

			paramType := "String"
			if _, ok := value.(float64); ok {
				paramType = "Numeric"
			} else if _, ok := value.(int); ok {
				paramType = "Numeric"
			}

			process.AddParameter(name, name, value, paramType)
		}
	}

	payload := map[string]interface{}{
		"Process": process,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", err
	}

	var result models.ProcessExecutionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	success := result.ProcessExecuteStatusCode == "CompletedSuccessfully"
	status := result.ProcessExecuteStatusCode
	errorLogFile := ""
	if result.ErrorLogFile != nil {
		errorLogFile = result.ErrorLogFile.Filename
	}

	return success, status, errorLogFile, nil
}

// SearchErrorLogFilenames searches for error log filenames containing a search string
func (ps *ProcessService) SearchErrorLogFilenames(ctx context.Context, searchString string, top int, descending bool) ([]string, error) {
	query := url.Values{}
	query.Set("$select", "Filename")
	query.Set("$filter", fmt.Sprintf("contains(tolower(Filename), tolower('%s'))", searchString))
	if top > 0 {
		query.Set("$top", fmt.Sprintf("%d", top))
	}
	if descending {
		query.Set("$orderby", "Filename desc")
	}
	endpoint := "/ErrorLogFiles?" + EncodeODataQuery(query)

	var response struct {
		Value []struct {
			Filename string `json:"Filename"`
		} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	filenames := make([]string, len(response.Value))
	for i, f := range response.Value {
		filenames[i] = f.Filename
	}

	return filenames, nil
}

// GetErrorLogFilenames gets error log filenames for a specific process
func (ps *ProcessService) GetErrorLogFilenames(ctx context.Context, processName string, top int, descending bool) ([]string, error) {
	if processName != "" {
		exists, err := ps.Exists(ctx, processName)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("'%s' is not a valid process", processName)
		}
		return ps.SearchErrorLogFilenames(ctx, processName, top, descending)
	}

	return ps.SearchErrorLogFilenames(ctx, "", top, descending)
}

// GetErrorLogFileContent retrieves the content of an error log file
func (ps *ProcessService) GetErrorLogFileContent(ctx context.Context, filename string) (string, error) {
	endpoint := fmt.Sprintf("/ErrorLogFiles('%s')/Content", url.PathEscape(filename))

	resp, err := ps.rest.Get(ctx, endpoint)
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

// GetProcessErrorLogs gets all ProcessErrorLog entries for a process
func (ps *ProcessService) GetProcessErrorLogs(ctx context.Context, processName string) ([]interface{}, error) {
	endpoint := fmt.Sprintf("/Processes('%s')/ErrorLogs", url.PathEscape(processName))

	var response struct {
		Value []interface{} `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetLastMessageFromProcessErrorLog gets the latest error log message from a process
func (ps *ProcessService) GetLastMessageFromProcessErrorLog(ctx context.Context, processName string) (string, error) {
	logs, err := ps.GetProcessErrorLogs(ctx, processName)
	if err != nil {
		return "", err
	}

	if len(logs) == 0 {
		return "", nil
	}

	lastLog := logs[len(logs)-1].(map[string]interface{})
	timestamp := lastLog["Timestamp"].(string)

	endpoint := fmt.Sprintf("/Processes('%s')/ErrorLogs('%s')/Content",
		url.PathEscape(processName),
		url.PathEscape(timestamp))

	resp, err := ps.rest.Get(ctx, endpoint)
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

// DebugProcess starts a debug session for a process
func (ps *ProcessService) DebugProcess(ctx context.Context, processName string, parameters map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/Processes('%s')/tm1.Debug?$expand=Breakpoints,Thread,CallStack($expand=Variables,Process($select=Name))",
		url.PathEscape(processName))

	payload := map[string]interface{}{}
	if len(parameters) > 0 {
		params := []map[string]interface{}{}
		for name, value := range parameters {
			params = append(params, map[string]interface{}{
				"Name":  name,
				"Value": value,
			})
		}
		payload["Parameters"] = params
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// DebugStepOver runs a single statement in the process (does not debug child processes)
func (ps *ProcessService) DebugStepOver(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/tm1.StepOver", url.PathEscape(debugID))

	resp, err := ps.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Small delay for TM1 to process
	time.Sleep(100 * time.Millisecond)

	return ps.getDebugContext(ctx, debugID)
}

// DebugStepIn runs a single statement and steps into child processes
func (ps *ProcessService) DebugStepIn(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/tm1.StepIn", url.PathEscape(debugID))

	resp, err := ps.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Small delay for TM1 to process
	time.Sleep(100 * time.Millisecond)

	return ps.getDebugContext(ctx, debugID)
}

// DebugStepOut resumes execution until current process finishes
func (ps *ProcessService) DebugStepOut(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/tm1.StepOut", url.PathEscape(debugID))

	resp, err := ps.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Small delay for TM1 to process
	time.Sleep(100 * time.Millisecond)

	return ps.getDebugContext(ctx, debugID)
}

// DebugContinue resumes execution until next breakpoint
func (ps *ProcessService) DebugContinue(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/tm1.Continue", url.PathEscape(debugID))

	resp, err := ps.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Small delay for TM1 to process
	time.Sleep(100 * time.Millisecond)

	return ps.getDebugContext(ctx, debugID)
}

// getDebugContext is a helper to retrieve debug context
func (ps *ProcessService) getDebugContext(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')?$expand=Breakpoints,Thread,CallStack($expand=Variables,Process($select=Name))",
		url.PathEscape(debugID))

	var result map[string]interface{}
	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DebugGetBreakpoints retrieves all breakpoints for a debug session
func (ps *ProcessService) DebugGetBreakpoints(ctx context.Context, debugID string) ([]*models.ProcessDebugBreakpoint, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/Breakpoints", url.PathEscape(debugID))

	var response struct {
		Value []*models.ProcessDebugBreakpoint `json:"value"`
	}

	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// DebugAddBreakpoint adds a breakpoint to a debug session
func (ps *ProcessService) DebugAddBreakpoint(ctx context.Context, debugID string, breakpoint *models.ProcessDebugBreakpoint) error {
	return ps.DebugAddBreakpoints(ctx, debugID, []*models.ProcessDebugBreakpoint{breakpoint})
}

// DebugAddBreakpoints adds multiple breakpoints to a debug session
func (ps *ProcessService) DebugAddBreakpoints(ctx context.Context, debugID string, breakpoints []*models.ProcessDebugBreakpoint) error {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/Breakpoints", url.PathEscape(debugID))

	bodyDicts := make([]map[string]interface{}, len(breakpoints))
	for i, bp := range breakpoints {
		bodyDicts[i] = bp.BodyAsDict()
	}

	bodyJSON, err := json.Marshal(bodyDicts)
	if err != nil {
		return fmt.Errorf("failed to marshal breakpoints: %w", err)
	}

	resp, err := ps.rest.Post(ctx, endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DebugRemoveBreakpoint removes a breakpoint from a debug session
func (ps *ProcessService) DebugRemoveBreakpoint(ctx context.Context, debugID string, breakpointID int) error {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/Breakpoints('%d')",
		url.PathEscape(debugID), breakpointID)

	resp, err := ps.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DebugUpdateBreakpoint updates a breakpoint in a debug session
func (ps *ProcessService) DebugUpdateBreakpoint(ctx context.Context, debugID string, breakpoint *models.ProcessDebugBreakpoint) error {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')/Breakpoints('%d')",
		url.PathEscape(debugID), breakpoint.BreakpointID)

	bodyJSON, err := json.Marshal(breakpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal breakpoint: %w", err)
	}

	resp, err := ps.rest.Patch(ctx, endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DebugGetVariableValues retrieves all variable values in a debug session
func (ps *ProcessService) DebugGetVariableValues(ctx context.Context, debugID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')?$expand=CallStack($expand=Variables)",
		url.PathEscape(debugID))

	var response map[string]interface{}
	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	variables := make(map[string]interface{})
	if callStack, ok := response["CallStack"].([]interface{}); ok && len(callStack) > 0 {
		if firstCall, ok := callStack[0].(map[string]interface{}); ok {
			if vars, ok := firstCall["Variables"].([]interface{}); ok {
				for _, v := range vars {
					if varMap, ok := v.(map[string]interface{}); ok {
						name := varMap["Name"].(string)
						value := varMap["Value"]
						variables[name] = value
					}
				}
			}
		}
	}

	return variables, nil
}

// DebugGetSingleVariableValue retrieves a specific variable value in a debug session
func (ps *ProcessService) DebugGetSingleVariableValue(ctx context.Context, debugID, variableName string) (interface{}, error) {
	query := url.Values{}
	query.Set("$expand", fmt.Sprintf("CallStack($expand=Variables($filter=tolower(Name) eq '%s';$select=Value))", strings.ToLower(variableName)))
	endpoint := fmt.Sprintf("/ProcessDebugContexts('%s')?%s", url.PathEscape(debugID), EncodeODataQuery(query))

	var response map[string]interface{}
	err := ps.rest.JSON(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	if callStack, ok := response["CallStack"].([]interface{}); ok && len(callStack) > 0 {
		if firstCall, ok := callStack[0].(map[string]interface{}); ok {
			if vars, ok := firstCall["Variables"].([]interface{}); ok && len(vars) > 0 {
				if varMap, ok := vars[0].(map[string]interface{}); ok {
					return varMap["Value"], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("'%s' not found in collection", variableName)
}

// EvaluateBooleanTIExpression evaluates a boolean TI expression
func (ps *ProcessService) EvaluateBooleanTIExpression(ctx context.Context, formula string) (bool, error) {
	formula = strings.TrimSuffix(strings.TrimSpace(formula), ";")

	prologProcedure := fmt.Sprintf(`
if (~%s);
  ProcessQuit;
endif;
`, formula)

	process := models.NewProcess("")
	process.PrologProcedure = prologProcedure

	_, status, _, err := ps.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return false, err
	}

	if status == "QuitCalled" {
		return false, nil
	} else if status == "CompletedSuccessfully" {
		return true, nil
	}

	return false, fmt.Errorf("unexpected TI return status: '%s' for expression: '%s'", status, formula)
}

// EvaluateTIExpression evaluates a TI expression and returns the result as a string
func (ps *ProcessService) EvaluateTIExpression(ctx context.Context, formula string) (string, error) {
	// Remove leading "=" if present
	if strings.HasPrefix(formula, "=") {
		formula = formula[1:]
	}

	// Ensure semicolon at end
	formula = strings.TrimSpace(formula)
	if !strings.HasSuffix(formula, ";") {
		formula += ";"
	}

	prologProcedure := "sFunc = " + formula + "\nsDebug='Stop';"

	process := models.NewProcess("")
	process.PrologProcedure = prologProcedure

	// Compile to check for syntax errors
	syntaxErrors, err := ps.CompileProcess(ctx, process)
	if err != nil {
		return "", err
	}

	if len(syntaxErrors) > 0 {
		return "", fmt.Errorf("syntax errors: %v", syntaxErrors)
	}

	// Create temporary process
	err = ps.Create(ctx, process)
	if err != nil {
		return "", err
	}
	defer ps.Delete(ctx, process.Name)

	// Start debug session
	debugInfo, err := ps.DebugProcess(ctx, process.Name, nil)
	if err != nil {
		return "", err
	}

	debugID := debugInfo["ID"].(string)

	// Add breakpoint on sFunc variable
	breakpoint := &models.ProcessDebugBreakpoint{
		BreakpointID: 1,
		Type:         "ProcessDebugContextDataBreakpoint",
		Enabled:      true,
		HitMode:      "BreakAlways",
		VariableName: "sFunc",
	}

	err = ps.DebugAddBreakpoint(ctx, debugID, breakpoint)
	if err != nil {
		return "", err
	}

	// Continue to breakpoint
	_, err = ps.DebugContinue(ctx, debugID)
	if err != nil {
		return "", err
	}

	// Get variable values
	variables, err := ps.DebugGetVariableValues(ctx, debugID)
	if err != nil {
		return "", err
	}

	// Continue to finish
	ps.DebugContinue(ctx, debugID)

	if result, ok := variables["sFunc"]; ok {
		return fmt.Sprintf("%v", result), nil
	}

	return "", fmt.Errorf("unknown error: no formula result found")
}
