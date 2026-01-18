package tm1

import (
	"testing"
)

func TestCellService_GetValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("get value from cell", func(t *testing.T) {
		// This would need a real cube setup to test
		t.Skip("requires test cube setup")
	})
}

func TestCellService_ExecuteMDX(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("execute simple MDX query", func(t *testing.T) {
		t.Skip("requires TM1 server connection")
	})
}

func TestCellService_ExecuteView(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("execute existing view", func(t *testing.T) {
		// This would need a real view to test
		t.Skip("requires test view setup")
	})
}

func TestCellService_WriteValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("write single value", func(t *testing.T) {
		// This would need a real cube and proper write permissions
		t.Skip("requires test cube with write access")
	})
}

func TestCellService_WriteValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("write multiple values", func(t *testing.T) {
		// This would need a real cube and proper write permissions
		t.Skip("requires test cube with write access")
	})
}

func TestCellService_CreateCellset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("create cellset from MDX", func(t *testing.T) {
		t.Skip("requires TM1 server connection")
	})
}

func TestCellService_ordinalToCoordinates(t *testing.T) {
	cs := &CellService{}

	tests := []struct {
		name              string
		ordinal           int
		axisCardinalities []int
		expected          []int
	}{
		{
			name:              "simple 2x3 grid",
			ordinal:           4,
			axisCardinalities: []int{2, 3},
			expected:          []int{1, 1},
		},
		{
			name:              "first cell",
			ordinal:           0,
			axisCardinalities: []int{2, 3},
			expected:          []int{0, 0},
		},
		{
			name:              "3D cube",
			ordinal:           13,
			axisCardinalities: []int{2, 3, 4},
			expected:          []int{1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cs.ordinalToCoordinates(tt.ordinal, tt.axisCardinalities)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d coordinates, got %d", len(tt.expected), len(result))
				return
			}

			for i, coord := range result {
				if coord != tt.expected[i] {
					t.Errorf("Coordinate %d: expected %d, got %d", i, tt.expected[i], coord)
				}
			}
		})
	}
}
