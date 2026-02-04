package tools

import "testing"

func TestValidateLookupParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            string
		lookupName    string
		idParamName   string
		nameParamName string
		wantErr       string
	}{
		{
			name:          "valid with id only",
			id:            "uuid-123",
			lookupName:    "",
			idParamName:   "profile_id",
			nameParamName: "name",
			wantErr:       "",
		},
		{
			name:          "valid with name only",
			id:            "",
			lookupName:    "my-profile",
			idParamName:   "profile_id",
			nameParamName: "name",
			wantErr:       "",
		},
		{
			name:          "error when both provided",
			id:            "uuid-123",
			lookupName:    "my-profile",
			idParamName:   "profile_id",
			nameParamName: "name",
			wantErr:       "cannot specify both profile_id and name; use one lookup method",
		},
		{
			name:          "error when neither provided",
			id:            "",
			lookupName:    "",
			idParamName:   "profile_id",
			nameParamName: "name",
			wantErr:       "either profile_id or name must be provided",
		},
		{
			name:          "uses custom param names in error",
			id:            "",
			lookupName:    "",
			idParamName:   "rule_type_id",
			nameParamName: "name",
			wantErr:       "either rule_type_id or name must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateLookupParams(tt.id, tt.lookupName, tt.idParamName, tt.nameParamName)
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
		wantErr  string
	}{
		{
			name:     "valid with id only",
			id:       "uuid-123",
			owner:    "",
			repoName: "",
			wantErr:  "",
		},
		{
			name:     "valid with owner and name",
			id:       "",
			owner:    "stacklok",
			repoName: "minder",
			wantErr:  "",
		},
		{
			name:     "error when neither provided",
			id:       "",
			owner:    "",
			repoName: "",
			wantErr:  "either repository_id or (owner and name) must be provided",
		},
		{
			name:     "error when both id and owner/name provided",
			id:       "uuid-123",
			owner:    "stacklok",
			repoName: "minder",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when id and only owner provided",
			id:       "uuid-123",
			owner:    "stacklok",
			repoName: "",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when id and only name provided",
			id:       "uuid-123",
			owner:    "",
			repoName: "minder",
			wantErr:  "cannot specify both repository_id and owner/name; use one lookup method",
		},
		{
			name:     "error when only owner provided",
			id:       "",
			owner:    "stacklok",
			repoName: "",
			wantErr:  "both owner and name are required for name-based lookup",
		},
		{
			name:     "error when only name provided",
			id:       "",
			owner:    "",
			repoName: "minder",
			wantErr:  "both owner and name are required for name-based lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateRepositoryLookupParams(tt.id, tt.owner, tt.repoName)
			if got != tt.wantErr {
				t.Errorf("ValidateRepositoryLookupParams() = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
