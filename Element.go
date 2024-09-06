package tm1go

import (
	"encoding/json"
)

// Define ElementType
type ElementType int

func (e ElementType) String() string {
	switch e {
	case Numeric:
		return "Numeric"
	case String:
		return "String"
	case Consolidated:
		return "Consolidated"
	default:
		return "Numeric"
	}
}

// Enum values for ElementType
const (
	Numeric ElementType = iota + 1
	String
	Consolidated
)

type Element struct {
	Name       string                 `json:"Name,omitempty"`
	UniqueName string                 `json:"UniqueName,omitempty"`
	Type       string                 `json:"Type,omitempty"`
	Level      int                    `json:"Level,omitempty"`
	Index      int                    `json:"Index,omitempty"`
	Attributes map[string]interface{} `json:"Attributes,omitempty"`
	Hierarchy  Hierarchy              `json:"Hierarchy,omitempty"`
}

func (e *Element) getBody() (string, error) {
	bodyAsDict := make(map[string]string)
	bodyAsDict["Name"] = e.Name
	bodyAsDict["Type"] = e.Type
	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
