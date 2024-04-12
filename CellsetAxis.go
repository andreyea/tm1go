package tm1go

type CellsetAxis struct {
	Ordinal     int         `json:"Ordinal"`
	Cardinality int         `json:"Cardinality"`
	Hierarchies []Hierarchy `json:"Hierarchies"`
	Tuples      []Tuple     `json:"Tuples"`
}
