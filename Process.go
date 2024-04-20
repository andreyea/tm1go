package tm1go

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

const BeginGeneratedStatements = "#****Begin: Generated Statements***"
const EndGeneratedStatements = "#****End: Generated Statements****"

// Process represents a TurboIntegrator process that can be used to manipulate TM1 data and metadata.
type Process struct {
	Name                string               `json:"Name"`                          // The name of a TurboIntegrator process.
	HasSecurityAccess   bool                 `json:"HasSecurityAccess,omitempty"`   // A Boolean that indicates whether the user security access to run this process.
	PrologProcedure     string               `json:"PrologProcedure,omitempty"`     // The code that is executed during the Prolog stage of the process.
	MetadataProcedure   string               `json:"MetadataProcedure,omitempty"`   // The code that is executed during the Metadata stage of the process.
	DataProcedure       string               `json:"DataProcedure,omitempty"`       // The code that is executed during the Data stage of the process.
	EpilogProcedure     string               `json:"EpilogProcedure,omitempty"`     // The code that is executed during the Epilog stage of the process.
	DataSource          ProcessDataSource    `json:"DataSource,omitempty"`          // The source of the data for the process. Assuming tm1.ProcessDataSource is a string for simplicity.
	Parameters          []ProcessParameter   `json:"Parameters,omitempty"`          // Parameters used by the process. Assuming Collection(tm1.ProcessParameter) is a slice of strings for simplicity.
	Variables           []ProcessVariable    `json:"Variables,omitempty"`           // Variables used by the process. Assuming Collection(tm1.ProcessVariable) is a slice of strings for simplicity.
	Attributes          Attribute            `json:"Attributes,omitempty"`          // The attributes of the process. Assuming tm1.Attributes is a string for simplicity.
	LocalizedAttributes []LocalizedAttribute `json:"LocalizedAttributes,omitempty"` // Translated string values for properties of the process.
	ErrorLogs           []ProcessErrorLog    `json:"ErrorLogs,omitempty"`           // A collection of error logs for the process.
	UIData              string               `json:"UIData,omitempty"`              // string "CubeAction=1511\fDataAction=1503\fCubeLogChanges=0\f"
	VariablesUIData     []interface{}        `json:"VariablesUIData,omitempty"`
}

func NewProcess(name string) *Process {
	p := &Process{
		Name:              name,
		HasSecurityAccess: false,
		DataSource:        ProcessDataSource{Type: "None"},
		UIData:            "CubeAction=1511\fDataAction=1503\fCubeLogChanges=0\f",
		Variables:         []ProcessVariable{},
		Parameters:        []ProcessParameter{},
		VariablesUIData:   []interface{}{},
	}
	return p
}

// ProcessErrorLog represents a collection of error logs for the process.
type ProcessErrorLog struct {
	Timestamp time.Time `json:"Timestamp,omitempty"` // The date and time of the process error.
	Content   []byte    `json:"Content,omitempty"`   // The content of the process error.
}

type ProcessSyntaxError struct {
	Procedure  string `json:"Procedure,omitempty"`
	LineNumber int    `json:"LineNumber,omitempty"`
	Message    string `json:"Message,omitempty"`
}

type NameValuePair struct {
	Name  string      `json:"Name,omitempty"`
	Value interface{} `json:"Value,omitempty"`
}

// ProcessParameter represents a parameter used by the process, which can have a value of different types.
type ProcessParameter struct {
	Name   string      `json:"Name,omitempty"`   // The name of the parameter, cannot be null.
	Prompt string      `json:"Prompt,omitempty"` // The prompt text for the parameter.
	Value  interface{} `json:"Value,omitempty"`  // The value of the parameter, can be Edm.Double or Edm.String. Use interface{} to accommodate both types.
	Type   string      `json:"Type,omitempty"`   // The type of the process variable, cannot be null. Assuming tm1.ProcessVariableType is represented as a string.
}

type ProcessDataSource struct {
	Type                    string `json:"Type,omitempty"`
	AsciiDecimalSeparator   string `json:"asciiDecimalSeparator,omitempty"`
	AsciiDelimiterChar      string `json:"asciiDelimiterChar,omitempty"`
	AsciiDelimiterType      string `json:"asciiDelimiterType,omitempty"` // "Character" or "FixedWidth"
	AsciiHeaderRecords      int    `json:"asciiHeaderRecords,omitempty"`
	AsciiQuoteCharacter     string `json:"asciiQuoteCharacter,omitempty"`
	AsciiThousandSeparator  string `json:"asciiThousandSeparator,omitempty"`
	DataSourceNameForClient string `json:"dataSourceNameForClient,omitempty"`
	DataSourceNameForServer string `json:"dataSourceNameForServer,omitempty"`
	UserName                string `json:"userName,omitempty"`
	Password                string `json:"password,omitempty"`
	Query                   string `json:"query,omitempty"`
	UsesUnicode             string `json:"usesUnicode,omitempty"`
	View                    string `json:"view,omitempty"`
	Subset                  string `json:"subset,omitempty"`
}

type ProcessVariable struct {
	Name      string `json:"Name"`                // The name of the process variable, cannot be null.
	Type      string `json:"Type"`                // The type of the process variable, cannot be null.
	Position  int    `json:"Position,omitempty"`  // The position of the variable in the data source.
	StartByte int    `json:"StartByte,omitempty"` // The starting byte position of the variable in the data source.
	EndByte   int    `json:"EndByte,omitempty"`   // The ending byte position of the variable in the data source.
}

type ErrorLogFile struct {
	Filename string `json:"Filename,omitempty"`
	Content  string `json:"Content,omitempty"`
}

type ProcessExecuteResult struct {
	ProcessExecuteStatusCode string       `json:"ProcessExecuteStatusCode"`
	ErrorLogFile             ErrorLogFile `json:"ErrorLogFile"`
}

func (p *Process) getBody() (string, error) {
	dataSource := ProcessDataSource{Type: p.DataSource.Type}

	switch p.DataSource.Type {
	case "ASCII":
		dataSource = ProcessDataSource{
			Type:                    p.DataSource.Type,
			AsciiDecimalSeparator:   p.DataSource.AsciiDecimalSeparator,
			AsciiDelimiterChar:      p.DataSource.AsciiDelimiterChar,
			AsciiDelimiterType:      p.DataSource.AsciiDelimiterType,
			AsciiHeaderRecords:      p.DataSource.AsciiHeaderRecords,
			AsciiQuoteCharacter:     p.DataSource.AsciiQuoteCharacter,
			AsciiThousandSeparator:  p.DataSource.AsciiThousandSeparator,
			DataSourceNameForClient: p.DataSource.DataSourceNameForClient,
			DataSourceNameForServer: p.DataSource.DataSourceNameForServer,
		}
		if p.DataSource.AsciiDelimiterType == "FixedWidth" {
			dataSource.AsciiDelimiterChar = ""
		}
	case "None":
		dataSource = ProcessDataSource{Type: "None"}
	case "ODBC", "TM1CubeView", "TM1DimensionSubset":
		dataSource = ProcessDataSource{
			Type:                    p.DataSource.Type,
			DataSourceNameForClient: p.DataSource.DataSourceNameForClient,
			DataSourceNameForServer: p.DataSource.DataSourceNameForServer,
			UserName:                p.DataSource.UserName,
			Password:                p.DataSource.Password,
			Query:                   p.DataSource.Query,
			UsesUnicode:             p.DataSource.UsesUnicode,
			View:                    p.DataSource.View,
			Subset:                  p.DataSource.Subset,
		}
	}

	params := interface{}(p.Parameters)
	if p.Parameters == nil {
		params = []ProcessParameter{}
	}
	vars := interface{}(p.Variables)
	if p.Variables == nil {
		vars = []ProcessVariable{}
	}
	varsUiData := interface{}(p.VariablesUIData)
	if p.VariablesUIData == nil {
		varsUiData = []interface{}{}
	}
	bodyAsDict := map[string]interface{}{
		"Name":              p.Name,
		"PrologProcedure":   AddGeneratedStringToCode(p.PrologProcedure),
		"MetadataProcedure": AddGeneratedStringToCode(p.MetadataProcedure),
		"DataProcedure":     AddGeneratedStringToCode(p.DataProcedure),
		"EpilogProcedure":   AddGeneratedStringToCode(p.EpilogProcedure),
		"HasSecurityAccess": p.HasSecurityAccess,
		"DataSource":        dataSource,
		"Parameters":        params,
		"Variables":         vars,
		"UIData":            p.UIData,
		"VariablesUIData":   varsUiData,
	}

	body, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (p *Process) AddParameter(name string, prompt string, value interface{}, parameterType string) {
	parameter := ProcessParameter{
		Name:   name,
		Prompt: prompt,
		Value:  value,
		Type:   parameterType,
	}
	p.Parameters = append(p.Parameters, parameter)
}

func (p *Process) RemoveParameter(name string) {
	for i, parameter := range p.Parameters {
		if parameter.Name == name {
			p.Parameters = append(p.Parameters[:i], p.Parameters[i+1:]...)
			break
		}
	}
}

func (p *Process) GetParameter(name string) *ProcessParameter {
	for _, parameter := range p.Parameters {
		if parameter.Name == name {
			return &parameter
		}
	}
	return nil
}

func AddGeneratedStringToCode(code string) string {
	pattern := `(?s)#\*\*\*\*Begin: Generated Statements(.*)#\*\*\*\*End: Generated Statements\*\*\*\*`
	matched, err := regexp.MatchString(pattern, code)
	if err != nil {
		return ""
	}
	if matched {
		return code
	} else {
		autoGeneratedStatements := fmt.Sprintf("%s\r\n%s\r\n", BeginGeneratedStatements, EndGeneratedStatements)
		return autoGeneratedStatements + code
	}
}
