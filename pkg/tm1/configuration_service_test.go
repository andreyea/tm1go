package tm1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestConfigurationServiceGetAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/Configuration" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"@odata.context":"x","ServerName":"S1"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewConfigurationService(rest)
	config, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if _, ok := config["@odata.context"]; ok {
		t.Fatal("GetAll() should remove @odata.context")
	}
	if config["ServerName"] != "S1" {
		t.Fatalf("GetAll() ServerName = %v, want S1", config["ServerName"])
	}
}

func TestConfigurationServiceGetStaticRequiresOpsAdmin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ActiveUser" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"u","Type":"User","Groups":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewConfigurationService(rest)
	_, err := svc.GetStatic(context.Background())
	if err == nil {
		t.Fatal("GetStatic() expected operations admin error")
	}
}

func TestConfigurationServiceUpdateStatic(t *testing.T) {
	var patchBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"ops","Type":"OperationsAdmin","Groups":[{"Name":"OperationsAdmin"}]}`))
		case r.Method == "PATCH" && r.URL.Path == "/StaticConfiguration":
			_ = json.NewDecoder(r.Body).Decode(&patchBody)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	rest.version = "11.8.0"
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	svc := NewConfigurationService(rest)
	payload := map[string]interface{}{"Administration": map[string]interface{}{"PerformanceMonitorOn": true}}
	if err := svc.UpdateStatic(context.Background(), payload); err != nil {
		t.Fatalf("UpdateStatic() error = %v", err)
	}
	admin, ok := patchBody["Administration"].(map[string]interface{})
	if !ok {
		t.Fatalf("UpdateStatic() payload missing Administration: %#v", patchBody)
	}
	if admin["PerformanceMonitorOn"] != true {
		t.Fatalf("UpdateStatic() payload wrong value: %#v", patchBody)
	}
}
