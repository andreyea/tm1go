package tm1

import (
	"encoding/json"
)

//ProcessExecuteResult describes result after a process finished execution
type ProcessExecuteResult struct {
	ProcessExecuteStatusCode string `json:ProcessExecuteStatusCode`
}

//ExecuteProcessResponse error response from executing process
type ExecuteProcessResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

//Process describes tm1 TI process
type Process struct {
	OdataContext      string `json:"@odata.context"`
	OdataEtag         string `json:"@odata.etag"`
	Name              string `json:"Name"`
	HasSecurityAccess bool   `json:"HasSecurityAccess"`
	PrologProcedure   string `json:"PrologProcedure"`
	MetadataProcedure string `json:"MetadataProcedure"`
	DataProcedure     string `json:"DataProcedure"`
	EpilogProcedure   string `json:"EpilogProcedure"`
	DataSource        struct {
		Type                    string `json:"Type"`
		ASCIIDecimalSeparator   string `json:"asciiDecimalSeparator"`
		ASCIIDelimiterChar      string `json:"asciiDelimiterChar"`
		ASCIIDelimiterType      string `json:"asciiDelimiterType"`
		ASCIIHeaderRecords      int    `json:"asciiHeaderRecords"`
		ASCIIQuoteCharacter     string `json:"asciiQuoteCharacter"`
		ASCIIThousandSeparator  string `json:"asciiThousandSeparator"`
		DataSourceNameForClient string `json:"dataSourceNameForClient"`
		DataSourceNameForServer string `json:"dataSourceNameForServer"`
	} `json:"DataSource"`
	Parameters []ProcessParameter `json:"Parameters"`
	Variables  []struct {
		Name      string `json:"Name"`
		Type      string `json:"Type"`
		Position  int    `json:"Position"`
		StartByte int    `json:"StartByte"`
		EndByte   int    `json:"EndByte"`
	} `json:"Variables"`
	Attributes struct {
		Caption string `json:"Caption"`
	} `json:"Attributes"`
}

//ProcessParameter describes one parameter in a TI process when received
type ProcessParameter struct {
	Name   string      `json:"Name"`
	Prompt string      `json:"Prompt"`
	Value  interface{} `json:"Value"`
	Type   string      `json:"Type"`
}

//ProcessParameterExecute describes one parameter in a TI process when sent
type ProcessParameterExecute struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

//ProcessesReponse describes response from getProcesses method
type ProcessesReponse struct {
	OdataContext string    `json:"@odata.context"`
	Value        []Process `json:"value"`
}

func (s Tm1Session) GetProcesses() ([]Process, error) {

	processes := ProcessesReponse{}
	res, err := s.Tm1SendHttpRequest("GET", "/Processes", nil)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(res, &processes)

	return processes.Value, nil
}

//ExecuteProcess runs a process in TM1
func (s Tm1Session) ExecuteProcess(process Process) (ProcessExecuteResult, error) {
	parameters := []ProcessParameterExecute{}

	for _, v := range process.Parameters {
		parameters = append(parameters, ProcessParameterExecute{v.Name, v.Value})
	}

	result := ProcessExecuteResult{}

	p1, _ := json.Marshal(parameters)
	payload := `{"Parameters":` + string(p1) + `}`

	res, err := s.Tm1SendHttpRequest("POST", "/Processes('"+process.Name+"')/tm1.ExecuteWithReturn", payload)
	if err != nil {
		return result, err
	}
	json.Unmarshal(res, &result)
	return result, nil
}
