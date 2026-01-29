package tm1

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// ExecuteMDXDataFrame executes an MDX query and returns the result as a gota DataFrame.
// dimensionNames is optional; when provided, it should match the coordinate order in the cellset.
func (cs *CellService) ExecuteMDXDataFrame(ctx context.Context, mdx string, cellProperties []string, sandboxName string, dimensionNames []string) (dataframe.DataFrame, error) {
	cellset, err := cs.ExecuteMDX(ctx, mdx, cellProperties, sandboxName)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	return CellsetToDataFrame(cellset, dimensionNames)
}

// ExecuteViewDataFrame executes a cube view and returns the result as a gota DataFrame.
// dimensionNames is optional; when provided, it should match the coordinate order in the cellset.
func (cs *CellService) ExecuteViewDataFrame(ctx context.Context, cubeName, viewName string, private bool, cellProperties []string, sandboxName string, dimensionNames []string) (dataframe.DataFrame, error) {
	cellset, err := cs.ExecuteView(ctx, cubeName, viewName, private, cellProperties, sandboxName)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	return CellsetToDataFrame(cellset, dimensionNames)
}

// CellsetToDataFrame converts a cellset into a gota DataFrame.
// dimensionNames is optional; when provided, it should match the coordinate order in the cellset.
func CellsetToDataFrame(cellset *Cellset, dimensionNames []string) (dataframe.DataFrame, error) {
	if cellset == nil || len(cellset.CellMap) == 0 {
		return dataframe.New(), nil
	}

	keys := make([]string, 0, len(cellset.CellMap))
	for key := range cellset.CellMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	maxCoordLen := 0
	propSet := make(map[string]struct{})

	for _, key := range keys {
		coords := splitCoordKey(key)
		if len(coords) > maxCoordLen {
			maxCoordLen = len(coords)
		}

		for prop := range cellset.CellMap[key] {
			propSet[prop] = struct{}{}
		}
	}

	dimNames := normalizeDimensionNames(dimensionNames, maxCoordLen)

	propNames := make([]string, 0, len(propSet))
	for prop := range propSet {
		if prop != "Value" && prop != "Ordinal" {
			propNames = append(propNames, prop)
		}
	}
	sort.Strings(propNames)

	coordCols := make([][]string, maxCoordLen)
	for i := range coordCols {
		coordCols[i] = make([]string, 0, len(keys))
	}

	propCols := make(map[string][]interface{})
	propCols["Value"] = make([]interface{}, 0, len(keys))
	for _, prop := range propNames {
		propCols[prop] = make([]interface{}, 0, len(keys))
	}

	for _, key := range keys {
		coords := splitCoordKey(key)
		for i := 0; i < maxCoordLen; i++ {
			value := ""
			if i < len(coords) {
				value = coords[i]
			}
			coordCols[i] = append(coordCols[i], value)
		}

		props := cellset.CellMap[key]
		propCols["Value"] = append(propCols["Value"], props["Value"])
		for _, prop := range propNames {
			propCols[prop] = append(propCols[prop], props[prop])
		}
	}

	seriesList := make([]series.Series, 0, maxCoordLen+1+len(propNames))
	for i, dimName := range dimNames {
		seriesList = append(seriesList, series.New(coordCols[i], series.String, dimName))
	}

	seriesList = append(seriesList, buildSeriesFromInterfaces("Value", propCols["Value"]))
	for _, prop := range propNames {
		seriesList = append(seriesList, buildSeriesFromInterfaces(prop, propCols[prop]))
	}

	return dataframe.New(seriesList...), nil
}

func splitCoordKey(coordKey string) []string {
	if coordKey == "" {
		return nil
	}

	parts := strings.Split(coordKey, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

func normalizeDimensionNames(dimensionNames []string, maxLen int) []string {
	if maxLen == 0 {
		return nil
	}

	names := make([]string, 0, maxLen)

	for i := 0; i < maxLen; i++ {
		if i < len(dimensionNames) && strings.TrimSpace(dimensionNames[i]) != "" {
			names = append(names, dimensionNames[i])
			continue
		}
		names = append(names, "Dim"+strconv.Itoa(i+1))
	}

	return names
}

func buildSeriesFromInterfaces(name string, values []interface{}) series.Series {
	if len(values) == 0 {
		return series.New([]string{}, series.String, name)
	}

	var (
		hasBool   bool
		hasNumber bool
		hasString bool
		hasOther  bool
	)

	for _, v := range values {
		if v == nil {
			continue
		}
		switch v.(type) {
		case bool:
			hasBool = true
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			hasNumber = true
		case string:
			hasString = true
		default:
			hasOther = true
		}
	}

	if hasOther || (hasString && (hasNumber || hasBool)) || (hasBool && hasNumber) {
		return toStringSeries(name, values)
	}

	if hasBool {
		return toBoolSeries(name, values)
	}

	if hasNumber {
		return toFloatSeries(name, values)
	}

	return toStringSeries(name, values)
}

func toStringSeries(name string, values []interface{}) series.Series {
	stringsCol := make([]string, len(values))
	for i, v := range values {
		if v == nil {
			stringsCol[i] = ""
			continue
		}
		stringsCol[i] = fmt.Sprint(v)
	}
	return series.New(stringsCol, series.String, name)
}

func toBoolSeries(name string, values []interface{}) series.Series {
	bools := make([]bool, len(values))
	for i, v := range values {
		if v == nil {
			bools[i] = false
			continue
		}
		b, ok := v.(bool)
		if ok {
			bools[i] = b
			continue
		}
		bools[i] = false
	}
	return series.New(bools, series.Bool, name)
}

func toFloatSeries(name string, values []interface{}) series.Series {
	floatVals := make([]float64, len(values))
	for i, v := range values {
		if v == nil {
			floatVals[i] = math.NaN()
			continue
		}
		switch n := v.(type) {
		case float64:
			floatVals[i] = n
		case float32:
			floatVals[i] = float64(n)
		case int:
			floatVals[i] = float64(n)
		case int8:
			floatVals[i] = float64(n)
		case int16:
			floatVals[i] = float64(n)
		case int32:
			floatVals[i] = float64(n)
		case int64:
			floatVals[i] = float64(n)
		case uint:
			floatVals[i] = float64(n)
		case uint8:
			floatVals[i] = float64(n)
		case uint16:
			floatVals[i] = float64(n)
		case uint32:
			floatVals[i] = float64(n)
		case uint64:
			floatVals[i] = float64(n)
		default:
			floatVals[i] = math.NaN()
		}
	}
	return series.New(floatVals, series.Float, name)
}
