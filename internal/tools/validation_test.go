package tools

import "testing"

func TestValidateLookupParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		id              string
		lookupName      string
		idParamName     string
		nameParamName   string
		auxiliaryParams map[string]string
		wantErr         string
	}{
		{
			name:            "valid with id only",
			id:              "uuid-123",
			lookupName:      "",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": ""},
			wantErr:         "",
		},
		{
			name:            "valid with name only",
			id:              "",
			lookupName:      "my-profile",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": ""},
			wantErr:         "",
		},
		{
			name:            "valid with name and auxiliary param",
			id:              "",
			lookupName:      "my-profile",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": "project-uuid"},
			wantErr:         "",
		},
		{
			name:            "valid with name and multiple auxiliary params",
			id:              "",
			lookupName:      "my-artifact",
			idParamName:     "artifact_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": "proj-123", "provider": "github"},
			wantErr:         "",
		},
		{
			name:            "error when both id and name provided",
			id:              "uuid-123",
			lookupName:      "my-profile",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": ""},
			wantErr:         "cannot specify both profile_id and name; use one lookup method",
		},
		{
			name:            "error when neither provided",
			id:              "",
			lookupName:      "",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": ""},
			wantErr:         "either profile_id or name must be provided",
		},
		{
			name:            "error when id provided with auxiliary param",
			id:              "uuid-123",
			lookupName:      "",
			idParamName:     "profile_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": "project-uuid"},
			wantErr:         "project_id not used with profile_id lookup; omit when using profile_id",
		},
		{
			name:            "error with artifact_id and provider",
			id:              "uuid-123",
			lookupName:      "",
			idParamName:     "artifact_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"provider": "github"},
			wantErr:         "provider not used with artifact_id lookup; omit when using artifact_id",
		},
		{
			name:            "uses custom param names in error",
			id:              "",
			lookupName:      "",
			idParamName:     "rule_type_id",
			nameParamName:   "name",
			auxiliaryParams: map[string]string{"project_id": ""},
			wantErr:         "either rule_type_id or name must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateLookupParams(tt.id, tt.lookupName, tt.idParamName, tt.nameParamName, tt.auxiliaryParams)
			if got != tt.wantErr {
				t.Errorf("ValidateLookupParams() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestValidateRepositoryLookupParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		id              string
		owner           string
		repoName        string
		auxiliaryParams map[string]string
		wantErr         string
	}{
		{
			name:            "valid with id only",
			id:              "uuid-123",
			owner:           "",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "",
		},
		{
			name:            "valid with owner and name",
			id:              "",
			owner:           "stacklok",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "",
		},
		{
			name:            "valid with owner, name, and provider",
			id:              "",
			owner:           "stacklok",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "", "provider": "github"},
			wantErr:         "",
		},
		{
			name:            "valid with owner, name, project_id and provider",
			id:              "",
			owner:           "stacklok",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "proj-123", "provider": "github"},
			wantErr:         "",
		},
		{
			name:            "error when neither provided",
			id:              "",
			owner:           "",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "either repository_id or (owner and name) must be provided",
		},
		{
			name:            "error when both id and owner/name provided",
			id:              "uuid-123",
			owner:           "stacklok",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:            "error when id and only owner provided",
			id:              "uuid-123",
			owner:           "stacklok",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:            "error when id and only name provided",
			id:              "uuid-123",
			owner:           "",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:            "error when id provided with provider",
			id:              "uuid-123",
			owner:           "",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "", "provider": "github"},
			wantErr:         "provider not used with repository_id lookup; omit when using repository_id",
		},
		{
			name:            "error when id provided with project_id",
			id:              "uuid-123",
			owner:           "",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "proj-123", "provider": ""},
			wantErr:         "project_id not used with repository_id lookup; omit when using repository_id",
		},
		{
			name:            "error when only owner provided",
			id:              "",
			owner:           "stacklok",
			repoName:        "",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "both owner and name are required for name-based lookup",
		},
		{
			name:            "error when only name provided",
			id:              "",
			owner:           "",
			repoName:        "minder",
			auxiliaryParams: map[string]string{"project_id": "", "provider": ""},
			wantErr:         "both owner and name are required for name-based lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateRepositoryLookupParams(tt.id, tt.owner, tt.repoName, tt.auxiliaryParams)
			if got != tt.wantErr {
				t.Errorf("ValidateRepositoryLookupParams() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
