package tm1go

type Level struct {
	Number      int    `json:"Number"`
	Name        string `json:"Name"`
	UniqueName  string `json:"UniqueName"`
	Cardinality int    `json:"Cardinality"`
	Type        int    `json:"Type"`
}
