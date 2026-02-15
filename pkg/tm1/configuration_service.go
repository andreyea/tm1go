package tm1

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// ConfigurationService handles TM1 server configuration endpoints.
type ConfigurationService struct {
	rest  *RestService
	users *UserService
}

// NewConfigurationService creates a new ConfigurationService instance.
func NewConfigurationService(rest *RestService) *ConfigurationService {
	return &ConfigurationService{rest: rest, users: NewUserService(rest)}
}

// GetAll returns /Configuration as a dictionary.
func (cs *ConfigurationService) GetAll(ctx context.Context) (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := cs.rest.JSON(ctx, "GET", "/Configuration", nil, &config); err != nil {
		return nil, err
	}
	delete(config, "@odata.context")
	return config, nil
}

// GetServerName returns /Configuration/ServerName/$value.
func (cs *ConfigurationService) GetServerName(ctx context.Context) (string, error) {
	return cs.getStringValue(ctx, "/Configuration/ServerName/$value")
}

// GetProductVersion returns /Configuration/ProductVersion/$value.
func (cs *ConfigurationService) GetProductVersion(ctx context.Context) (string, error) {
	return cs.getStringValue(ctx, "/Configuration/ProductVersion/$value")
}

// GetAdminHost returns /Configuration/AdminHost/$value.
// Deprecated in TM1 v12.
func (cs *ConfigurationService) GetAdminHost(ctx context.Context) (string, error) {
	if err := cs.requirePreV12(); err != nil {
		return "", err
	}
	return cs.getStringValue(ctx, "/Configuration/AdminHost/$value")
}

// GetDataDirectory returns /Configuration/DataBaseDirectory/$value.
// Deprecated in TM1 v12.
func (cs *ConfigurationService) GetDataDirectory(ctx context.Context) (string, error) {
	if err := cs.requirePreV12(); err != nil {
		return "", err
	}
	return cs.getStringValue(ctx, "/Configuration/DataBaseDirectory/$value")
}

// GetStatic returns /StaticConfiguration. Requires ops admin.
func (cs *ConfigurationService) GetStatic(ctx context.Context) (map[string]interface{}, error) {
	if err := cs.requireOpsAdmin(ctx); err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := cs.rest.JSON(ctx, "GET", "/StaticConfiguration", nil, &config); err != nil {
		return nil, err
	}
	delete(config, "@odata.context")
	return config, nil
}

// GetActive returns /ActiveConfiguration. Requires ops admin.
func (cs *ConfigurationService) GetActive(ctx context.Context) (map[string]interface{}, error) {
	if err := cs.requireOpsAdmin(ctx); err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := cs.rest.JSON(ctx, "GET", "/ActiveConfiguration", nil, &config); err != nil {
		return nil, err
	}
	delete(config, "@odata.context")
	return config, nil
}

// UpdateStatic patches /StaticConfiguration. Requires ops admin.
func (cs *ConfigurationService) UpdateStatic(ctx context.Context, configuration map[string]interface{}) error {
	if err := cs.requireOpsAdmin(ctx); err != nil {
		return err
	}
	return cs.rest.JSON(ctx, "PATCH", "/StaticConfiguration", configuration, nil)
}

func (cs *ConfigurationService) getStringValue(ctx context.Context, endpoint string) (string, error) {
	resp, err := cs.rest.Get(ctx, endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (cs *ConfigurationService) requireOpsAdmin(ctx context.Context) error {
	user, err := cs.users.GetCurrent(ctx)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("operations admin privileges required")
	}
	if strings.EqualFold(user.Type, "Admin") || strings.EqualFold(user.Type, "OperationsAdmin") {
		return nil
	}
	return fmt.Errorf("operations admin privileges required")
}

func (cs *ConfigurationService) requirePreV12() error {
	version := strings.TrimSpace(cs.rest.version)
	if version == "" {
		return nil
	}
	if IsV1GreaterOrEqualToV2(version, "12.0.0") {
		return fmt.Errorf("operation is deprecated and unavailable in TM1 version 12.0.0+")
	}
	return nil
}
