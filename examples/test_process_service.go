package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreyea/tm1go/pkg/models"
	"github.com/andreyea/tm1go/pkg/tm1"
)

func main() {
	// Create TM1 config
	cfg := tm1.Config{
		Address:  "localhost",
		Port:     8882,
		User:     "admin",
		Password: "",
		SSL:      true,
		Logging:  true,
	}

	// Create TM1Service
	svc, err := tm1.NewTM1Service(cfg)
	if err != nil {
		log.Fatalf("Failed to create TM1 service: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Example 1: Get all process names
	fmt.Println("\n=== Example 1: Get All Process Names ===")
	names, err := svc.Processes.GetAllNames(ctx, false)
	if err != nil {
		log.Printf("Error getting process names: %v", err)
	} else {
		fmt.Printf("Found %d processes\n", len(names))
		if len(names) > 0 {
			fmt.Printf("First few: %v\n", names[:min(5, len(names))])
		}
	}

	// Example 2: Get all process names (skip control processes)
	fmt.Println("\n=== Example 2: Get Process Names (Skip Control) ===")
	names, err = svc.Processes.GetAllNames(ctx, true)
	if err != nil {
		log.Printf("Error getting process names: %v", err)
	} else {
		fmt.Printf("Found %d non-control processes\n", len(names))
	}

	// Example 3: Check if a process exists
	fmt.Println("\n=== Example 3: Check Process Existence ===")
	processName := "Bedrock.Server.Wait"
	exists, err := svc.Processes.Exists(ctx, processName)
	if err != nil {
		log.Printf("Error checking process: %v", err)
	} else {
		fmt.Printf("Process '%s' exists: %v\n", processName, exists)
	}

	// Example 4: Create a simple process
	fmt.Println("\n=== Example 4: Create Process ===")
	process := models.NewProcess("tm1go.TestProcess")
	process.PrologProcedure = `
		sMessage = 'Hello from tm1go!';
		ASCIIOutput('tm1go.log', sMessage);
	`
	process.AddParameter("pParam1", "Parameter 1", "DefaultValue", "String")

	err = svc.Processes.Create(ctx, process)
	if err != nil {
		log.Printf("Error creating process: %v", err)
	} else {
		fmt.Println("Process created successfully")
	}

	// Example 5: Get a process
	fmt.Println("\n=== Example 5: Get Process ===")
	retrievedProcess, err := svc.Processes.Get(ctx, "tm1go.TestProcess")
	if err != nil {
		log.Printf("Error getting process: %v", err)
	} else {
		fmt.Printf("Retrieved process: %s\n", retrievedProcess.Name)
		fmt.Printf("Prolog procedure length: %d characters\n", len(retrievedProcess.PrologProcedure))
		fmt.Printf("Parameters count: %d\n", len(retrievedProcess.Parameters))
	}

	// Example 6: Update a process
	fmt.Println("\n=== Example 6: Update Process ===")
	if retrievedProcess != nil {
		retrievedProcess.EpilogProcedure = `
			sMessage = 'Goodbye from tm1go!';
			ASCIIOutput('tm1go.log', sMessage);
		`
		err = svc.Processes.Update(ctx, retrievedProcess)
		if err != nil {
			log.Printf("Error updating process: %v", err)
		} else {
			fmt.Println("Process updated successfully")
		}
	}

	// Example 7: Execute a process
	fmt.Println("\n=== Example 7: Execute Process ===")
	parameters := map[string]interface{}{
		"pParam1": "TestValue",
	}
	err = svc.Processes.Execute(ctx, "tm1go.TestProcess", parameters, nil, false)
	if err != nil {
		log.Printf("Error executing process: %v", err)
	} else {
		fmt.Println("Process executed successfully")
	}

	// Example 8: Execute with return status
	fmt.Println("\n=== Example 8: Execute With Return ===")
	success, status, errorLog, err := svc.Processes.ExecuteWithReturn(ctx, "tm1go.TestProcess", parameters, nil, false)
	if err != nil {
		log.Printf("Error executing process: %v", err)
	} else {
		fmt.Printf("Success: %v, Status: %s\n", success, status)
		if errorLog != "" {
			fmt.Printf("Error log: %s\n", errorLog)
		}
	}

	// Example 9: Search processes by name
	fmt.Println("\n=== Example 9: Search Process Names ===")
	matches, err := svc.Processes.SearchStringInName(ctx, "Bedrock", []string{"Server"}, "and", true)
	if err != nil {
		log.Printf("Error searching processes: %v", err)
	} else {
		fmt.Printf("Found %d processes matching criteria\n", len(matches))
		for _, name := range matches[:min(5, len(matches))] {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Example 10: Search processes by code content
	fmt.Println("\n=== Example 10: Search Process Code ===")
	matches, err = svc.Processes.SearchStringInCode(ctx, "CellGet", true)
	if err != nil {
		log.Printf("Error searching process code: %v", err)
	} else {
		fmt.Printf("Found %d processes containing 'CellGet'\n", len(matches))
	}

	// Example 11: Compile a process
	fmt.Println("\n=== Example 11: Compile Process ===")
	syntaxErrors, err := svc.Processes.Compile(ctx, "tm1go.TestProcess")
	if err != nil {
		log.Printf("Error compiling process: %v", err)
	} else {
		if len(syntaxErrors) > 0 {
			fmt.Printf("Found %d syntax errors\n", len(syntaxErrors))
		} else {
			fmt.Println("No syntax errors found")
		}
	}

	// Example 12: Get error log filenames
	fmt.Println("\n=== Example 12: Get Error Log Filenames ===")
	errorLogs, err := svc.Processes.GetErrorLogFilenames(ctx, "", 10, true)
	if err != nil {
		log.Printf("Error getting error logs: %v", err)
	} else {
		fmt.Printf("Found %d recent error logs\n", len(errorLogs))
		for _, logFile := range errorLogs[:min(3, len(errorLogs))] {
			fmt.Printf("  - %s\n", logFile)
		}
	}

	// Example 13: Create process with data source
	fmt.Println("\n=== Example 13: Create Process with Data Source ===")
	processWithDS := models.NewProcess("tm1go.TestProcessWithDS")
	processWithDS.DataSource = &models.ProcessDataSource{
		Type:                    "#tm1.ProcessDataSource",
		DataSourceNameForServer: "data.csv",
		DataSourceNameForClient: "data.csv",
		ASCIIDelimiterChar:      ",",
		ASCIIDelimiterType:      "Character",
		ASCIIHeaderRecords:      1,
		ASCIIQuoteCharacter:     "\"",
	}
	processWithDS.MetadataProcedure = `
		# Metadata procedure
	`
	processWithDS.DataProcedure = `
		# Data procedure
	`

	err = svc.Processes.Create(ctx, processWithDS)
	if err != nil {
		log.Printf("Error creating process with data source: %v", err)
	} else {
		fmt.Println("Process with data source created successfully")
	}

	// Cleanup: Delete test processes
	fmt.Println("\n=== Cleanup: Delete Test Processes ===")
	for _, testProcess := range []string{"tm1go.TestProcess", "tm1go.TestProcessWithDS"} {
		err = svc.Processes.Delete(ctx, testProcess)
		if err != nil {
			log.Printf("Error deleting process %s: %v", testProcess, err)
		} else {
			fmt.Printf("Deleted process: %s\n", testProcess)
		}
	}

	fmt.Println("\n=== All Examples Completed ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
