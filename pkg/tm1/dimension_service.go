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

// DimensionService handles operations for TM1 Dimensions
type DimensionService struct {
	rest        *RestService
	hierarchies *HierarchyService
	subsets     *SubsetService
}

// NewDimensionService creates a new DimensionService instance
func NewDimensionService(rest *RestService) *DimensionService {
	return &DimensionService{
		rest:        rest,
		hierarchies: NewHierarchyService(rest),
		subsets:     NewSubsetService(rest),
	}
}

// Create creates a new dimension in TM1
func (ds *DimensionService) Create(ctx context.Context, dimension *models.Dimension) error {
	// Check if dimension already exists
	exists, err := ds.Exists(ctx, dimension.Name)
	if err != nil {
		return fmt.Errorf("failed to check dimension existence: %w", err)
	}
	if exists {
		return fmt.Errorf("dimension '%s' already exists", dimension.Name)
	}

	// Create dimension with hierarchies, elements, and edges
	url := "/Dimensions"
	body := strings.NewReader(dimension.Body())

	resp, err := ds.rest.Post(ctx, url, body)
	if err != nil {
		return fmt.Errorf("failed to create dimension: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Create element attributes for each hierarchy
	for _, hierarchy := range dimension.Hierarchies {
		if len(hierarchy.ElementAttributes) > 0 {
			if err := ds.hierarchies.CreateElementAttributes(ctx, dimension.Name, hierarchy.Name, hierarchy.ElementAttributes); err != nil {
				// If element attributes fail, try to clean up by deleting the dimension
				ds.Delete(ctx, dimension.Name)
				return fmt.Errorf("failed to create element attributes: %w", err)
			}
		}
	}

	return nil
}

// Get retrieves a dimension by name
func (ds *DimensionService) Get(ctx context.Context, dimensionName string) (*models.Dimension, error) {
	url := fmt.Sprintf("/Dimensions('%s')?$expand=Hierarchies($expand=*)",
		url.PathEscape(dimensionName))

	resp, err := ds.rest.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get dimension: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	dimension, err := models.DimensionFromJSON(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dimension: %w", err)
	}

	return dimension, nil
}

// Update updates an existing dimension
func (ds *DimensionService) Update(ctx context.Context, dimension *models.Dimension, keepExistingAttributes bool) error {
	// Get list of hierarchies to be removed
	existingHierarchies, err := ds.hierarchies.GetAllNames(ctx, dimension.Name)
	if err != nil {
		return fmt.Errorf("failed to get existing hierarchies: %w", err)
	}

	hierarchiesToRemove := make(map[string]bool)
	for _, h := range existingHierarchies {
		hierarchiesToRemove[strings.ToLower(h)] = true
	}

	// Remove hierarchies that exist in the dimension object from removal list
	for _, h := range dimension.HierarchyNames() {
		delete(hierarchiesToRemove, strings.ToLower(h))
	}

	// Update all hierarchies except the implicitly maintained 'Leaves' hierarchy
	for _, hierarchy := range dimension.Hierarchies {
		if strings.EqualFold(hierarchy.Name, "Leaves") {
			continue
		}

		exists, err := ds.hierarchies.Exists(ctx, dimension.Name, hierarchy.Name)
		if err != nil {
			return fmt.Errorf("failed to check hierarchy existence: %w", err)
		}

		if exists {
			if err := ds.hierarchies.Update(ctx, &hierarchy, keepExistingAttributes); err != nil {
				return fmt.Errorf("failed to update hierarchy '%s': %w", hierarchy.Name, err)
			}
		} else {
			if err := ds.hierarchies.Create(ctx, &hierarchy); err != nil {
				return fmt.Errorf("failed to create hierarchy '%s': %w", hierarchy.Name, err)
			}
		}
	}

	// Handle edge case: elements in leaves hierarchy that don't exist in other hierarchies
	if dimension.HasHierarchy("Leaves") {
		leavesHierarchy := dimension.GetHierarchy("Leaves")
		if leavesHierarchy != nil && len(leavesHierarchy.Elements) > 0 {
			// Get the default hierarchy (same name as dimension)
			defaultHierarchy := dimension.GetHierarchy(dimension.Name)
			if defaultHierarchy != nil {
				// Add missing elements from leaves to default hierarchy
				for _, element := range leavesHierarchy.Elements {
					// Check if element exists in default hierarchy
					found := false
					for _, defElement := range defaultHierarchy.Elements {
						if strings.EqualFold(element.Name, defElement.Name) {
							found = true
							break
						}
					}
					if !found {
						defaultHierarchy.AddElement(element)
					}
				}
			}
		}
	}

	// Delete hierarchies that have been removed from the dimension object
	for hierarchyName := range hierarchiesToRemove {
		if err := ds.hierarchies.Delete(ctx, dimension.Name, hierarchyName); err != nil {
			return fmt.Errorf("failed to delete hierarchy '%s': %w", hierarchyName, err)
		}
	}

	return nil
}

// UpdateOrCreate updates a dimension if it exists, otherwise creates it
func (ds *DimensionService) UpdateOrCreate(ctx context.Context, dimension *models.Dimension) error {
	exists, err := ds.Exists(ctx, dimension.Name)
	if err != nil {
		return fmt.Errorf("failed to check dimension existence: %w", err)
	}

	if exists {
		return ds.Update(ctx, dimension, false)
	}
	return ds.Create(ctx, dimension)
}

// Delete deletes a dimension
func (ds *DimensionService) Delete(ctx context.Context, dimensionName string) error {
	url := fmt.Sprintf("/Dimensions('%s')", url.PathEscape(dimensionName))

	resp, err := ds.rest.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete dimension: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

// Exists checks if a dimension exists
func (ds *DimensionService) Exists(ctx context.Context, dimensionName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%s')", url.PathEscape(dimensionName))

	resp, err := ds.rest.Get(ctx, url)
	if err != nil {
		// Check if it's a 404 error
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check dimension existence: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return true, nil
}

// GetAllNames retrieves all dimension names
func (ds *DimensionService) GetAllNames(ctx context.Context, skipControlDims bool) ([]string, error) {
	endpoint := "Dimensions"
	if skipControlDims {
		endpoint = "ModelDimensions()"
	}

	url := fmt.Sprintf("/%s?$select=Name", endpoint)

	resp, err := ds.rest.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get dimension names: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, entry := range result.Value {
		names[i] = entry.Name
	}

	return names, nil
}

// GetNumberOfDimensions returns the count of dimensions
func (ds *DimensionService) GetNumberOfDimensions(ctx context.Context, skipControlDims bool) (int, error) {
	if skipControlDims {
		names, err := ds.GetAllNames(ctx, true)
		if err != nil {
			return 0, err
		}
		return len(names), nil
	}

	resp, err := ds.rest.Get(ctx, "/Dimensions/$count")
	if err != nil {
		return 0, fmt.Errorf("failed to get dimension count: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(string(body), "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse count: %w", err)
	}

	return count, nil
}

// UsesAlternateHierarchies checks if a dimension uses alternate hierarchies
func (ds *DimensionService) UsesAlternateHierarchies(ctx context.Context, dimensionName string) (bool, error) {
	hierarchyNames, err := ds.hierarchies.GetAllNames(ctx, dimensionName)
	if err != nil {
		return false, fmt.Errorf("failed to get hierarchy names: %w", err)
	}

	if len(hierarchyNames) > 1 {
		return true, nil
	}

	// Check if the single hierarchy name differs from dimension name (case-insensitive)
	if len(hierarchyNames) == 1 {
		return !strings.EqualFold(dimensionName, hierarchyNames[0]), nil
	}

	return false, nil
}

// GetAll retrieves all dimensions
func (ds *DimensionService) GetAll(ctx context.Context, skipControlDims bool) ([]*models.Dimension, error) {
	names, err := ds.GetAllNames(ctx, skipControlDims)
	if err != nil {
		return nil, err
	}

	dimensions := make([]*models.Dimension, 0, len(names))
	for _, name := range names {
		dim, err := ds.Get(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get dimension '%s': %w", name, err)
		}
		dimensions = append(dimensions, dim)
	}

	return dimensions, nil
}
