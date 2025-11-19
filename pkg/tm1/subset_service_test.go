package tm1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreyea/tm1go/pkg/models"
)

func TestSubsetServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/Dimensions('Region')/Hierarchies('Region')/Subsets") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["Name"] != "TopRegions" {
			t.Errorf("Expected Name 'TopRegions', got %v", body["Name"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	subset := models.NewDynamicSubset("Region", "Region", "TopRegions", "{[Region].[Region].Members}")

	err := service.Create(context.Background(), subset, false)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestSubsetServiceCreateStatic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["Name"] != "StaticSubset" {
			t.Errorf("Expected Name 'StaticSubset', got %v", body["Name"])
		}

		// Check elements binding exists for static subset
		if _, ok := body["Elements@odata.bind"]; !ok {
			t.Error("Expected Elements@odata.bind for static subset")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	subset := models.NewStaticSubset("Region", "Region", "StaticSubset", []string{"North", "South", "East"})

	err := service.Create(context.Background(), subset, false)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestSubsetServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		response := map[string]interface{}{
			"Name":       "TopRegions",
			"Expression": "{[Region].[Region].Members}",
			"Alias":      "",
			"Hierarchy": map[string]interface{}{
				"Name": "Region",
				"Dimension": map[string]interface{}{
					"Name": "Region",
				},
			},
			"Elements": []interface{}{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	subset, err := service.Get(context.Background(), "TopRegions", "Region", "Region", false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if subset.Name != "TopRegions" {
		t.Errorf("Expected name 'TopRegions', got %s", subset.Name)
	}
	if subset.Expression != "{[Region].[Region].Members}" {
		t.Errorf("Expected expression, got %s", subset.Expression)
	}
	if !subset.IsDynamic() {
		t.Error("Expected dynamic subset")
	}
}

func TestSubsetServiceGetStatic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"Name":       "StaticSubset",
			"Expression": "",
			"Alias":      "",
			"Hierarchy": map[string]interface{}{
				"Name": "Region",
				"Dimension": map[string]interface{}{
					"Name": "Region",
				},
			},
			"Elements": []interface{}{
				map[string]interface{}{"Name": "North"},
				map[string]interface{}{"Name": "South"},
				map[string]interface{}{"Name": "East"},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	subset, err := service.Get(context.Background(), "StaticSubset", "Region", "", false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !subset.IsStatic() {
		t.Error("Expected static subset")
	}

	if len(subset.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(subset.Elements))
	}

	expectedElements := []string{"North", "South", "East"}
	for i, elem := range subset.Elements {
		if elem != expectedElements[i] {
			t.Errorf("Expected element %s at index %d, got %s", expectedElements[i], i, elem)
		}
	}
}

func TestSubsetServiceGetAllNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		response := map[string]interface{}{
			"value": []map[string]interface{}{
				{"Name": "Subset1"},
				{"Name": "Subset2"},
				{"Name": "Subset3"},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	names, err := service.GetAllNames(context.Background(), "Region", "Region", false)
	if err != nil {
		t.Fatalf("GetAllNames failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	expected := []string{"Subset1", "Subset2", "Subset3"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected name %s at index %d, got %s", expected[i], i, name)
		}
	}
}

func TestSubsetServiceUpdate(t *testing.T) {
	patchCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "Elements/$ref") {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == "PATCH" {
			patchCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"Name": "StaticSubset"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	subset := models.NewStaticSubset("Region", "Region", "StaticSubset", []string{"North", "South"})

	err := service.Update(context.Background(), subset, false)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !patchCalled {
		t.Error("Expected PATCH request")
	}
}

func TestSubsetServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/Dimensions('Region')/Hierarchies('Region')/Subsets('TopRegions')") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	err := service.Delete(context.Background(), "TopRegions", "Region", "Region", false)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSubsetServiceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "('Existing')") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"Name": "Existing"})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	exists, err := service.Exists(context.Background(), "Existing", "Region", "Region", false)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected subset to exist")
	}

	exists, err = service.Exists(context.Background(), "NonExisting", "Region", "Region", false)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected subset to not exist")
	}
}

func TestSubsetServiceMakeStatic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "tm1.SaveAs") {
			t.Errorf("Expected tm1.SaveAs in path: %s", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["MakeStatic"] != true {
			t.Error("Expected MakeStatic to be true")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"Name": "DynamicSubset"})
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	err := service.MakeStatic(context.Background(), "DynamicSubset", "Region", "Region", false)
	if err != nil {
		t.Fatalf("MakeStatic failed: %v", err)
	}
}

func TestSubsetServiceUpdateOrCreate(t *testing.T) {
	existingSubset := "Existing"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check exists
		if r.Method == "GET" && strings.Contains(r.URL.Path, existingSubset) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"Name": existingSubset})
			return
		}
		if r.Method == "GET" && strings.Contains(r.URL.Path, "New") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Update (PATCH)
		if r.Method == "PATCH" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"Name": existingSubset})
			return
		}

		// Create (POST)
		if r.Method == "POST" && !strings.Contains(r.URL.Path, "SaveAs") {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"Name": "New"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	// Test update existing
	subset := models.NewDynamicSubset("Region", "Region", existingSubset, "{[Region].Members}")
	err := service.UpdateOrCreate(context.Background(), subset, false)
	if err != nil {
		t.Fatalf("UpdateOrCreate (update) failed: %v", err)
	}

	// Test create new
	newSubset := models.NewDynamicSubset("Region", "Region", "New", "{[Region].Members}")
	err = service.UpdateOrCreate(context.Background(), newSubset, false)
	if err != nil {
		t.Fatalf("UpdateOrCreate (create) failed: %v", err)
	}
}

func TestSubsetServicePrivateSubsets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for PrivateSubsets in URL
		if !strings.Contains(r.URL.Path, "PrivateSubsets") {
			t.Errorf("Expected PrivateSubsets in URL path: %s", r.URL.Path)
		}

		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"Name": "PrivateSubset"})
		} else if r.Method == "GET" {
			response := map[string]interface{}{
				"value": []map[string]interface{}{
					{"Name": "Private1"},
					{"Name": "Private2"},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	// Test create private subset
	subset := models.NewDynamicSubset("Region", "Region", "PrivateSubset", "{[Region].Members}")
	err := service.Create(context.Background(), subset, true)
	if err != nil {
		t.Fatalf("Create private subset failed: %v", err)
	}

	// Test get all private subset names
	names, err := service.GetAllNames(context.Background(), "Region", "Region", true)
	if err != nil {
		t.Fatalf("GetAllNames private subsets failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("Expected 2 private subsets, got %d", len(names))
	}
}

func TestSubsetServiceGetElementNamesStatic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"Name":       "StaticSubset",
			"Expression": "",
			"Hierarchy": map[string]interface{}{
				"Name": "Region",
				"Dimension": map[string]interface{}{
					"Name": "Region",
				},
			},
			"Elements": []interface{}{
				map[string]interface{}{"Name": "North"},
				map[string]interface{}{"Name": "South"},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	elements, err := service.GetElementNames(context.Background(), "Region", "Region", "StaticSubset", false)
	if err != nil {
		t.Fatalf("GetElementNames failed: %v", err)
	}

	if len(elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(elements))
	}
}

func TestSubsetServiceGetElementNamesDynamic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"Name":       "DynamicSubset",
			"Expression": "{[Region].Members}",
			"Hierarchy": map[string]interface{}{
				"Name": "Region",
				"Dimension": map[string]interface{}{
					"Name": "Region",
				},
			},
			"Elements": []interface{}{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	// Should return an error for dynamic subsets
	_, err := service.GetElementNames(context.Background(), "Region", "Region", "DynamicSubset", false)
	if err == nil {
		t.Fatal("Expected error for dynamic subset element retrieval")
	}

	if !strings.Contains(err.Error(), "MDX execution") {
		t.Errorf("Expected MDX execution error, got: %v", err)
	}
}

func TestSubsetServiceDeleteElementsFromStaticSubset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "Elements/$ref") {
			t.Errorf("Expected Elements/$ref in path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := Config{BaseURL: server.URL}
	rest, _ := NewRestService(cfg)
	service := NewSubsetService(rest)

	err := service.DeleteElementsFromStaticSubset(context.Background(), "Region", "Region", "StaticSubset", false)
	if err != nil {
		t.Fatalf("DeleteElementsFromStaticSubset failed: %v", err)
	}
}

func TestSubsetModelStaticBody(t *testing.T) {
	subset := models.NewStaticSubset("Region", "Region", "TestSubset", []string{"North", "South", "East"})
	subset.Alias = "TestAlias"

	body, err := subset.Body()
	if err != nil {
		t.Fatalf("Body() failed: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(body), &bodyMap); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if bodyMap["Name"] != "TestSubset" {
		t.Errorf("Expected Name 'TestSubset', got %v", bodyMap["Name"])
	}

	if bodyMap["Alias"] != "TestAlias" {
		t.Errorf("Expected Alias 'TestAlias', got %v", bodyMap["Alias"])
	}

	if _, ok := bodyMap["Elements@odata.bind"]; !ok {
		t.Error("Expected Elements@odata.bind in body")
	}

	if _, ok := bodyMap["Expression"]; ok {
		t.Error("Did not expect Expression in static subset body")
	}
}

func TestSubsetModelDynamicBody(t *testing.T) {
	subset := models.NewDynamicSubset("Region", "Region", "DynSubset", "{[Region].Members}")

	body, err := subset.Body()
	if err != nil {
		t.Fatalf("Body() failed: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(body), &bodyMap); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if bodyMap["Expression"] != "{[Region].Members}" {
		t.Errorf("Expected Expression, got %v", bodyMap["Expression"])
	}

	if _, ok := bodyMap["Elements@odata.bind"]; ok {
		t.Error("Did not expect Elements@odata.bind in dynamic subset body")
	}
}

func TestSubsetModelAddElements(t *testing.T) {
	subset := models.NewSubset("Region", "Region", "TestSubset")

	subset.AddElements("North", "South")

	if len(subset.Elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(subset.Elements))
	}

	if !subset.IsStatic() {
		t.Error("Expected subset to be static after adding elements")
	}

	subset.AddElements("East", "West")

	if len(subset.Elements) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(subset.Elements))
	}
}

func TestSubsetModelSetExpression(t *testing.T) {
	subset := models.NewStaticSubset("Region", "Region", "TestSubset", []string{"North", "South"})

	subset.SetExpression("{[Region].Members}")

	if !subset.IsDynamic() {
		t.Error("Expected subset to be dynamic after setting expression")
	}

	if subset.Expression != "{[Region].Members}" {
		t.Errorf("Expected expression, got %s", subset.Expression)
	}

	if len(subset.Elements) != 0 {
		t.Error("Expected elements to be cleared after setting expression")
	}
}
