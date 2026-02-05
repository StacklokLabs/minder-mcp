package tools

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//nolint:gocyclo // complexity is inherent to the number of supported filter parameters
func (t *Tools) listEvaluationHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")
	profileName := req.GetString("profile_name", "")
	entityType := req.GetString("entity_type", "")
	entityName := req.GetString("entity_name", "")
	evalStatus := req.GetString("evaluation_status", "")
	remediationStatus := req.GetString("remediation_status", "")
	alertStatus := req.GetString("alert_status", "")
	fromStr := req.GetString("from", "")
	toStr := req.GetString("to", "")
	cursor := req.GetString("cursor", "")
	pageSize := req.GetInt("page_size", 0)
	labelFilter := req.GetString("label_filter", "*") // Default to "*" to include all profiles

	// Parse time filters once
	var fromTime, toTime *timestamppb.Timestamp
	if fromStr != "" {
		if ts, err := time.Parse(time.RFC3339, fromStr); err == nil {
			fromTime = timestamppb.New(ts)
		}
	}
	if toStr != "" {
		if ts, err := time.Parse(time.RFC3339, toStr); err == nil {
			toTime = timestamppb.New(ts)
		}
	}

	// Use multi-project aggregation when no project_id specified
	// Note: pagination only works within a single project when aggregating
	evaluations, err := forEachProject(
		ctx, client, projectID,
		func(ctx context.Context, projID string) ([]*minderv1.EvaluationHistory, error) {
			reqProto := &minderv1.ListEvaluationHistoryRequest{
				Context: &minderv1.Context{
					Project: &projID,
				},
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
			if labelFilter != "" {
				reqProto.LabelFilter = []string{labelFilter}
			}
			if fromTime != nil {
				reqProto.From = fromTime
			}
			if toTime != nil {
				reqProto.To = toTime
			}

			// Add pagination parameters (advanced cursor)
			if cursor != "" || pageSize > 0 {
				reqProto.Cursor = &minderv1.Cursor{}
				if cursor != "" {
					reqProto.Cursor.Cursor = cursor
				}
				if pageSize > 0 && pageSize <= 100 {
					reqProto.Cursor.Size = uint32(pageSize) //nolint:gosec // pageSize is bounded by schema validation (1-100)
				}
			}

			resp, err := client.EvalResults().ListEvaluationHistory(ctx, reqProto)
			if err != nil {
				return nil, err
			}
			return resp.Data, nil
		})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	// Build response (pagination info not reliable when aggregating multiple projects)
	result := map[string]any{
		"results": evaluations,
	}

	return marshalResult(result)
}
