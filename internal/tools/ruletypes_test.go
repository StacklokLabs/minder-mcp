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

func TestListRuleTypes(t *testing.T) {
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
			name: "returns rule types successfully",
			mockSetup: func(m *mockMinderClient) {
				m.ruleTypes.listResp = &minderv1.ListRuleTypesResponse{
					RuleTypes: []*minderv1.RuleType{
						{Name: "secret_scanning"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "secret_scanning",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.ruleTypes.listErr = status.Error(codes.PermissionDenied, "no access")
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

			result, err := tools.listRuleTypes(context.Background(), req)
			if err != nil {
				t.Fatalf("listRuleTypes() returned Go error: %v", err)
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

func TestGetRuleType(t *testing.T) {
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
			name: "gets rule type by ID",
			mockSetup: func(m *mockMinderClient) {
				m.ruleTypes.getByIDResp = &minderv1.GetRuleTypeByIdResponse{
					RuleType: &minderv1.RuleType{Name: "my-rule"},
				}
			},
			params:     map[string]any{"rule_type_id": "rt-123"},
			wantErr:    false,
			wantInResp: "my-rule",
		},
		{
			name: "gets rule type by name",
			mockSetup: func(m *mockMinderClient) {
				m.ruleTypes.getByNameResp = &minderv1.GetRuleTypeByNameResponse{
					RuleType: &minderv1.RuleType{Name: "named-rule"},
				}
			},
			params:     map[string]any{"name": "named-rule"},
			wantErr:    false,
			wantInResp: "named-rule",
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
				m.ruleTypes.getByIDErr = status.Error(codes.NotFound, "not found")
			},
			params:      map[string]any{"rule_type_id": "nonexistent"},
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

			result, err := tools.getRuleType(context.Background(), req)
			if err != nil {
				t.Fatalf("getRuleType() returned Go error: %v", err)
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
