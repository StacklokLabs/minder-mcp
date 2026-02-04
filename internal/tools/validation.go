// Package tools provides MCP tool implementations for Minder operations.
package tools

import (
	"fmt"
	"strings"
)

// ValidateLookupParams validates lookup parameters for resources that support
// both ID and name-based lookups. Exactly one of id or name must be provided.
// If any auxiliary params are provided with ID-based lookup, returns an error
// since auxiliary params (like project_id, provider) only apply to name-based lookups.
// auxiliaryParams is a map of param names to their values.
// Returns an error message if validation fails, or empty string if valid.
func ValidateLookupParams(id, name, idParamName, nameParamName string, auxiliaryParams map[string]string) string {
	hasID := id != ""
	hasName := name != ""

	if !hasID && !hasName {
		return fmt.Sprintf("either %s or %s must be provided", idParamName, nameParamName)
	}
	if hasID && hasName {
		return fmt.Sprintf("cannot specify both %s and %s; use one lookup method", idParamName, nameParamName)
	}
	if hasID {
		// Check for any auxiliary params that shouldn't be used with ID lookup
		var providedAux []string
		for paramName, value := range auxiliaryParams {
			if value != "" {
				providedAux = append(providedAux, paramName)
			}
		}
		if len(providedAux) > 0 {
			return fmt.Sprintf("%s not used with %s lookup; omit when using %s",
				strings.Join(providedAux, ", "), idParamName, idParamName)
		}
	}
	return ""
}

// ValidateRepositoryLookupParams validates repository lookup parameters.
// Either repository_id OR (owner AND name) must be provided, not both.
// If auxiliary params are provided with ID-based lookup, returns an error since
// they only apply to name-based lookups.
// auxiliaryParams is a map of param names to their values.
// Returns an error message if validation fails, or empty string if valid.
func ValidateRepositoryLookupParams(id, owner, name string, auxiliaryParams map[string]string) string {
	hasID := id != ""
	hasOwnerName := owner != "" || name != ""

	if !hasID && !hasOwnerName {
		return "either repository_id or (owner and name) must be provided"
	}
	if hasID && hasOwnerName {
		return "cannot specify both repository_id and owner/name; use one lookup method"
	}
	if hasID {
		// Check for any auxiliary params that shouldn't be used with ID lookup
		var providedAux []string
		for paramName, value := range auxiliaryParams {
			if value != "" {
				providedAux = append(providedAux, paramName)
			}
		}
		if len(providedAux) > 0 {
			return fmt.Sprintf("%s not used with repository_id lookup; omit when using repository_id",
				strings.Join(providedAux, ", "))
		}
	}
	if !hasID && (owner == "" || name == "") {
		return "both owner and name are required for name-based lookup"
	}
	return ""
}
