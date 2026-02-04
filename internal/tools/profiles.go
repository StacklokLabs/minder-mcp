package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listProfiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")
	labelFilter := req.GetString("label_filter", "")

	reqProto := &minderv1.ListProfilesRequest{}
	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}
	if labelFilter != "" {
		reqProto.LabelFilter = labelFilter
	}

	resp, err := client.Profiles().ListProfiles(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(resp.Profiles)
}

func (t *Tools) getProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profileID := req.GetString("profile_id", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	// Validate parameters
	if errMsg := ValidateLookupParams(profileID, name, "profile_id", "name", map[string]string{
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

	var profile *minderv1.Profile

	if profileID != "" {
		// Lookup by ID
		resp, err := client.Profiles().GetProfileById(ctx, &minderv1.GetProfileByIdRequest{
			Id: profileID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		profile = resp.Profile
	} else {
		// Lookup by name
		reqProto := &minderv1.GetProfileByNameRequest{
			Name: name,
		}
		if projectID != "" {
			reqProto.Context = &minderv1.Context{
				Project: &projectID,
			}
		}
		resp, err := client.Profiles().GetProfileByName(ctx, reqProto)
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		profile = resp.Profile
	}

	return marshalResult(profile)
}

func (t *Tools) getProfileStatusByName(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	reqProto := &minderv1.GetProfileStatusByNameRequest{
		Name: name,
	}
	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}

	resp, err := client.Profiles().GetProfileStatusByName(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(resp)
}
