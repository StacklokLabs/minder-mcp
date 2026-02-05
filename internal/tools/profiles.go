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

	// Use multi-project aggregation when no project_id specified
	profiles, err := forEachProject(ctx, client, projectID, func(ctx context.Context, projID string) ([]*minderv1.Profile, error) {
		reqProto := &minderv1.ListProfilesRequest{
			Context: &minderv1.Context{
				Project: &projID,
			},
		}
		if labelFilter != "" {
			reqProto.LabelFilter = labelFilter
		}
		resp, err := client.Profiles().ListProfiles(ctx, reqProto)
		if err != nil {
			return nil, err
		}
		return resp.Profiles, nil
	})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(profiles)
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
		// Lookup by ID - no project context needed
		resp, err := client.Profiles().GetProfileById(ctx, &minderv1.GetProfileByIdRequest{
			Id: profileID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		profile = resp.Profile
	} else {
		// Lookup by name - search across projects if none specified
		profile, err = findInProjects(ctx, client, projectID, func(ctx context.Context, projID string) (*minderv1.Profile, error) {
			resp, err := client.Profiles().GetProfileByName(ctx, &minderv1.GetProfileByNameRequest{
				Name: name,
				Context: &minderv1.Context{
					Project: &projID,
				},
			})
			if err != nil {
				return nil, err
			}
			return resp.Profile, nil
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
	}

	return marshalResult(profile)
}

func (t *Tools) getProfileStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	if profileID != "" {
		// Lookup by ID - no project context needed
		resp, err := client.Profiles().GetProfileStatusById(ctx, &minderv1.GetProfileStatusByIdRequest{
			Id:  profileID,
			All: true, // Always request detailed per-rule evaluation results
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		return marshalResult(resp)
	}

	// Lookup by name - search across projects if none specified
	resp, err := findInProjects(
		ctx, client, projectID,
		func(ctx context.Context, projID string) (*minderv1.GetProfileStatusByNameResponse, error) {
			return client.Profiles().GetProfileStatusByName(ctx, &minderv1.GetProfileStatusByNameRequest{
				Name: name,
				All:  true,
				Context: &minderv1.Context{
					Project: &projID,
				},
			})
		})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(resp)
}
