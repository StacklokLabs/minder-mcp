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

func TestListRepositories(t *testing.T) {
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
			name: "returns repositories successfully",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.listResp = &minderv1.ListRepositoriesResponse{
					Results: []*minderv1.Repository{
						{Name: "test-repo", Owner: "test-owner"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "test-repo",
		},
		{
			name: "returns repositories with pagination",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.listResp = &minderv1.ListRepositoriesResponse{
					Results: []*minderv1.Repository{
						{Name: "paginated-repo"},
					},
					Cursor: "next-page-cursor",
				}
			},
			params:     map[string]any{"limit": 10},
			wantErr:    false,
			wantInResp: "next_cursor",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.listErr = status.Error(codes.Unavailable, "service down")
			},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "unavailable",
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

			result, err := tools.listRepositories(context.Background(), req)
			if err != nil {
				t.Fatalf("listRepositories() returned Go error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result, got success")
				}
				text := getResultText(t, result)
				if !strings.Contains(strings.ToLower(text), strings.ToLower(tt.errContains)) {
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

func TestGetRepository(t *testing.T) {
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
			name: "gets repository by ID",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.getByIDResp = &minderv1.GetRepositoryByIdResponse{
					Repository: &minderv1.Repository{Name: "my-repo", Owner: "my-owner"},
				}
			},
			params:     map[string]any{"repository_id": "repo-123"},
			wantErr:    false,
			wantInResp: "my-repo",
		},
		{
			name: "gets repository by owner/name",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.getByNameResp = &minderv1.GetRepositoryByNameResponse{
					Repository: &minderv1.Repository{Name: "named-repo", Owner: "named-owner"},
				}
			},
			params:     map[string]any{"owner": "named-owner", "name": "named-repo"},
			wantErr:    false,
			wantInResp: "named-repo",
		},
		{
			name:        "error when neither ID nor owner/name provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "must be provided",
		},
		{
			name:        "error when only owner provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{"owner": "some-owner"},
			wantErr:     true,
			errContains: "both owner and name",
		},
		{
			name: "handles not found error",
			mockSetup: func(m *mockMinderClient) {
				m.repositories.getByIDErr = status.Error(codes.NotFound, "repo not found")
			},
			params:      map[string]any{"repository_id": "nonexistent"},
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

			result, err := tools.getRepository(context.Background(), req)
			if err != nil {
				t.Fatalf("getRepository() returned Go error: %v", err)
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
