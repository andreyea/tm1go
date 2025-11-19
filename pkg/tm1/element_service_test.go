package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreyea/tm1go/pkg/models"
)

func setupTestService(handler http.HandlerFunc) (*ElementService, *httptest.Server) {
	server := httptest.NewServer(handler)
	config := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(config)
	rest.SetBaseURL(server.URL)
	service := NewElementService(rest)
	return service, server
}

func TestElementService_Get(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/Elements('Element1')") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Name":"Element1","Type":"Numeric","Index":0}`))
	}))
	defer server.Close()

	element, err := service.Get(context.Background(), "TestDim", "TestDim", "Element1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if element.Name != "Element1" {
		t.Errorf("expected Element1, got %s", element.Name)
	}
}

func TestElementService_Create(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	element := models.Element{
		Name: "NewElement",
		Type: "Numeric",
	}

	err := service.Create(context.Background(), "TestDim", "TestDim", element)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestElementService_Update(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	element := models.Element{
		Name: "Element1",
		Type: "Numeric",
	}

	err := service.Update(context.Background(), "TestDim", "TestDim", element)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestElementService_Delete(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := service.Delete(context.Background(), "TestDim", "TestDim", "Element1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestElementService_Exists(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check exact path for "Existing" element
		if r.URL.Path == "/Dimensions('TestDim')/Hierarchies('TestDim')/Elements('Existing')" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Name":"Existing"}`))
		} else if r.URL.Path == "/Dimensions('TestDim')/Hierarchies('TestDim')/Elements('NonExisting')" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code":"404","message":"Not Found"}}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	exists, err := service.Exists(context.Background(), "TestDim", "TestDim", "Existing")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected element to exist")
	}

	exists, err = service.Exists(context.Background(), "TestDim", "TestDim", "NonExisting")
	if err != nil {
		t.Fatalf("Exists check for non-existing failed: %v", err)
	}
	if exists {
		t.Error("expected element to not exist")
	}
}

func TestElementService_GetElementNames(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Element1"},{"Name":"Element2"},{"Name":"Element3"}]}`))
	}))
	defer server.Close()

	names, err := service.GetElementNames(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetElementNames failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("expected 3 elements, got %d", len(names))
	}
}

func TestElementService_GetLeafElementNames(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "Type") {
			t.Errorf("expected filter in query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Leaf1"},{"Name":"Leaf2"}]}`))
	}))
	defer server.Close()

	names, err := service.GetLeafElementNames(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetLeafElementNames failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 leaf elements, got %d", len(names))
	}
}

func TestElementService_GetConsolidatedElementNames(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "Type") {
			t.Errorf("expected filter in query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Total"}]}`))
	}))
	defer server.Close()

	names, err := service.GetConsolidatedElementNames(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetConsolidatedElementNames failed: %v", err)
	}

	if len(names) != 1 {
		t.Errorf("expected 1 consolidated element, got %d", len(names))
	}
}

func TestElementService_GetNumberOfElements(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "$count") {
			t.Errorf("expected $count in path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`42`))
	}))
	defer server.Close()

	count, err := service.GetNumberOfElements(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetNumberOfElements failed: %v", err)
	}

	if count != 42 {
		t.Errorf("expected 42 elements, got %d", count)
	}
}

func TestElementService_GetEdges(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[
			{"ParentName":"Total","ComponentName":"Q1","Weight":1.0},
			{"ParentName":"Total","ComponentName":"Q2","Weight":1.0}
		]}`))
	}))
	defer server.Close()

	edges, err := service.GetEdges(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}

	if len(edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(edges))
	}
}

func TestElementService_GetElementAttributes(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[
			{"Name":"Caption","AttributeType":"Alias"},
			{"Name":"Weight","AttributeType":"Numeric"}
		]}`))
	}))
	defer server.Close()

	attrs, err := service.GetElementAttributes(context.Background(), "TestDim", "TestDim")
	if err != nil {
		t.Fatalf("GetElementAttributes failed: %v", err)
	}

	if len(attrs) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(attrs))
	}
}

func TestElementService_CreateElementAttribute(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	attr := models.ElementAttribute{
		Name:          "NewAttr",
		AttributeType: "String",
	}

	err := service.CreateElementAttribute(context.Background(), "TestDim", "TestDim", attr)
	if err != nil {
		t.Fatalf("CreateElementAttribute failed: %v", err)
	}
}

func TestElementService_AddEdges(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	edges := map[[2]string]float64{
		{"Total", "Q1"}: 1.0,
		{"Total", "Q2"}: 1.0,
	}

	err := service.AddEdges(context.Background(), "TestDim", "TestDim", edges)
	if err != nil {
		t.Fatalf("AddEdges failed: %v", err)
	}
}

func TestElementService_RemoveEdge(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := service.RemoveEdge(context.Background(), "TestDim", "TestDim", "Total", "Q1")
	if err != nil {
		t.Fatalf("RemoveEdge failed: %v", err)
	}
}

func TestElementService_GetElementsByLevel(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "Level") {
			t.Errorf("expected Level filter in query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Leaf1"},{"Name":"Leaf2"}]}`))
	}))
	defer server.Close()

	names, err := service.GetElementsByLevel(context.Background(), "TestDim", "TestDim", 0)
	if err != nil {
		t.Fatalf("GetElementsByLevel failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 elements at level 0, got %d", len(names))
	}
}

func TestElementService_GetElementsFilteredByWildcard(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Product1"},{"Name":"Product2"}]}`))
	}))
	defer server.Close()

	names, err := service.GetElementsFilteredByWildcard(context.Background(), "TestDim", "TestDim", "Product", nil)
	if err != nil {
		t.Fatalf("GetElementsFilteredByWildcard failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 elements, got %d", len(names))
	}
}

func TestElementService_GetParents(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Total"},{"Name":"SubTotal"}]}`))
	}))
	defer server.Close()

	parents, err := service.GetParents(context.Background(), "TestDim", "TestDim", "Element1")
	if err != nil {
		t.Fatalf("GetParents failed: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("expected 2 parents, got %d", len(parents))
	}
}

func TestElementService_ElementIsParent(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Total"}]}`))
	}))
	defer server.Close()

	isParent, err := service.ElementIsParent(context.Background(), "TestDim", "TestDim", "Total", "Q1")
	if err != nil {
		t.Fatalf("ElementIsParent failed: %v", err)
	}

	if !isParent {
		t.Error("expected Total to be parent of Q1")
	}
}

func TestElementService_GetLevelNames(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Level 0"},{"Name":"Level 1"},{"Name":"Level 2"}]}`))
	}))
	defer server.Close()

	names, err := service.GetLevelNames(context.Background(), "TestDim", "TestDim", true)
	if err != nil {
		t.Fatalf("GetLevelNames failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("expected 3 levels, got %d", len(names))
	}

	// When descending is true, order should be reversed
	if names[0] != "Level 2" {
		t.Errorf("expected Level 2 first in descending order, got %s", names[0])
	}
}

func TestElementService_GetElementTypes(t *testing.T) {
	service, server := setupTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[
			{"Name":"Element1","Type":"Numeric"},
			{"Name":"Element2","Type":"String"},
			{"Name":"Total","Type":"Consolidated"}
		]}`))
	}))
	defer server.Close()

	types, err := service.GetElementTypes(context.Background(), "TestDim", "TestDim", false)
	if err != nil {
		t.Fatalf("GetElementTypes failed: %v", err)
	}

	if len(types) != 3 {
		t.Errorf("expected 3 element types, got %d", len(types))
	}

	if types["Element1"] != "Numeric" {
		t.Errorf("expected Element1 to be Numeric, got %s", types["Element1"])
	}
}
