package tm1go

import (
	"encoding/json"
	"fmt"
)

type ViewService struct {
	rest   *RestService
	object *ObjectService
}

func NewViewService(rest *RestService, object *ObjectService) *ViewService {
	return &ViewService{rest: rest, object: object}
}

func (vs *ViewService) GetAll(cubeName string, privateViews bool) ([]View, error) {
	viewType := "Views"
	if privateViews {
		viewType = "PrivateViews"
	}

	url := fmt.Sprintf("/Cubes('%s')/%s?$expand="+
		"tm1.NativeView/Rows/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias), "+
		"tm1.NativeView/Columns/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias), "+
		"tm1.NativeView/Titles/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias), "+
		"tm1.NativeView/Titles/Selected($select=Name;$expand=Hierarchy($select=Name;$expand=Dimension($select=Name)))",
		cubeName, viewType)

	response, err := vs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[ViewWrapper]{}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	sliceOfViews := make([]View, 0, len(result.Value))
	for _, viewWrapper := range result.Value {
		sliceOfViews = append(sliceOfViews, viewWrapper.View)
	}
	return sliceOfViews, nil
}

// Get retrieves a view
func (vs *ViewService) Get(cubeName, viewName string, private bool) (View, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}
	url := fmt.Sprintf("/Cubes('%s')/%s('%s')?$expand=*", cubeName, viewType, viewName)
	response, err := vs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	view := ViewWrapper{}
	err = json.NewDecoder(response.Body).Decode(&view)
	if err != nil {
		return nil, err
	}
	return view.View, nil
}

// Delete a view
func (vs *ViewService) Delete(cubeName, viewName string, private bool) error {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}
	url := fmt.Sprintf("/Cubes('%s')/%s('%s')", cubeName, viewType, viewName)
	_, err := vs.rest.DELETE(url, nil, 0, nil)
	return err
}

// Execute a view. Returns cellset id
func (vs *ViewService) Execute(cubeName, viewName string, private bool) (string, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}
	url := fmt.Sprintf("/Cubes('%s')/%s('%s')/tm1.Execute", cubeName, viewType, viewName)
	response, err := vs.rest.POST(url, "", nil, 0, nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	result := Cellset{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// Exists checks if a view exists
func (vs *ViewService) Exists(cubeName, viewName string, private bool) (bool, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}
	url := fmt.Sprintf("/Cubes('%s')/%s('%s')", cubeName, viewType, viewName)
	return vs.object.Exists(url)
}

// Create a new view
func (vs *ViewService) Create(cubeName string, view View, private bool) error {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}
	url := fmt.Sprintf("/Cubes('%s')/%s", cubeName, viewType)
	viewBody, err := view.getBody(true)
	if err != nil {
		return err
	}
	_, err = vs.rest.POST(url, viewBody, nil, 0, nil)
	return err
}
