package tm1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// SubsetService handles operations for TM1 Subsets (dynamic and static)
type SubsetService struct {
	rest *RestService
}

// NewSubsetService creates a new SubsetService instance
func NewSubsetService(rest *RestService) *SubsetService {
	return &SubsetService{
		rest: rest,
	}
}

// Create creates a subset on the TM1 Server
func (ss *SubsetService) Create(ctx context.Context, subset *models.Subset, private bool) error {
	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s",
		url.PathEscape(subset.DimensionName),
		url.PathEscape(subset.HierarchyName),
		subsetsType)

	body, err := subset.Body()
	if err != nil {
		return fmt.Errorf("failed to serialize subset: %w", err)
	}

	_, err = ss.rest.Post(ctx, endpoint, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create subset: %w", err)
	}

	return nil
}

// Get retrieves a subset from the TM1 Server
func (ss *SubsetService) Get(ctx context.Context, subsetName, dimensionName, hierarchyName string, private bool) (*models.Subset, error) {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')?$expand=Hierarchy($select=Dimension,Name),Elements($select=Name)&$select=*,Alias",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType,
		url.PathEscape(subsetName))

	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get subset: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode subset response: %w", err)
	}

	subset, err := models.SubsetFromDict(result)
	if err != nil {
		return nil, err
	}

	return subset, nil
}

// GetAllNames retrieves names of all private or public subsets in a hierarchy
func (ss *SubsetService) GetAllNames(ctx context.Context, dimensionName, hierarchyName string, private bool) ([]string, error) {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s?$select=Name",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType)

	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get subset names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// Update updates a subset on the TM1 Server
func (ss *SubsetService) Update(ctx context.Context, subset *models.Subset, private bool) error {
	// If static subset, delete existing elements first
	if subset.IsStatic() {
		if err := ss.DeleteElementsFromStaticSubset(ctx, subset.DimensionName, subset.HierarchyName, subset.Name, private); err != nil {
			// Ignore error if elements don't exist
			_ = err
		}
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')",
		url.PathEscape(subset.DimensionName),
		url.PathEscape(subset.HierarchyName),
		subsetsType,
		url.PathEscape(subset.Name))

	body, err := subset.Body()
	if err != nil {
		return fmt.Errorf("failed to serialize subset: %w", err)
	}

	_, err = ss.rest.Patch(ctx, endpoint, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to update subset: %w", err)
	}

	return nil
}

// MakeStatic converts a dynamic subset into a static subset on the TM1 Server
func (ss *SubsetService) MakeStatic(ctx context.Context, subsetName, dimensionName, hierarchyName string, private bool) error {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	payload := map[string]interface{}{
		"Name":        subsetName,
		"MakePrivate": private,
		"MakeStatic":  true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')/tm1.SaveAs",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType,
		url.PathEscape(subsetName))

	_, err = ss.rest.Post(ctx, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to make subset static: %w", err)
	}

	return nil
}

// UpdateOrCreate updates if exists, else creates
func (ss *SubsetService) UpdateOrCreate(ctx context.Context, subset *models.Subset, private bool) error {
	exists, err := ss.Exists(ctx, subset.Name, subset.DimensionName, subset.HierarchyName, private)
	if err != nil {
		return err
	}

	if exists {
		return ss.Update(ctx, subset, private)
	}

	return ss.Create(ctx, subset, private)
}

// Delete deletes an existing subset on the TM1 Server
func (ss *SubsetService) Delete(ctx context.Context, subsetName, dimensionName, hierarchyName string, private bool) error {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType,
		url.PathEscape(subsetName))

	_, err := ss.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to delete subset: %w", err)
	}

	return nil
}

// Exists checks if a private or public subset exists
func (ss *SubsetService) Exists(ctx context.Context, subsetName, dimensionName, hierarchyName string, private bool) (bool, error) {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType,
		url.PathEscape(subsetName))

	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

// DeleteElementsFromStaticSubset deletes all elements from a static subset
func (ss *SubsetService) DeleteElementsFromStaticSubset(ctx context.Context, dimensionName, hierarchyName, subsetName string, private bool) error {
	subsetsType := "Subsets"
	if private {
		subsetsType = "PrivateSubsets"
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/%s('%s')/Elements/$ref",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		subsetsType,
		url.PathEscape(subsetName))

	_, err := ss.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to delete subset elements: %w", err)
	}

	return nil
}

// GetElementNames gets elements from existing (dynamic or static) subset
// For static subsets, returns the elements directly
// For dynamic subsets, returns an error as MDX execution is needed (use ElementService.ExecuteSetMDX)
func (ss *SubsetService) GetElementNames(ctx context.Context, dimensionName, hierarchyName, subsetName string, private bool) ([]string, error) {
	subset, err := ss.Get(ctx, subsetName, dimensionName, hierarchyName, private)
	if err != nil {
		return nil, err
	}

	if subset.IsStatic() {
		return subset.Elements, nil
	}

	// For dynamic subsets, we would need to execute the MDX expression using ElementService
	// This creates a circular dependency, so return an error indicating the user should use ElementService
	return nil, fmt.Errorf("dynamic subset element retrieval requires MDX execution - use ElementService.ExecuteSetMDX with expression: %s", subset.Expression)
}
