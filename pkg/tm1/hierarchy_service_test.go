package tm1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupHierarchyTestService(handler http.HandlerFunc) (*HierarchyService, *httptest.Server) {
	server := httptest.NewServer(handler)
	config := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(config)
	rest.SetBaseURL(server.URL)
	service := NewHierarchyService(rest)
	return service, server
}

func TestHierarchyService_GetElementAttributeNames(t *testing.T) {
	numericType := 0
	stringType := 1
	aliasType := 2

	tests := []struct {
		name          string
		attributeType *int
		expectedQuery string
	}{
		{
			name:          "all attributes",
			attributeType: nil,
			expectedQuery: "$expand=ElementAttributes($select=Name)",
		},
		{
			name:          "numeric attributes only",
			attributeType: &numericType,
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%200)",
		},
		{
			name:          "string attributes only",
			attributeType: &stringType,
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%201)",
		},
		{
			name:          "alias attributes only",
			attributeType: &aliasType,
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%202)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, server := setupHierarchyTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/Dimensions('Department')/Hierarchies('Department')" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.URL.RawQuery != tt.expectedQuery {
					t.Errorf("unexpected query: got %s want %s", r.URL.RawQuery, tt.expectedQuery)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"Name":"Department","ElementAttributes":[{"Name":"Code"},{"Name":"Description"}]}`))
			}))
			defer server.Close()

			names, err := service.GetElementAttributeNames(context.Background(), "Department", "Department", tt.attributeType)
			if err != nil {
				t.Fatalf("GetElementAttributeNames failed: %v", err)
			}

			if len(names) != 2 {
				t.Fatalf("expected 2 names, got %d", len(names))
			}
			if names[0] != "Code" || names[1] != "Description" {
				t.Fatalf("unexpected names: %v", names)
			}
		})
	}
}

func TestHierarchyService_GetElementAttributeNames_DefaultHierarchy(t *testing.T) {
	service, server := setupHierarchyTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/Dimensions('Department')/Hierarchies('Department')") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ElementAttributes":[{"Name":"Alias"}]}`))
	}))
	defer server.Close()

	names, err := service.GetElementAttributeNames(context.Background(), "Department", "", nil)
	if err != nil {
		t.Fatalf("GetElementAttributeNames failed: %v", err)
	}
	if len(names) != 1 || names[0] != "Alias" {
		t.Fatalf("unexpected names: %v", names)
	}
}

func TestHierarchyService_GetAttributeNameWrappers(t *testing.T) {
	tests := []struct {
		name          string
		expectedQuery string
		call          func(service *HierarchyService) ([]string, error)
	}{
		{
			name:          "numeric wrapper",
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%200)",
			call: func(service *HierarchyService) ([]string, error) {
				return service.GetNumericAttributeNames(context.Background(), "Department", "Department")
			},
		},
		{
			name:          "string wrapper",
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%201)",
			call: func(service *HierarchyService) ([]string, error) {
				return service.GetStringAttributeNames(context.Background(), "Department", "Department")
			},
		},
		{
			name:          "alias wrapper",
			expectedQuery: "$expand=ElementAttributes($select=Name;$filter=Type%20eq%202)",
			call: func(service *HierarchyService) ([]string, error) {
				return service.GetAliasAttributeNames(context.Background(), "Department", "Department")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, server := setupHierarchyTestService(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != tt.expectedQuery {
					t.Errorf("unexpected query: got %s want %s", r.URL.RawQuery, tt.expectedQuery)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ElementAttributes":[{"Name":"A1"}]}`))
			}))
			defer server.Close()

			names, err := tt.call(service)
			if err != nil {
				t.Fatalf("wrapper call failed: %v", err)
			}
			if len(names) != 1 || names[0] != "A1" {
				t.Fatalf("unexpected names: %v", names)
			}
		})
	}
}
