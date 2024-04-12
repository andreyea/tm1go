package tm1go

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type DataFrame struct {
	Columns map[string][]interface{}
	Headers []string
}

func NewDataFrame(columns []string) *DataFrame {
	df := &DataFrame{
		Columns: make(map[string][]interface{}),
		Headers: columns,
	}
	for _, col := range columns {
		df.Columns[col] = make([]interface{}, 0)
	}
	return df
}

func (df *DataFrame) AddRow(row []interface{}) error {
	if len(row) != len(df.Headers) {
		return fmt.Errorf("the number of values in the row does not match the number of columns")
	}
	for i, value := range row {
		colName := df.Headers[i]
		df.Columns[colName] = append(df.Columns[colName], value)
	}
	return nil
}

// SortByColumns sorts the DataFrame based on the specified columns
func (df *DataFrame) SortByColumns(columnNames []string) error {
	indices := make([]int, len(df.Columns[df.Headers[0]]))
	for i := range indices {
		indices[i] = i
	}

	// Reverse iterate over the columnNames to ensure the primary sort key takes precedence
	for i := len(columnNames) - 1; i >= 0; i-- {
		columnName := columnNames[i]
		column, ok := df.Columns[columnName]
		if !ok {
			return fmt.Errorf("column %s does not exist", columnName)
		}

		sort.SliceStable(indices, func(i, j int) bool {
			// Compare the values in the specified column. Extend these cases for other types as needed
			switch column[indices[i]].(type) {
			case int:
				return column[indices[i]].(int) < column[indices[j]].(int)
			case float64:
				return column[indices[i]].(float64) < column[indices[j]].(float64)
			case string:
				return strings.Compare(column[indices[i]].(string), column[indices[j]].(string)) < 0
			default:
				return false // Extend to handle other types
			}
		})
	}

	// Reorder the columns based on sorted indices.
	sortedColumns := make(map[string][]interface{})
	for _, header := range df.Headers {
		sortedColumns[header] = make([]interface{}, len(indices))
		for newIndex, oldIndex := range indices {
			sortedColumns[header][newIndex] = df.Columns[header][oldIndex]
		}
	}

	df.Columns = sortedColumns

	return nil
}

// SortByColumn sorts the DataFrame based on the specified column
func (df *DataFrame) SortByColumn(columnName string) error {
	// Retrieve the column to sort by.
	column, ok := df.Columns[columnName]
	if !ok {
		return fmt.Errorf("column %s does not exist", columnName)
	}

	// Create a slice of indices and sort it based on the values in the specified column
	indices := make([]int, len(column))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		// Handle sorting for ints and strings
		switch column[i].(type) {
		case int:
			return column[i].(int) < column[j].(int)
		case string:
			return column[i].(string) < column[j].(string)
		default:
			return false // This could be extended to handle other types
		}
	})

	// Create a new map to hold the sorted columns
	sortedColumns := make(map[string][]interface{})
	for _, header := range df.Headers {
		sortedColumns[header] = make([]interface{}, len(column))
		for newIndex, oldIndex := range indices {
			sortedColumns[header][newIndex] = df.Columns[header][oldIndex]
		}
	}

	df.Columns = sortedColumns

	return nil
}

// ExportToCSV exports the DataFrame to a CSV file with the given filename
func (df *DataFrame) ExportToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the headers (column names) as the first row
	if err := writer.Write(df.Headers); err != nil {
		return fmt.Errorf("error writing headers to CSV: %w", err)
	}

	// Iterate over the rows and write each one
	numRows := len(df.Columns[df.Headers[0]])
	for i := 0; i < numRows; i++ {
		row := make([]string, len(df.Headers))
		for j, header := range df.Headers {
			// Convert each value to string
			row[j] = fmt.Sprintf("%v", df.Columns[header][i])
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row to CSV: %w", err)
		}
	}

	return nil
}

// AddColumn adds a new column to the DataFrame with the provided name and values.
// If the number of values doesn't match the number of rows in the DataFrame,
// it returns an error without adding the column.
func (df *DataFrame) AddColumn(name string, values []interface{}) error {
	// Check if the column already exists
	if _, exists := df.Columns[name]; exists {
		return fmt.Errorf("column %s already exists", name)
	}

	// Ensure the new column has the correct number of rows
	expectedRows := len(df.Columns[df.Headers[0]])
	if len(values) != expectedRows {
		return fmt.Errorf("expected %d values, got %d", expectedRows, len(values))
	}

	// Add the new column
	df.Columns[name] = values
	df.Headers = append(df.Headers, name)

	return nil
}

// RowCount returns the number of rows in the DataFrame.
func (df *DataFrame) RowCount() int {
	if len(df.Headers) == 0 {
		return 0 // No columns, hence no rows
	}
	firstColumnName := df.Headers[0]
	return len(df.Columns[firstColumnName])
}

// DeleteRow deletes the row at the specified index from the DataFrame
// If the index is out of range, it returns an error
func (df *DataFrame) DeleteRow(index int) error {
	// Check if the index is out of range
	numRows := df.RowCount()
	if index < 0 || index >= numRows {
		return fmt.Errorf("index out of range: %d", index)
	}

	// Delete the row from each column
	for _, col := range df.Headers {
		df.Columns[col] = append(df.Columns[col][:index], df.Columns[col][index+1:]...)
	}

	return nil
}

// BuildMDX converts dataframes to MDX query
func (df *DataFrame) ToMDX(cubeName string) (string, error) {
	if len(df.Headers) < 3 { // Need at least one dimension column and one value column
		return "", fmt.Errorf("dataFrame must contain at least 3 columns")
	}

	// Find uniform columns which will be on where clause
	uniformColumnIndices := df.FindUniformColumnIndices()

	rowCount := df.RowCount()
	axis := ""
	whereSlice := make([]string, 0)
	for i := 0; i < rowCount; i++ {

		tuple := "("

		for j := 0; j < len(df.Headers)-1; j++ {
			// Check if the column is uniform
			if SliceContains(uniformColumnIndices, j) {
				dim, hier := ExtractDimensionHierarchyFromString(df.Headers[j])
				whereSlice = append(whereSlice, fmt.Sprintf("[%s].[%s].[%v]", dim, hier, df.Columns[df.Headers[j]][i]))
			} else {
				dim, hier := ExtractDimensionHierarchyFromString(df.Headers[j])
				member := fmt.Sprintf("[%s].[%s].[%v]", dim, hier, df.Columns[df.Headers[j]][i])
				tuple += member + ","
			}
		}

		tuple = strings.TrimRight(tuple, ",") + "),"
		axis += tuple
	}
	axis = "{" + strings.TrimRight(axis, ",") + "} ON 0"
	whereString := ""
	if len(whereSlice) > 0 {
		whereSlice = UniqueStrings(whereSlice)
		whereString = " WHERE (" + strings.Join(whereSlice, ",") + ")"
	}
	// Construct the MDX query
	mdx := fmt.Sprintf("SELECT %s FROM [%s]%s", axis, cubeName, whereString)
	return mdx, nil
}

// FindUniformColumnIndicesParallel finds indices of uniform columns using multiple Go routines
func (df *DataFrame) FindUniformColumnIndicesParallel() []int {
	var wg sync.WaitGroup
	uniformColumnIndicesChan := make(chan int, len(df.Headers))

	for idx, colName := range df.Headers {
		wg.Add(1)
		// Launch a Go routine for each column
		go func(idx int, colName string) {
			defer wg.Done()
			column := df.Columns[colName]
			isUniform := true

			if len(column) > 1 {
				firstValue := column[0]
				for _, value := range column[1:] {
					if value != firstValue {
						isUniform = false
						break
					}
				}
			}

			if isUniform {
				uniformColumnIndicesChan <- idx
			}
		}(idx, colName)
	}

	// Close the channel once all Go routines have finished
	go func() {
		wg.Wait()
		close(uniformColumnIndicesChan)
	}()

	var uniformColumnIndices []int
	for idx := range uniformColumnIndicesChan {
		uniformColumnIndices = append(uniformColumnIndices, idx)
	}

	return uniformColumnIndices
}

// FindUniformColumnIndices searches for uniform columns
func (df *DataFrame) FindUniformColumnIndices() []int {
	uniformColumnIndices := make([]int, 0, len(df.Headers)) // Allocate with expected capacity to minimize reallocations

	for idx, colName := range df.Headers {
		column := df.Columns[colName]
		isUniform := true

		// Directly include columns with 0 or 1 values as uniform
		if len(column) <= 1 {
			uniformColumnIndices = append(uniformColumnIndices, idx)
			continue
		}

		// Use the first value as a reference point for comparison
		firstValue := column[0]
		for _, value := range column[1:] {
			// If a different value is found, mark as not uniform and break
			if value != firstValue {
				isUniform = false
				break
			}
		}

		// If the column is uniform, include its index
		if isUniform {
			uniformColumnIndices = append(uniformColumnIndices, idx)
		}
	}

	return uniformColumnIndices
}

// Convert cellset to dataframe
func CellSetToDataFrame(cellset *Cellset) (*DataFrame, error) {
	columnsCount := cellset.Axes[0].Cardinality
	headers := make([]string, 0)
	for _, axis := range cellset.Axes {
		for _, hierarchy := range axis.Hierarchies {
			headers = append(headers, hierarchy.UniqueName)
		}
	}
	headers = append(headers, "Value")

	df := NewDataFrame(headers)
	for i, cell := range cellset.Cells {
		row := make([]interface{}, 0, len(headers))

		y := i / columnsCount
		x := i % columnsCount

		for _, m := range cellset.Axes[0].Tuples[x].Members {
			row = append(row, m.Name)
		}

		if len(cellset.Axes) > 1 {
			for _, m := range cellset.Axes[1].Tuples[y].Members {
				row = append(row, m.Name)
			}
		}

		if len(cellset.Axes) > 2 {
			for _, m := range cellset.Axes[2].Tuples[0].Members {
				row = append(row, m.Name)
			}
		}

		row = append(row, cell.Value)
		err := df.AddRow(row)
		if err != nil {
			return nil, fmt.Errorf("error adding row %d to dataframe: %w", i, err)
		}
	}

	return df, nil
}
