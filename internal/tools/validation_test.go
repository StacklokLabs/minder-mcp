package tools

import "testing"

func TestValidateLookupParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		id                 string
		lookupName         string
		idParamName        string
		nameParamName      string
		auxiliaryParam     string
		auxiliaryParamName string
		wantErr            string
	}{
		{
			name:               "valid with id only",
			id:                 "uuid-123",
			lookupName:         "",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "",
			auxiliaryParamName: "project_id",
			wantErr:            "",
		},
		{
			name:               "valid with name only",
			id:                 "",
			lookupName:         "my-profile",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "",
			auxiliaryParamName: "project_id",
			wantErr:            "",
		},
		{
			name:               "valid with name and auxiliary param",
			id:                 "",
			lookupName:         "my-profile",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "project-uuid",
			auxiliaryParamName: "project_id",
			wantErr:            "",
		},
		{
			name:               "error when both id and name provided",
			id:                 "uuid-123",
			lookupName:         "my-profile",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "",
			auxiliaryParamName: "project_id",
			wantErr:            "cannot specify both profile_id and name; use one lookup method",
		},
		{
			name:               "error when neither provided",
			id:                 "",
			lookupName:         "",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "",
			auxiliaryParamName: "project_id",
			wantErr:            "either profile_id or name must be provided",
		},
		{
			name:               "error when id provided with auxiliary param",
			id:                 "uuid-123",
			lookupName:         "",
			idParamName:        "profile_id",
			nameParamName:      "name",
			auxiliaryParam:     "project-uuid",
			auxiliaryParamName: "project_id",
			wantErr:            "project_id is not used with profile_id lookup; omit project_id when using profile_id",
		},
		{
			name:               "error with artifact_id and provider",
			id:                 "uuid-123",
			lookupName:         "",
			idParamName:        "artifact_id",
			nameParamName:      "name",
			auxiliaryParam:     "github",
			auxiliaryParamName: "provider",
			wantErr:            "provider is not used with artifact_id lookup; omit provider when using artifact_id",
		},
		{
			name:               "uses custom param names in error",
			id:                 "",
			lookupName:         "",
			idParamName:        "rule_type_id",
			nameParamName:      "name",
			auxiliaryParam:     "",
			auxiliaryParamName: "project_id",
			wantErr:            "either rule_type_id or name must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateLookupParams(tt.id, tt.lookupName, tt.idParamName, tt.nameParamName, tt.auxiliaryParam, tt.auxiliaryParamName)
			if got != tt.wantErr {
				t.Errorf("ValidateLookupParams() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestValidateRepositoryLookupParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		owner    string
		repoName string
		provider string
		wantErr  string
	}{
		{
			name:     "valid with id only",
			id:       "uuid-123",
			owner:    "",
			repoName: "",
			provider: "",
			wantErr:  "",
		},
		{
			name:     "valid with owner and name",
			id:       "",
			owner:    "stacklok",
			repoName: "minder",
			provider: "",
			wantErr:  "",
		},
		{
			name:     "valid with owner, name, and provider",
			id:       "",
			owner:    "stacklok",
			repoName: "minder",
			provider: "github",
			wantErr:  "",
		},
		{
			name:     "error when neither provided",
			id:       "",
			owner:    "",
			repoName: "",
			provider: "",
			wantErr:  "either repository_id or (owner and name) must be provided",
		},
		{
			name:     "error when both id and owner/name provided",
			id:       "uuid-123",
			owner:    "stacklok",
			repoName: "minder",
			provider: "",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when id and only owner provided",
			id:       "uuid-123",
			owner:    "stacklok",
			repoName: "",
			provider: "",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when id and only name provided",
			id:       "uuid-123",
			owner:    "",
			repoName: "minder",
			provider: "",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when id provided with provider",
			id:       "uuid-123",
			owner:    "",
			repoName: "",
			provider: "github",
			wantErr:  "provider is not used with repository_id lookup; omit provider when using repository_id",
		},
		{
			name:     "error when only owner provided",
			id:       "",
			owner:    "stacklok",
			repoName: "",
			provider: "",
			wantErr:  "both owner and name are required for name-based lookup",
		},
		{
			name:     "error when only name provided",
			id:       "",
			owner:    "",
			repoName: "minder",
			provider: "",
			wantErr:  "both owner and name are required for name-based lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateRepositoryLookupParams(tt.id, tt.owner, tt.repoName, tt.provider)
			if got != tt.wantErr {
				t.Errorf("ValidateRepositoryLookupParams() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
