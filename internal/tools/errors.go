// Package tools provides MCP tool implementations for Minder operations.
package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// marshalResult converts a value to pretty-printed JSON and returns it as an MCP tool result.
// On marshal failure, returns an error result (not a Go error).
//
//nolint:unparam // error return matches tool handler signature for direct return
func marshalResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// MapGRPCError converts a gRPC error to a user-friendly error message.
//
//nolint:gocyclo // Switch statement on error codes is readable and complete
func MapGRPCError(err error) string {
	if err == nil {
		return ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}

	switch st.Code() {
	case codes.OK:
		return ""
	case codes.Canceled:
		return "Request was canceled"
	case codes.Unknown:
		return "Unknown error: " + st.Message()
	case codes.InvalidArgument:
		return "Invalid argument: " + st.Message()
	case codes.DeadlineExceeded:
		return "Request timed out"
	case codes.NotFound:
		return "Not found: " + st.Message()
	case codes.AlreadyExists:
		return "Already exists: " + st.Message()
	case codes.PermissionDenied:
		return "Permission denied: " + st.Message()
	case codes.ResourceExhausted:
		return "Resource exhausted: " + st.Message()
	case codes.FailedPrecondition:
		return "Failed precondition: " + st.Message()
	case codes.Aborted:
		return "Operation aborted: " + st.Message()
	case codes.OutOfRange:
		return "Out of range: " + st.Message()
	case codes.Unimplemented:
		return "Operation not implemented"
	case codes.Internal:
		return "Internal server error"
	case codes.Unavailable:
		return "Service unavailable"
	case codes.DataLoss:
		return "Data loss error"
	case codes.Unauthenticated:
		return "Authentication required: " + st.Message()
	default:
		return st.Message()
	}
}
