package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// UserService handles operations for TM1 users.
type UserService struct {
	rest *RestService
}

// NewUserService creates a new UserService instance.
func NewUserService(rest *RestService) *UserService {
	return &UserService{rest: rest}
}

// GetAll retrieves all users.
func (us *UserService) GetAll(ctx context.Context) ([]*models.User, error) {
	endpoint := "/Users?$expand=Groups"
	var response struct {
		Value []*models.User `json:"value"`
	}
	if err := us.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetActive retrieves active users.
func (us *UserService) GetActive(ctx context.Context) ([]*models.User, error) {
	query := url.Values{}
	query.Set("$filter", "IsActive eq true")
	query.Set("$expand", "Groups")

	endpoint := "/Users?" + EncodeODataQuery(query)
	var response struct {
		Value []*models.User `json:"value"`
	}
	if err := us.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// IsActive checks whether a user is currently active.
func (us *UserService) IsActive(ctx context.Context, userName string) (bool, error) {
	endpoint := fmt.Sprintf("/Users('%s')/IsActive", url.PathEscape(userName))
	var response struct {
		Value bool `json:"value"`
	}
	if err := us.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return false, err
	}
	return response.Value, nil
}

// Disconnect disconnects a user session.
func (us *UserService) Disconnect(ctx context.Context, userName string) error {
	endpoint := fmt.Sprintf("/Users('%s')/tm1.Disconnect", url.PathEscape(userName))
	resp, err := us.rest.Post(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// DisconnectAll disconnects all active users except the current user.
// Requires admin privileges.
func (us *UserService) DisconnectAll(ctx context.Context) ([]string, error) {
	currentUser, err := us.GetCurrent(ctx)
	if err != nil {
		return nil, err
	}
	if !currentUser.IsAdmin() {
		return nil, fmt.Errorf("admin privileges required")
	}

	activeUsers, err := us.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	disconnected := make([]string, 0, len(activeUsers))
	for _, user := range activeUsers {
		if caseAndSpaceInsensitiveEquals(currentUser.Name, user.Name) {
			continue
		}
		if err := us.Disconnect(ctx, user.Name); err != nil {
			return nil, err
		}
		disconnected = append(disconnected, user.Name)
	}
	return disconnected, nil
}

// GetCurrent retrieves the currently authenticated user.
func (us *UserService) GetCurrent(ctx context.Context) (*models.User, error) {
	var user models.User
	if err := us.rest.JSON(ctx, "GET", "/ActiveUser?$expand=Groups", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func caseAndSpaceInsensitiveEquals(a, b string) bool {
	normalize := func(s string) string {
		s = strings.ReplaceAll(s, " ", "")
		s = strings.TrimSpace(s)
		return strings.ToLower(s)
	}
	return normalize(a) == normalize(b)
}
