package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/andreyea/tm1go/pkg/models"
)

func TestProcessServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Processes" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	process := models.NewProcess("TestProcess")
	process.PrologProcedure = "sMessage = 'Hello World';"

	err := ps.Create(ctx, process)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
}

func TestProcessServiceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Processes('ExistingProcess')" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Name":"ExistingProcess"}`))
		} else if r.URL.Path == "/Processes('NonExistingProcess')" {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	exists, err := ps.Exists(ctx, "ExistingProcess")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}

	exists, err = ps.Exists(ctx, "NonExistingProcess")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}
}

func TestProcessServiceGetAllNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return fewer processes when filter is present
		if strings.Contains(r.URL.RawQuery, "filter") {
			w.Write([]byte(`{"value":[{"Name":"Process1"},{"Name":"Process3"}]}`))
		} else {
			w.Write([]byte(`{"value":[{"Name":"Process1"},{"Name":"}Process2"},{"Name":"Process3"}]}`))
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	// Test with control processes
	names, err := ps.GetAllNames(ctx, false)
	if err != nil {
		t.Errorf("GetAllNames() error = %v", err)
	}
	if len(names) != 3 {
		t.Errorf("GetAllNames() count = %d, want 3", len(names))
	}

	// Test without control processes
	names, err = ps.GetAllNames(ctx, true)
	if err != nil {
		t.Errorf("GetAllNames(skipControl) error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("GetAllNames(skipControl) count = %d, want 2", len(names))
	}
}

func TestProcessServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Processes('TestProcess')" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	err := ps.Delete(ctx, "TestProcess")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestProcessModel(t *testing.T) {
	process := models.NewProcess("TestProcess")

	if process.Name != "TestProcess" {
		t.Errorf("Name = %v, want TestProcess", process.Name)
	}

	// Test adding parameter
	process.AddParameter("pParam1", "Prompt1", "Value1", "String")
	if len(process.Parameters) != 1 {
		t.Errorf("Parameters count = %d, want 1", len(process.Parameters))
	}
	if process.Parameters[0].Name != "pParam1" {
		t.Errorf("Parameter name = %v, want pParam1", process.Parameters[0].Name)
	}

	// Test removing parameter
	process.RemoveParameter("pParam1")
	if len(process.Parameters) != 0 {
		t.Errorf("Parameters count after removal = %d, want 0", len(process.Parameters))
	}

	// Test adding variable
	process.AddVariable("vVar1", "String", 1)
	if len(process.Variables) != 1 {
		t.Errorf("Variables count = %d, want 1", len(process.Variables))
	}
	if process.Variables[0].Name != "vVar1" {
		t.Errorf("Variable name = %v, want vVar1", process.Variables[0].Name)
	}
}

func TestProcessServiceSearchStringInCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Process1"},{"Name":"Process2"}]}`))
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	names, err := ps.SearchStringInCode(ctx, "CellGet", false)
	if err != nil {
		t.Errorf("SearchStringInCode() error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("SearchStringInCode() count = %d, want 2", len(names))
	}
}

func TestProcessServiceSearchStringInName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":[{"Name":"Bedrock.Server.Wait"},{"Name":"Bedrock.Cube.Data.Copy"}]}`))
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	ps := NewProcessService(rest)
	ctx := context.Background()

	names, err := ps.SearchStringInName(ctx, "Bedrock", []string{"Server", "Cube"}, "or", false)
	if err != nil {
		t.Errorf("SearchStringInName() error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("SearchStringInName() count = %d, want 2", len(names))
	}
}
