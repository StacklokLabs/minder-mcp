// Package tools provides MCP tool implementations for Minder operations.
package tools

import "fmt"

// ValidateLookupParams validates that exactly one of id or name is provided.
// Returns an error message if validation fails, or empty string if valid.
func ValidateLookupParams(id, name, idParamName, nameParamName string) string {
	hasID := id != ""
	hasName := name != ""
	if !hasID && !hasName {
		return fmt.Sprintf("either %s or %s must be provided", idParamName, nameParamName)
	}
	if hasID && hasName {
		return fmt.Sprintf("cannot specify both %s and %s; use one lookup method", idParamName, nameParamName)
	}
	return ""
}

// ValidateRepositoryLookupParams validates repository lookup parameters.
// Either repository_id OR (owner AND name) must be provided, not both.
// Returns an error message if validation fails, or empty string if valid.
func ValidateRepositoryLookupParams(id, owner, name string) string {
	hasID := id != ""
	hasOwnerName := owner != "" || name != ""
	if !hasID && !hasOwnerName {
		return "either repository_id or (owner and name) must be provided"
	}
	if hasID && hasOwnerName {
		return "cannot specify both repository_id and owner/name; use one lookup method"
	}
	if !hasID && (owner == "" || name == "") {
		return "both owner and name are required for name-based lookup"
	}
	return ""
}
