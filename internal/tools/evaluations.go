package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (t *Tools) listEvaluationHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	projectID := req.GetString("project_id", "")
	profileName := req.GetString("profile_name", "")
	entityType := req.GetString("entity_type", "")
	entityName := req.GetString("entity_name", "")
	evalStatus := req.GetString("evaluation_status", "")
	remediationStatus := req.GetString("remediation_status", "")
	alertStatus := req.GetString("alert_status", "")
	fromStr := req.GetString("from", "")
	toStr := req.GetString("to", "")

	reqProto := &minderv1.ListEvaluationHistoryRequest{}

	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}

	if profileName != "" {
		reqProto.ProfileName = []string{profileName}
	}

	if entityType != "" {
		reqProto.EntityType = []string{entityType}
	}

	if entityName != "" {
		reqProto.EntityName = []string{entityName}
	}

	if evalStatus != "" {
		reqProto.Status = []string{evalStatus}
	}

	if remediationStatus != "" {
		reqProto.Remediation = []string{remediationStatus}
	}

	if alertStatus != "" {
		reqProto.Alert = []string{alertStatus}
	}

	if fromStr != "" {
		if ts, err := time.Parse(time.RFC3339, fromStr); err == nil {
			reqProto.From = timestamppb.New(ts)
		}
	}

	if toStr != "" {
		if ts, err := time.Parse(time.RFC3339, toStr); err == nil {
			reqProto.To = timestamppb.New(ts)
		}
	}

	resp, err := client.EvalResults().ListEvaluationHistory(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
