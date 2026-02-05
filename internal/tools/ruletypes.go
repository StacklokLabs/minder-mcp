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

	// Use multi-project aggregation when no project_id specified
	ruleTypes, err := forEachProject(ctx, client, projectID, func(ctx context.Context, projID string) ([]*minderv1.RuleType, error) {
		resp, err := client.RuleTypes().ListRuleTypes(ctx, &minderv1.ListRuleTypesRequest{
			Context: &minderv1.Context{
				Project: &projID,
			},
		})
		if err != nil {
			return nil, err
		}
		return resp.RuleTypes, nil
	})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(ruleTypes)
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
		// Lookup by ID - no project context needed
		resp, err := client.RuleTypes().GetRuleTypeById(ctx, &minderv1.GetRuleTypeByIdRequest{
			Id: ruleTypeID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		ruleType = resp.RuleType
	} else {
		// Lookup by name - search across projects if none specified
		ruleType, err = findInProjects(ctx, client, projectID, func(ctx context.Context, projID string) (*minderv1.RuleType, error) {
			resp, err := client.RuleTypes().GetRuleTypeByName(ctx, &minderv1.GetRuleTypeByNameRequest{
				Name: name,
				Context: &minderv1.Context{
					Project: &projID,
				},
			})
			if err != nil {
				return nil, err
			}
			return resp.RuleType, nil
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
	}

	return marshalResult(ruleType)
}
