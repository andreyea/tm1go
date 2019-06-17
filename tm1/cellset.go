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

//Member of a tuple
type Member struct {
	Name          string `json:"Name"`
	UniqueName    string `json:"UniqueName"`
	Type          string `json:"Type"`
	Ordinal       int    `json:"Ordinal"`
	IsPlaceholder bool   `json:"IsPlaceholder"`
	Weight        int    `json:"Weight"`
}

//Tuple makes up a column or a row
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

type Tm1Matrix [][]Tm1MatrixCell

type Tm1MatrixCell struct {
	NValue float64
	SValue string
}

//CreateMatrix method converts a cellset to a two dimensional matrix
func (c Cellset) CreateMatrix() Tm1Matrix {
	matrix := Tm1Matrix{}

	columnsNumber := c.Axes[0].Cardinality
	row := []Tm1MatrixCell{}
	for i, v := range c.Cells {

		NValue := 0.0
		SValue := ""
		switch t := v.Value.(type) {
		case float64:
			NValue = t
		case int64:
			NValue = float64(t)
		case string:
			SValue = string(t)
		default:
			return nil
		}

		if i%columnsNumber == 0 {
			matrix = append(matrix, row)
			row = []Tm1MatrixCell{}
		}
		row = append(row, Tm1MatrixCell{NValue: NValue, SValue: SValue})
	}
	return matrix
}
