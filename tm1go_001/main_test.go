package tm1go_test

import (
	"os"
	"testing"

	"github.com/andreyea/tm1go"
)

var tm1ServiceT *tm1go.TM1Service

func TestMain(m *testing.M) {
	// Setup
	setup()

	// Run tests
	code := m.Run()

	// Teardown
	teardown()

	// Exit with the exit code from running the tests
	os.Exit(code)
}

func setup() {
	// Initialize the client
	config := tm1go.TM1ServiceConfig{}
	config.Load("config_for_testing.json")
	tm1ServiceT = tm1go.NewTM1Service(config)
}

func teardown() {
	// Logout
	if err := tm1ServiceT.Logout(); err != nil {
		panic("Failed to logout: " + err.Error())
	}
}
