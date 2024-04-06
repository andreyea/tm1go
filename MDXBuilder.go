package tm1go

import (
	"fmt"
	"strconv"
	"strings"
)

type MDXBuilder struct {
	With  []string
	Cube  string
	Axes  []MDXAxis
	Where []MDXMember
}

func (b *MDXBuilder) AddCube(cubeName string) {
	b.Cube = cubeName
}

func NewMDXBuilder(cube string) *MDXBuilder {
	return &MDXBuilder{
		Cube: cube,
	}
}

func (b *MDXBuilder) AddWithStatement(member string) {
	b.With = append(b.With, member)
}

func (b *MDXBuilder) AddExpressionOnColumns(expression string) {
	if len(b.Axes) == 0 {
		b.Axes = append(b.Axes, MDXAxis{})
	}
	b.Axes[0].CustomExpression = expression
}

func (b *MDXBuilder) AddExpressionOnRows(expression string) {
	if len(b.Axes) == 1 {
		b.Axes = append(b.Axes, MDXAxis{})
	}
	b.Axes[1].CustomExpression = expression
}

func (b *MDXBuilder) AddMemberOnColumns(dimension, hierarchy, name string) {
	if len(b.Axes) == 0 {
		b.Axes = append(b.Axes, MDXAxis{})
	}
	member := MDXMember{
		Dimension: dimension,
		Hierarchy: hierarchy,
		Name:      name,
	}
	tuple := MDXTuple{
		Members: []MDXMember{member},
	}
	b.Axes[0].Tuples = append(b.Axes[0].Tuples, tuple)
}

func (b *MDXBuilder) AddMemberOnRows(dimension, hierarchy, name string) {
	if len(b.Axes) == 1 {
		b.Axes = append(b.Axes, MDXAxis{})
	}
	member := MDXMember{
		Dimension: dimension,
		Hierarchy: hierarchy,
		Name:      name,
	}
	tuple := MDXTuple{
		Members: []MDXMember{member},
	}
	b.Axes[1].Tuples = append(b.Axes[1].Tuples, tuple)
}

func (b *MDXBuilder) AddMemberOnWhere(dimension, hierarchy, name string) {
	member := MDXMember{
		Dimension: dimension,
		Hierarchy: hierarchy,
		Name:      name,
	}
	b.Where = append(b.Where, member)
}

func (b *MDXBuilder) ToString() (string, error) {
	if len(b.Axes) == 0 {
		return "", fmt.Errorf("MDXBuilder must have at least one axis")
	}

	if b.Cube == "" {
		return "", fmt.Errorf("MDXBuilder must have a cube")
	}

	var mdx string
	if len(b.With) > 0 {
		mdx += "WITH " + strings.Join(b.With, ", ") + " "
	}

	mdx = "SELECT "

	for i, axis := range b.Axes {
		expression := axis.ToString()
		if expression != "" {
			mdx += axis.ToString() + " ON " + strconv.Itoa(i) + ","
		}
	}
	mdx = mdx[:len(mdx)-1] + " FROM [" + b.Cube + "] "

	if len(b.Where) > 0 {
		mdx += "WHERE ("
		for _, member := range b.Where {
			mdx += member.ToString() + ","
		}
		mdx = mdx[:len(mdx)-1] + ")"
	}

	return mdx, nil
}

type MDXAxis struct {
	SuppressZeroes   bool
	IgnoreBadTuples  bool
	Tuples           []MDXTuple
	CustomExpression string
}

func (a *MDXAxis) ToString() string {
	if a.CustomExpression != "" {
		return a.CustomExpression
	}
	var axis string
	if a.SuppressZeroes {
		axis += "NON EMPTY "
	}
	if a.IgnoreBadTuples {
		axis += "TM1IGNORE_BADTUPLES "
	}
	for i, tuple := range a.Tuples {
		if i == 0 {
			axis += "{"
		}
		axis += tuple.ToString() + ","
	}
	if axis != "" {
		axis = axis[:len(axis)-1] + "}"
	}
	return axis
}

type MDXMember struct {
	Dimension string
	Hierarchy string
	Name      string
}

func (m *MDXMember) ToString() string {
	return "[" + m.Dimension + "].[" + m.Hierarchy + "].[" + m.Name + "]"
}

type MDXTuple struct {
	Members []MDXMember
}

func (t *MDXTuple) ToString() string {
	var tuple string
	for _, member := range t.Members {
		tuple += member.ToString() + ", "
	}
	tuple = "(" + tuple[:len(tuple)-2] + ")"
	return tuple
}
