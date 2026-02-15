package tm1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// SecurityService handles TM1 user/group security operations.
type SecurityService struct {
	rest    *RestService
	process *ProcessService
	cells   *CellService
	users   *UserService
}

// NewSecurityService creates a new SecurityService instance.
func NewSecurityService(rest *RestService) *SecurityService {
	return &SecurityService{
		rest:    rest,
		process: NewProcessService(rest),
		cells:   NewCellService(rest),
		users:   NewUserService(rest),
	}
}

// DetermineActualUserName resolves a user name case/space-insensitively against TM1 objects.
func (ss *SecurityService) DetermineActualUserName(ctx context.Context, userName string) (string, error) {
	return ss.determineActualObjectName(ctx, "Users", userName)
}

// DetermineActualGroupName resolves a group name case/space-insensitively against TM1 objects.
func (ss *SecurityService) DetermineActualGroupName(ctx context.Context, groupName string) (string, error) {
	return ss.determineActualObjectName(ctx, "Groups", groupName)
}

// CreateUser creates a user on TM1. Requires security admin privileges.
func (ss *SecurityService) CreateUser(ctx context.Context, user *models.User) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	body := ss.buildUserPayload(user)
	return ss.rest.JSON(ctx, "POST", "/Users", body, nil)
}

// CreateGroup creates a security group. Requires security admin privileges.
func (ss *SecurityService) CreateGroup(ctx context.Context, groupName string) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	return ss.rest.JSON(ctx, "POST", "/Groups", map[string]string{"Name": groupName}, nil)
}

// GetUser retrieves a user definition from TM1.
func (ss *SecurityService) GetUser(ctx context.Context, userName string) (*models.User, error) {
	actual, err := ss.DetermineActualUserName(ctx, userName)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("$select", "Name,FriendlyName,Password,Type,Enabled")
	query.Set("$expand", "Groups")
	endpoint := fmt.Sprintf("/Users('%s')?%s", url.PathEscape(actual), EncodeODataQuery(query))

	var user models.User
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetCurrentUser retrieves user and group assignments of the current session.
func (ss *SecurityService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	query := url.Values{}
	query.Set("$select", "Name,FriendlyName,Password,Type,Enabled")
	query.Set("$expand", "Groups")

	var user models.User
	if err := ss.rest.JSON(ctx, "GET", "/ActiveUser?"+EncodeODataQuery(query), nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user. Requires security admin privileges.
func (ss *SecurityService) UpdateUser(ctx context.Context, user *models.User) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	actualName, err := ss.DetermineActualUserName(ctx, user.Name)
	if err != nil {
		return err
	}
	user.Name = actualName

	currentGroups, err := ss.GetGroups(ctx, user.Name)
	if err != nil {
		return err
	}
	desiredGroups := user.GroupNames()
	for _, currentGroup := range currentGroups {
		if !containsInsensitive(desiredGroups, currentGroup) {
			if err := ss.RemoveUserFromGroup(ctx, currentGroup, user.Name); err != nil {
				return err
			}
		}
	}

	endpoint := fmt.Sprintf("/Users('%s')", url.PathEscape(user.Name))
	body := ss.buildUserPayload(user)
	return ss.rest.JSON(ctx, "PATCH", endpoint, body, nil)
}

// UpdateUserPassword updates only the user password.
func (ss *SecurityService) UpdateUserPassword(ctx context.Context, userName, password string) error {
	endpoint := fmt.Sprintf("/Users('%s')", url.PathEscape(userName))
	return ss.rest.JSON(ctx, "PATCH", endpoint, map[string]string{"Password": password}, nil)
}

// DeleteUser deletes a user. Requires security admin privileges.
func (ss *SecurityService) DeleteUser(ctx context.Context, userName string) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	actualName, err := ss.DetermineActualUserName(ctx, userName)
	if err != nil {
		return err
	}

	resp, err := ss.rest.Delete(ctx, fmt.Sprintf("/Users('%s')", url.PathEscape(actualName)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// DeleteGroup deletes a group. Requires security admin privileges.
func (ss *SecurityService) DeleteGroup(ctx context.Context, groupName string) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	actualName, err := ss.DetermineActualGroupName(ctx, groupName)
	if err != nil {
		return err
	}

	resp, err := ss.rest.Delete(ctx, fmt.Sprintf("/Groups('%s')", url.PathEscape(actualName)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// GetAllUsers gets all users from TM1.
func (ss *SecurityService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := url.Values{}
	query.Set("$select", "Name,FriendlyName,Password,Type,Enabled")
	query.Set("$expand", "Groups")

	var response struct {
		Value []*models.User `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", "/Users?"+EncodeODataQuery(query), nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

// GetAllUserNames gets all user names from TM1.
func (ss *SecurityService) GetAllUserNames(ctx context.Context) ([]string, error) {
	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", "/Users?$select=Name", nil, &response); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(response.Value))
	for _, u := range response.Value {
		names = append(names, u.Name)
	}
	return names, nil
}

// GetUsersFromGroup gets all users belonging to a group.
func (ss *SecurityService) GetUsersFromGroup(ctx context.Context, groupName string) ([]*models.User, error) {
	actualGroup, err := ss.DetermineActualGroupName(ctx, groupName)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/Groups('%s')?$expand=Users($select=Name,FriendlyName,Password,Type,Enabled;$expand=Groups)", url.PathEscape(actualGroup))
	var response struct {
		Users []*models.User `json:"Users"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Users, nil
}

// GetUserNamesFromGroup gets all user names belonging to a group.
func (ss *SecurityService) GetUserNamesFromGroup(ctx context.Context, groupName string) ([]string, error) {
	actualGroup, err := ss.DetermineActualGroupName(ctx, groupName)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/Groups('%s')?$expand=Users($expand=Groups)", url.PathEscape(actualGroup))
	var response struct {
		Users []struct {
			Name string `json:"Name"`
		} `json:"Users"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(response.Users))
	for _, u := range response.Users {
		names = append(names, u.Name)
	}
	return names, nil
}

// GetGroups gets groups assigned to a user.
func (ss *SecurityService) GetGroups(ctx context.Context, userName string) ([]string, error) {
	actualName, err := ss.DetermineActualUserName(ctx, userName)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("/Users('%s')/Groups", url.PathEscape(actualName))

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	groups := make([]string, 0, len(response.Value))
	for _, g := range response.Value {
		groups = append(groups, g.Name)
	}
	return groups, nil
}

// AddUserToGroups adds user memberships. Requires security admin privileges.
func (ss *SecurityService) AddUserToGroups(ctx context.Context, userName string, groups []string) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	actualUser, err := ss.DetermineActualUserName(ctx, userName)
	if err != nil {
		return err
	}

	bind := make([]string, 0, len(groups))
	for _, group := range groups {
		actualGroup, groupErr := ss.DetermineActualGroupName(ctx, group)
		if groupErr != nil {
			return groupErr
		}
		bind = append(bind, fmt.Sprintf("Groups('%s')", actualGroup))
	}

	payload := map[string]interface{}{
		"Name":              actualUser,
		"Groups@odata.bind": bind,
	}

	endpoint := fmt.Sprintf("/Users('%s')", url.PathEscape(actualUser))
	return ss.rest.JSON(ctx, "PATCH", endpoint, payload, nil)
}

// RemoveUserFromGroup removes a user from a group. Requires security admin privileges.
func (ss *SecurityService) RemoveUserFromGroup(ctx context.Context, groupName, userName string) error {
	ok, err := ss.isSecurityAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("security admin privileges required")
	}

	actualUser, err := ss.DetermineActualUserName(ctx, userName)
	if err != nil {
		return err
	}
	actualGroup, err := ss.DetermineActualGroupName(ctx, groupName)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/Users('%s')/Groups?$id=Groups('%s')", url.PathEscape(actualUser), url.PathEscape(actualGroup))
	resp, err := ss.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// GetAllGroups gets all group names from TM1.
func (ss *SecurityService) GetAllGroups(ctx context.Context) ([]string, error) {
	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", "/Groups?$select=Name", nil, &response); err != nil {
		return nil, err
	}
	groups := make([]string, 0, len(response.Value))
	for _, g := range response.Value {
		groups = append(groups, g.Name)
	}
	return groups, nil
}

// SecurityRefresh runs TI command SecurityRefresh; Requires admin privileges.
func (ss *SecurityService) SecurityRefresh(ctx context.Context) error {
	ok, err := ss.isAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("admin privileges required")
	}

	process := models.NewProcess("")
	process.PrologProcedure = "SecurityRefresh;"

	success, status, _, err := ss.process.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("process did not complete successfully: %s", status)
	}
	return nil
}

// UserExists checks if a user exists.
func (ss *SecurityService) UserExists(ctx context.Context, userName string) (bool, error) {
	endpoint := fmt.Sprintf("/Users('%s')", url.PathEscape(userName))
	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return true, nil
}

// GroupExists checks if a group exists.
func (ss *SecurityService) GroupExists(ctx context.Context, groupName string) (bool, error) {
	endpoint := fmt.Sprintf("/Groups('%s')", url.PathEscape(groupName))
	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return true, nil
}

// GetCustomSecurityGroups returns all non-built-in security groups.
func (ss *SecurityService) GetCustomSecurityGroups(ctx context.Context) ([]string, error) {
	groups, err := ss.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}

	builtin := map[string]struct{}{
		normalizeCaseSpace("Admin"):           {},
		normalizeCaseSpace("DataAdmin"):       {},
		normalizeCaseSpace("SecurityAdmin"):   {},
		normalizeCaseSpace("OperationsAdmin"): {},
		normalizeCaseSpace("}tp_Everyone"):    {},
	}

	result := make([]string, 0)
	seen := map[string]struct{}{}
	for _, g := range groups {
		n := normalizeCaseSpace(g)
		if _, isBuiltin := builtin[n]; isBuiltin {
			continue
		}
		if _, already := seen[n]; already {
			continue
		}
		seen[n] = struct{}{}
		result = append(result, g)
	}
	sort.Strings(result)
	return result, nil
}

// GetReadOnlyUsers retrieves users flagged as read-only in }ClientProperties cube.
func (ss *SecurityService) GetReadOnlyUsers(ctx context.Context) ([]string, error) {
	mdx := `
SELECT
{[}ClientProperties].[ReadOnlyUser]} ON COLUMNS,
NON EMPTY {[}Clients].MEMBERS} ON ROWS
FROM [}ClientProperties]
`
	cellset, err := ss.cells.ExecuteMDX(ctx, mdx, []string{"Ordinal", "Value"}, "")
	if err != nil {
		return nil, err
	}

	readOnlyUsers := make([]string, 0)
	for coord, props := range cellset.CellMap {
		value := props["Value"]
		if !isTruthy(value) {
			continue
		}
		parts := strings.Split(coord, ",")
		if len(parts) == 0 {
			continue
		}
		candidate := strings.TrimSpace(parts[len(parts)-1])
		if candidate != "" {
			readOnlyUsers = append(readOnlyUsers, candidate)
		}
	}

	return UniqueStrings(readOnlyUsers), nil
}

func (ss *SecurityService) determineActualObjectName(ctx context.Context, objectClass, objectName string) (string, error) {
	var endpoint string
	switch objectClass {
	case "Users":
		endpoint = "/Users?$select=Name"
	case "Groups":
		endpoint = "/Groups?$select=Name"
	default:
		return "", fmt.Errorf("unsupported object class: %s", objectClass)
	}

	var response struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return "", err
	}

	target := normalizeCaseSpace(objectName)
	for _, entry := range response.Value {
		if normalizeCaseSpace(entry.Name) == target {
			return entry.Name, nil
		}
	}
	return objectName, nil
}

func (ss *SecurityService) buildUserPayload(user *models.User) map[string]interface{} {
	payload := map[string]interface{}{
		"Name":         user.Name,
		"FriendlyName": user.FriendlyName,
		"Enabled":      user.Enabled,
		"Type":         user.Type,
	}
	if payload["FriendlyName"] == "" {
		payload["FriendlyName"] = user.Name
	}
	if user.Password != "" {
		payload["Password"] = user.Password
	}

	groups := user.GroupNames()
	groupBinds := make([]string, 0, len(groups))
	for _, g := range groups {
		groupBinds = append(groupBinds, fmt.Sprintf("Groups('%s')", g))
	}
	payload["Groups@odata.bind"] = groupBinds

	return payload
}

func (ss *SecurityService) isAdmin(ctx context.Context) (bool, error) {
	current, err := ss.users.GetCurrent(ctx)
	if err != nil {
		return false, err
	}
	return current.IsAdmin(), nil
}

func (ss *SecurityService) isSecurityAdmin(ctx context.Context) (bool, error) {
	current, err := ss.users.GetCurrent(ctx)
	if err != nil {
		return false, err
	}
	return current.IsSecurityAdmin(), nil
}

func containsInsensitive(items []string, candidate string) bool {
	for _, item := range items {
		if normalizeCaseSpace(item) == normalizeCaseSpace(candidate) {
			return true
		}
	}
	return false
}

func normalizeCaseSpace(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	return strings.ToLower(value)
}

func isTruthy(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(v))
		return trimmed == "true" || trimmed == "1" || trimmed == "y" || trimmed == "yes"
	case float64:
		return v != 0
	case float32:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case int32:
		return v != 0
	default:
		return false
	}
}
