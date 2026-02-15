package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// ThreadService handles TM1 thread APIs (removed as of TM1 v12).
type ThreadService struct {
	rest *RestService
}

// NewThreadService creates a new ThreadService instance.
func NewThreadService(rest *RestService) *ThreadService {
	return &ThreadService{rest: rest}
}

// GetAll returns all currently running threads.
func (ts *ThreadService) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	if err := ts.requirePreV12(); err != nil {
		return nil, err
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ts.rest.JSON(ctx, "GET", "/Threads", nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetActive returns non-idle threads.
func (ts *ThreadService) GetActive(ctx context.Context) ([]map[string]interface{}, error) {
	if err := ts.requirePreV12(); err != nil {
		return nil, err
	}

	endpoint := "/Threads?$filter=Function ne 'GET /Threads' and State ne 'Idle'"
	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ts.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// Cancel cancels a running thread.
func (ts *ThreadService) Cancel(ctx context.Context, threadID int) error {
	if err := ts.requirePreV12(); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/Threads('%s')/tm1.CancelOperation", url.PathEscape(fmt.Sprintf("%d", threadID)))
	resp, err := ts.rest.Post(ctx, endpoint, strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// CancelAllRunning cancels all running non-system threads.
func (ts *ThreadService) CancelAllRunning(ctx context.Context) ([]map[string]interface{}, error) {
	runningThreads, err := ts.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	canceled := make([]map[string]interface{}, 0)
	for _, thread := range runningThreads {
		state, _ := thread["State"].(string)
		typeName, _ := thread["Type"].(string)
		name, _ := thread["Name"].(string)
		function, _ := thread["Function"].(string)
		if state == "Idle" || typeName == "System" || name == "Pseudo" || function == "GET /Threads" || function == "GET /api/v1/Threads" {
			continue
		}

		id, ok := toInt(thread["ID"])
		if !ok {
			continue
		}
		if err := ts.Cancel(ctx, id); err != nil {
			return nil, err
		}
		canceled = append(canceled, thread)
	}
	return canceled, nil
}

func (ts *ThreadService) requirePreV12() error {
	version := strings.TrimSpace(ts.rest.version)
	if version == "" {
		return nil
	}
	if IsV1GreaterOrEqualToV2(version, "12.0.0") {
		return fmt.Errorf("threads are removed as of TM1 version 12.0.0")
	}
	return nil
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}
