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

func TestListDataSources(t *testing.T) {
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
			name: "returns data sources successfully",
			mockSetup: func(m *mockMinderClient) {
				m.dataSources.listResp = &minderv1.ListDataSourcesResponse{
					DataSources: []*minderv1.DataSource{
						{Name: "osv-data"},
					},
				}
			},
			params:     map[string]any{},
			wantErr:    false,
			wantInResp: "osv-data",
		},
		{
			name: "handles gRPC error",
			mockSetup: func(m *mockMinderClient) {
				m.dataSources.listErr = status.Error(codes.Unavailable, "service down")
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

			result, err := tools.listDataSources(context.Background(), req)
			if err != nil {
				t.Fatalf("listDataSources() returned Go error: %v", err)
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

func TestGetDataSource(t *testing.T) {
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
			name: "gets data source by ID",
			mockSetup: func(m *mockMinderClient) {
				m.dataSources.getByIDResp = &minderv1.GetDataSourceByIdResponse{
					DataSource: &minderv1.DataSource{Name: "my-datasource"},
				}
			},
			params:     map[string]any{"data_source_id": "ds-123"},
			wantErr:    false,
			wantInResp: "my-datasource",
		},
		{
			name: "gets data source by name",
			mockSetup: func(m *mockMinderClient) {
				m.dataSources.getByNameResp = &minderv1.GetDataSourceByNameResponse{
					DataSource: &minderv1.DataSource{Name: "named-ds"},
				}
			},
			params:     map[string]any{"name": "named-ds"},
			wantErr:    false,
			wantInResp: "named-ds",
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
				m.dataSources.getByIDErr = status.Error(codes.NotFound, "not found")
			},
			params:      map[string]any{"data_source_id": "nonexistent"},
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

			result, err := tools.getDataSource(context.Background(), req)
			if err != nil {
				t.Fatalf("getDataSource() returned Go error: %v", err)
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
