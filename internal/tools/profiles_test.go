package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Note: Helper functions are in test_helpers_test.go

func TestListProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func(*mockMinderClient)
		params      map[string]any
		wantErr     bool
		errContains string
		wantInResp  string
	}{
		{
			name: "returns profiles successfully",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.listResp = &minderv1.ListProfilesResponse{
					Profiles: []*minderv1.Profile{
						{Name: "test-profile", Id: ptr("prof-123")},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "test-profile",
		},
		{
			name: "returns profiles with project filter",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.listResp = &minderv1.ListProfilesResponse{
					Profiles: []*minderv1.Profile{
						{Name: "filtered-profile"},
					},
				}
			},
			params:     map[string]any{"project_id": "proj-456"},
			wantErr:    false,
			wantInResp: "filtered-profile",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.listErr = status.Error(codes.PermissionDenied, "access denied")
			},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "Permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := newMockClient()
			tt.mockSetup(mockClient)

			tools := newTestTools(mockClient)

			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params

			result, err := tools.listProfiles(context.Background(), req)
			if err != nil {
				t.Fatalf("listProfiles() returned Go error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result, got success")
				}
				text := getResultText(t, result)
				if !strings.Contains(text, tt.errContains) {
					t.Errorf("error %q does not contain %q", text, tt.errContains)
				}
			} else {
				if result.IsError {
					t.Errorf("expected success, got error: %s", getResultText(t, result))
				}
				text := getResultText(t, result)
				if !strings.Contains(text, tt.wantInResp) {
					t.Errorf("response %q does not contain %q", text, tt.wantInResp)
				}
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func(*mockMinderClient)
		params      map[string]any
		wantErr     bool
		errContains string
		wantInResp  string
	}{
		{
			name: "gets profile by ID",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.getByIDResp = &minderv1.GetProfileByIdResponse{
					Profile: &minderv1.Profile{Name: "my-profile", Id: ptr("prof-123")},
				}
			},
			params:     map[string]any{"profile_id": "prof-123"},
			wantErr:    false,
			wantInResp: "my-profile",
		},
		{
			name: "gets profile by name",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.getByNameResp = &minderv1.GetProfileByNameResponse{
					Profile: &minderv1.Profile{Name: "named-profile"},
				}
			},
			params:     map[string]any{"name": "named-profile"},
			wantErr:    false,
			wantInResp: "named-profile",
		},
		{
			name:        "error when neither ID nor name provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "must be provided",
		},
		{
			name:        "error when both ID and name provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{"profile_id": "123", "name": "test"},
			wantErr:     true,
			errContains: "cannot specify both",
		},
		{
			name: "handles not found error",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.getByIDErr = status.Error(codes.NotFound, "profile not found")
			},
			params:      map[string]any{"profile_id": "nonexistent"},
			wantErr:     true,
			errContains: "Not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := newMockClient()
			tt.mockSetup(mockClient)

			tools := newTestTools(mockClient)

			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params

			result, err := tools.getProfile(context.Background(), req)
			if err != nil {
				t.Fatalf("getProfile() returned Go error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result, got success")
				}
				text := getResultText(t, result)
				if !strings.Contains(text, tt.errContains) {
					t.Errorf("error %q does not contain %q", text, tt.errContains)
				}
			} else {
				if result.IsError {
					t.Errorf("expected success, got error: %s", getResultText(t, result))
				}
				text := getResultText(t, result)
				if !strings.Contains(text, tt.wantInResp) {
					t.Errorf("response %q does not contain %q", text, tt.wantInResp)
				}
			}
		})
	}
}

func TestGetProfileStatusByName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func(*mockMinderClient)
		params      map[string]any
		wantErr     bool
		errContains string
		checkReq    func(*testing.T, *mockMinderClient)
	}{
		{
			name: "gets status successfully with All flag set",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.getStatusResp = &minderv1.GetProfileStatusByNameResponse{
					ProfileStatus: &minderv1.ProfileStatus{
						ProfileName: "test-profile",
					},
				}
			},
			params:  map[string]any{"name": "test-profile"},
			wantErr: false,
			checkReq: func(t *testing.T, m *mockMinderClient) {
				t.Helper()
				if m.profiles.getStatusReq == nil {
					t.Fatal("expected request to be captured")
				}
				if !m.profiles.getStatusReq.All {
					t.Error("expected All flag to be true for detailed evaluation results")
				}
				if m.profiles.getStatusReq.Name != "test-profile" {
					t.Errorf("expected name %q, got %q", "test-profile", m.profiles.getStatusReq.Name)
				}
			},
		},
		{
			name:        "error when name not provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.profiles.getStatusErr = status.Error(codes.NotFound, "profile not found")
			},
			params:      map[string]any{"name": "missing"},
			wantErr:     true,
			errContains: "Not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := newMockClient()
			tt.mockSetup(mockClient)

			tools := newTestTools(mockClient)

			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.params

			result, err := tools.getProfileStatusByName(context.Background(), req)
			if err != nil {
				t.Fatalf("getProfileStatusByName() returned Go error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result, got success")
				}
				text := getResultText(t, result)
				if !strings.Contains(text, tt.errContains) {
					t.Errorf("error %q does not contain %q", text, tt.errContains)
				}
			} else {
				if result.IsError {
					t.Errorf("expected success, got error: %s", getResultText(t, result))
				}
				if tt.checkReq != nil {
					tt.checkReq(t, mockClient)
				}
			}
		})
	}
}
