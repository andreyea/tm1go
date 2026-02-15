package models

import (
	"fmt"
	"strings"
	"time"
)

// Chore represents a TM1 chore.
type Chore struct {
	Name          string      `json:"Name"`
	StartTime     string      `json:"StartTime,omitempty"`
	DSTSensitive  bool        `json:"DSTSensitive,omitempty"`
	Active        bool        `json:"Active,omitempty"`
	ExecutionMode string      `json:"ExecutionMode,omitempty"`
	Frequency     string      `json:"Frequency,omitempty"`
	Tasks         []ChoreTask `json:"Tasks,omitempty"`
}

// ChoreTask represents a process task in a TM1 chore.
type ChoreTask struct {
	Step           int                  `json:"Step,omitempty"`
	Process        *NamedObject         `json:"Process,omitempty"`
	ProcessBinding string               `json:"Process@odata.bind,omitempty"`
	Parameters     []ChoreTaskParameter `json:"Parameters,omitempty"`
}

// NamedObject captures objects with a Name field in TM1 responses.
type NamedObject struct {
	Name string `json:"Name"`
}

// ChoreTaskParameter represents one process parameter value in a chore task.
type ChoreTaskParameter struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

// ProcessName resolves the task process name from response or request shape.
func (t ChoreTask) ProcessName() string {
	if t.Process != nil && t.Process.Name != "" {
		return t.Process.Name
	}
	if t.ProcessBinding == "" {
		return ""
	}
	// Expected: Processes('ProcessName')
	if strings.HasPrefix(t.ProcessBinding, "Processes('") && strings.HasSuffix(t.ProcessBinding, "')") {
		return t.ProcessBinding[len("Processes('") : len(t.ProcessBinding)-2]
	}
	return t.ProcessBinding
}

// ToRequestBody converts the task to request format (Process@odata.bind + Parameters).
func (t ChoreTask) ToRequestBody() map[string]interface{} {
	body := map[string]interface{}{
		"Process@odata.bind": fmt.Sprintf("Processes('%s')", t.ProcessName()),
		"Parameters":         t.Parameters,
	}
	return body
}

// Equal compares two ChoreTask values by process name and parameters.
func (t ChoreTask) Equal(other ChoreTask) bool {
	if t.ProcessName() != other.ProcessName() {
		return false
	}
	if len(t.Parameters) != len(other.Parameters) {
		return false
	}
	for i := range t.Parameters {
		if t.Parameters[i].Name != other.Parameters[i].Name {
			return false
		}
		if fmt.Sprintf("%v", t.Parameters[i].Value) != fmt.Sprintf("%v", other.Parameters[i].Value) {
			return false
		}
	}
	return true
}

// ParseChoreTime parses TM1 chore StartTime values.
func ParseChoreTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04Z",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse chore time: %s", value)
}
