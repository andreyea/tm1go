package tm1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// ElementService provides methods to interact with TM1 elements
type ElementService struct {
	rest *RestService
}

// NewElementService creates a new ElementService instance
func NewElementService(rest *RestService) *ElementService {
	return &ElementService{
		rest: rest,
	}
}

// Get retrieves an element by name
func (es *ElementService) Get(ctx context.Context, dimensionName, hierarchyName, elementName string) (*models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')?$expand=*",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(elementName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get element: %w", err)
	}
	defer resp.Body.Close()

	var element models.Element
	if err := json.NewDecoder(resp.Body).Decode(&element); err != nil {
		return nil, fmt.Errorf("decode element: %w", err)
	}

	return &element, nil
}

// Create creates a new element in a hierarchy
func (es *ElementService) Create(ctx context.Context, dimensionName, hierarchyName string, element models.Element) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	body, err := json.Marshal(element)
	if err != nil {
		return fmt.Errorf("marshal element: %w", err)
	}

	resp, err := es.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create element: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Update updates an existing element
func (es *ElementService) Update(ctx context.Context, dimensionName, hierarchyName string, element models.Element) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(element.Name),
	)

	body, err := json.Marshal(element)
	if err != nil {
		return fmt.Errorf("marshal element: %w", err)
	}

	resp, err := es.rest.Patch(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update element: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// UpdateOrCreate updates an element if it exists, otherwise creates it
func (es *ElementService) UpdateOrCreate(ctx context.Context, dimensionName, hierarchyName string, element models.Element) error {
	exists, err := es.Exists(ctx, dimensionName, hierarchyName, element.Name)
	if err != nil {
		return fmt.Errorf("check element existence: %w", err)
	}

	if exists {
		return es.Update(ctx, dimensionName, hierarchyName, element)
	}
	return es.Create(ctx, dimensionName, hierarchyName, element)
}

// Exists checks if an element exists in a hierarchy
func (es *ElementService) Exists(ctx context.Context, dimensionName, hierarchyName, elementName string) (bool, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(elementName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("check element existence: %w", err)
	}
	defer resp.Body.Close()

	return true, nil
}

// Delete deletes an element from a hierarchy
func (es *ElementService) Delete(ctx context.Context, dimensionName, hierarchyName, elementName string) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(elementName),
	)

	resp, err := es.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("delete element: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetElements retrieves all elements in a hierarchy
func (es *ElementService) GetElements(ctx context.Context, dimensionName, hierarchyName string) ([]models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name,Type",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Element `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode elements: %w", err)
	}

	return result.Value, nil
}

// GetElementNames retrieves all element names in a hierarchy
func (es *ElementService) GetElementNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get element names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode element names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetLeafElements retrieves all leaf (non-consolidated) elements
func (es *ElementService) GetLeafElements(ctx context.Context, dimensionName, hierarchyName string) ([]models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type+ne+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get leaf elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Element `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode leaf elements: %w", err)
	}

	return result.Value, nil
}

// GetLeafElementNames retrieves all leaf element names
func (es *ElementService) GetLeafElementNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type+ne+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get leaf element names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode leaf element names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetConsolidatedElements retrieves all consolidated elements
func (es *ElementService) GetConsolidatedElements(ctx context.Context, dimensionName, hierarchyName string) ([]models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type+eq+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get consolidated elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Element `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode consolidated elements: %w", err)
	}

	return result.Value, nil
}

// GetConsolidatedElementNames retrieves all consolidated element names
func (es *ElementService) GetConsolidatedElementNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type+eq+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get consolidated element names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode consolidated element names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetNumericElements retrieves all numeric elements
func (es *ElementService) GetNumericElements(ctx context.Context, dimensionName, hierarchyName string) ([]models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type+eq+1",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get numeric elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Element `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode numeric elements: %w", err)
	}

	return result.Value, nil
}

// GetNumericElementNames retrieves all numeric element names
func (es *ElementService) GetNumericElementNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type+eq+1",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get numeric element names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode numeric element names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetStringElements retrieves all string elements
func (es *ElementService) GetStringElements(ctx context.Context, dimensionName, hierarchyName string) ([]models.Element, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type+eq+2",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get string elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Element `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode string elements: %w", err)
	}

	return result.Value, nil
}

// GetStringElementNames retrieves all string element names
func (es *ElementService) GetStringElementNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type+eq+2",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get string element names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode string element names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetNumberOfElements retrieves the count of all elements
func (es *ElementService) GetNumberOfElements(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements/$count",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get number of elements: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetNumberOfConsolidatedElements retrieves the count of consolidated elements
func (es *ElementService) GetNumberOfConsolidatedElements(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type+eq+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get number of consolidated elements: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetNumberOfLeafElements retrieves the count of leaf elements
func (es *ElementService) GetNumberOfLeafElements(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type+ne+3",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get number of leaf elements: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetNumberOfNumericElements retrieves the count of numeric elements
func (es *ElementService) GetNumberOfNumericElements(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type+eq+1",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get number of numeric elements: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetNumberOfStringElements retrieves the count of string elements
func (es *ElementService) GetNumberOfStringElements(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type+eq+2",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get number of string elements: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetEdges retrieves all edges (parent-child relationships) in a hierarchy
func (es *ElementService) GetEdges(ctx context.Context, dimensionName, hierarchyName string) (map[[2]string]float64, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Edges?$select=ParentName,ComponentName,Weight",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get edges: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.Edge `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode edges: %w", err)
	}

	edges := make(map[[2]string]float64)
	for _, edge := range result.Value {
		key := [2]string{edge.ParentName, edge.ComponentName}
		edges[key] = edge.Weight
	}

	return edges, nil
}

// GetElementAttributes retrieves all element attributes in a hierarchy
func (es *ElementService) GetElementAttributes(ctx context.Context, dimensionName, hierarchyName string) ([]models.ElementAttribute, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/ElementAttributes",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get element attributes: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.ElementAttribute `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode element attributes: %w", err)
	}

	return result.Value, nil
}

// GetElementAttributeNames retrieves all element attribute names in a hierarchy
func (es *ElementService) GetElementAttributeNames(ctx context.Context, dimensionName, hierarchyName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/ElementAttributes?$select=Name",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get element attribute names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode element attribute names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// CreateElementAttribute creates an element attribute in a hierarchy
func (es *ElementService) CreateElementAttribute(ctx context.Context, dimensionName, hierarchyName string, attribute models.ElementAttribute) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/ElementAttributes",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	body, err := json.Marshal(attribute)
	if err != nil {
		return fmt.Errorf("marshal element attribute: %w", err)
	}

	resp, err := es.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create element attribute: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// DeleteElementAttribute deletes an element attribute from a hierarchy
func (es *ElementService) DeleteElementAttribute(ctx context.Context, dimensionName, hierarchyName, attributeName string) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('}ElementAttributes_%s')/Hierarchies('}ElementAttributes_%s')/Elements('%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(attributeName),
	)

	resp, err := es.rest.Delete(ctx, endpoint)
	if err != nil {
		// Fail silently if attribute doesn't exist (404)
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("delete element attribute: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// AddEdges adds edges to a hierarchy
func (es *ElementService) AddEdges(ctx context.Context, dimensionName, hierarchyName string, edges map[[2]string]float64) error {
	if hierarchyName == "" {
		hierarchyName = dimensionName
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Edges",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	edgeList := make([]map[string]interface{}, 0, len(edges))
	for key, weight := range edges {
		edgeList = append(edgeList, map[string]interface{}{
			"ParentName":    key[0],
			"ComponentName": key[1],
			"Weight":        weight,
		})
	}

	body, err := json.Marshal(edgeList)
	if err != nil {
		return fmt.Errorf("marshal edges: %w", err)
	}

	resp, err := es.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("add edges: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// RemoveEdge removes a single edge from a hierarchy
func (es *ElementService) RemoveEdge(ctx context.Context, dimensionName, hierarchyName, parent, component string) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')/Edges(ParentName='%s',ComponentName='%s')",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(parent),
		url.PathEscape(parent),
		url.PathEscape(component),
	)

	resp, err := es.rest.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("remove edge: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// AddElements adds multiple elements to a hierarchy
func (es *ElementService) AddElements(ctx context.Context, dimensionName, hierarchyName string, elements []models.Element) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	body, err := json.Marshal(elements)
	if err != nil {
		return fmt.Errorf("marshal elements: %w", err)
	}

	resp, err := es.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("add elements: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// AddElementAttributes adds multiple element attributes to a hierarchy
func (es *ElementService) AddElementAttributes(ctx context.Context, dimensionName, hierarchyName string, attributes []models.ElementAttribute) error {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/ElementAttributes",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	body, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("marshal element attributes: %w", err)
	}

	resp, err := es.rest.Post(ctx, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("add element attributes: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetElementsByLevel retrieves elements filtered by level
func (es *ElementService) GetElementsByLevel(ctx context.Context, dimensionName, hierarchyName string, level int) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Level+eq+%d",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		level,
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get elements by level: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode elements: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetElementsFilteredByWildcard retrieves elements filtered by wildcard pattern
func (es *ElementService) GetElementsFilteredByWildcard(ctx context.Context, dimensionName, hierarchyName, wildcard string, level *int) ([]string, error) {
	filter := fmt.Sprintf("contains(tolower(replace(Name,' ','')),tolower(replace('%s',' ','')))", wildcard)
	if level != nil {
		filter = fmt.Sprintf("%s and Level eq %d", filter, *level)
	}

	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=%s",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.QueryEscape(filter),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get elements by wildcard: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode elements: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetMembersUnderConsolidation retrieves all members under a consolidation element
func (es *ElementService) GetMembersUnderConsolidation(ctx context.Context, dimensionName, hierarchyName, consolidation string) ([]string, error) {
	return es.getMembersUnderConsolidation(ctx, dimensionName, hierarchyName, consolidation, 99, false)
}

// GetLeavesUnderConsolidation retrieves all leaf elements under a consolidation
func (es *ElementService) GetLeavesUnderConsolidation(ctx context.Context, dimensionName, hierarchyName, consolidation string) ([]string, error) {
	return es.getMembersUnderConsolidation(ctx, dimensionName, hierarchyName, consolidation, 99, true)
}

// getMembersUnderConsolidation is a helper to get members under consolidation
func (es *ElementService) getMembersUnderConsolidation(ctx context.Context, dimensionName, hierarchyName, consolidation string, depth int, leavesOnly bool) ([]string, error) {
	// Build URL with recursive expansion
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')?$select=Name,Type&$expand=Components(",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(consolidation),
	)

	for i := 0; i < depth-1; i++ {
		endpoint += "$select=Name,Type;$expand=Components("
	}
	endpoint = endpoint[:len(endpoint)-1] + strings.Repeat(")", depth)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get members under consolidation: %w", err)
	}
	defer resp.Body.Close()

	var consolidationTree map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&consolidationTree); err != nil {
		return nil, fmt.Errorf("decode consolidation tree: %w", err)
	}

	members := []string{}
	var getMembers func(element map[string]interface{})
	getMembers = func(element map[string]interface{}) {
		elemType, _ := element["Type"].(string)
		elemName, _ := element["Name"].(string)

		if elemType == "Numeric" || elemType == "String" {
			members = append(members, elemName)
		} else if elemType == "Consolidated" {
			if !leavesOnly {
				members = append(members, elemName)
			}
			if components, ok := element["Components"].([]interface{}); ok {
				for _, comp := range components {
					if compMap, ok := comp.(map[string]interface{}); ok {
						getMembers(compMap)
					}
				}
			}
		}
	}

	getMembers(consolidationTree)
	return members, nil
}

// GetEdgesUnderConsolidation retrieves all edges under a consolidation element
func (es *ElementService) GetEdgesUnderConsolidation(ctx context.Context, dimensionName, hierarchyName, consolidation string, maxDepth int) (map[[2]string]float64, error) {
	if maxDepth == 0 {
		maxDepth = 99
	}

	// Build URL with recursive expansion
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')?$select=Edges&$expand=Edges($expand=Component(",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(consolidation),
	)

	for i := 0; i < maxDepth-1; i++ {
		if i == 0 {
			endpoint += "$select=Edges;$expand=Edges($expand=Component("
		} else {
			endpoint += "$select=Edges;$expand=Edges($expand=Component("
		}
	}
	endpoint = endpoint[:len(endpoint)-1] + strings.Repeat(")", maxDepth*2-1)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get edges under consolidation: %w", err)
	}
	defer resp.Body.Close()

	var consolidationTree map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&consolidationTree); err != nil {
		return nil, fmt.Errorf("decode consolidation tree: %w", err)
	}

	edges := make(map[[2]string]float64)
	var getEdges func(subTrees []interface{})
	getEdges = func(subTrees []interface{}) {
		for _, subTree := range subTrees {
			if subTreeMap, ok := subTree.(map[string]interface{}); ok {
				parentName, _ := subTreeMap["ParentName"].(string)
				componentName, _ := subTreeMap["ComponentName"].(string)
				weight, _ := subTreeMap["Weight"].(float64)

				edges[[2]string{parentName, componentName}] = weight

				if component, ok := subTreeMap["Component"].(map[string]interface{}); ok {
					if componentEdges, ok := component["Edges"].([]interface{}); ok {
						getEdges(componentEdges)
					}
				}
			}
		}
	}

	if edgesList, ok := consolidationTree["Edges"].([]interface{}); ok {
		getEdges(edgesList)
	}

	return edges, nil
}

// GetParents retrieves the parent elements of an element
func (es *ElementService) GetParents(ctx context.Context, dimensionName, hierarchyName, elementName string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements('%s')/Parents?$select=Name",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
		url.PathEscape(elementName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get parents: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode parents: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	return names, nil
}

// GetParentsOfAllElements retrieves parents for all elements in a hierarchy
func (es *ElementService) GetParentsOfAllElements(ctx context.Context, dimensionName, hierarchyName string) (map[string][]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$expand=Parents($select=Name)",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get parents of all elements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name    string `json:"Name"`
			Parents []struct {
				Name string `json:"Name"`
			} `json:"Parents"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode parents: %w", err)
	}

	parentsMap := make(map[string][]string)
	for _, child := range result.Value {
		parents := make([]string, len(child.Parents))
		for i, parent := range child.Parents {
			parents[i] = parent.Name
		}
		parentsMap[child.Name] = parents
	}

	return parentsMap, nil
}

// ElementIsParent checks if an element is a direct parent of another element
func (es *ElementService) ElementIsParent(ctx context.Context, dimensionName, hierarchyName, parentName, elementName string) (bool, error) {
	parents, err := es.GetParents(ctx, dimensionName, hierarchyName, elementName)
	if err != nil {
		return false, err
	}

	for _, parent := range parents {
		if strings.EqualFold(parent, parentName) {
			return true, nil
		}
	}

	return false, nil
}

// GetLevelNames retrieves level names in a hierarchy
func (es *ElementService) GetLevelNames(ctx context.Context, dimensionName, hierarchyName string, descending bool) ([]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Levels?$select=Name",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get level names: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode level names: %w", err)
	}

	names := make([]string, len(result.Value))
	for i, item := range result.Value {
		names[i] = item.Name
	}

	if descending {
		// Reverse the slice
		for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
			names[i], names[j] = names[j], names[i]
		}
	}

	return names, nil
}

// GetLevelsCount retrieves the count of levels in a hierarchy
func (es *ElementService) GetLevelsCount(ctx context.Context, dimensionName, hierarchyName string) (int, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Levels/$count",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get levels count: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}

	return count, nil
}

// GetElementTypes retrieves a map of element names to their types
func (es *ElementService) GetElementTypes(ctx context.Context, dimensionName, hierarchyName string, skipConsolidations bool) (map[string]string, error) {
	endpoint := fmt.Sprintf(
		"/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name,Type",
		url.PathEscape(dimensionName),
		url.PathEscape(hierarchyName),
	)

	if skipConsolidations {
		endpoint += "&$filter=Type+ne+3"
	}

	resp, err := es.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get element types: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			Name string `json:"Name"`
			Type string `json:"Type"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode element types: %w", err)
	}

	types := make(map[string]string)
	for _, item := range result.Value {
		types[item.Name] = item.Type
	}

	return types, nil
}

// GetElementPrincipalName retrieves the principal (server-side) name of an element
func (es *ElementService) GetElementPrincipalName(ctx context.Context, dimensionName, hierarchyName, elementName string) (string, error) {
	element, err := es.Get(ctx, dimensionName, hierarchyName, elementName)
	if err != nil {
		return "", err
	}
	return element.Name, nil
}
