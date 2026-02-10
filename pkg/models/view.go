package models

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ViewDefinition represents a TM1 view (NativeView or MDXView)
type ViewDefinition interface {
	GetType() string
	GetName() string
	Body(static bool) (string, error)
}

// ViewWrapper allows polymorphic unmarshalling based on @odata.type
type ViewWrapper struct {
	View ViewDefinition
}

// UnmarshalJSON selects the correct view implementation based on @odata.type
func (vw *ViewWrapper) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	switch tmp["@odata.type"] {
	case "#ibm.tm1.api.v1.NativeView":
		vw.View = &NativeView{}
	case "#ibm.tm1.api.v1.MDXView":
		vw.View = &MDXView{}
	default:
		return fmt.Errorf("unknown view type")
	}

	return json.Unmarshal(data, vw.View)
}

// ViewAxis represents an axis definition in a native view
// Subset is decoded using SubsetFromDict to populate DimensionName/HierarchyName.
type ViewAxis struct {
	Subset *Subset `json:"Subset,omitempty"`
}

// UnmarshalJSON custom decoding for axis with subset payloads
func (va *ViewAxis) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if subsetRaw, ok := raw["Subset"].(map[string]interface{}); ok {
		subset, err := SubsetFromDict(subsetRaw)
		if err != nil {
			return err
		}
		va.Subset = subset
	}

	return nil
}

// ViewTitleAxis represents a title axis (subset + selected element)
type ViewTitleAxis struct {
	Subset   *Subset              `json:"Subset,omitempty"`
	Selected *ViewSelectedElement `json:"Selected,omitempty"`
}

// UnmarshalJSON custom decoding for title axis
func (vta *ViewTitleAxis) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if subsetRaw, ok := raw["Subset"].(map[string]interface{}); ok {
		subset, err := SubsetFromDict(subsetRaw)
		if err != nil {
			return err
		}
		vta.Subset = subset
	}

	if selectedRaw, ok := raw["Selected"].(map[string]interface{}); ok {
		selected, err := ViewSelectedElementFromDict(selectedRaw)
		if err != nil {
			return err
		}
		vta.Selected = selected
	}

	return nil
}

// ViewSelectedElement holds the selected element for titles
// DimensionName and HierarchyName are used when creating a view payload.
type ViewSelectedElement struct {
	Name          string `json:"Name"`
	DimensionName string `json:"-"`
	HierarchyName string `json:"-"`
}

type subsetBody struct {
	Name       string   `json:"Name"`
	Alias      string   `json:"Alias,omitempty"`
	Hierarchy  string   `json:"Hierarchy@odata.bind"`
	Elements   []string `json:"Elements@odata.bind,omitempty"`
	Expression string   `json:"Expression,omitempty"`
}

// ViewSelectedElementFromDict builds a ViewSelectedElement from API responses
func ViewSelectedElementFromDict(dict map[string]interface{}) (*ViewSelectedElement, error) {
	selected := &ViewSelectedElement{}

	if name, ok := dict["Name"].(string); ok {
		selected.Name = name
	}

	if hierarchy, ok := dict["Hierarchy"].(map[string]interface{}); ok {
		if hierName, ok := hierarchy["Name"].(string); ok {
			selected.HierarchyName = hierName
		}
		if dim, ok := hierarchy["Dimension"].(map[string]interface{}); ok {
			if dimName, ok := dim["Name"].(string); ok {
				selected.DimensionName = dimName
			}
		}
	}

	if selected.HierarchyName == "" {
		selected.HierarchyName = selected.DimensionName
	}

	return selected, nil
}

// NativeView represents a TM1 NativeView
// Columns/Rows/Titles subsets use models.Subset
// Selected uses ViewSelectedElement for title selection.
type NativeView struct {
	Type                 string          `json:"@odata.type,omitempty"`
	Cube                 *Cube           `json:"Cube,omitempty"`
	Name                 string          `json:"Name"`
	Columns              []ViewAxis      `json:"Columns,omitempty"`
	Rows                 []ViewAxis      `json:"Rows,omitempty"`
	Titles               []ViewTitleAxis `json:"Titles,omitempty"`
	SuppressEmptyColumns bool            `json:"SuppressEmptyColumns,omitempty"`
	SuppressEmptyRows    bool            `json:"SuppressEmptyRows,omitempty"`
	FormatString         string          `json:"FormatString,omitempty"`
}

func (v *NativeView) GetType() string {
	return v.Type
}

func (v *NativeView) GetName() string {
	return v.Name
}

// Body returns the JSON representation for a NativeView create/update request
func (v *NativeView) Body(static bool) (string, error) {
	type axisBody struct {
		Subset   subsetBody `json:"Subset"`
		Selected string     `json:"Selected@odata.bind,omitempty"`
	}

	type nativeViewBody struct {
		Type                 string     `json:"@odata.type"`
		Name                 string     `json:"Name"`
		Columns              []axisBody `json:"Columns"`
		Rows                 []axisBody `json:"Rows"`
		Titles               []axisBody `json:"Titles,omitempty"`
		SuppressEmptyColumns bool       `json:"SuppressEmptyColumns,omitempty"`
		SuppressEmptyRows    bool       `json:"SuppressEmptyRows,omitempty"`
		FormatString         string     `json:"FormatString,omitempty"`
	}

	viewType := v.Type
	if viewType == "" {
		viewType = "#ibm.tm1.api.v1.NativeView"
	}

	body := nativeViewBody{
		Type:                 viewType,
		Name:                 v.Name,
		SuppressEmptyColumns: v.SuppressEmptyColumns,
		SuppressEmptyRows:    v.SuppressEmptyRows,
		FormatString:         v.FormatString,
	}

	body.Columns = make([]axisBody, 0, len(v.Columns))
	for _, column := range v.Columns {
		if column.Subset == nil {
			return "", fmt.Errorf("column subset is required")
		}
		subsetBody, err := buildSubsetBody(column.Subset, static)
		if err != nil {
			return "", err
		}
		body.Columns = append(body.Columns, axisBody{Subset: subsetBody})
	}

	body.Rows = make([]axisBody, 0, len(v.Rows))
	for _, row := range v.Rows {
		if row.Subset == nil {
			return "", fmt.Errorf("row subset is required")
		}
		subsetBody, err := buildSubsetBody(row.Subset, static)
		if err != nil {
			return "", err
		}
		body.Rows = append(body.Rows, axisBody{Subset: subsetBody})
	}

	if len(v.Titles) > 0 {
		body.Titles = make([]axisBody, 0, len(v.Titles))
		for _, title := range v.Titles {
			if title.Subset == nil {
				return "", fmt.Errorf("title subset is required")
			}
			subsetBody, err := buildSubsetBody(title.Subset, static)
			if err != nil {
				return "", err
			}
			axis := axisBody{Subset: subsetBody}
			if title.Selected != nil && title.Selected.Name != "" {
				binding, err := buildSelectedBinding(title.Selected)
				if err != nil {
					return "", err
				}
				axis.Selected = binding
			}
			body.Titles = append(body.Titles, axis)
		}
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// MDXView represents a TM1 MDX View
// Meta is optional and can contain Aliases, ContextSets, and ExpandAboves.
type MDXView struct {
	Cube *Cube  `json:"Cube,omitempty"`
	Type string `json:"@odata.type"`
	Name string `json:"Name"`
	MDX  string `json:"MDX"`
	Meta struct {
		Aliases      map[string]string            `json:"Aliases"`
		ContextSets  map[string]map[string]string `json:"ContextSets"`
		ExpandAboves map[string]bool              `json:"ExpandAboves"`
	} `json:"Meta,omitempty"`
}

func (v *MDXView) GetType() string {
	return v.Type
}

func (v *MDXView) GetName() string {
	return v.Name
}

// Body returns the JSON representation for an MDX view create/update request
func (v *MDXView) Body(static bool) (string, error) {
	bodyAsDict := make(map[string]interface{})
	viewType := v.Type
	if viewType == "" {
		viewType = "#ibm.tm1.api.v1.MDXView"
	}

	bodyAsDict["@odata.type"] = viewType
	bodyAsDict["Name"] = v.Name
	bodyAsDict["MDX"] = v.MDX
	bodyAsDict["Meta"] = v.Meta

	jsonData, err := json.Marshal(bodyAsDict)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func buildSubsetBody(subset *Subset, static bool) (subsetBody, error) {
	if subset.DimensionName == "" || subset.HierarchyName == "" {
		return subsetBody{}, fmt.Errorf("subset dimension and hierarchy are required")
	}

	body := subsetBody{
		Name:      subset.Name,
		Alias:     subset.Alias,
		Hierarchy: fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')", url.PathEscape(subset.DimensionName), url.PathEscape(subset.HierarchyName)),
	}

	if static && len(subset.Elements) > 0 {
		elementsBind := make([]string, len(subset.Elements))
		for i, element := range subset.Elements {
			elementsBind[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
				url.PathEscape(subset.DimensionName),
				url.PathEscape(subset.HierarchyName),
				url.PathEscape(element))
		}
		body.Elements = elementsBind
	} else if subset.Expression != "" {
		body.Expression = subset.Expression
	}

	return body, nil
}

func buildSelectedBinding(selected *ViewSelectedElement) (string, error) {
	if selected.DimensionName == "" || selected.HierarchyName == "" || selected.Name == "" {
		return "", fmt.Errorf("selected element must include dimension, hierarchy, and name")
	}

	return fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')",
		url.PathEscape(selected.DimensionName),
		url.PathEscape(selected.HierarchyName),
		url.PathEscape(selected.Name)), nil
}
