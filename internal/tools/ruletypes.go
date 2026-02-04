package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listRuleTypes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")

	reqProto := &minderv1.ListRuleTypesRequest{}
	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}

	resp, err := client.RuleTypes().ListRuleTypes(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(resp.RuleTypes)
}

func (t *Tools) getRuleType(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ruleTypeID := req.GetString("rule_type_id", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	// Validate parameters
	if errMsg := ValidateLookupParams(ruleTypeID, name, "rule_type_id", "name", map[string]string{
		"project_id": projectID,
	}); errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	var ruleType *minderv1.RuleType

	if ruleTypeID != "" {
		// Lookup by ID
		resp, err := client.RuleTypes().GetRuleTypeById(ctx, &minderv1.GetRuleTypeByIdRequest{
			Id: ruleTypeID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		ruleType = resp.RuleType
	} else {
		// Lookup by name
		reqProto := &minderv1.GetRuleTypeByNameRequest{
			Name: name,
		}
		if projectID != "" {
			reqProto.Context = &minderv1.Context{
				Project: &projectID,
			}
		}
		resp, err := client.RuleTypes().GetRuleTypeByName(ctx, reqProto)
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		ruleType = resp.RuleType
	}

	return marshalResult(ruleType)
}
