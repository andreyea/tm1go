package tm1go

import (
	"encoding/json"
	"fmt"
)

type NativeView struct {
	Type    string `json:"@odata.type,omitempty"`
	Cube    Cube   `json:"Cube,omitempty"`
	Name    string `json:"Name"`
	Columns []struct {
		Subset Subset `json:"Subset,omitempty"`
	} `json:"Columns,omitempty"`
	Rows []struct {
		Subset Subset `json:"Subset,omitempty"`
	} `json:"Rows,omitempty"`
	Titles []struct {
		Subset   Subset  `json:"Subset,omitempty"`
		Selected Element `json:"Selected,omitempty"`
	} `json:"Titles,omitempty"`
	SuppressEmptyColumns bool   `json:"SuppressEmptyColumns,omitempty"`
	SuppressEmptyRows    bool   `json:"SuppressEmptyRows,omitempty"`
	FormatString         string `json:"FormatString,omitempty"`
}

func (v *NativeView) GetType() string {
	return v.Type
}

func (v *NativeView) GetName() string {
	return v.Name
}

func (v *NativeView) getBody(static bool) (string, error) {

	type subsetBody struct {
		Name        string   `json:"Name"`
		Alias       string   `json:"Alias,omitempty"`
		Hierarchy   string   `json:"Hierarchy@odata.bind"`
		Elements    []string `json:"Elements@odata.bind,omitempty"`
		Expression  string   `json:"Expression,omitempty"`
		ExpandAbove bool     `json:"ExpandAbove,omitempty"`
	}

	type axisBody struct {
		Subset   subsetBody `json:"Subset"`
		Selected string     `json:"Selected@odata.bind,omitempty"`
	}

	type nativeViewBody struct {
		Type    string     `json:"@odata.type"`
		Name    string     `json:"Name"`
		Columns []axisBody `json:"Columns"`
		Rows    []axisBody `json:"Rows"`
		Titles  []axisBody `json:"Titles,omitempty"`
	}

	body := nativeViewBody{
		Name: v.Name,
		Type: v.Type,
	}

	body.Columns = make([]axisBody, 0)
	for _, column := range v.Columns {
		columnBody := axisBody{
			Subset: subsetBody{
				Name:        column.Subset.Name,
				Alias:       column.Subset.Alias,
				Hierarchy:   fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')", column.Subset.Hierarchy.Dimension.Name, column.Subset.Hierarchy.Name),
				ExpandAbove: column.Subset.ExpandAbove,
			},
		}
		if len(column.Subset.Elements) > 0 && static {
			elementsBind := make([]string, len(column.Subset.Elements))
			for i, element := range column.Subset.Elements {
				elementsBind[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", column.Subset.Hierarchy.Dimension.Name, column.Subset.Hierarchy.Name, element.Name)
			}
			columnBody.Subset.Elements = elementsBind
		} else {
			columnBody.Subset.Expression = column.Subset.Expression
		}
		body.Columns = append(body.Columns, columnBody)
	}

	body.Rows = make([]axisBody, 0)
	for _, row := range v.Rows {
		rowBody := axisBody{
			Subset: subsetBody{
				Name:        row.Subset.Name,
				Alias:       row.Subset.Alias,
				Hierarchy:   fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')", row.Subset.Hierarchy.Dimension.Name, row.Subset.Hierarchy.Name),
				ExpandAbove: row.Subset.ExpandAbove,
			},
		}
		if len(row.Subset.Elements) > 0 && static {
			elementsBind := make([]string, len(row.Subset.Elements))
			for i, element := range row.Subset.Elements {
				elementsBind[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", row.Subset.Hierarchy.Dimension.Name, row.Subset.Hierarchy.Name, element.Name)
			}
			rowBody.Subset.Elements = elementsBind
		} else {
			rowBody.Subset.Expression = row.Subset.Expression
		}
		body.Rows = append(body.Rows, rowBody)
	}

	body.Titles = make([]axisBody, 0)
	for _, title := range v.Titles {
		titleBody := axisBody{
			Subset: subsetBody{
				Name:      title.Subset.Name,
				Alias:     title.Subset.Alias,
				Hierarchy: fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')", title.Subset.Hierarchy.Dimension.Name, title.Subset.Hierarchy.Name),
			},
		}
		if len(title.Subset.Elements) > 0 && static {
			elementsBind := make([]string, len(title.Subset.Elements))
			for i, element := range title.Subset.Elements {
				elementsBind[i] = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", title.Subset.Hierarchy.Dimension.Name, title.Subset.Hierarchy.Name, element.Name)
			}
			titleBody.Subset.Elements = elementsBind
		} else {
			titleBody.Subset.Expression = title.Subset.Expression
		}

		titleBody.Selected = fmt.Sprintf("Dimensions('%s')/Hierarchies('%s')/Elements('%s')", title.Selected.Hierarchy.Dimension.Name, title.Selected.Hierarchy.Name, title.Selected.Name)
		body.Titles = append(body.Titles, titleBody)
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
