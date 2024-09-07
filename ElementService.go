package tm1go

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ElementService struct {
	rest   *RestService
	object *ObjectService
}

func NewElementService(rest *RestService, object *ObjectService) *ElementService {
	return &ElementService{rest: rest, object: object}
}

// MDXExecuteParams is a struct for MDX query parameters
type MDXExecuteParams struct {
	MDX               string
	TopRecords        *int
	MemberProperties  []string
	ParentProperties  []string
	ElementProperties []string
}

// Get returns an element from the specified dimension and hierarchy
func (es *ElementService) Get(dimensionName string, hierarchyName string, elementName string) (*Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')?$expand=*", dimensionName, hierarchyName, elementName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	element := &Element{}
	err = json.NewDecoder(response.Body).Decode(&element)
	if err != nil {
		return nil, err
	}
	return element, nil
}

// Create creates a new element in the specified dimension and hierarchy
func (es *ElementService) Create(dimensionName string, hierarchyName string, element *Element) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements", dimensionName, hierarchyName)
	body, err := element.getBody()
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = es.rest.POST(url, string(jsonBody), nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Update updates an element in the specified dimension and hierarchy
func (es *ElementService) Update(dimensionName string, hierarchyName string, element *Element) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')", dimensionName, hierarchyName, element.Name)
	body, err := element.getBody()
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(body)
	_, err = es.rest.PATCH(url, string(jsonBody), nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// Exists checks if an element exists in the specified dimension and hierarchy
func (es *ElementService) Exists(dimensionName string, hierarchyName string, elementName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')", dimensionName, hierarchyName, elementName)
	return es.object.Exists(url)
}

// UpdateOrCreate updates or creates an element in the specified dimension and hierarchy
func (es *ElementService) UpdateOrCreate(dimensionName string, hierarchyName string, element *Element) error {
	exists, err := es.Exists(dimensionName, hierarchyName, element.Name)
	if err != nil {
		return err
	}
	if exists {
		err = es.Update(dimensionName, hierarchyName, element)
		if err != nil {
			return err
		}
	} else {
		err = es.Create(dimensionName, hierarchyName, element)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes an element from the specified dimension and hierarchy
func (es *ElementService) Delete(dimensionName string, hierarchyName string, elementName string) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')", dimensionName, hierarchyName, elementName)
	_, err := es.rest.DELETE(url, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetElements returns all elements from the specified dimension and hierarchy
func (es *ElementService) GetElements(dimensionName string, hierarchyName string) ([]Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name,Type", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// get_elements_dataframe

// GetEdges returns all edges from the specified dimension and hierarchy
func (es *ElementService) GetEdges(dimensionName string, hierarchyName string) ([]Edge, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Edges?$select=ParentName,ComponentName,Weight", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Edge]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetElement returns an element from the specified dimension and hierarchy
func (es *ElementService) GetLeafElements(dimensionName string, hierarchyName string) ([]Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type ne 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetLeafElementNames returns the names of all leaf elements from the specified dimension and hierarchy
func (es *ElementService) GetLeafElementNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type ne 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, element := range result.Value {
		names = append(names, element.Name)
	}
	return names, nil
}

// GetConsolidatedElements returns all consolidated elements from the specified dimension and hierarchy
func (es *ElementService) GetConsolidatedElements(dimensionName string, hierarchyName string) ([]Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetConsolidatedElementNames returns the names of all consolidated elements from the specified dimension and hierarchy
func (es *ElementService) GetConsolidatedElementNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type eq 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, element := range result.Value {
		names = append(names, element.Name)
	}
	return names, nil
}

// GetNumericElements returns all numeric elements from the specified dimension and hierarchy
func (es *ElementService) GetNumericElements(dimensionName string, hierarchyName string) ([]Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq 1", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetNumericElementNames returns the names of all numeric elements from the specified dimension and hierarchy
func (es *ElementService) GetNumericElementNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type eq 1", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, element := range result.Value {
		names = append(names, element.Name)
	}
	return names, nil
}

// GetStringElements returns all string elements from the specified dimension and hierarchy
func (es *ElementService) GetStringElements(dimensionName string, hierarchyName string) ([]Element, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$expand=*&$filter=Type eq 2", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetStringElementNames returns the names of all string elements from the specified dimension and hierarchy
func (es *ElementService) GetStringElementNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name&$filter=Type eq 2", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, element := range result.Value {
		names = append(names, element.Name)
	}
	return names, nil
}

// GetElementNames returns the names of all elements from the specified dimension and hierarchy
func (es *ElementService) GetElementNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements?$select=Name", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Element]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, element := range result.Value {
		names = append(names, element.Name)
	}
	return names, nil
}

// GetElementCount returns the number of elements in the specified dimension and hierarchy
func (es *ElementService) GetNumberOfElements(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetNumberOfElementsByType returns the number of elements of the specified type in the specified dimension and hierarchy
func (es *ElementService) GetNumberOfConsolidatedElements(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type eq 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetNumberOfLeafElements returns the number of leaf elements in the specified dimension and hierarchy
func (es *ElementService) GetNumberOfLeafElements(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type ne 3", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetNumberOfConsolidatedElements returns the number of consolidated elements in the specified dimension and hierarchy
func (es *ElementService) GetNumberOfNumericElements(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type eq 1", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetNumberOfStringElements returns the number of string elements in the specified dimension and hierarchy
func (es *ElementService) GetNumberOfStringElements(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements/$count?$filter=Type eq 2", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Getlevelnames returns the names of all levels from the specified dimension and hierarchy
func (es *ElementService) GetLevelNames(dimensionName string, hierarchyName string) ([]string, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Levels?$select=Name", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[Level]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, level := range result.Value {
		names = append(names, level.Name)
	}
	return names, nil
}

// GetLevelCount returns the number of levels in the specified dimension and hierarchy
func (es *ElementService) GetLevelsCount(dimensionName string, hierarchyName string) (int, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Levels/$count", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	count := 0
	err = json.NewDecoder(response.Body).Decode(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// AttributeExists checks if an attribute cube exists for the specified dimension and hierarchy
func (es *ElementService) AttributeCubeExists(dimensionName string) (bool, error) {
	url := fmt.Sprintf("/Cubes('%s')", "}ElementAttributes_"+dimensionName)
	return es.object.Exists(url)
}

// _retrieve_mdx_rows_and_cell_values_as_string_set
// _retrieve_mdx_rows_and_values
// get_alias_element_attributes

// GetElementAttributes returns all element attributes from the specified dimension and hierarchy
func (es *ElementService) GetElementAttributes(dimensionName string, hierarchyName string) ([]ElementAttribute, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/ElementAttributes", dimensionName, hierarchyName)
	response, err := es.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := ValueArray[ElementAttribute]{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

// CreateElementAttribute creates a new element attribute in the specified dimension and hierarchy
func (es *ElementService) CreateElementAttribute(dimensionName string, hierarchyName string, elementAttribute *ElementAttribute) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/ElementAttributes", dimensionName, hierarchyName)
	body, err := elementAttribute.getBody()
	if err != nil {
		return err
	}
	_, err = es.rest.POST(url, body, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteElementAttribute deletes an element attribute from the specified dimension and hierarchy
func (es *ElementService) DeleteElementAttribute(dimensionName string, hierarchyName string, elementAttributeName string) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/ElementAttributes('%s')", dimensionName, hierarchyName, elementAttributeName)
	_, err := es.rest.DELETE(url, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// ExecuteSetMDX executes an MDX query and returns the result as a CellsetAxis
func (es *ElementService) ExecuteSetMDX(params MDXExecuteParams) (*CellsetAxis, error) {
	top := ""
	if params.TopRecords != nil {
		top = fmt.Sprintf("$top=%d;", *params.TopRecords)
	}

	// Set default member properties if none are provided
	if params.MemberProperties == nil || len(params.MemberProperties) == 0 {
		params.MemberProperties = []string{"Name"}
	} else {
		// Process member properties, removing spaces for attributes
		for i, prop := range params.MemberProperties {
			if strings.HasPrefix(prop, "Attributes/") {
				params.MemberProperties[i] = strings.ReplaceAll(prop, " ", "")
			}
		}
	}

	// Process element properties
	if params.ElementProperties != nil && len(params.ElementProperties) > 0 {
		for i, prop := range params.ElementProperties {
			if strings.HasPrefix(prop, "Attributes/") {
				params.ElementProperties[i] = strings.ReplaceAll(prop, " ", "")
			}
		}
	}

	// Process parent properties
	if params.ParentProperties != nil && len(params.ParentProperties) > 0 {
		for i, prop := range params.ParentProperties {
			if strings.HasPrefix(prop, "Attributes/") {
				params.ParentProperties[i] = strings.ReplaceAll(prop, " ", "")
			}
		}
	}

	// Join properties for URL construction
	memberPropertiesJoined := strings.Join(params.MemberProperties, ",")
	selectMemberProperties := "$select=" + memberPropertiesJoined

	propertiesToExpand := []string{}
	if len(params.ParentProperties) > 0 {
		selectParentProperties := "Parent($select=" + strings.Join(params.ParentProperties, ",") + ")"
		propertiesToExpand = append(propertiesToExpand, selectParentProperties)
	}

	if len(params.ElementProperties) > 0 {
		selectElementProperties := "Element($select=" + strings.Join(params.ElementProperties, ",") + ")"
		propertiesToExpand = append(propertiesToExpand, selectElementProperties)
	}

	expandProperties := ""
	if len(propertiesToExpand) > 0 {
		expandProperties = ";$expand=" + strings.Join(propertiesToExpand, ",")
	}

	url := "/ExecuteMDXSetExpression?$expand=Tuples(" + top +
		"$expand=Members(" + selectMemberProperties +
		expandProperties + "))"

	payload := map[string]string{"MDX": params.MDX}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := es.rest.POST(url, string(payloadBytes), nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var cellsetAxis = &CellsetAxis{}
	if err := json.NewDecoder(response.Body).Decode(&cellsetAxis); err != nil {
		return nil, err
	}

	return cellsetAxis, nil
}

// RemoveEdge removes an edge from the specified dimension and hierarchy
func (es *ElementService) RemoveEdge(dimensionName string, hierarchyName string, parentName string, componentName string) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements('%s')/Edges(ParentName='%s',ComponentName='%s')", dimensionName, hierarchyName, parentName, parentName, componentName)
	_, err := es.rest.DELETE(url, nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// AddEdge adds an edge to the specified dimension and hierarchy
func (es *ElementService) AddEdges(dimensionName string, hierarchyName string, edges []Edge) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Edges", dimensionName, hierarchyName)
	body := make([]string, 0)
	for _, edge := range edges {
		edgeBody := edge.getBody()
		body = append(body, edgeBody)
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = es.rest.POST(url, string(bodyJson), nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// AddElements adds an element to the specified dimension and hierarchy
func (es *ElementService) AddElements(dimensionName string, hierarchyName string, elements []Element) error {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')/Elements", dimensionName, hierarchyName)
	body := make([]map[string]interface{}, 0)
	for _, element := range elements {
		elementBody, err := element.getBody()
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		body = append(body, elementBody)
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = es.rest.POST(url, string(bodyJson), nil, 0, nil)
	if err != nil {
		return err
	}
	return nil
}

// HierarchyExists checks if a hierarchy exists in the specified dimension
func (es *ElementService) HierarchyExists(dimensionName string, hierarchyName string) (bool, error) {
	url := fmt.Sprintf("/Dimensions('%s')/Hierarchies('%s')", dimensionName, hierarchyName)
	return es.object.Exists(url)
}

// add_element_attributes
// get_parents
// get_parents_of_all_elements
// get_element_principal_name
// _get_mdx_set_cardinality
// _build_drill_intersection_mdx
// element_is_parent
// element_is_ancestor
// _element_is_ancestor_ti
// get_process_service
// _get_hierarchy_service
// _get_subset_service
// _get_process_service
// _get_cell_service
// get_element_attribute_names
// get_elements_filtered_by_attribute
// get_all_leaf_element_identifiers
// get_elements_by_level
// get_elements_filtered_by_wildcard
// get_all_element_identifiers
// get_element_identifiers
// get_attribute_of_elements
// _extract_dict_from_rows_and_values
// get_leaves_under_consolidation
// get_edges_under_consolidation
// get_members_under_consolidation
// get_element_types
// get_element_types_from_all_hierarchies
// delete_elements
// delete_elements_use_ti
// delete_edges
// delete_edges_use_ti
