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

// HierarchyService handles operations for TM1 Hierarchies
type HierarchyService struct {
	rest *RestService
}

// NewHierarchyService creates a new HierarchyService instance
func NewHierarchyService(rest *RestService) *HierarchyService {
	return &HierarchyService{
		rest: rest,
	}
}

// Create creates a new hierarchy
func (hs *HierarchyService) Create(ctx context.Context, hierarchy *models.Hierarchy) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies",
		url.PathEscape(hierarchy.DimensionName))

	bodyJSON, _ := json.Marshal(hierarchy)
	body := strings.NewReader(string(bodyJSON))
	resp, err := hs.rest.Post(ctx, url, body)
	if err != nil {
		return fmt.Errorf("failed to create hierarchy: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

// Update updates an existing hierarchy
func (hs *HierarchyService) Update(ctx context.Context, hierarchy *models.Hierarchy, keepExistingAttributes bool) error {
	// Implementation will be added when we implement full hierarchy service
	return fmt.Errorf("hierarchy update not yet implemented")
}

// Delete deletes a hierarchy
func (hs *HierarchyService) Delete(ctx context.Context, dimensionName, hierarchyName string) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName))

	resp, err := hs.rest.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete hierarchy: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

// Exists checks if a hierarchy exists
func (hs *HierarchyService) Exists(ctx context.Context, dimensionName, hierarchyName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName))

	resp, err := hs.rest.Get(ctx, url)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check hierarchy existence: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return true, nil
}

// GetAllNames retrieves all hierarchy names for a dimension
func (hs *HierarchyService) GetAllNames(ctx context.Context, dimensionName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies?$select=Name",
		url.PathEscape(dimensionName))

	resp, err := hs.rest.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy names: %w", err)
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

// CreateElementAttributes creates element attributes for a hierarchy
func (hs *HierarchyService) CreateElementAttributes(ctx context.Context, dimensionName, hierarchyName string, attributes []models.ElementAttribute) error {
	for _, attr := range attributes {
		url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/ElementAttributes",
			url.PathEscape(dimensionName),
			url.PathEscape(hierarchyName))

		attrJSON, _ := json.Marshal(attr)
		body := strings.NewReader(string(attrJSON))

		resp, err := hs.rest.Post(ctx, url, body)
		if err != nil {
			return fmt.Errorf("failed to create element attribute '%s': %w", attr.Name, err)
		}
		resp.Body.Close()
	}

	return nil
}
