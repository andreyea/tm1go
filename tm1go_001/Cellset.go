package tm1go

type Cellset struct {
	Cube  Cube          `json:"Cube,omitempty"`
	Cells []Cell        `json:"Cells,omitempty"`
	Axes  []CellsetAxis `json:"Axes,omitempty"`
	ID    string        `json:"ID,omitempty"`
}
