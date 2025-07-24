package tm1go

import (
	"encoding/json"
	"fmt"
	"time"
)

type Cube struct {
	OdataContext      string      `json:"@odata.context"`
	OdataEtag         string      `json:"@odata.etag"`
	Name              string      `json:"Name"`
	Rules             string      `json:"Rules"`
	DrillthroughRules string      `json:"DrillthroughRules"`
	LastSchemaUpdate  time.Time   `json:"LastSchemaUpdate"`
	LastDataUpdate    time.Time   `json:"LastDataUpdate"`
	Dimensions        []Dimension `json:"Dimensions"`
	Views             []View      `json:"Views"`
	PrivateViews      []View      `json:"PrivateViews"`
}

type RuleSyntaxError struct {
	LineNumber int    `json:"LineNumber"`
	Message    string `json:"Message"`
}

type CubeBody struct {
	Name       string
	Rules      string
	Dimensions []string
}

func (c *Cube) getBody() (string, error) {
	bodyAsDict := make(map[string]interface{})
	bodyAsDict["Name"] = c.Name
	bodyAsDict["Dimensions@odata.bind"] = []string{}

	for _, dimension := range c.Dimensions {
		bodyAsDict["Dimensions@odata.bind"] = append(bodyAsDict["Dimensions@odata.bind"].([]string), fmt.Sprintf("Dimensions('%v')", dimension.Name))
	}

	if c.Rules != "" {
		bodyAsDict["Rules"] = c.Rules
	}

	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
