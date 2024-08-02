package tm1go_test

import (
	"testing"

	"github.com/andreyea/tm1go"
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

func TestProcessService_GetAll(t *testing.T) {
	type args struct {
		skipControlProcesses bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get all processes",
			args:    args{skipControlProcesses: false},
			wantErr: false,
		},
		{
			name:    "Get all processes except control processes",
			args:    args{skipControlProcesses: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process, err := tm1ServiceT.ProcessService.GetAll(tt.args.skipControlProcesses)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && process == nil {
				t.Errorf("TestProcessService.GetAll() error = %v, wantErr %v", "No processes were returned", tt.wantErr)
			}
		})
	}
}

func TestProcessService_GetAllNames(t *testing.T) {
	type args struct {
		skipControlProcesses bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get all processes names",
			args:    args{skipControlProcesses: false},
			wantErr: false,
		},
		{
			name:    "Get all processes names except control processes",
			args:    args{skipControlProcesses: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process, err := tm1ServiceT.ProcessService.GetAllNames(tt.args.skipControlProcesses)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetAllNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && process == nil {
				t.Errorf("TestProcessService.GetAllNames() error = %v, wantErr %v", "No processes were returned", tt.wantErr)
			}
		})
	}
}

func TestProcessService_SearchStringInCode(t *testing.T) {
	type args struct {
		skipControlProcesses bool
		searchString         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Search string in all processes",
			args:    args{skipControlProcesses: false, searchString: "sleep"},
			wantErr: false,
		},
		{
			name:    "Search string in all processes except control processes",
			args:    args{skipControlProcesses: true, searchString: "sleep"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.ProcessService.SearchStringInCode(tt.args.searchString, tt.args.skipControlProcesses)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.SearchStringInCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessService_Create(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Create a dummy process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ProcessService.Create(tt.args.process)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_Update(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	process.PrologProcedure = "s = 2;"
	process.EpilogProcedure = "s = 3;"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Update a dummy process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ProcessService.Update(tt.args.process)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_Delete(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Delete a dummy process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ProcessService.Delete(tt.args.process.Name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestProcessService_Exists(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Check if a dummy process exists",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := tm1ServiceT.ProcessService.Exists(tt.args.process.Name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Exists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !exists {
				t.Errorf("TestProcessService.Exists() error = %v. Process created but not found by the test, wantErr %v", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_Compile(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Compile existing process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.ProcessService.Compile(tt.args.process.Name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Compile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})

}

func TestProcessService_CompileProcess(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Compile unbound process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.ProcessService.CompileProcess(tt.args.process)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.CompileProcess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessService_Execute(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Execute a dummy process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm1ServiceT.ProcessService.Execute(tt.args.process.Name, nil, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})

}
func TestProcessService_ExecuteWithReturn(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}
	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Execute unbound process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tm1ServiceT.ProcessService.ExecuteWithReturn(tt.args.process.Name, nil, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.ExecuteWithReturn() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result.ProcessExecuteStatusCode != tm1go.CompletedSuccessfully {
				t.Errorf("TestProcessService.ExecuteWithReturn() error = %v, wantErr %v. Process did not execute successfully", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})

}

func TestProcessService_ExecuteTICode(t *testing.T) {
	type args struct {
		code string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Execute TI code",
			args:    args{code: "s=123;"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tm1ServiceT.ProcessService.ExecuteTICode(tt.args.code, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.ExecuteTICode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result.ProcessExecuteStatusCode != tm1go.CompletedSuccessfully {
				t.Errorf("TestProcessService.ExecuteTICode() error = %v, wantErr %v. Process did not execute successfully", err, tt.wantErr)
			}
		})
	}
}

func TestProcessService_SearchErrorLogFilenames(t *testing.T) {
	type args struct {
		searchString string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Search string in errors file names",
			args:    args{searchString: "sleep"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.ProcessService.SearchErrorLogFilenames(tt.args.searchString, 5, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.SearchErrorLogFilenames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessService_GetErrorLogFileNames(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}

	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1;"

	// Create process
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Update() error = %v", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get error log file names for a process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm1ServiceT.ProcessService.GetErrorLogFileNames(tt.args.process.Name, 5, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetErrorLogFileNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_GetErrorLogFileContent(t *testing.T) {
	type args struct {
		logName string
	}

	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1"

	// Create process with errors
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Create() error = %v", err)
	}

	// Execute process to generate errors
	_ = tm1ServiceT.ProcessService.Execute(process.Name, nil, 0)

	logs, err := tm1ServiceT.ProcessService.GetErrorLogFileNames(process.Name, 1, false)
	if err != nil {
		t.Errorf("TestProcessService.GetErrorLogFileNames() error = %v", err)
	}
	if len(logs) == 0 {
		t.Errorf("TestProcessService.GetErrorLogFileNames() error = %v. No error log files found", err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get conent of the first error log file",
			args:    args{logName: logs[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tm1ServiceT.ProcessService.GetErrorLogFileContent(tt.args.logName)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetErrorLogFileContent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if content == "" {
				t.Errorf("TestProcessService.GetErrorLogFileContent() error = %v, wantErr %v. The test should get some content from process logs. Received empty string.", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_GetProcessErrorLogs(t *testing.T) {
	type args struct {
		process *tm1go.Process
	}

	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1"

	// Create process with errors
	err := tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Create() error = %v", err)
	}

	// Execute process to generate errors
	_ = tm1ServiceT.ProcessService.Execute(process.Name, nil, 0)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get error logs for a process",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := tm1ServiceT.ProcessService.GetProcessErrorLogs(tt.args.process.Name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetProcessErrorLogs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(logs) == 0 {
				t.Errorf("TestProcessService.GetProcessErrorLogs() error = %v, wantErr %v. The test should get logs for the process.", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}

func TestProcessService_GetLastMessageFromProcessErrorLog(t *testing.T) {
	tm1Version, err := tm1ServiceT.ConfigurationService.GetProductVersion()
	if err != nil {
		t.Errorf("TestProcessService.GetLastMessageFromProcessErrorLog() error (GetProductVersion) = %v", err)
	}
	if tm1go.IsV1GreaterOrEqualToV2(tm1Version, "12.0.0") {
		t.Skip("TestProcessService.GetLastMessageFromProcessErrorLog() skipped for TM1 version 12.0.0 and above")
	}

	type args struct {
		process *tm1go.Process
	}

	process := tm1go.NewProcess("!!!TestProcess!!!")
	process.PrologProcedure = "s = 1"

	// Create process with errors
	err = tm1ServiceT.ProcessService.Create(process)
	if err != nil {
		t.Errorf("TestProcessService.Create() error = %v", err)
	}

	// Execute process to generate errors
	_ = tm1ServiceT.ProcessService.Execute(process.Name, nil, 0)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Get last message from process error log",
			args:    args{process: process},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := tm1ServiceT.ProcessService.GetLastMessageFromProcessErrorLog(tt.args.process.Name)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestProcessService.GetLastMessageFromProcessErrorLog() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(logs) == 0 {
				t.Errorf("TestProcessService.GetLastMessageFromProcessErrorLog() error = %v, wantErr %v. The test should get logs for the process.", err, tt.wantErr)
			}
		})
	}
	t.Cleanup(func() {
		// Delete process
		tm1ServiceT.ProcessService.Delete(process.Name)
	})
}
