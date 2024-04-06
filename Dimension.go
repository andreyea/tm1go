package tm1go

import (
	"encoding/json"
	"strings"
)

type Dimension struct {
	OdataContext           string                 `json:"@odata.context,omitempty"`
	OdataEtag              string                 `json:"@odata.etag,omitempty"`
	Name                   string                 `json:"Name,omitempty"`
	UniqueName             string                 `json:"UniqueName,omitempty"`
	AllLeavesHierarchyName string                 `json:"AllLeavesHierarchyName,omitempty"`
	Attributes             map[string]interface{} `json:"Attributes,omitempty"`
	Hierarchies            []Hierarchy            `json:"Hierarchies,omitempty"`
}

// GetBody method to create body map
func (d *Dimension) getBody(includeLeavesHierarchy bool) (string, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = d.Name
	bodyAsDict["UniqueName"] = d.UniqueName
	if d.Attributes != nil && len(d.Attributes) > 0 {
		bodyAsDict["Attributes"] = d.Attributes
	}
	if d.Hierarchies != nil && len(d.Hierarchies) > 0 {
		var hierarchies []string
		for _, hierarchy := range d.Hierarchies {
			if strings.ToLower(hierarchy.Name) != "leaves" || includeLeavesHierarchy {
				hierarchyBody, err := hierarchy.getBody(false)
				if err != nil {
					return "", err
				}
				hierarchies = append(hierarchies, hierarchyBody)
			}
		}
		bodyAsDict["Hierarchies"] = hierarchies
	}
	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
