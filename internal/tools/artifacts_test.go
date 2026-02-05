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

func TestListArtifacts(t *testing.T) {
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
			name: "returns artifacts successfully",
			mockSetup: func(m *mockMinderClient) {
				m.artifacts.listResp = &minderv1.ListArtifactsResponse{
					Results: []*minderv1.Artifact{
						{Name: "my-image:latest"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "my-image",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.artifacts.listErr = status.Error(codes.PermissionDenied, "no access")
			},
			params:      map[string]any{"project_id": "test-project"},
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

			result, err := tools.listArtifacts(context.Background(), req)
			if err != nil {
				t.Fatalf("listArtifacts() returned Go error: %v", err)
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

func TestGetArtifact(t *testing.T) {
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
			name: "gets artifact by ID",
			mockSetup: func(m *mockMinderClient) {
				m.artifacts.getByIDResp = &minderv1.GetArtifactByIdResponse{
					Artifact: &minderv1.Artifact{Name: "artifact-by-id"},
				}
			},
			params:     map[string]any{"artifact_id": "art-123"},
			wantErr:    false,
			wantInResp: "artifact-by-id",
		},
		{
			name: "gets artifact by name",
			mockSetup: func(m *mockMinderClient) {
				m.artifacts.getByNameResp = &minderv1.GetArtifactByNameResponse{
					Artifact: &minderv1.Artifact{Name: "artifact-by-name"},
				}
			},
			params:     map[string]any{"name": "artifact-by-name"},
			wantErr:    false,
			wantInResp: "artifact-by-name",
		},
		{
			name:        "error when neither ID nor name provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "must be provided",
		},
		{
			name: "handles not found error",
			mockSetup: func(m *mockMinderClient) {
				m.artifacts.getByIDErr = status.Error(codes.NotFound, "not found")
			},
			params:      map[string]any{"artifact_id": "nonexistent"},
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

			result, err := tools.getArtifact(context.Background(), req)
			if err != nil {
				t.Fatalf("getArtifact() returned Go error: %v", err)
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
