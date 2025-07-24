package tm1go

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

func (h *Hierarchy) getBody(includeElementAttributes bool) (map[string]interface{}, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = h.Name

	if h.Elements != nil && len(h.Elements) > 0 {
		elements := make([]map[string]interface{}, 0)
		for _, element := range h.Elements {
			elementBody, err := element.getBody()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elementBody)
		}
		bodyAsDict["Elements"] = elements
	}

	if h.Edges != nil && len(h.Edges) > 0 {
		edges := make([]map[string]interface{}, 0)
		for _, edge := range h.Edges {
			edgeAsDict := map[string]interface{}{
				"ParentName":    edge.ParentName,
				"ComponentName": edge.ComponentName,
				"Weight":        edge.Weight,
			}
			edges = append(edges, edgeAsDict)
		}
		bodyAsDict["Edges"] = edges
	}

	if includeElementAttributes && h.ElementAttributes != nil && len(h.ElementAttributes) > 0 {
		bodyAsDict["ElementAttributes"] = h.ElementAttributes
	}

	return bodyAsDict, nil
}
