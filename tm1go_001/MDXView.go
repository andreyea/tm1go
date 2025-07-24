package tm1go

import "encoding/json"

type MDXView struct {
	Cube Cube   `json:"Cube,omitempty"`
	Type string `json:"@odata.type"`
	Name string `json:"Name"`
	MDX  string `json:"MDX"`
	Meta struct {
		Aliases      map[string]string            `json:"Aliases"`
		ContextSets  map[string]map[string]string `json:"ContextSets"`
		ExpandAboves map[string]bool              `json:"ExpandAboves"`
	} `json:"Meta,omitempty"`
}

func (v *MDXView) GetType() string {
	return v.Type
}

func (v *MDXView) GetName() string {
	return v.Name
}

func (v *MDXView) getBody(static bool) (string, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["@odata.type"] = v.Type
	bodyAsDict["Name"] = v.Name
	bodyAsDict["MDX"] = v.MDX
	bodyAsDict["Meta"] = v.Meta

	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
