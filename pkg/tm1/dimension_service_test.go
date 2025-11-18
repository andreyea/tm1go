package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/andreyea/tm1go/pkg/models"
)

func TestDimensionServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Dimensions('TestDim')":
			if r.Method == "GET" {
				w.WriteHeader(http.StatusNotFound)
			}
		case "/Dimensions":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.SetBaseURL(server.URL)

	ds := NewDimensionService(rest)
	ctx := context.Background()

	dim := models.NewDimension("TestDim")
	err := ds.Create(ctx, dim)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
}

func TestDimensionServiceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Dimensions('ExistingDim')" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Name":"ExistingDim"}`))
		} else if r.URL.Path == "/Dimensions('NonExistingDim')" {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, err := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ds := NewDimensionService(rest)
	ctx := context.Background()

	exists, err := ds.Exists(ctx, "ExistingDim")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}

	exists, err = ds.Exists(ctx, "NonExistingDim")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}
}

func TestDimensionServiceGetAllNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Dimensions" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":[{"Name":"Dim1"},{"Name":"Dim2"},{"Name":"Dim3"}]}`))
		} else if r.URL.Path == "/ModelDimensions()" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":[{"Name":"Dim1"},{"Name":"Dim2"}]}`))
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ds := NewDimensionService(rest)
	ctx := context.Background()

	// Test with control dims
	names, err := ds.GetAllNames(ctx, false)
	if err != nil {
		t.Errorf("GetAllNames() error = %v", err)
	}
	if len(names) != 3 {
		t.Errorf("GetAllNames() count = %d, want 3", len(names))
	}

	// Test without control dims
	names, err = ds.GetAllNames(ctx, true)
	if err != nil {
		t.Errorf("GetAllNames(skipControl) error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("GetAllNames(skipControl) count = %d, want 2", len(names))
	}
}

func TestDimensionServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Dimensions('TestDim')" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ds := NewDimensionService(rest)
	ctx := context.Background()

	err := ds.Delete(ctx, "TestDim")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestDimensionModel(t *testing.T) {
	dim := models.NewDimension("TestDim")

	if dim.Name != "TestDim" {
		t.Errorf("Name = %v, want TestDim", dim.Name)
	}

	hierarchy := models.NewHierarchy("TestHier", "TestDim")
	dim.AddHierarchy(*hierarchy)

	if len(dim.Hierarchies) != 1 {
		t.Errorf("Hierarchies count = %d, want 1", len(dim.Hierarchies))
	}

	if !dim.HasHierarchy("TestHier") {
		t.Error("HasHierarchy() = false, want true")
	}

	names := dim.HierarchyNames()
	if len(names) != 1 || names[0] != "TestHier" {
		t.Errorf("HierarchyNames() = %v, want [TestHier]", names)
	}
}
