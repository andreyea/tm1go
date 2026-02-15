package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

// JobService handles TM1 Job objects (introduced in TM1 v12).
type JobService struct {
	rest *RestService
}

// NewJobService creates a new JobService instance.
func NewJobService(rest *RestService) *JobService {
	return &JobService{rest: rest}
}

// GetAll returns all currently running jobs.
func (js *JobService) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	if err := js.requireVersion12(); err != nil {
		return nil, err
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := js.rest.JSON(ctx, "GET", "/Jobs", nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// Cancel cancels a running job by ID.
func (js *JobService) Cancel(ctx context.Context, jobID interface{}) error {
	if err := js.requireVersion12(); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/Jobs('%s')/tm1.Cancel", url.PathEscape(fmt.Sprintf("%v", jobID)))
	resp, err := js.rest.Post(ctx, endpoint, strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// CancelAll cancels all currently running jobs.
func (js *JobService) CancelAll(ctx context.Context) ([]map[string]interface{}, error) {
	jobs, err := js.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	canceled := make([]map[string]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if err := js.Cancel(ctx, job["ID"]); err != nil {
			return nil, err
		}
		canceled = append(canceled, job)
	}
	return canceled, nil
}

// GetAsDataFrame returns jobs as a gota DataFrame.
func (js *JobService) GetAsDataFrame(ctx context.Context) (dataframe.DataFrame, error) {
	jobs, err := js.GetAll(ctx)
	if err != nil {
		return dataframe.DataFrame{}, err
	}
	return dataframe.LoadMaps(jobs), nil
}

func (js *JobService) requireVersion12() error {
	version := strings.TrimSpace(js.rest.version)
	if version == "" {
		return fmt.Errorf("TM1 server version is unknown")
	}
	if !IsV1GreaterOrEqualToV2(version, "12.0.0") {
		return fmt.Errorf("operation requires TM1 version >= 12.0.0, current version: %s", version)
	}
	return nil
}
