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

func TestListProviders(t *testing.T) {
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
			name: "returns providers successfully",
			mockSetup: func(m *mockMinderClient) {
				m.providers.listResp = &minderv1.ListProvidersResponse{
					Providers: []*minderv1.Provider{
						{Name: "github"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "github",
		},
		{
			name: "returns providers with pagination cursor",
			mockSetup: func(m *mockMinderClient) {
				m.providers.listResp = &minderv1.ListProvidersResponse{
					Providers: []*minderv1.Provider{
						{Name: "gitlab"},
					},
					Cursor: "next-cursor",
				}
			},
			params:     map[string]any{"limit": 10},
			wantErr:    false,
			wantInResp: "has_more",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.providers.listErr = status.Error(codes.Internal, "internal error")
			},
			params:      map[string]any{"project_id": "test-project"},
			wantErr:     true,
			errContains: "Internal server error",
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

			result, err := tools.listProviders(context.Background(), req)
			if err != nil {
				t.Fatalf("listProviders() returned Go error: %v", err)
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

func TestGetProvider(t *testing.T) {
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
			name: "gets provider by name",
			mockSetup: func(m *mockMinderClient) {
				m.providers.getResp = &minderv1.GetProviderResponse{
					Provider: &minderv1.Provider{Name: "github"},
				}
			},
			params:     map[string]any{"name": "github"},
			wantErr:    false,
			wantInResp: "github",
		},
		{
			name:        "error when name not provided",
			mockSetup:   func(_ *mockMinderClient) {},
			params:      map[string]any{},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "handles not found error",
			mockSetup: func(m *mockMinderClient) {
				m.providers.getErr = status.Error(codes.NotFound, "provider not found")
			},
			params:      map[string]any{"name": "nonexistent"},
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

			result, err := tools.getProvider(context.Background(), req)
			if err != nil {
				t.Fatalf("getProvider() returned Go error: %v", err)
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
