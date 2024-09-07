package tm1go

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

func (e *Element) getBody() (map[string]interface{}, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = e.Name
	bodyAsDict["Type"] = e.Type

	if e.UniqueName != "" {
		bodyAsDict["UniqueName"] = e.UniqueName
	}
	if e.Level != 0 {
		bodyAsDict["Level"] = e.Level
	}
	if e.Index != 0 {
		bodyAsDict["Index"] = e.Index
	}
	if e.Attributes != nil && len(e.Attributes) > 0 {
		bodyAsDict["Attributes"] = e.Attributes
	}

	return bodyAsDict, nil
}
