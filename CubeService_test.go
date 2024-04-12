package tm1go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCubeService_Get(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL
		expectedURL := "/Cubes('TestCube')?$expand=Dimensions($select=Name)"
		if r.URL.Path != expectedURL {
			t.Errorf("Expected URL: %s, got: %s", expectedURL, r.URL.Path)
		}

		// Send a mock response
		cube := Cube{
			Dimensions: []Dimension{{Name: "Dimension1"}, {Name: "Dimension2"}},
		}
		jsonBytes, _ := json.Marshal(cube)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	}))
	defer server.Close()

	// Create a new CubeService with the mock server URL
	cs := &CubeService{
		rest: NewRestClient(TM1ServiceConfig{BaseURL: "https://localhost:8010",
			User:     "admin",
			Password: ""}),
	}

	// Call the Get method
	cube, err := cs.Get("TestCube")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the returned cube
	expectedCube := &Cube{
		Dimensions: []Dimension{{Name: "Dimension1"}, {Name: "Dimension2"}},
	}
	if !reflect.DeepEqual(cube, expectedCube) {
		t.Errorf("Expected cube: %+v, got: %+v", expectedCube, cube)
	}
}
