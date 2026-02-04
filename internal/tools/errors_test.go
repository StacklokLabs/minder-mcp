package tools

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMarshalResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		wantJSON    string
		wantIsError bool
	}{
		{
			name:        "simple map",
			input:       map[string]string{"key": "value"},
			wantJSON:    `"key": "value"`,
			wantIsError: false,
		},
		{
			name:        "slice of strings",
			input:       []string{"a", "b", "c"},
			wantJSON:    `"a"`,
			wantIsError: false,
		},
		{
			name:        "nested struct",
			input:       struct{ Name string }{Name: "test"},
			wantJSON:    `"Name": "test"`,
			wantIsError: false,
		},
		{
			name:        "nil value",
			input:       nil,
			wantJSON:    "null",
			wantIsError: false,
		},
		{
			name:        "channel cannot be marshaled",
			input:       make(chan int),
			wantJSON:    "",
			wantIsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := marshalResult(tt.input)
			if err != nil {
				t.Fatalf("marshalResult() returned error: %v", err)
			}
			if result == nil {
				t.Fatal("marshalResult() returned nil result")
			}

			if tt.wantIsError {
				if !result.IsError {
					t.Error("expected error result, got success")
				}
			} else {
				if result.IsError {
					t.Error("expected success result, got error")
				}
				// Check that the JSON contains expected content
				if len(result.Content) == 0 {
					t.Fatal("result has no content")
				}
				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Fatalf("content is not TextContent, got %T", result.Content[0])
				}
				if !strings.Contains(textContent.Text, tt.wantJSON) {
					t.Errorf("JSON output %q does not contain %q", textContent.Text, tt.wantJSON)
				}
			}
		})
	}
}

func TestMapGRPCError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "nil error",
			err:     nil,
			wantMsg: "",
		},
		{
			name:    "not found",
			err:     status.Error(codes.NotFound, "resource missing"),
			wantMsg: "Not found: resource missing",
		},
		{
			name:    "permission denied",
			err:     status.Error(codes.PermissionDenied, "access denied"),
			wantMsg: "Permission denied: access denied",
		},
		{
			name:    "unauthenticated",
			err:     status.Error(codes.Unauthenticated, "invalid token"),
			wantMsg: "Authentication required: invalid token",
		},
		{
			name:    "internal error",
			err:     status.Error(codes.Internal, "something broke"),
			wantMsg: "Internal server error",
		},
		{
			name:    "deadline exceeded",
			err:     status.Error(codes.DeadlineExceeded, "timeout"),
			wantMsg: "Request timed out",
		},
		{
			name:    "invalid argument",
			err:     status.Error(codes.InvalidArgument, "bad input"),
			wantMsg: "Invalid argument: bad input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapGRPCError(tt.err)
			if got != tt.wantMsg {
				t.Errorf("MapGRPCError() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}
