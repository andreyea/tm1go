package tm1go_test

import (
	"testing"
)

func TestProcessService_Get(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		wantErr     bool
	}{
		{
			name:        "Valid process name",
			processName: "}bedrock.server.wait",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process, err := tm1ServiceT.ProcessService.Get(tt.processName)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && process == nil {
				t.Errorf("TestProcessService.Get() error = %v, wantErr %v", "No process was returned", tt.wantErr)
			}
		})
	}

}
