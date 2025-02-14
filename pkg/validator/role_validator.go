package validator

import "strings"

// RoleValidator handles role-based email validation
type RoleValidator struct {
	rolePrefixes []string
}

// NewRoleValidator creates a new instance of RoleValidator
func NewRoleValidator() *RoleValidator {
	return &RoleValidator{
		rolePrefixes: []string{
			"admin",
			"support",
			"info",
			"sales",
			"contact",
			"help",
			"marketing",
			"team",
			"billing",
			"office",
		},
	}
}

// Validate checks if the email address is role-based
func (v *RoleValidator) Validate(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := strings.ToLower(parts[0])
	for _, prefix := range v.rolePrefixes {
		if localPart == prefix {
			return true
		}
	}
	return false
}
