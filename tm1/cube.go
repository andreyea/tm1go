package tm1

import (
	"encoding/json"
	"errors"
	"fmt"
)

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
	res, err := s.Tm1SendHttpRequest("GET", "/Cubes('"+cubeName+"')?$expand=Dimensions($select=Name)", nil)
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

//CellPutN sends a value to tm1 cube
func (s Tm1Session) CellPutN(value float64, cubeName string, elements ...string) error {

	//get cube details
	cube, err := s.GetCube(cubeName)
	if err != nil {
		return err
	}

	//check number of elements passed
	if len(elements) != len(cube.Dimensions) {
		return errors.New("Invalid number of elements provided. Expected: " + string(len(cube.Dimensions)) + ". Received: " + string(len(elements)))
	}

	//Loop through dimensions and create tuple
	tuple := ""
	for k, v := range cube.Dimensions {
		if k == len(cube.Dimensions)-1 {
			tuple = tuple + `"Dimensions('` + v.Name + `')/Hierarchies('` + v.Name + `')/Elements('` + elements[k] + `')"`
		} else {
			tuple = tuple + `"Dimensions('` + v.Name + `')/Hierarchies('` + v.Name + `')/Elements('` + elements[k] + `')",`
		}

	}

	payload := `
	{
		"Cells":[
		{"Tuple@odata.bind": [
			` + tuple + `
		]}],
		"Value":"` + fmt.Sprintf("%f", value) + `"
	}
	`

	_, err = s.Tm1SendHttpRequest("POST", "/Cubes('"+cubeName+"')/tm1.Update", payload)

	if err != nil {
		return err
	}
	return nil
}

//ExecuteView returns cellset of a cube view
func (s Tm1Session) ExecuteView(cubeName, viewName string) (cellset Cellset, err error) {

	cont, err := s.Tm1SendHttpRequest("POST", "/Cubes('"+cubeName+"')/Views('"+viewName+"')/tm1.Execute?$expand=Axes($expand=Hierarchies,Tuples($expand=Members)),Cells,Cube($select=Name;$expand=Dimensions($select=Name))", "{}")
	if err != nil {
		return cellset, err
	}
	err = json.Unmarshal(cont, &cellset)
	if err != nil {
		return cellset, err
	}
	return cellset, nil
}
