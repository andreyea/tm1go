package tm1

import "encoding/json"

//Cube
type Cube struct {
	OdataEtag         string      `json:"@odata.etag"`
	Name              string      `json:"Name"`
	Dimensions        []Dimension `json:"Dimensions"`
	Rules             string      `json:"Rules"`
	DrillthroughRules interface{} `json:"DrillthroughRules"`
	LastSchemaUpdate  string   `json:"LastSchemaUpdate"`
	LastDataUpdate   string     `json:"LastDataUpdate"`
}

//CubesResponse
type CubesResponse struct {
	OdataContext string `json:"@odata.context"`
	Value        []Cube `json:"value"`
}


//GetCubes function
func (s Tm1Session) GetCubes() ([]Cube, error) {
	cubes := CubesResponse{}
	req, err := s.NewRequest("GET", "/Cubes", nil)
	if err != nil {

		return nil, err
	}

	content, err := s.Do(req)
	_ = json.Unmarshal(content, &cubes)
	return cubes.Value, nil
}



//CubeCreate creates new cube
func (s Tm1Session) CubeCreate(cube Cube) error {

	var dims string

	for i, v := range cube.Dimensions {
		if len(cube.Dimensions) == i {
			dims = dims + `"Dimensions('` + v.Name + `')"`
		} else {
			dims = dims + `"Dimensions('` + v.Name + `')",`
		}

	}

	payload := `
	{
		"Name": "` + cube.Name + `",
		"Dimensions@odata.bind": [` + dims + `]
	}
	`

	_, err := s.Tm1SendHttpRequest("POST", "/Cubes", payload)

	if err != nil {
		return err
	}
	return nil
}