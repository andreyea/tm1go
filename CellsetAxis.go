package tm1go

type CellsetAxis struct {
	Ordinal     int     `json:"Ordinal"`
	Cardinality int     `json:"Cardinality"`
	Tuples      []Tuple `json:"Tuples"`
}
