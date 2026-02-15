package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/andreyea/tm1go/pkg/models"
)

// ChoreService handles operations for TM1 chores.
type ChoreService struct {
	rest *RestService
}

// NewChoreService creates a new ChoreService instance.
func NewChoreService(rest *RestService) *ChoreService {
	return &ChoreService{rest: rest}
}

const choreExpand = "Tasks($expand=*,Process($select=Name),Chore($select=Name))"

// Get retrieves one chore from TM1.
func (cs *ChoreService) Get(ctx context.Context, choreName string) (*models.Chore, error) {
	query := url.Values{}
	query.Set("$expand", choreExpand)

	endpoint := fmt.Sprintf("/Chores('%s')?%s", url.PathEscape(choreName), EncodeODataQuery(query))
	var chore models.Chore
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &chore); err != nil {
		return nil, err
	}

	return &chore, nil
}

// GetAll retrieves all chores from TM1.
func (cs *ChoreService) GetAll(ctx context.Context) ([]*models.Chore, error) {
	query := url.Values{}
	query.Set("$expand", choreExpand)
	endpoint := "/Chores?" + EncodeODataQuery(query)

	var response struct {
		Value []*models.Chore `json:"value"`
	}
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetAllNames retrieves all chore names from TM1.
func (cs *ChoreService) GetAllNames(ctx context.Context) ([]string, error) {
	endpoint := "/Chores?$select=Name"

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}

	names := make([]string, len(response.Value))
	for i, c := range response.Value {
		names[i] = c.Name
	}

	return names, nil
}

// Create creates a chore in TM1.
func (cs *ChoreService) Create(ctx context.Context, chore *models.Chore) error {
	if err := cs.rest.JSON(ctx, "POST", "/Chores", chore, nil); err != nil {
		return err
	}

	if chore.DSTSensitive && chore.StartTime != "" {
		startTime, err := models.ParseChoreTime(chore.StartTime)
		if err != nil {
			return err
		}
		if err = cs.SetLocalStartTime(ctx, chore.Name, startTime); err != nil {
			return err
		}
	}

	if chore.Active {
		if err := cs.Activate(ctx, chore.Name); err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a chore from TM1.
func (cs *ChoreService) Delete(ctx context.Context, choreName string) error {
	resp, err := cs.rest.Delete(ctx, fmt.Sprintf("/Chores('%s')", url.PathEscape(choreName)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// Exists checks if a chore exists.
func (cs *ChoreService) Exists(ctx context.Context, choreName string) (bool, error) {
	endpoint := fmt.Sprintf("/Chores('%s')", url.PathEscape(choreName))
	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

// SearchForProcessName returns chores containing the specified process in tasks.
func (cs *ChoreService) SearchForProcessName(ctx context.Context, processName string) ([]*models.Chore, error) {
	normalized := strings.ToLower(strings.ReplaceAll(processName, " ", ""))
	normalized = strings.ReplaceAll(normalized, "'", "''")

	query := url.Values{}
	query.Set("$filter", fmt.Sprintf("Tasks/any(t: replace(tolower(t/Process/Name), ' ', '') eq '%s')", normalized))
	query.Set("$expand", choreExpand)

	var response struct {
		Value []*models.Chore `json:"value"`
	}
	endpoint := "/Chores?" + EncodeODataQuery(query)
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}

	return response.Value, nil
}

// SearchForParameterValue returns chores that contain a matching parameter string value.
func (cs *ChoreService) SearchForParameterValue(ctx context.Context, parameterValue string) ([]*models.Chore, error) {
	needle := strings.ToLower(parameterValue)
	needle = strings.ReplaceAll(needle, "'", "''")

	query := url.Values{}
	query.Set("$filter", fmt.Sprintf("Tasks/any(t: t/Parameters/any(p: isof(p/Value, Edm.String) and contains(tolower(p/Value), '%s')))", needle))
	query.Set("$expand", choreExpand)

	var response struct {
		Value []*models.Chore `json:"value"`
	}
	endpoint := "/Chores?" + EncodeODataQuery(query)
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}

	return response.Value, nil
}

// Update updates an existing chore and all its tasks.
func (cs *ChoreService) Update(ctx context.Context, chore *models.Chore) error {
	return cs.withDeactivatedChore(ctx, chore.Name, &chore.Active, func() error {
		patchBody := map[string]interface{}{
			"Name":          chore.Name,
			"StartTime":     chore.StartTime,
			"DSTSensitive":  chore.DSTSensitive,
			"Active":        chore.Active,
			"ExecutionMode": chore.ExecutionMode,
			"Frequency":     chore.Frequency,
		}
		endpoint := fmt.Sprintf("/Chores('%s')", url.PathEscape(chore.Name))
		if err := cs.rest.JSON(ctx, "PATCH", endpoint, patchBody, nil); err != nil {
			return err
		}

		oldCount, err := cs.getTasksCount(ctx, chore.Name)
		if err != nil {
			return err
		}

		for i, taskNew := range chore.Tasks {
			if i >= oldCount {
				if err = cs.addTask(ctx, chore.Name, taskNew); err != nil {
					return err
				}
				continue
			}

			taskOld, taskErr := cs.getTask(ctx, chore.Name, i)
			if taskErr != nil {
				return taskErr
			}
			if !taskNew.Equal(*taskOld) {
				taskNew.Step = i
				if err = cs.updateTask(ctx, chore.Name, taskNew); err != nil {
					return err
				}
			}
		}

		for j := len(chore.Tasks); j < oldCount; j++ {
			if err := cs.deleteTask(ctx, chore.Name, len(chore.Tasks)); err != nil {
				return err
			}
		}

		if chore.DSTSensitive && chore.StartTime != "" {
			startTime, parseErr := models.ParseChoreTime(chore.StartTime)
			if parseErr != nil {
				return parseErr
			}
			if err := cs.setLocalStartTimeUnsafe(ctx, chore.Name, startTime); err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateOrCreate updates an existing chore or creates it if absent.
func (cs *ChoreService) UpdateOrCreate(ctx context.Context, chore *models.Chore) error {
	exists, err := cs.Exists(ctx, chore.Name)
	if err != nil {
		return err
	}

	if exists {
		return cs.Update(ctx, chore)
	}
	return cs.Create(ctx, chore)
}

// Activate activates a chore.
func (cs *ChoreService) Activate(ctx context.Context, choreName string) error {
	resp, err := cs.rest.Post(ctx, fmt.Sprintf("/Chores('%s')/tm1.Activate", url.PathEscape(choreName)), strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// Deactivate deactivates a chore.
func (cs *ChoreService) Deactivate(ctx context.Context, choreName string) error {
	resp, err := cs.rest.Post(ctx, fmt.Sprintf("/Chores('%s')/tm1.Deactivate", url.PathEscape(choreName)), strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// SetLocalStartTime sets server local start time for a chore.
func (cs *ChoreService) SetLocalStartTime(ctx context.Context, choreName string, dt time.Time) error {
	return cs.withDeactivatedChore(ctx, choreName, nil, func() error {
		return cs.setLocalStartTimeUnsafe(ctx, choreName, dt)
	})
}

func (cs *ChoreService) setLocalStartTimeUnsafe(ctx context.Context, choreName string, dt time.Time) error {
	data := map[string]string{
		"StartDate": fmt.Sprintf("%d-%d-%d", dt.Year(), int(dt.Month()), dt.Day()),
		"StartTime": fmt.Sprintf("%02d:%02d:%02d", dt.Hour(), dt.Minute(), dt.Second()),
	}

	return cs.rest.JSON(ctx, "POST", fmt.Sprintf("/Chores('%s')/tm1.SetServerLocalStartTime", url.PathEscape(choreName)), data, nil)
}

// ExecuteChore executes a chore.
func (cs *ChoreService) ExecuteChore(ctx context.Context, choreName string) error {
	resp, err := cs.rest.Post(ctx, fmt.Sprintf("/Chores('%s')/tm1.Execute", url.PathEscape(choreName)), strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

func (cs *ChoreService) getTasksCount(ctx context.Context, choreName string) (int, error) {
	resp, err := cs.rest.Get(ctx, fmt.Sprintf("/Chores('%s')/Tasks/$count", url.PathEscape(choreName)))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var count int
	if _, err = fmt.Fscanf(resp.Body, "%d", &count); err != nil {
		return 0, err
	}
	return count, nil
}

func (cs *ChoreService) getTask(ctx context.Context, choreName string, step int) (*models.ChoreTask, error) {
	query := url.Values{}
	query.Set("$expand", "*,Process($select=Name),Chore($select=Name)")
	endpoint := fmt.Sprintf("/Chores('%s')/Tasks(%d)?%s", url.PathEscape(choreName), step, EncodeODataQuery(query))

	var task models.ChoreTask
	if err := cs.rest.JSON(ctx, "GET", endpoint, nil, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (cs *ChoreService) deleteTask(ctx context.Context, choreName string, step int) error {
	resp, err := cs.rest.Delete(ctx, fmt.Sprintf("/Chores('%s')/Tasks(%d)", url.PathEscape(choreName), step))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

func (cs *ChoreService) addTask(ctx context.Context, choreName string, task models.ChoreTask) error {
	body := task.ToRequestBody()
	return cs.rest.JSON(ctx, "POST", fmt.Sprintf("/Chores('%s')/Tasks", url.PathEscape(choreName)), body, nil)
}

func (cs *ChoreService) updateTask(ctx context.Context, choreName string, task models.ChoreTask) error {
	endpoint := fmt.Sprintf("/Chores('%s')/Tasks(%d)", url.PathEscape(choreName), task.Step)
	body := task.ToRequestBody()
	return cs.rest.JSON(ctx, "PATCH", endpoint, body, nil)
}

func (cs *ChoreService) withDeactivatedChore(ctx context.Context, choreName string, reactivate *bool, fn func() error) error {
	chore, err := cs.Get(ctx, choreName)
	if err != nil {
		return err
	}

	shouldReactivate := chore.Active
	if reactivate != nil {
		shouldReactivate = *reactivate
	}

	wasActive := chore.Active
	if wasActive {
		if err = cs.Deactivate(ctx, choreName); err != nil {
			return err
		}
	}

	fnErr := fn()
	if shouldReactivate {
		activateErr := cs.Activate(ctx, choreName)
		if fnErr != nil {
			if activateErr != nil {
				return fmt.Errorf("%w (reactivation failed: %v)", fnErr, activateErr)
			}
			return fnErr
		}
		return activateErr
	}

	return fnErr
}
