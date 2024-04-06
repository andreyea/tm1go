package tm1go

type Cell struct {
	Ordinal             int           `json:"Ordinal,omitempty"`
	Status              string        `json:"Status,omitempty"`
	Value               interface{}   `json:"Value,omitempty"`
	FormatString        string        `json:"FormatString,omitempty"`
	FormattedValue      string        `json:"FormattedValue,omitempty"`
	Updateable          int           `json:"Updateable,omitempty"`
	RuleDerived         bool          `json:"RuleDerived,omitempty"`
	Annotated           bool          `json:"Annotated,omitempty"`
	Consolidated        bool          `json:"Consolidated,omitempty"`
	NullIntersected     bool          `json:"NullIntersected,omitempty"`
	Language            int           `json:"Language,omitempty"`
	HasPicklist         bool          `json:"HasPicklist,omitempty"`
	PicklistValues      []interface{} `json:"PicklistValues,omitempty"`
	HasDrillthrough     bool          `json:"HasDrillthrough,omitempty"`
	Members             []Member      `json:"Members,omitempty"`
	DrillthroughScripts []interface{} `json:"DrillthroughScripts,omitempty"`
	Annotations         []interface{} `json:"Annotations,omitempty"`
}
