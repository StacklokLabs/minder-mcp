package tools

import (
	"context"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

// listAllProjects returns all accessible projects for the current user.
func listAllProjects(ctx context.Context, client MinderClient) ([]*minderv1.Project, error) {
	resp, err := client.Projects().ListProjects(ctx, &minderv1.ListProjectsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

// forEachProject executes a function for each project, collecting results.
// If projectID is provided, only that project is used.
// If projectID is empty, all accessible projects are iterated.
func forEachProject[T any](
	ctx context.Context,
	client MinderClient,
	projectID string,
	fn func(ctx context.Context, projectID string) ([]T, error),
) ([]T, error) {
	if projectID != "" {
		// Single project specified
		return fn(ctx, projectID)
	}

	// Get all projects
	projects, err := listAllProjects(ctx, client)
	if err != nil {
		return nil, err
	}

	// Aggregate results from all projects
	var allResults []T
	for _, project := range projects {
		results, err := fn(ctx, project.ProjectId)
		if err != nil {
			// Log but continue with other projects
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// findInProjects searches for an item across all projects using a finder function.
// Returns the first match found. If projectID is provided, only searches that project.
func findInProjects[T any](
	ctx context.Context,
	client MinderClient,
	projectID string,
	fn func(ctx context.Context, projectID string) (T, error),
) (T, error) {
	var zero T

	if projectID != "" {
		// Single project specified
		return fn(ctx, projectID)
	}

	// Get all projects
	projects, err := listAllProjects(ctx, client)
	if err != nil {
		return zero, err
	}

	// Search each project
	var lastErr error
	for _, project := range projects {
		result, err := fn(ctx, project.ProjectId)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	// Return last error if nothing found
	return zero, lastErr
}
