package models

import (
	"encoding/json"
)

// Process represents a TM1 TurboIntegrator process
type Process struct {
	Name              string             `json:"Name"`
	HasSecurityAccess bool               `json:"HasSecurityAccess,omitempty"`
	PrologProcedure   string             `json:"PrologProcedure,omitempty"`
	MetadataProcedure string             `json:"MetadataProcedure,omitempty"`
	DataProcedure     string             `json:"DataProcedure,omitempty"`
	EpilogProcedure   string             `json:"EpilogProcedure,omitempty"`
	DataSource        *ProcessDataSource `json:"DataSource,omitempty"`
	Parameters        []ProcessParameter `json:"Parameters,omitempty"`
	Variables         []ProcessVariable  `json:"Variables,omitempty"`
	Attributes        map[string]string  `json:"Attributes,omitempty"`
	UIData            string             `json:"UIData,omitempty"`
	VariablesUIData   []ProcessUIData    `json:"VariablesUIData,omitempty"`
}

// ProcessDataSource represents the data source configuration for a process
type ProcessDataSource struct {
	Type                    string `json:"@odata.type,omitempty"`
	DataSourceNameForServer string `json:"dataSourceNameForServer,omitempty"`
	DataSourceNameForClient string `json:"dataSourceNameForClient,omitempty"`
	ASCIIDecimalSeparator   string `json:"asciiDecimalSeparator,omitempty"`
	ASCIIDelimiterChar      string `json:"asciiDelimiterChar,omitempty"`
	ASCIIDelimiterType      string `json:"asciiDelimiterType,omitempty"`
	ASCIIHeaderRecords      int    `json:"asciiHeaderRecords,omitempty"`
	ASCIIQuoteCharacter     string `json:"asciiQuoteCharacter,omitempty"`
	ASCIIThousandSeparator  string `json:"asciiThousandSeparator,omitempty"`
	View                    string `json:"view,omitempty"`
	Query                   string `json:"query,omitempty"`
	UserName                string `json:"userName,omitempty"`
	Password                string `json:"password,omitempty"`
	UsesUnicode             bool   `json:"usesUnicode,omitempty"`
	Subset                  string `json:"subset,omitempty"`
	JSONRootPointer         string `json:"jsonRootPointer,omitempty"`
	JSONVariableMapping     string `json:"jsonVariableMapping,omitempty"`
}

// ProcessParameter represents a parameter in a TM1 process
type ProcessParameter struct {
	Name   string      `json:"Name"`
	Prompt string      `json:"Prompt,omitempty"`
	Value  interface{} `json:"Value,omitempty"`
	Type   string      `json:"Type,omitempty"` // "Numeric" or "String"
}

// ProcessVariable represents a variable in a TM1 process
type ProcessVariable struct {
	Name      string `json:"Name"`
	Type      string `json:"Type"` // "Numeric" or "String"
	Position  int    `json:"Position,omitempty"`
	StartByte int    `json:"StartByte,omitempty"`
	EndByte   int    `json:"EndByte,omitempty"`
}

// ProcessUIData represents UI data for process variables
type ProcessUIData struct {
	Name  string `json:"Name"`
	Value string `json:"Value,omitempty"`
}

// ProcessDebugBreakpoint represents a breakpoint for process debugging
type ProcessDebugBreakpoint struct {
	BreakpointID  int    `json:"ID,omitempty"`
	Type          string `json:"@odata.type"`
	Enabled       bool   `json:"Enabled"`
	HitMode       string `json:"HitMode,omitempty"`
	VariableName  string `json:"VariableName,omitempty"`
	LineNumber    int    `json:"LineNumber,omitempty"`
	ProcedureType string `json:"ProcedureType,omitempty"`
}

// ProcessExecutionResult represents the result of a process execution
type ProcessExecutionResult struct {
	ProcessExecuteStatusCode string        `json:"ProcessExecuteStatusCode"`
	ErrorLogFile             *ErrorLogFile `json:"ErrorLogFile,omitempty"`
}

// ErrorLogFile represents an error log file reference
type ErrorLogFile struct {
	Filename string `json:"Filename"`
}

// NewProcess creates a new Process instance
func NewProcess(name string) *Process {
	return &Process{
		Name:       name,
		Parameters: make([]ProcessParameter, 0),
		Variables:  make([]ProcessVariable, 0),
		Attributes: make(map[string]string),
	}
}

// AddParameter adds a parameter to the process
func (p *Process) AddParameter(name, prompt string, value interface{}, paramType string) {
	param := ProcessParameter{
		Name:   name,
		Prompt: prompt,
		Value:  value,
		Type:   paramType,
	}
	p.Parameters = append(p.Parameters, param)
}

// RemoveParameter removes a parameter by name
func (p *Process) RemoveParameter(name string) {
	for i, param := range p.Parameters {
		if param.Name == name {
			p.Parameters = append(p.Parameters[:i], p.Parameters[i+1:]...)
			return
		}
	}
}

// AddVariable adds a variable to the process
func (p *Process) AddVariable(name, varType string, position int) {
	variable := ProcessVariable{
		Name:     name,
		Type:     varType,
		Position: position,
	}
	p.Variables = append(p.Variables, variable)
}

// Body returns the JSON representation for API requests
func (p *Process) Body() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// ProcessFromJSON creates a Process from JSON data
func ProcessFromJSON(jsonData []byte) (*Process, error) {
	var process Process
	err := json.Unmarshal(jsonData, &process)
	if err != nil {
		return nil, err
	}
	return &process, nil
}

// DropParameterTypes removes Type field from parameters (for TM1 < v11)
func (p *Process) DropParameterTypes() {
	for i := range p.Parameters {
		p.Parameters[i].Type = ""
	}
}

// ProcessDebugBreakpointFromJSON creates a ProcessDebugBreakpoint from JSON
func ProcessDebugBreakpointFromJSON(jsonData []byte) (*ProcessDebugBreakpoint, error) {
	var breakpoint ProcessDebugBreakpoint
	err := json.Unmarshal(jsonData, &breakpoint)
	if err != nil {
		return nil, err
	}
	return &breakpoint, nil
}

// Body returns the JSON representation for breakpoint API requests
func (b *ProcessDebugBreakpoint) Body() string {
	data, _ := json.Marshal(b)
	return string(data)
}

// BodyAsDict returns the breakpoint as a map for batch operations
func (b *ProcessDebugBreakpoint) BodyAsDict() map[string]interface{} {
	result := make(map[string]interface{})
	data, _ := json.Marshal(b)
	json.Unmarshal(data, &result)
	return result
}
