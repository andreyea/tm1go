package tm1go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type ProcessService struct {
	rest   *RestService
	object *ObjectService
}

func NewProcessService(rest *RestService, object *ObjectService) *ProcessService {
	return &ProcessService{rest: rest, object: object}
}

// Get retrieves a process from TM1
func (ps *ProcessService) Get(processName string) (*Process, error) {
	url := fmt.Sprintf("/Processes('%v')?$select=*,UIData,VariablesUIData,DataSource/dataSourceNameForServer,DataSource/dataSourceNameForClient,DataSource/asciiDecimalSeparator,DataSource/asciiDelimiterChar,DataSource/asciiDelimiterType,DataSource/asciiHeaderRecords,DataSource/asciiQuoteCharacter,DataSource/asciiThousandSeparator,DataSource/view,DataSource/query,DataSource/userName,DataSource/password,DataSource/usesUnicode,DataSource/subset", processName)
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	process := &Process{}
	err = json.NewDecoder(response.Body).Decode(process)
	if err != nil {
		return nil, err
	}
	return process, nil
}

// GetAll retrieves all processes from TM1
func (ps *ProcessService) GetAll(skipControlProcesses bool) ([]Process, error) {
	url := "/Processes?$select=*,UIData,VariablesUIData," +
		"DataSource/dataSourceNameForServer," +
		"DataSource/dataSourceNameForClient," +
		"DataSource/asciiDecimalSeparator," +
		"DataSource/asciiDelimiterChar," +
		"DataSource/asciiDelimiterType," +
		"DataSource/asciiHeaderRecords," +
		"DataSource/asciiQuoteCharacter," +
		"DataSource/asciiThousandSeparator," +
		"DataSource/view," +
		"DataSource/query," +
		"DataSource/userName," +
		"DataSource/password," +
		"DataSource/usesUnicode," +
		"DataSource/subset"

	if !skipControlProcesses {
		modelProcessFilter := "&$filter=startswith(Name,'}') eq false and startswith(Name,'{') eq false"
		url += modelProcessFilter
	}

	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Process]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetAllNames retrieves all process names from TM1
func (ps *ProcessService) GetAllNames(skipControlProcesses bool) ([]string, error) {
	url := "/Processes?$select=Name"
	if !skipControlProcesses {
		modelProcessFilter := "&$filter=startswith(Name,'}') eq false and startswith(Name,'{') eq false"
		url += modelProcessFilter
	}
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Process]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(result.Value))
	for _, process := range result.Value {
		names = append(names, process.Name)
	}
	return names, nil
}

// SearchStringInCode searches for a string in the code of all processes in TM1
func (ps *ProcessService) SearchStringInCode(searchString string, skipControlProcesses bool) ([]string, error) {
	searchString = strings.ToLower(strings.ReplaceAll(searchString, " ", ""))
	url := fmt.Sprintf("/Processes?$select=Name&$filter="+
		"contains(tolower(replace(PrologProcedure, ' ', '')),'%v') "+
		"or contains(tolower(replace(MetadataProcedure, ' ', '')),'%v') "+
		"or contains(tolower(replace(DataProcedure, ' ', '')),'%v') "+
		"or contains(tolower(replace(EpilogProcedure, ' ', '')),'%v')", searchString, searchString, searchString, searchString)
	if !skipControlProcesses {
		modelProcessFilter := "&$filter=startswith(Name,'}') eq false and startswith(Name,'{') eq false"
		url += modelProcessFilter
	}
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Process]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(result.Value))
	for _, process := range result.Value {
		names = append(names, process.Name)
	}
	return names, nil
}

// Create process in TM1
func (ps *ProcessService) Create(process *Process) error {
	url := "/Processes"
	processBody, err := process.getBody()
	if err != nil {
		return err
	}
	_, err = ps.rest.POST(url, processBody, nil, 0, nil)
	return err
}

// Update process in TM1
func (ps *ProcessService) Update(process *Process) error {
	url := fmt.Sprintf("/Processes('%v')", process.Name)
	processBody, err := process.getBody()
	if err != nil {
		return err
	}
	_, err = ps.rest.PATCH(url, processBody, nil, 0, nil)
	return err
}

// Delete process from TM1
func (ps *ProcessService) Delete(processName string) error {
	url := fmt.Sprintf("/Processes('%v')", processName)
	_, err := ps.rest.DELETE(url, nil, 0, nil)
	return err
}

// Exists checks if the process exists in TM1
func (ps *ProcessService) Exists(processName string) (bool, error) {
	url := fmt.Sprintf("/Processes('%v')", processName)
	return ps.object.Exists(url)
}

// Compile process in TM1
func (ps *ProcessService) Compile(processName string) ([]ProcessSyntaxError, error) {
	url := fmt.Sprintf("/Processes('%v')/tm1.Compile", processName)
	response, err := ps.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[ProcessSyntaxError]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Compile unbound process
func (ps *ProcessService) CompileProcess(process Process) ([]ProcessSyntaxError, error) {
	url := "/CompileProcess"
	processBody, err := process.getBody()
	if err != nil {
		return nil, err
	}
	response, err := ps.rest.POST(url, "{\"Process\":"+processBody+"}", nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[ProcessSyntaxError]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// Execute process
func (ps *ProcessService) Execute(processName string, parameters map[string]interface{}, timeout time.Duration) error {
	url := fmt.Sprintf("/Processes('%v')/tm1.Execute", processName)
	parametersBody := "{}"
	if parameters != nil {
		parametersBody = "{\"Parameters\": ["
		for key, value := range parameters {
			parametersBody += fmt.Sprintf("{\"Name\":\"%v\",\"Value\":\"%v\"},", key, value)
		}
		parametersBody = strings.TrimSuffix(parametersBody, ",")
		parametersBody += "]}"
	}
	_, err := ps.rest.POST(url, parametersBody, nil, timeout, nil)
	return err
}

// Execute unboud process
func (ps *ProcessService) ExecuteProcessWithReturn(process Process, parameters map[string]interface{}, timeout time.Duration) (*ProcessExecuteResult, error) {
	url := "/ExecuteProcessWithReturn?$expand=*"
	for parameterName, parameterValue := range parameters {
		p := process.GetParameter(parameterName)
		paramType := ""
		if p != nil {
			paramType = p.Type
		} else {
			switch parameterValue.(type) {
			case string:
				paramType = "String"
			default:
				paramType = "Numeric"
			}
		}
		process.RemoveParameter(parameterName)
		process.AddParameter(parameterName, parameterName, parameterValue, paramType)
	}
	processBody, err := process.getBody()
	if err != nil {
		return nil, err
	}
	payload := fmt.Sprintf("{\"Process\":%v}", processBody)
	response, err := ps.rest.POST(url, payload, nil, timeout, nil)
	if err != nil {
		return nil, err
	}
	result := ProcessExecuteResult{}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Execute process with return
func (ps *ProcessService) ExecuteWithReturn(processName string, parameters map[string]interface{}, timeout time.Duration) (*ProcessExecuteResult, error) {
	url := fmt.Sprintf("/Processes('%v')/tm1.ExecuteWithReturn?$expand=*", processName)
	parametersBody := "{}"
	if parameters != nil {
		parametersBody = "{\"Parameters\": ["
		for key, value := range parameters {
			parametersBody += fmt.Sprintf("{\"Name\":\"%v\",\"Value\":\"%v\"},", key, value)
		}
		parametersBody = strings.TrimSuffix(parametersBody, ",")
		parametersBody += "]}"
	}
	response, err := ps.rest.POST(url, parametersBody, nil, timeout, nil)
	if err != nil {
		return nil, err
	}
	result := ProcessExecuteResult{}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Execute TI code
func (ps *ProcessService) ExecuteTICode(prologCode string, epilogCode string) (*ProcessExecuteResult, error) {
	processName := "}TM1go" + RandomString(10)
	process := NewProcess(processName)
	process.PrologProcedure = prologCode
	process.EpilogProcedure = epilogCode
	result, err := ps.ExecuteProcessWithReturn(*process, nil, 0)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Search error log filenames for given search string like a datestamp e.g. 20231201
func (ps *ProcessService) SearchErrorLogFilenames(searchString string, top int, descending bool) ([]string, error) {
	url := fmt.Sprintf("/ErrorLogFiles?select=Filename&$filter=contains(tolower(Filename), tolower('%v'))", searchString)
	if top > 0 {
		url += fmt.Sprintf("&$top=%v", top)
	}
	if descending {
		url += "&$orderby=Filename desc"
	}
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[map[string]string]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	filenames := make([]string, 0, len(result.Value))
	for _, file := range result.Value {
		filenames = append(filenames, file["Filename"])
	}
	return filenames, nil
}

func (ps *ProcessService) GetErrorLogFilenams(processName string, top int, descending bool) ([]string, error) {
	exists, err := ps.Exists(processName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Process %s does not exist", processName)
	}
	searchString := processName
	return ps.SearchErrorLogFilenames(searchString, top, descending)
}

func (ps *ProcessService) GetErrorLogFileContent(filename string) (string, error) {
	url := fmt.Sprintf("/ErrorLogFiles('%v')/Content", filename)
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	content := ""
	if response.Body != nil {
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, response.Body)
		if err != nil {
			return "", err
		}
		content = buf.String()
	}
	return content, nil
}

// Get all ProcessErrorLog entries for a process
func (ps *ProcessService) GetProcessErrorLogs(processName string) ([]ProcessErrorLog, error) {
	url := fmt.Sprintf("/Processes('%v')/ErrorLogs", processName)
	response, err := ps.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[ProcessErrorLog]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (ps *ProcessService) GetLastMessageFromProcessErrorLog(processName string) (string, error) {
	if isV1GreaterOrEqualToV2(ps.rest.version, "12.0.0") {
		err := fmt.Errorf("GetLastMessageFromProcessErrorLog is deprecated in version 12.0.0")
		return "", err
	}
	logs, err := ps.GetProcessErrorLogs(processName)
	if err != nil {
		return "", err
	}
	if len(logs) > 0 {
		lastLog := logs[len(logs)-1]
		url := fmt.Sprintf("/Processes('%v')/ErrorLogs('%v')/Content", processName, lastLog.Timestamp)
		response, err := ps.rest.GET(url, nil, 0, nil)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()
		content := ""
		if response.Body != nil {
			buf := new(bytes.Buffer)
			_, err := io.Copy(buf, response.Body)
			if err != nil {
				return "", err
			}
			content = buf.String()
		}
		return content, nil
	}
	return "", nil
}
