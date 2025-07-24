package tm1go

import (
	"encoding/json"
	"fmt"
)

type Subset struct {
	Name        string                 `json:"Name,omitempty"`
	UniqueName  string                 `json:"UniqueName,omitempty"`
	Expression  string                 `json:"Expression,omitempty"`
	Attributes  map[string]interface{} `json:"Attributes,omitempty"`
	Alias       string                 `json:"Alias,omitempty"`
	Hierarchy   Hierarchy              `json:"Hierarchy,omitempty"`
	Elements    []Element              `json:"Elements,omitempty"`
	ExpandAbove bool                   `json:"ExpandAbove,omitempty"`
}

func (s *Subset) getBodyAsStatic() string {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = s.Name
	if s.Alias != "" {
		bodyAsDict["Alias"] = s.Alias
	}
	bodyAsDict["Hierarchy@odata.bind"] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')", s.Hierarchy.Dimension.Name, s.Hierarchy.Name)
	if len(s.Elements) > 0 {
		elementsBind := make([]string, len(s.Elements))
		for i, element := range s.Elements {
			elementsBind[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", s.Hierarchy.Dimension.Name, s.Hierarchy.Name, element.Name)
		}
		bodyAsDict["Elements@odata.bind"] = elementsBind
	}
	jsonData, _ := json.Marshal(bodyAsDict)

	return string(jsonData)
}
