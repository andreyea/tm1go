package tm1

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ElementServiceImpl implements the ElementService interface
type ElementServiceImpl struct {
	client Client
}

// NewElementServiceImpl creates a new ElementService implementation
func NewElementServiceImpl(client Client) ElementService {
	return &ElementServiceImpl{
		client: client,
	}
}

// Get retrieves an element from the specified dimension and hierarchy
func (es *ElementServiceImpl) Get(dimensionName, hierarchyName, elementName string) (*Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')?$expand=*",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		url.QueryEscape(elementName))

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get element: %w", err)
	}

	var element Element
	if err := json.Unmarshal(resp.Body, &element); err != nil {
		return nil, fmt.Errorf("failed to unmarshal element: %w", err)
	}

	return &element, nil
}

// Create creates a new element in the specified dimension and hierarchy
func (es *ElementServiceImpl) Create(dimensionName, hierarchyName string, element *Element) error {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	body, err := json.Marshal(element)
	if err != nil {
		return fmt.Errorf("failed to marshal element: %w", err)
	}

	_, err = es.client.POST(urlPath, body, nil)
	if err != nil {
		return fmt.Errorf("failed to create element: %w", err)
	}

	return nil
}

// Update updates an element in the specified dimension and hierarchy
func (es *ElementServiceImpl) Update(dimensionName, hierarchyName string, element *Element) error {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		url.QueryEscape(element.Name))

	body, err := json.Marshal(element)
	if err != nil {
		return fmt.Errorf("failed to marshal element: %w", err)
	}

	_, err = es.client.PATCH(urlPath, body, nil)
	if err != nil {
		return fmt.Errorf("failed to update element: %w", err)
	}

	return nil
}

// Delete deletes an element from the specified dimension and hierarchy
func (es *ElementServiceImpl) Delete(dimensionName, hierarchyName, elementName string) error {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		url.QueryEscape(elementName))

	_, err := es.client.DELETE(urlPath, nil)
	if err != nil {
		return fmt.Errorf("failed to delete element: %w", err)
	}

	return nil
}

// Exists checks if an element exists in the specified dimension and hierarchy
func (es *ElementServiceImpl) Exists(dimensionName, hierarchyName, elementName string) (bool, error) {
	_, err := es.Get(dimensionName, hierarchyName, elementName)
	if err != nil {
		if restErr, ok := err.(*TM1RestException); ok && restErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetElements returns all elements from the specified dimension and hierarchy
func (es *ElementServiceImpl) GetElements(dimensionName, hierarchyName string) ([]Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name,Type",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get elements: %w", err)
	}

	var result ValueArray[Element]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal elements: %w", err)
	}

	return result.Value, nil
}

// GetElementNames returns the names of all elements from the specified dimension and hierarchy
func (es *ElementServiceImpl) GetElementNames(dimensionName, hierarchyName string) ([]string, error) {
	elements, err := es.GetElements(dimensionName, hierarchyName)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(elements))
	for i, element := range elements {
		names[i] = element.Name
	}

	return names, nil
}

// GetLeafElements returns all leaf (non-consolidated) elements
func (es *ElementServiceImpl) GetLeafElements(dimensionName, hierarchyName string) ([]Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type ne %d",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		ElementTypeConsolidated)

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf elements: %w", err)
	}

	var result ValueArray[Element]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaf elements: %w", err)
	}

	return result.Value, nil
}

// GetConsolidatedElements returns all consolidated elements
func (es *ElementServiceImpl) GetConsolidatedElements(dimensionName, hierarchyName string) ([]Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq %d",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		ElementTypeConsolidated)

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get consolidated elements: %w", err)
	}

	var result ValueArray[Element]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consolidated elements: %w", err)
	}

	return result.Value, nil
}

// GetNumericElements returns all numeric elements
func (es *ElementServiceImpl) GetNumericElements(dimensionName, hierarchyName string) ([]Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq %d",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		ElementTypeNumeric)

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get numeric elements: %w", err)
	}

	var result ValueArray[Element]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal numeric elements: %w", err)
	}

	return result.Value, nil
}

// GetStringElements returns all string elements
func (es *ElementServiceImpl) GetStringElements(dimensionName, hierarchyName string) ([]Element, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq %d",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		ElementTypeString)

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get string elements: %w", err)
	}

	var result ValueArray[Element]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string elements: %w", err)
	}

	return result.Value, nil
}

// GetNumberOfElements returns the count of elements in the hierarchy
func (es *ElementServiceImpl) GetNumberOfElements(dimensionName, hierarchyName string) (int, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get element count: %w", err)
	}

	count, err := strconv.Atoi(string(resp.Body))
	if err != nil {
		return 0, fmt.Errorf("failed to parse element count: %w", err)
	}

	return count, nil
}

// GetEdges returns all edges (parent-child relationships) in the hierarchy
func (es *ElementServiceImpl) GetEdges(dimensionName, hierarchyName string) ([]Edge, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Edges?$select=ParentName,ComponentName,Weight",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	var result ValueArray[Edge]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
	}

	return result.Value, nil
}

// AddEdges adds multiple edges to the hierarchy
func (es *ElementServiceImpl) AddEdges(dimensionName, hierarchyName string, edges []Edge) error {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Edges",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	body, err := json.Marshal(edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	_, err = es.client.POST(urlPath, body, nil)
	if err != nil {
		return fmt.Errorf("failed to add edges: %w", err)
	}

	return nil
}

// RemoveEdge removes an edge from the hierarchy
func (es *ElementServiceImpl) RemoveEdge(dimensionName, hierarchyName, parentName, componentName string) error {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')/Edges(ParentName='%s',ComponentName='%s')",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName),
		url.QueryEscape(parentName),
		url.QueryEscape(parentName),
		url.QueryEscape(componentName))

	_, err := es.client.DELETE(urlPath, nil)
	if err != nil {
		return fmt.Errorf("failed to remove edge: %w", err)
	}

	return nil
}

// GetElementAttributes returns all element attributes for the hierarchy
func (es *ElementServiceImpl) GetElementAttributes(dimensionName, hierarchyName string) ([]ElementAttribute, error) {
	urlPath := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/ElementAttributes",
		url.QueryEscape(dimensionName),
		url.QueryEscape(hierarchyName))

	resp, err := es.client.GET(urlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get element attributes: %w", err)
	}

	var result ValueArray[ElementAttribute]
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal element attributes: %w", err)
	}

	return result.Value, nil
}

// ExecuteSetMDX executes an MDX SET expression and returns the result
func (es *ElementServiceImpl) ExecuteSetMDX(params MDXExecuteParams) (*CellsetAxis, error) {
	// Build the URL with proper OData parameters
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("/ExecuteMDXSetExpression?$expand=Tuples(")

	if params.TopRecords != nil {
		urlBuilder.WriteString(fmt.Sprintf("$top=%d;", *params.TopRecords))
	}

	// Set default member properties if none provided
	memberProperties := params.MemberProperties
	if len(memberProperties) == 0 {
		memberProperties = []string{"Name"}
	}

	// Process attribute properties (remove spaces)
	for i, prop := range memberProperties {
		if strings.HasPrefix(prop, "Attributes/") {
			memberProperties[i] = strings.ReplaceAll(prop, " ", "")
		}
	}

	urlBuilder.WriteString("$expand=Members($select=")
	urlBuilder.WriteString(strings.Join(memberProperties, ","))

	// Add parent and element expansions if specified
	expansions := []string{}
	if len(params.ParentProperties) > 0 {
		parentProps := make([]string, len(params.ParentProperties))
		copy(parentProps, params.ParentProperties)
		for i, prop := range parentProps {
			if strings.HasPrefix(prop, "Attributes/") {
				parentProps[i] = strings.ReplaceAll(prop, " ", "")
			}
		}
		expansions = append(expansions, "Parent($select="+strings.Join(parentProps, ",")+")")
	}

	if len(params.ElementProperties) > 0 {
		elementProps := make([]string, len(params.ElementProperties))
		copy(elementProps, params.ElementProperties)
		for i, prop := range elementProps {
			if strings.HasPrefix(prop, "Attributes/") {
				elementProps[i] = strings.ReplaceAll(prop, " ", "")
			}
		}
		expansions = append(expansions, "Element($select="+strings.Join(elementProps, ",")+")")
	}

	if len(expansions) > 0 {
		urlBuilder.WriteString(";$expand=")
		urlBuilder.WriteString(strings.Join(expansions, ","))
	}

	urlBuilder.WriteString("))")

	// Create the request payload
	payload := map[string]string{"MDX": params.MDX}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MDX payload: %w", err)
	}

	resp, err := es.client.POST(urlBuilder.String(), body, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute MDX: %w", err)
	}

	var cellsetAxis CellsetAxis
	if err := json.Unmarshal(resp.Body, &cellsetAxis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MDX result: %w", err)
	}

	return &cellsetAxis, nil
}

// UpdateElementService to return the implementation instead of nil
func NewElementService(client Client) ElementService {
	return NewElementServiceImpl(client)
}
