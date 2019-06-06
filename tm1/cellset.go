package tm1

//Cell is a single cell of a cellset
type Cell struct {
	Ordinal        int         `json:"Ordinal"`
	Value          interface{} `json:"Value"`
	FormattedValue string      `json:"FormattedValue"`
}

//Axis of a view/mdx query
type Axis struct {
	Ordinal     int         `json:"Ordinal"`
	Cardinality int         `json:"Cardinality"`
	Hierarchies []Hierarchy `json:"Hierarchies"`
	Tuples      []Tuple     `json:"Tuples"`
}



//Member
type Member struct {
	Name          string `json:"Name"`
	UniqueName    string `json:"UniqueName"`
	Type          string `json:"Type"`
	Ordinal       int    `json:"Ordinal"`
	IsPlaceholder bool   `json:"IsPlaceholder"`
	Weight        int    `json:"Weight"`
}

//Tuple
type Tuple struct {
	Ordinal int      `json:"Ordinal"`
	Members []Member `json:"Members"`
}

//Cellset
type Cellset struct {
	OdataContext string `json:"@odata.context"`
	ID           string `json:"ID"`
	Cube         Cube   `json:"Cube"`
	Axes         []Axis `json:"Axes"`
	Cells        []Cell `json:"Cells"`
}