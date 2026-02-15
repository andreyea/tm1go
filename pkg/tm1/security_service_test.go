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

func TestSecurityServiceDetermineActualNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/Users" && strings.Contains(r.URL.RawQuery, "$select=Name"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"JohnDoe"},{"Name":"Admin User"}]}`))
		case r.Method == "GET" && r.URL.Path == "/Groups" && strings.Contains(r.URL.RawQuery, "$select=Name"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"Finance Team"},{"Name":"Admin"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSecurityService(rest)

	actualUser, err := service.DetermineActualUserName(context.Background(), "john doe")
	if err != nil {
		t.Fatalf("DetermineActualUserName() error = %v", err)
	}
	if actualUser != "JohnDoe" {
		t.Fatalf("DetermineActualUserName() = %s, want JohnDoe", actualUser)
	}

	actualGroup, err := service.DetermineActualGroupName(context.Background(), "financeteam")
	if err != nil {
		t.Fatalf("DetermineActualGroupName() error = %v", err)
	}
	if actualGroup != "Finance Team" {
		t.Fatalf("DetermineActualGroupName() = %s, want 'Finance Team'", actualGroup)
	}
}

func TestSecurityServiceCreateGroupRequiresSecurityAdmin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ActiveUser" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"user","Type":"User","Groups":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSecurityService(rest)
	err := service.CreateGroup(context.Background(), "Finance")
	if err == nil {
		t.Fatal("CreateGroup() expected security admin error")
	}
}

func TestSecurityServiceCreateUser(t *testing.T) {
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"secadmin","Type":"SecurityAdmin","Groups":[{"Name":"SecurityAdmin"}]}`))
		case r.Method == "POST" && r.URL.Path == "/Users":
			created = true
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSecurityService(rest)
	enabled := true
	err := service.CreateUser(context.Background(), &models.User{
		Name:         "john",
		FriendlyName: "John",
		Password:     "pw",
		Enabled:      &enabled,
		Type:         models.UserTypeUser,
		Groups:       []models.NamedObject{{Name: "Finance"}},
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if !created {
		t.Fatal("CreateUser() expected POST /Users")
	}
}

func TestSecurityServiceGetCustomSecurityGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/Groups" && strings.Contains(r.URL.RawQuery, "$select=Name") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":[{"Name":"Admin"},{"Name":"DataAdmin"},{"Name":"SecurityAdmin"},{"Name":"OperationsAdmin"},{"Name":"}tp_Everyone"},{"Name":"Finance"},{"Name":"Sales Team"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSecurityService(rest)
	groups, err := service.GetCustomSecurityGroups(context.Background())
	if err != nil {
		t.Fatalf("GetCustomSecurityGroups() error = %v", err)
	}
	if len(groups) != 2 || groups[0] != "Finance" || groups[1] != "Sales Team" {
		t.Fatalf("GetCustomSecurityGroups() = %#v, want [Finance Sales Team]", groups)
	}
}

func TestSecurityServiceSecurityRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/ActiveUser":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Name":"admin","Type":"Admin","Groups":[{"Name":"Admin"}]}`))
		case r.Method == "POST" && r.URL.Path == "/ExecuteProcessWithReturn":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ProcessExecuteStatusCode":"CompletedSuccessfully"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{Address: "localhost", Port: 8882, SSL: false}
	rest, _ := NewRestService(cfg)
	mockBaseURL, _ := url.Parse(server.URL)
	rest.SetBaseURL(mockBaseURL.String())

	service := NewSecurityService(rest)
	if err := service.SecurityRefresh(context.Background()); err != nil {
		t.Fatalf("SecurityRefresh() error = %v", err)
	}
}
