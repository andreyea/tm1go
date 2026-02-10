package tm1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// ViewService handles operations for TM1 Views
// Supports both public and private views.
type ViewService struct {
	rest *RestService
}

// NewViewService creates a new ViewService instance
func NewViewService(rest *RestService) *ViewService {
	return &ViewService{rest: rest}
}

// GetAll retrieves all views in a cube
func (vs *ViewService) GetAll(ctx context.Context, cubeName string, privateViews bool) ([]models.ViewDefinition, error) {
	viewType := "Views"
	if privateViews {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s?$expand="+
		"tm1.NativeView/Rows/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias),"+
		"tm1.NativeView/Columns/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias),"+
		"tm1.NativeView/Titles/Subset($expand=Hierarchy($select=Name;"+
		"$expand=Dimension($select=Name)),Elements($select=Name);"+
		"$select=Expression,UniqueName,Name,Alias),"+
		"tm1.NativeView/Titles/Selected($select=Name;$expand=Hierarchy($select=Name;$expand=Dimension($select=Name)))",
		url.PathEscape(cubeName), viewType)

	var result struct {
		Value []models.ViewWrapper `json:"value"`
	}

	if err := vs.rest.JSON(ctx, "GET", endpoint, nil, &result); err != nil {
		return nil, err
	}

	views := make([]models.ViewDefinition, 0, len(result.Value))
	for _, viewWrapper := range result.Value {
		views = append(views, viewWrapper.View)
	}

	return views, nil
}

// Get retrieves a view by name
func (vs *ViewService) Get(ctx context.Context, cubeName, viewName string, private bool) (models.ViewDefinition, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s('%s')?$expand=*", url.PathEscape(cubeName), viewType, url.PathEscape(viewName))
	wrapper := models.ViewWrapper{}
	if err := vs.rest.JSON(ctx, "GET", endpoint, nil, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.View, nil
}

// Delete deletes a view by name
func (vs *ViewService) Delete(ctx context.Context, cubeName, viewName string, private bool) error {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s('%s')", url.PathEscape(cubeName), viewType, url.PathEscape(viewName))
	resp, err := vs.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// Execute executes a view and returns the cellset ID
func (vs *ViewService) Execute(ctx context.Context, cubeName, viewName string, private bool) (string, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s('%s')/tm1.Execute", url.PathEscape(cubeName), viewType, url.PathEscape(viewName))
	resp, err := vs.rest.Post(ctx, endpoint, strings.NewReader(""))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result Cellset
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// Exists checks if a view exists
func (vs *ViewService) Exists(ctx context.Context, cubeName, viewName string, private bool) (bool, error) {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s('%s')", url.PathEscape(cubeName), viewType, url.PathEscape(viewName))
	resp, err := vs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// Create creates a new view
func (vs *ViewService) Create(ctx context.Context, cubeName string, view models.ViewDefinition, private bool) error {
	viewType := "Views"
	if private {
		viewType = "PrivateViews"
	}

	endpoint := fmt.Sprintf("/Cubes('%s')/%s", url.PathEscape(cubeName), viewType)
	viewBody, err := view.Body(true)
	if err != nil {
		return err
	}

	resp, err := vs.rest.Post(ctx, endpoint, strings.NewReader(viewBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}
