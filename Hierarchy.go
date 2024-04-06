package tm1go

import "encoding/json"

type Hierarchy struct {
	Name              string                 `json:"Name"`
	UniqueName        string                 `json:"UniqueName,omitempty"`
	Elements          []Element              `json:"Elements,omitempty"`
	Edges             []Edge                 `json:"Edges,omitempty"`
	Visible           bool                   `json:"Visible,omitempty"`
	Cardinality       int                    `json:"Cardinality,omitempty"`
	Attributes        map[string]interface{} `json:"Attributes,omitempty"`
	Dimension         Dimension              `json:"Dimension,omitempty"`
	Subsets           []Subset               `json:"Subset,omitempty"`
	PrivateSubsets    []Subset               `json:"PrivateSubsets,omitempty"`
	SessionSubsets    []Subset               `json:"SessionSubsets,omitempty"`
	Members           []Member               `json:"Members,omitempty"`
	AllMember         []Member               `json:"AllMember,omitempty"`
	DefaultMember     Member                 `json:"DefaultMember,omitempty"`
	Levels            []Level                `json:"Levels,omitempty"`
	ElementAttributes []ElementAttribute     `json:"ElementAttributes,omitempty"`
}

func (h *Hierarchy) getBody(includeElementAttributes bool) (string, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = h.Name

	if h.Elements != nil && len(h.Elements) > 0 {
		bodyAsDict["Elements"] = make([]map[string]interface{}, 0)
		for _, element := range h.Elements {
			elementBody, err := element.getBody()
			if err != nil {
				return "", err
			}
			bodyAsDict["Elements"] = append(bodyAsDict["Elements"].([]interface{}), elementBody)
		}
	}

	if h.Edges != nil && len(h.Edges) > 0 {
		bodyAsDict["Edges"] = make([]map[string]interface{}, 0)
		for _, edge := range h.Edges {
			edgeAsDict := make(map[string]interface{})
			edgeAsDict["ParentName"] = edge.ParentName
			edgeAsDict["ComponentName"] = edge.ComponentName
			edgeAsDict["Weight"] = edge.Weight
			bodyAsDict["Edges"] = append(bodyAsDict["Edges"].([]map[string]interface{}), edgeAsDict)
		}
	}

	if includeElementAttributes && h.ElementAttributes != nil && len(h.ElementAttributes) > 0 {
		bodyAsDict["ElementAttributes"] = h.Attributes
	}

	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
