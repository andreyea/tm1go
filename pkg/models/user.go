package models

import "strings"

const (
	// UserTypeUser is a standard TM1 user.
	UserTypeUser = "User"
	// UserTypeSecurityAdmin is a TM1 security admin.
	UserTypeSecurityAdmin = "SecurityAdmin"
	// UserTypeDataAdmin is a TM1 data admin.
	UserTypeDataAdmin = "DataAdmin"
	// UserTypeAdmin is a full TM1 admin.
	UserTypeAdmin = "Admin"
	// UserTypeOperationsAdmin is a TM1 operations admin.
	UserTypeOperationsAdmin = "OperationsAdmin"
)

// User represents a TM1 user.
type User struct {
	Name         string        `json:"Name"`
	FriendlyName string        `json:"FriendlyName,omitempty"`
	Password     string        `json:"Password,omitempty"`
	Enabled      *bool         `json:"Enabled,omitempty"`
	Type         string        `json:"Type,omitempty"`
	Groups       []NamedObject `json:"Groups,omitempty"`
}

// GroupNames returns the user group names.
func (u *User) GroupNames() []string {
	if u == nil || len(u.Groups) == 0 {
		return nil
	}
	names := make([]string, 0, len(u.Groups))
	for _, g := range u.Groups {
		names = append(names, g.Name)
	}
	return names
}

// IsAdmin indicates whether the user has TM1 admin type.
func (u *User) IsAdmin() bool {
	if u == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(u.Type), UserTypeAdmin)
}

// IsSecurityAdmin indicates whether the user is security admin or full admin.
func (u *User) IsSecurityAdmin() bool {
	if u == nil {
		return false
	}
	userType := strings.TrimSpace(u.Type)
	return strings.EqualFold(userType, UserTypeSecurityAdmin) || strings.EqualFold(userType, UserTypeAdmin)
}
