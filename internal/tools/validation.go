// Package tools provides MCP tool implementations for Minder operations.
package tools

import "fmt"

// ValidateLookupParams validates lookup parameters for resources that support
// both ID and name-based lookups. Exactly one of id or name must be provided.
// If auxiliaryParam is provided with ID-based lookup, returns an error since
// auxiliary params (like project_id) only apply to name-based lookups.
// Returns an error message if validation fails, or empty string if valid.
func ValidateLookupParams(id, name, idParamName, nameParamName, auxiliaryParam, auxiliaryParamName string) string {
	hasID := id != ""
	hasName := name != ""
	hasAuxiliary := auxiliaryParam != ""

	if !hasID && !hasName {
		return fmt.Sprintf("either %s or %s must be provided", idParamName, nameParamName)
	}
	if hasID && hasName {
		return fmt.Sprintf("cannot specify both %s and %s; use one lookup method", idParamName, nameParamName)
	}
	if hasID && hasAuxiliary {
		return fmt.Sprintf("%s is not used with %s lookup; omit %s when using %s",
			auxiliaryParamName, idParamName, auxiliaryParamName, idParamName)
	}
	return ""
}

// ValidateRepositoryLookupParams validates repository lookup parameters.
// Either repository_id OR (owner AND name) must be provided, not both.
// If provider is provided with ID-based lookup, returns an error since
// provider only applies to name-based lookups.
// Returns an error message if validation fails, or empty string if valid.
func ValidateRepositoryLookupParams(id, owner, name, provider string) string {
	hasID := id != ""
	hasOwnerName := owner != "" || name != ""
	hasProvider := provider != ""

	if !hasID && !hasOwnerName {
		return "either repository_id or (owner and name) must be provided"
	}
	if hasID && hasOwnerName {
		return "cannot specify both repository_id and owner/name; use one lookup method"
	}
	if hasID && hasProvider {
		return "provider is not used with repository_id lookup; omit provider when using repository_id"
	}
	if !hasID && (owner == "" || name == "") {
		return "both owner and name are required for name-based lookup"
	}
	return ""
}
