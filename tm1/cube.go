package tm1

import "encoding/json"

//Cube
type Cube struct {
	OdataEtag         string      `json:"@odata.etag"`
	Name              string      `json:"Name"`
	Dimensions        []Dimension `json:"Dimensions"`
	Rules             string      `json:"Rules"`
	DrillthroughRules interface{} `json:"DrillthroughRules"`
	LastSchemaUpdate  string      `json:"LastSchemaUpdate"`
	LastDataUpdate    string      `json:"LastDataUpdate"`
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

//GetCube method gets a cube from tm1
func (s Tm1Session) GetCube(cubeName string) (Cube, error) {
	cube := Cube{}
	res, err := s.Tm1SendHttpRequest("GET", "/Cubes('"+cubeName+"')", nil)
	if err != nil {
		return cube, err
	}

	json.Unmarshal(res, &cube)

	return cube, nil
}

//CubeDestroy deletes tm1 cube
func (s Tm1Session) CubeDestroy(cubeName string) error {
	_, err := s.Tm1SendHttpRequest("DELETE", "/Cubes('"+cubeName+"')", "{}")
	if err != nil {
		return err
	}
	return nil
}

//CubeExists check if a cube exists in tm1
func (s Tm1Session) CubeExists(cubeName string) (bool, error) {
	_, err := s.GetCube(cubeName)
	if err != nil {
		return false, err
	}
	return true, nil
}

//CubeCreate creates new cube
func (s Tm1Session) CubeCreate(cube Cube) error {

	var dims string

	for i, v := range cube.Dimensions {

		//Check if the dimension exists in tm1
		checkDim, _ := s.DimensionExists(v.Name)
		//if dimension doesn't exist, create one
		if !checkDim {
			newDim, _ := DimensionCreate(v.Name)
			s.DimensionCreate(newDim)
		}

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

//CubeCreate creates local cube struct
func CubeCreate(cubeName string) (Cube, error) {
	cube := Cube{Name: cubeName}
	return cube, nil
}
