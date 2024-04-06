package tm1go

import "encoding/json"

type ElementAttribute struct {
	Name string `json:"Name"`
	Type string `json:"Type"`
}

func (e ElementAttribute) getBody() (string, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = e.Name
	bodyAsDict["Type"] = e.Type
	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
