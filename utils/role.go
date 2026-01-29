package utils

import "slices"

func HasRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}

func IsAdmin(roles []string) bool {
	return HasRole(roles, "admin")
}

func IsViewer(roles []string) bool {
	return HasRole(roles, "viewer")
}

func IsOperator(roles []string) bool {
	return HasRole(roles, "operator")
}
