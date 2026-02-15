package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// SessionService handles TM1 session APIs.
type SessionService struct {
	rest  *RestService
	users *UserService
}

// NewSessionService creates a new SessionService instance.
func NewSessionService(rest *RestService) *SessionService {
	return &SessionService{rest: rest, users: NewUserService(rest)}
}

// GetAll returns all sessions, optionally expanded with user and thread info.
func (ss *SessionService) GetAll(ctx context.Context, includeUser bool, includeThreads bool) ([]map[string]interface{}, error) {
	endpoint := "/Sessions"
	if includeUser || includeThreads {
		expands := make([]string, 0, 2)
		if includeUser {
			expands = append(expands, "User")
		}
		if includeThreads {
			expands = append(expands, "Threads")
		}
		endpoint += "?$expand=" + strings.Join(expands, ",")
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetCurrent returns the current active session.
func (ss *SessionService) GetCurrent(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := ss.rest.JSON(ctx, "GET", "/ActiveSession", nil, &response); err != nil {
		return nil, err
	}

	if value, ok := response["value"]; ok {
		if asMap, ok := value.(map[string]interface{}); ok {
			return asMap, nil
		}
	}
	return response, nil
}

// GetThreadsForCurrent gets threads for the current active session.
func (ss *SessionService) GetThreadsForCurrent(ctx context.Context, excludeIdle bool) ([]map[string]interface{}, error) {
	query := "Function ne 'GET /ActiveSession/Threads'"
	if excludeIdle {
		query += " and State ne 'Idle'"
	}
	q := url.Values{}
	q.Set("$filter", query)
	endpoint := "/ActiveSession/Threads?" + EncodeODataQuery(q)

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// Close closes a session by ID.
func (ss *SessionService) Close(ctx context.Context, sessionID interface{}) error {
	endpoint := fmt.Sprintf("/Sessions('%s')/tm1.Close", url.PathEscape(fmt.Sprintf("%v", sessionID)))
	resp, err := ss.rest.Post(ctx, endpoint, strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// CloseAll closes all sessions except the current user's sessions. Requires admin.
func (ss *SessionService) CloseAll(ctx context.Context) ([]map[string]interface{}, error) {
	currentUser, err := ss.users.GetCurrent(ctx)
	if err != nil {
		return nil, err
	}
	if !currentUser.IsAdmin() {
		return nil, fmt.Errorf("admin privileges required")
	}

	sessions, err := ss.GetAll(ctx, true, true)
	if err != nil {
		return nil, err
	}

	closed := make([]map[string]interface{}, 0)
	for _, session := range sessions {
		userObj, ok := session["User"].(map[string]interface{})
		if !ok || userObj == nil {
			continue
		}
		userNameRaw, ok := userObj["Name"]
		if !ok {
			continue
		}
		userName, ok := userNameRaw.(string)
		if !ok {
			continue
		}
		if caseAndSpaceInsensitiveEquals(currentUser.Name, userName) {
			continue
		}

		if err := ss.Close(ctx, session["ID"]); err != nil {
			return nil, err
		}
		closed = append(closed, session)
	}

	return closed, nil
}
