package tools

import (
	"errors"
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
			name:    "OK code",
			err:     status.Error(codes.OK, ""),
			wantMsg: "",
		},
		{
			name:    "canceled",
			err:     status.Error(codes.Canceled, "client canceled"),
			wantMsg: "Request was canceled",
		},
		{
			name:    "unknown",
			err:     status.Error(codes.Unknown, "mystery error"),
			wantMsg: "Unknown error: mystery error",
		},
		{
			name:    "invalid argument",
			err:     status.Error(codes.InvalidArgument, "bad input"),
			wantMsg: "Invalid argument: bad input",
		},
		{
			name:    "deadline exceeded",
			err:     status.Error(codes.DeadlineExceeded, "timeout"),
			wantMsg: "Request timed out",
		},
		{
			name:    "not found",
			err:     status.Error(codes.NotFound, "resource missing"),
			wantMsg: "Not found: resource missing",
		},
		{
			name:    "already exists",
			err:     status.Error(codes.AlreadyExists, "duplicate entry"),
			wantMsg: "Already exists: duplicate entry",
		},
		{
			name:    "permission denied",
			err:     status.Error(codes.PermissionDenied, "access denied"),
			wantMsg: "Permission denied: access denied",
		},
		{
			name:    "resource exhausted",
			err:     status.Error(codes.ResourceExhausted, "quota exceeded"),
			wantMsg: "Resource exhausted: quota exceeded",
		},
		{
			name:    "failed precondition",
			err:     status.Error(codes.FailedPrecondition, "invalid state"),
			wantMsg: "Failed precondition: invalid state",
		},
		{
			name:    "aborted",
			err:     status.Error(codes.Aborted, "transaction aborted"),
			wantMsg: "Operation aborted: transaction aborted",
		},
		{
			name:    "out of range",
			err:     status.Error(codes.OutOfRange, "index out of bounds"),
			wantMsg: "Out of range: index out of bounds",
		},
		{
			name:    "unimplemented",
			err:     status.Error(codes.Unimplemented, "not supported"),
			wantMsg: "Operation not implemented",
		},
		{
			name:    "internal error",
			err:     status.Error(codes.Internal, "something broke"),
			wantMsg: "Internal server error",
		},
		{
			name:    "unavailable",
			err:     status.Error(codes.Unavailable, "server down"),
			wantMsg: "Service unavailable",
		},
		{
			name:    "data loss",
			err:     status.Error(codes.DataLoss, "corruption detected"),
			wantMsg: "Data loss error",
		},
		{
			name:    "unauthenticated",
			err:     status.Error(codes.Unauthenticated, "invalid token"),
			wantMsg: "Authentication required: invalid token",
		},
		{
			name:    "non-gRPC error",
			err:     errors.New("plain error"),
			wantMsg: "plain error",
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
