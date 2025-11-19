package tm1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// HierarchyService provides methods to interact with TM1 hierarchies
type HierarchyService struct {
	rest     *RestService
	elements *ElementService
	subsets  *SubsetService
}

// NewHierarchyService creates a new HierarchyService instance
func NewHierarchyService(rest *RestService) *HierarchyService {
	return &HierarchyService{
		rest:     rest,
		elements: NewElementService(rest),
		subsets:  NewSubsetService(rest),
	}
}

// Create creates a new hierarchy in an existing dimension
func (hs *HierarchyService) Create(ctx context.Context, hierarchy *models.Hierarchy) error {
	if hierarchy.DimensionName == "" {
		return fmt.Errorf("dimension name is required")
	}
	if hierarchy.Name == "" {
		return fmt.Errorf("hierarchy name is required")
	}

	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies", url.PathEscape(hierarchy.DimensionName))

	body, err := json.Marshal(hierarchy)
	if err != nil {
		return fmt.Errorf("marshal hierarchy: %w", err)
	}

	resp, err := hs.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create hierarchy: %w", err)
	}
	defer resp.Body.Close()

	// Update element attributes if any
	if len(hierarchy.ElementAttributes) > 0 {
		if err := hs.UpdateElementAttributes(ctx, hierarchy, false); err != nil {
			return fmt.Errorf("update element attributes: %w", err)
		}
	}

	return nil
}

// Get retrieves a hierarchy by name
func (hs *HierarchyService) Get(ctx context.Context, dimensionName, hierarchyName string) (*models.Hierarchy, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')?$expand=Edges,Elements,ElementAttributes,Subsets,DefaultMember",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get hierarchy: %w", err)
	}
	defer resp.Body.Close()

	var hierarchy models.Hierarchy
	if err := json.NewDecoder(resp.Body).Decode(&hierarchy); err != nil {
		return nil, fmt.Errorf("decode hierarchy: %w", err)
	}

	hierarchy.DimensionName = dimensionName
	return &hierarchy, nil
}

// GetAllNames retrieves all hierarchy names in a dimension
func (hs *HierarchyService) GetAllNames(ctx context.Context, dimensionName string) ([]string, error) {
	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies?$select=Name", url.PathEscape(dimensionName))

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get all hierarchy names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// Update updates an existing hierarchy
func (hs *HierarchyService) Update(ctx context.Context, hierarchy *models.Hierarchy, keepExistingAttributes bool) error {
	if hierarchy.DimensionName == "" {
		return fmt.Errorf("dimension name is required")
	}
	if hierarchy.Name == "" {
		return fmt.Errorf("hierarchy name is required")
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(hierarchy.DimensionName),
		url.PathEscape(hierarchy.Name),
	)

	body, err := json.Marshal(hierarchy)
	if err != nil {
		return fmt.Errorf("marshal hierarchy: %w", err)
	}

	resp, err := hs.rest.Patch(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update hierarchy: %w", err)
	}
	defer resp.Body.Close()

	// Update element attributes
	if err := hs.UpdateElementAttributes(ctx, hierarchy, keepExistingAttributes); err != nil {
		return fmt.Errorf("update element attributes: %w", err)
	}

	return nil
}

// UpdateOrCreate updates a hierarchy if it exists, otherwise creates it
func (hs *HierarchyService) UpdateOrCreate(ctx context.Context, hierarchy *models.Hierarchy) error {
	exists, err := hs.Exists(ctx, hierarchy.DimensionName, hierarchy.Name)
	if err != nil {
		return fmt.Errorf("check hierarchy existence: %w", err)
	}

	if exists {
		return hs.Update(ctx, hierarchy, false)
	}
	return hs.Create(ctx, hierarchy)
}

// Exists checks if a hierarchy exists in a dimension
func (hs *HierarchyService) Exists(ctx context.Context, dimensionName, hierarchyName string) (bool, error) {
	endpoint := fmt.Sprintf("/Dimensions('%s')/Hierarchies?$select=Name", url.PathEscape(dimensionName))

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("check hierarchy existence: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	for _, item := range result.Value {
		if strings.EqualFold(item.Name, hierarchyName) {
			return true, nil
		}
	}

	return false, nil
}

// Delete deletes a hierarchy from a dimension
func (hs *HierarchyService) Delete(ctx context.Context, dimensionName, hierarchyName string) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := hs.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("delete hierarchy: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetHierarchySummary retrieves summary statistics for a hierarchy
func (hs *HierarchyService) GetHierarchySummary(ctx context.Context, dimensionName, hierarchyName string) (map[string]int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')?$expand=Edges/$count,Elements/$count,"+
			"ElementAttributes/$count,Members/$count,Levels/$count&$select=Cardinality",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get hierarchy summary: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	summary := make(map[string]int)
	properties := []string{"Elements", "Edges", "ElementAttributes", "Members", "Levels"}
	for _, prop := range properties {
		key := prop + "@odata.count"
		if count, ok := result[key].(float64); ok {
			summary[prop] = int(count)
		}
	}

	return summary, nil
}

// UpdateElementAttributes updates the element attributes of a hierarchy
func (hs *HierarchyService) UpdateElementAttributes(ctx context.Context, hierarchy *models.Hierarchy, keepExistingAttributes bool) error {
	// Get existing attributes
	existingAttrs, err := hs.elements.GetElementAttributes(ctx, hierarchy.DimensionName, hierarchy.Name)
	if err != nil {
		// If 404, no existing attributes
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			existingAttrs = []models.ElementAttribute{}
		} else {
			return fmt.Errorf("get existing attributes: %w", err)
		}
	}

	// Build maps for comparison
	existingAttrMap := make(map[string]models.ElementAttribute)
	for _, attr := range existingAttrs {
		existingAttrMap[strings.ToLower(strings.ReplaceAll(attr.Name, " ", ""))] = attr
	}

	newAttrMap := make(map[string]models.ElementAttribute)
	for _, attr := range hierarchy.ElementAttributes {
		newAttrMap[strings.ToLower(strings.ReplaceAll(attr.Name, " ", ""))] = attr
	}

	// Determine attributes to create, update, and delete
	var attrsToCreate []models.ElementAttribute
	var attrsToUpdate []models.ElementAttribute
	var attrsToDelete []string

	for key, newAttr := range newAttrMap {
		if existingAttr, exists := existingAttrMap[key]; exists {
			if existingAttr.AttributeType != newAttr.AttributeType {
				attrsToUpdate = append(attrsToUpdate, newAttr)
			}
		} else {
			attrsToCreate = append(attrsToCreate, newAttr)
		}
	}

	if !keepExistingAttributes {
		for key, existingAttr := range existingAttrMap {
			if _, exists := newAttrMap[key]; !exists {
				attrsToDelete = append(attrsToDelete, existingAttr.Name)
			}
		}
	}

	// Create new attributes
	for _, attr := range attrsToCreate {
		if err := hs.elements.CreateElementAttribute(ctx, hierarchy.DimensionName, hierarchy.Name, attr); err != nil {
			return fmt.Errorf("create attribute %s: %w", attr.Name, err)
		}
	}

	// Delete attributes
	for _, attrName := range attrsToDelete {
		if err := hs.elements.DeleteElementAttribute(ctx, hierarchy.DimensionName, hierarchy.Name, attrName); err != nil {
			return fmt.Errorf("delete attribute %s: %w", attrName, err)
		}
	}

	// Update attributes (delete and recreate)
	for _, attr := range attrsToUpdate {
		if err := hs.elements.DeleteElementAttribute(ctx, hierarchy.DimensionName, hierarchy.Name, attr.Name); err != nil {
			return fmt.Errorf("delete attribute for update %s: %w", attr.Name, err)
		}
		if err := hs.elements.CreateElementAttribute(ctx, hierarchy.DimensionName, hierarchy.Name, attr); err != nil {
			return fmt.Errorf("recreate attribute %s: %w", attr.Name, err)
		}
	}

	return nil
}

// GetDefaultMember retrieves the default member of a hierarchy
func (hs *HierarchyService) GetDefaultMember(ctx context.Context, dimensionName, hierarchyName string) (string, error) {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/DefaultMember",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		return "", fmt.Errorf("get default member: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if len(body) == 0 {
		return "", nil
	}

	var result struct {
		Name string `json:"Name"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Name, nil
}

// UpdateDefaultMember updates the default member of a hierarchy
func (hs *HierarchyService) UpdateDefaultMember(ctx context.Context, dimensionName, hierarchyName, memberName string) error {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	payload := map[string]string{
		"DefaultMemberName": memberName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := hs.rest.Patch(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update default member: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// RemoveAllEdges removes all edges from a hierarchy
func (hs *HierarchyService) RemoveAllEdges(ctx context.Context, dimensionName, hierarchyName string) error {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	payload := map[string]interface{}{
		"Edges": []interface{}{},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := hs.rest.Patch(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("remove all edges: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// RemoveEdgesUnderConsolidation removes all edges under a specific consolidation element
func (hs *HierarchyService) RemoveEdgesUnderConsolidation(ctx context.Context, dimensionName, hierarchyName, consolidationElement string) error {
	// Get the hierarchy
	hierarchy, err := hs.Get(ctx, dimensionName, hierarchyName)
	if err != nil {
		return fmt.Errorf("get hierarchy: %w", err)
	}

	// Get members under consolidation
	members, err := hs.elements.GetMembersUnderConsolidation(ctx, dimensionName, hierarchyName, consolidationElement)
	if err != nil {
		return fmt.Errorf("get members under consolidation: %w", err)
	}

	// Create a set of members to remove
	membersSet := make(map[string]bool)
	for _, member := range members {
		membersSet[strings.ToLower(strings.ReplaceAll(member, " ", ""))] = true
	}
	membersSet[strings.ToLower(strings.ReplaceAll(consolidationElement, " ", ""))] = true

	// Filter edges
	var newEdges []models.Edge
	for _, edge := range hierarchy.Edges {
		parentKey := strings.ToLower(strings.ReplaceAll(edge.ParentName, " ", ""))
		childKey := strings.ToLower(strings.ReplaceAll(edge.ComponentName, " ", ""))

		if !membersSet[parentKey] || !membersSet[childKey] {
			newEdges = append(newEdges, edge)
		}
	}

	hierarchy.Edges = newEdges

	return hs.Update(ctx, hierarchy, false)
}

// AddEdges adds edges to a hierarchy
func (hs *HierarchyService) AddEdges(ctx context.Context, dimensionName, hierarchyName string, edges map[[2]string]float64) error {
	return hs.elements.AddEdges(ctx, dimensionName, hierarchyName, edges)
}

// AddElements adds elements to a hierarchy
func (hs *HierarchyService) AddElements(ctx context.Context, dimensionName, hierarchyName string, elements []models.Element) error {
	return hs.elements.AddElements(ctx, dimensionName, hierarchyName, elements)
}

// AddElementAttributes adds element attributes to a hierarchy
func (hs *HierarchyService) AddElementAttributes(ctx context.Context, dimensionName, hierarchyName string, attributes []models.ElementAttribute) error {
	return hs.elements.AddElementAttributes(ctx, dimensionName, hierarchyName, attributes)
}

// CreateElementAttributes creates element attributes for a hierarchy (alias for AddElementAttributes)
func (hs *HierarchyService) CreateElementAttributes(ctx context.Context, dimensionName, hierarchyName string, attributes []models.ElementAttribute) error {
	return hs.AddElementAttributes(ctx, dimensionName, hierarchyName, attributes)
}

// IsBalanced checks if a hierarchy is balanced
func (hs *HierarchyService) IsBalanced(ctx context.Context, dimensionName, hierarchyName string) (bool, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Structure/$value",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := hs.rest.Get(ctx, endpoint)
	if err != nil {
		return false, fmt.Errorf("get hierarchy structure: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read response: %w", err)
	}

	structure := string(body)
	// 0 = balanced, 2 = unbalanced
	switch structure {
	case "0":
		return true, nil
	case "2":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected structure value: %s", structure)
	}
}

// Elements returns the ElementService associated with this HierarchyService
func (hs *HierarchyService) Elements() *ElementService {
	return hs.elements
}

// Subsets returns the SubsetService associated with this HierarchyService
func (hs *HierarchyService) Subsets() *SubsetService {
	return hs.subsets
}
