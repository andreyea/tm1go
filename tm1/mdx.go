package tm1

import "encoding/json"

//ExecuteMdx runs mdx query
func (s Tm1Session) ExecuteMdx(mdx string) (cellset Cellset, err error) {
	payload := `{"MDX":"` + mdx + `"}`

	cont, err := s.Tm1SendHttpRequest("POST", "/ExecuteMDX?$expand=Axes($expand=Hierarchies,Tuples($expand=Members)),Cells,Cube($select=Name;$expand=Dimensions($select=Name))", payload)
	if err != nil {
		return cellset, err
	}
	err = json.Unmarshal(cont, &cellset)
	if err != nil {
		return cellset, err
	}
	return cellset, nil
}
