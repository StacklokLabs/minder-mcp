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

func TestListProjects(t *testing.T) {
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
			name: "returns all projects when no parent specified",
			mockSetup: func(m *mockMinderClient) {
				m.projects.listResp = &minderv1.ListProjectsResponse{
					Projects: []*minderv1.Project{
						{Name: "root-project", ProjectId: "proj-123"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "root-project",
		},
		{
			name: "returns child projects when parent ID specified",
			mockSetup: func(m *mockMinderClient) {
				m.projects.listChildResp = &minderv1.ListChildProjectsResponse{
					Projects: []*minderv1.Project{
						{Name: "child-project", ProjectId: "child-123"},
					},
				}
			},
			params:     map[string]any{"project_id": "parent-123"},
			wantErr:    false,
			wantInResp: "child-project",
		},
		{
			name: "handles gRPC error for list all",
			mockSetup: func(m *mockMinderClient) {
				m.projects.listErr = status.Error(codes.Unauthenticated, "invalid token")
			},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "Authentication required",
		},
		{
			name: "handles gRPC error for list children",
			mockSetup: func(m *mockMinderClient) {
				m.projects.listChildErr = status.Error(codes.NotFound, "parent not found")
			},
			params:      map[string]any{"project_id": "missing-parent"},
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

			result, err := tools.listProjects(context.Background(), req)
			if err != nil {
				t.Fatalf("listProjects() returned Go error: %v", err)
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
