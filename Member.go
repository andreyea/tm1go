package tm1go

type Member struct {
	Name             string                 `json:"Name"`
	UniqueName       string                 `json:"UniqueName"`
	Type             string                 `json:"Type"`
	Ordinal          int                    `json:"Ordinal"`
	IsPlaceholder    bool                   `json:"IsPlaceholder"`
	Weight           int                    `json:"Weight"`
	Attributes       map[string]interface{} `json:"Attributes"`
	DisplayInfo      int                    `json:"DisplayInfo"`
	DisplayInfoAbove int                    `json:"DisplayInfoAbove"`
}
