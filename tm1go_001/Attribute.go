package tm1go

type LocalizedAttribute struct {
	LocaleID   string      `json:"LocaleId"`
	Attributes []Attribute `json:"Attributes"`
}

type Attribute map[string]interface{}
