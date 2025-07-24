package tm1go

import "fmt"

type Edge struct {
	ParentName    string  `json:"ParentName"`
	ComponentName string  `json:"ComponentName"`
	Weight        float64 `json:"Weight"`
}

func (e *Edge) getBody() string {
	return `{
		"ParentName": "` + e.ParentName + `",
		"ComponentName": "` + e.ComponentName + `",
		"Weight": ` + fmt.Sprintf("%f", e.Weight) + `
	}`
}

//convert float to string
