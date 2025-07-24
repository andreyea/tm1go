package tm1go_test

import (
	"testing"

	"github.com/andreyea/tm1go"
)

func TestCellService_CreateCellSet(t *testing.T) {
	type args struct {
		mdx     string
		sandbox string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Create a cellset for base sandbox",
			args:    args{mdx: "SELECT [d2].members ON COLUMNS , [d1].members ON ROWS FROM [3D]  WHERE ([Measure].[Measure].[Value])", sandbox: ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tm1ServiceT.CellService.CreateCellSet(tt.args.mdx, tt.args.sandbox)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestCellService.CreateCellSet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if id == "" {
				t.Errorf("TestCellService.CreateCellSet() error = %v, wantErr %v. Empty cellset id.", err, tt.wantErr)
			}

		})
	}
}

func TestCellService_DeleteCellSet(t *testing.T) {
	// Prepare test by creating a cellset
	cellsetID, err := tm1ServiceT.CellService.CreateCellSet("SELECT [Time].members ON COLUMNS , [Account].members ON ROWS FROM [Balance Sheet]", "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		wantErr   bool
	}{
		{
			name:      "Delete existing cellset",
			cellsetID: cellsetID,
			wantErr:   false,
		},
		{
			name:      "Delete non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CellService.DeleteCellSet(tt.cellsetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.DeleteCellSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCellService_ExtractCellSetCount(t *testing.T) {
	// Prepare test by creating a cellset
	mdx := "SELECT [Time].members ON COLUMNS , [Account].members ON ROWS FROM [Balance Sheet]"
	cellsetID, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Extract count from valid cellset",
			cellsetID: cellsetID,
			wantCount: 0, // Replace with expected count
			wantErr:   false,
		},
		{
			name:      "Extract count from non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			wantCount: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tm1ServiceT.CellService.ExtractCellSetCount(tt.cellsetID, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.ExtractCellSetCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Check if the count is greater than or equal to the expected count
			if count <= tt.wantCount {
				t.Errorf("CellService.ExtractCellSetCount() = %v, want %v", count, tt.wantCount)
			}
		})
	}

	// Clean up by deleting the cellset
	err = tm1ServiceT.CellService.DeleteCellSet(cellsetID)
	if err != nil {
		t.Fatalf("Failed to delete cellset: %v", err)
	}
}

func TestCellService_ExtractCellsetCellsRaw(t *testing.T) {
	// Prepare test by creating a cellset
	mdx := `
		SELECT 
		{[Measure].[Measure].[Value]} 
		ON COLUMNS, 
		{[Line].[Line].[e1],[Line].[Line].[e2],[Line].[Line].[e3]} 
		ON ROWS 
		FROM [2D] 
	`

	cellsetID, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		wantErr   bool
	}{
		{
			name:      "Extract cells from valid cellset",
			cellsetID: cellsetID,
			wantErr:   false,
		},
		{
			name:      "Extract cells from non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cells, err := tm1ServiceT.CellService.ExtractCellsetCellsRaw(tt.cellsetID, []string{"Value"}, 0, 0, false, false, false, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.ExtractCellsetCellsRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cells == nil {
					t.Errorf("CellService.ExtractCellsetCellsRaw() returned nil cells for valid cellset")
				}
				// Add more specific checks on the returned cells if needed
				// For example, check if the number of cells matches expected count
				// expectedCount := ... // Set this based on your MDX query
				// if len(cells) != expectedCount {
				//     t.Errorf("CellService.ExtractCellsetCellsRaw() returned %d cells, want %d", len(cells), expectedCount)
				// }
			}
		})
	}

	// Clean up by deleting the cellset
	err = tm1ServiceT.CellService.DeleteCellSet(cellsetID)
	if err != nil {
		t.Fatalf("Failed to delete cellset: %v", err)
	}
}

func TestCellService_ExtractCellsetCellsAsync(t *testing.T) {
	// Prepare test by creating a cellset
	mdx := "SELECT [Time].members ON COLUMNS , [Account].members ON ROWS FROM [Balance Sheet]"
	cellsetID, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		wantErr   bool
	}{
		{
			name:      "Extract cells asynchronously from valid cellset",
			cellsetID: cellsetID,
			wantErr:   false,
		},
		{
			name:      "Extract cells asynchronously from non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cells, err := tm1ServiceT.CellService.ExtractCellsetCellsAsync(tt.cellsetID, []string{"Value", "FormattedValue"}, "", 5)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.ExtractCellsetCellsAsync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cells == nil {
					t.Errorf("CellService.ExtractCellsetCellsAsync() returned nil cells for valid cellset")
				}
				// Add more specific checks on the returned cells if needed
				// For example, check if the number of cells matches expected count
				// expectedCount := ... // Set this based on your MDX query
				// if len(cells) != expectedCount {
				//     t.Errorf("CellService.ExtractCellsetCellsAsync() returned %d cells, want %d", len(cells), expectedCount)
				// }
			}
		})

	}

}

func TestCellService_UpdateCellset(t *testing.T) {
	// Prepare test by creating a cellset
	mdx := `
		SELECT 
		{[Measure].[Measure].[Value]} 
		ON COLUMNS, 
		{[Line].[Line].[e1],[Line].[Line].[e2],[Line].[Line].[e3]} 
		ON ROWS 
		FROM [2D] 
	`
	cellsetID, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		updates   []interface{}
		wantErr   bool
	}{
		{
			name:      "Update valid cellset",
			cellsetID: cellsetID,
			updates:   []interface{}{1.1, 2.123, 3.3},
			wantErr:   false,
		},
		{
			name:      "Update non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			updates:   []interface{}{1},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CellService.UpdateCellset(tt.cellsetID, tt.updates, 0, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.UpdateCellset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the update by extracting the cells and checking their values
				// Execute mdx again
				cellsetID2, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
				if err != nil {
					t.Fatalf("Failed to create cellset: %v", err)
				}
				cells, err := tm1ServiceT.CellService.ExtractCellsetCellsRaw(cellsetID2, []string{"Value"}, 0, 0, false, false, false, "")
				if err != nil {
					t.Errorf("Failed to extract cells after update: %v", err)
					return
				}

				for i, update := range tt.updates {
					updateValue, ok := update.(float64)
					if !ok {
						t.Errorf("Update value at index %d is not a float64", i)
						continue
					}
					cellValue, ok := cells[i].Value.(float64)
					if !ok {
						t.Errorf("Cell value at ordinal %d is not a float64", i)
						continue
					}
					if updateValue != cellValue {
						t.Errorf("Cell at ordinal %d was not updated correctly.", i)
					}
				}
			}
		})
	}

	// Clean up by deleting the cellset
	err = tm1ServiceT.CellService.DeleteCellSet(cellsetID)
	if err != nil {
		t.Fatalf("Failed to delete cellset: %v", err)
	}
}

func TestCellService_UpdateCellsetMDX(t *testing.T) {
	mdx := `
	SELECT 
	{[Measure].[Measure].[Value]} 
	ON COLUMNS, 
	{[Line].[Line].[e1],[Line].[Line].[e2],[Line].[Line].[e3]} 
	ON ROWS 
	FROM [2D] 
`
	tests := []struct {
		name    string
		mdx     string
		updates []interface{}
		wantErr bool
	}{
		{
			name:    "Update using valid mdx",
			mdx:     mdx,
			updates: []interface{}{1.1, 2.123, 3.3},
			wantErr: false,
		},
		{
			name:    "Try to update with invalid mdx",
			mdx:     "select columns and rows....",
			updates: []interface{}{1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CellService.UpdateCellsetMDX(tt.mdx, tt.updates, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.UpdateCellsetMDX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the update by extracting the cells and checking their values
				// Execute mdx again
				cellsetID2, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
				if err != nil {
					t.Fatalf("Failed to create cellset: %v", err)
				}
				cells, err := tm1ServiceT.CellService.ExtractCellsetCellsRaw(cellsetID2, []string{"Value"}, 0, 0, false, false, false, "")
				if err != nil {
					t.Errorf("Failed to extract cells after update: %v", err)
					return
				}

				for i, update := range tt.updates {
					updateValue, ok := update.(float64)
					if !ok {
						t.Errorf("Update value at index %d is not a float64", i)
						continue
					}
					cellValue, ok := cells[i].Value.(float64)
					if !ok {
						t.Errorf("Cell value at ordinal %d is not a float64", i)
						continue
					}
					if updateValue != cellValue {
						t.Errorf("Cell at ordinal %d was not updated correctly. Got %v, want %v", i, cellValue, updateValue)
					}
				}
			}
		})
	}
}

func TestCellService_CellGet(t *testing.T) {
	tests := []struct {
		name      string
		cubeName  string
		elements  []string
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "Get existing cell",
			cubeName:  "2D",
			elements:  []string{"E1", "Value"},
			wantValue: 1.0,
			wantErr:   false,
		},
		{
			name:      "Get cell from non-existent cube",
			cubeName:  "NonExistentCube",
			elements:  []string{"Element1", "Element2"},
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "Get cell with invalid elements",
			cubeName:  "Balance Sheet",
			elements:  []string{"InvalidYear", "InvalidAccount"},
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allParams := append([]string{tt.cubeName}, tt.elements...)
			gotValue, err := tm1ServiceT.CellService.CellGet(allParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.CellGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				_, ok := gotValue.(float64)
				if !ok {
					t.Errorf("CellService.CellGet() - unable to convert recieved result to float64")
				}
			}
		})
	}
}

func TestCellService_CellPut(t *testing.T) {
	tests := []struct {
		name     string
		cubeName string
		elements []interface{}
		value    interface{}
		wantErr  bool
	}{
		{
			name:     "Put value in existing cell",
			cubeName: "2D",
			elements: []interface{}{"E1", "Value"},
			value:    110000.01,
			wantErr:  false,
		},
		{
			name:     "Put value in non-existent cube",
			cubeName: "NonExistentCube",
			elements: []interface{}{"Element1", "Element2"},
			value:    100.0,
			wantErr:  true,
		},
		{
			name:     "Put value with invalid elements",
			cubeName: "Balance Sheet",
			elements: []interface{}{"InvalidYear", "InvalidAccount"},
			value:    100.0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allParams := append([]interface{}{tt.value}, append([]interface{}{tt.cubeName}, tt.elements...)...)
			err := tm1ServiceT.CellService.CellPut(allParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.CellPut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the update by getting the cell value
				// Convert tt.elements to []string
				elements := make([]string, len(tt.elements))
				for i, element := range tt.elements {
					elements[i] = element.(string)
				}
				allParams := append([]string{tt.cubeName}, elements...)
				gotValue, err := tm1ServiceT.CellService.CellGet(allParams...)
				if err != nil {
					t.Errorf("Failed to get cell after put: %v", err)
					return
				}
				if gotValue.(float64) != tt.value.(float64) {
					t.Errorf("CellPut() failed to update. Got %v, want %v", gotValue, tt.value)
				}
			}
		})
	}
}

func TestCellService_ExtractCellsetAxesAndCube(t *testing.T) {
	// Prepare test by creating a cellset
	mdx := `
	SELECT 
	{[Measure].[Measure].[Value]} 
	ON COLUMNS, 
	{[Line].[Line].[e1],[Line].[Line].[e2],[Line].[Line].[e3]} 
	ON ROWS 
	FROM [2D] 
`
	cellsetID, err := tm1ServiceT.CellService.CreateCellSet(mdx, "")
	if err != nil {
		t.Fatalf("Failed to create cellset: %v", err)
	}

	tests := []struct {
		name      string
		cellsetID string
		wantCube  string
		wantAxes  int
		wantErr   bool
	}{
		{
			name:      "Extract axes and cube from valid cellset",
			cellsetID: cellsetID,
			wantCube:  "2D",
			wantAxes:  2, // Columns, Rows
			wantErr:   false,
		},
		{
			name:      "Extract from non-existent cellset",
			cellsetID: "nonExistentCellsetID",
			wantCube:  "",
			wantAxes:  0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cellset, err := tm1ServiceT.CellService.ExtractCellsetAxesAndCube(tt.cellsetID, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.ExtractCellsetAxesAndCube() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cellset.Cube.Name != tt.wantCube {
					t.Errorf("CellService.ExtractCellsetAxesAndCube() cube = %v, want %v", cellset.Cube.Name, tt.wantCube)
				}
				if len(cellset.Axes) != tt.wantAxes {
					t.Errorf("CellService.ExtractCellsetAxesAndCube() got %v axes, want %v", len(cellset.Axes), tt.wantAxes)
				}

				// Check the content of axes
				expectedDimensions := []string{"Measure", "Line"}
				for i, axis := range cellset.Axes {
					if axis.Hierarchies[0].Dimension.Name != expectedDimensions[i] {
						t.Errorf("Axis %d dimension = %v, want %v", i, axis.Hierarchies[0].Dimension.Name, expectedDimensions[i])
					}
				}
			}
		})
	}

	// Clean up by deleting the cellset
	err = tm1ServiceT.CellService.DeleteCellSet(cellsetID)
	if err != nil {
		t.Fatalf("Failed to delete cellset: %v", err)
	}
}

func TestCellService_UpdateCellsetFromDataframe(t *testing.T) {

	// Simulate a dataframe with a slice of maps
	// Create a DataFrame with three lines
	df := &tm1go.DataFrame{
		Headers: []string{"Line", "Measure", "Value"},
		Columns: map[string][]interface{}{
			"Line":    {"e1", "e2", "e3"},
			"Measure": {"Value", "Value", "Value"},
			"Value":   {1000000.0, 1200000.0, 1500000.0},
		},
	}

	tests := []struct {
		name      string
		cubeName  string
		dataframe *tm1go.DataFrame
		wantErr   bool
	}{
		{
			name:      "Update valid cube from dataframe",
			cubeName:  "2D",
			dataframe: df,
			wantErr:   false,
		},
		{
			name:      "Update non-existent cube",
			cubeName:  "nonExistentCellsetID",
			dataframe: df,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CellService.UpdateCellsetFromDataframe(tt.cubeName, tt.dataframe, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.UpdateCellsetFromDataframe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCellService_ExecuteMdxToDataframe(t *testing.T) {
	// Prepare test MDX query
	mdx := `
	SELECT 
	{[Measure].[Measure].[Value]} 
	ON COLUMNS, 
	{[Line].[Line].[e1],[Line].[Line].[e2],[Line].[Line].[e3]} 
	ON ROWS 
	FROM [2D] 
`

	tests := []struct {
		name        string
		mdx         string
		sandboxName string
		wantErr     bool
	}{
		{
			name:        "Valid MDX query",
			mdx:         mdx,
			sandboxName: "",
			wantErr:     false,
		},
		{
			name:        "Invalid MDX query",
			mdx:         "SELECT invalid MDX",
			sandboxName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			df, err := tm1ServiceT.CellService.ExecuteMdxToDataframe(tt.mdx, tt.sandboxName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.ExecuteMdxToDataframe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if df.RowCount() == 0 {
					t.Errorf("CellService.ExecuteMdxToDataframe() returned empty dataframe")
				}
			}
		})
	}
}
func TestCellService_UpdateCellsetFromDataframeViaBlob(t *testing.T) {
	// Prepare test data
	cubeName := "2D"
	df := tm1go.NewDataFrame([]string{"Line", "Measure", "Value"})
	_ = df.AddRow([]interface{}{"e1", "Value", 21000000.0})
	_ = df.AddRow([]interface{}{"e2", "Value", 2800000.0})
	_ = df.AddRow([]interface{}{"e3", "Value", 212200000.0})
	_ = df.AddRow([]interface{}{"e4", "Value", 200000.0})

	tests := []struct {
		name        string
		cubeName    string
		df          *tm1go.DataFrame
		sandboxName string
		wantErr     bool
	}{
		{
			name:        "Valid update",
			cubeName:    cubeName,
			df:          df,
			sandboxName: "",
			wantErr:     false,
		},
		{
			name:        "Invalid cube name",
			cubeName:    "NonExistentCube",
			df:          df,
			sandboxName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.CellService.UpdateCellsetFromDataframeViaBlob(tt.cubeName, tt.df, tt.sandboxName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellService.UpdateCellsetFromDataframeViaBlob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
