package tools

import (
	"context"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/grpc"
)

// mockMinderClient implements MinderClient for testing.
type mockMinderClient struct {
	health       *mockHealthService
	profiles     *mockProfileService
	repositories *mockRepositoryService
	ruleTypes    *mockRuleTypeService
	dataSources  *mockDataSourceService
	providers    *mockProvidersService
	projects     *mockProjectsService
	artifacts    *mockArtifactService
	evalResults  *mockEvalResultsService
}

func newMockClient() *mockMinderClient {
	return &mockMinderClient{
		health:       &mockHealthService{},
		profiles:     &mockProfileService{},
		repositories: &mockRepositoryService{},
		ruleTypes:    &mockRuleTypeService{},
		dataSources:  &mockDataSourceService{},
		providers:    &mockProvidersService{},
		projects:     &mockProjectsService{},
		artifacts:    &mockArtifactService{},
		evalResults:  &mockEvalResultsService{},
	}
}

func (*mockMinderClient) Close() error                                     { return nil }
func (m *mockMinderClient) Health() minderv1.HealthServiceClient           { return m.health }
func (m *mockMinderClient) Profiles() minderv1.ProfileServiceClient        { return m.profiles }
func (m *mockMinderClient) Repositories() minderv1.RepositoryServiceClient { return m.repositories }
func (m *mockMinderClient) RuleTypes() minderv1.RuleTypeServiceClient      { return m.ruleTypes }
func (m *mockMinderClient) DataSources() minderv1.DataSourceServiceClient  { return m.dataSources }
func (m *mockMinderClient) Providers() minderv1.ProvidersServiceClient     { return m.providers }
func (m *mockMinderClient) Projects() minderv1.ProjectsServiceClient       { return m.projects }
func (m *mockMinderClient) Artifacts() minderv1.ArtifactServiceClient      { return m.artifacts }
func (m *mockMinderClient) EvalResults() minderv1.EvalResultsServiceClient { return m.evalResults }

// Mock service implementations

type mockHealthService struct {
	minderv1.HealthServiceClient
	checkResp *minderv1.CheckHealthResponse
	checkErr  error
}

func (m *mockHealthService) CheckHealth(_ context.Context, _ *minderv1.CheckHealthRequest, _ ...grpc.CallOption) (*minderv1.CheckHealthResponse, error) {
	// By default return healthy status if no response/error is set
	if m.checkResp == nil && m.checkErr == nil {
		return &minderv1.CheckHealthResponse{Status: "OK"}, nil
	}
	return m.checkResp, m.checkErr
}

type mockProfileService struct {
	minderv1.ProfileServiceClient
	listResp      *minderv1.ListProfilesResponse
	listErr       error
	getByIDResp   *minderv1.GetProfileByIdResponse
	getByIDErr    error
	getByNameResp *minderv1.GetProfileByNameResponse
	getByNameErr  error
	getStatusResp *minderv1.GetProfileStatusByNameResponse
	getStatusErr  error
}

func (m *mockProfileService) ListProfiles(_ context.Context, _ *minderv1.ListProfilesRequest, _ ...grpc.CallOption) (*minderv1.ListProfilesResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockProfileService) GetProfileById(_ context.Context, _ *minderv1.GetProfileByIdRequest, _ ...grpc.CallOption) (*minderv1.GetProfileByIdResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockProfileService) GetProfileByName(_ context.Context, _ *minderv1.GetProfileByNameRequest, _ ...grpc.CallOption) (*minderv1.GetProfileByNameResponse, error) {
	return m.getByNameResp, m.getByNameErr
}

func (m *mockProfileService) GetProfileStatusByName(_ context.Context, _ *minderv1.GetProfileStatusByNameRequest, _ ...grpc.CallOption) (*minderv1.GetProfileStatusByNameResponse, error) {
	return m.getStatusResp, m.getStatusErr
}

type mockRepositoryService struct {
	minderv1.RepositoryServiceClient
	listResp      *minderv1.ListRepositoriesResponse
	listErr       error
	getByIDResp   *minderv1.GetRepositoryByIdResponse
	getByIDErr    error
	getByNameResp *minderv1.GetRepositoryByNameResponse
	getByNameErr  error
}

func (m *mockRepositoryService) ListRepositories(_ context.Context, _ *minderv1.ListRepositoriesRequest, _ ...grpc.CallOption) (*minderv1.ListRepositoriesResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockRepositoryService) GetRepositoryById(_ context.Context, _ *minderv1.GetRepositoryByIdRequest, _ ...grpc.CallOption) (*minderv1.GetRepositoryByIdResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockRepositoryService) GetRepositoryByName(_ context.Context, _ *minderv1.GetRepositoryByNameRequest, _ ...grpc.CallOption) (*minderv1.GetRepositoryByNameResponse, error) {
	return m.getByNameResp, m.getByNameErr
}

type mockRuleTypeService struct {
	minderv1.RuleTypeServiceClient
	listResp      *minderv1.ListRuleTypesResponse
	listErr       error
	getByIDResp   *minderv1.GetRuleTypeByIdResponse
	getByIDErr    error
	getByNameResp *minderv1.GetRuleTypeByNameResponse
	getByNameErr  error
}

func (m *mockRuleTypeService) ListRuleTypes(_ context.Context, _ *minderv1.ListRuleTypesRequest, _ ...grpc.CallOption) (*minderv1.ListRuleTypesResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockRuleTypeService) GetRuleTypeById(_ context.Context, _ *minderv1.GetRuleTypeByIdRequest, _ ...grpc.CallOption) (*minderv1.GetRuleTypeByIdResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockRuleTypeService) GetRuleTypeByName(_ context.Context, _ *minderv1.GetRuleTypeByNameRequest, _ ...grpc.CallOption) (*minderv1.GetRuleTypeByNameResponse, error) {
	return m.getByNameResp, m.getByNameErr
}

type mockDataSourceService struct {
	minderv1.DataSourceServiceClient
	listResp      *minderv1.ListDataSourcesResponse
	listErr       error
	getByIDResp   *minderv1.GetDataSourceByIdResponse
	getByIDErr    error
	getByNameResp *minderv1.GetDataSourceByNameResponse
	getByNameErr  error
}

func (m *mockDataSourceService) ListDataSources(_ context.Context, _ *minderv1.ListDataSourcesRequest, _ ...grpc.CallOption) (*minderv1.ListDataSourcesResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockDataSourceService) GetDataSourceById(_ context.Context, _ *minderv1.GetDataSourceByIdRequest, _ ...grpc.CallOption) (*minderv1.GetDataSourceByIdResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockDataSourceService) GetDataSourceByName(_ context.Context, _ *minderv1.GetDataSourceByNameRequest, _ ...grpc.CallOption) (*minderv1.GetDataSourceByNameResponse, error) {
	return m.getByNameResp, m.getByNameErr
}

type mockProvidersService struct {
	minderv1.ProvidersServiceClient
	listResp *minderv1.ListProvidersResponse
	listErr  error
	getResp  *minderv1.GetProviderResponse
	getErr   error
}

func (m *mockProvidersService) ListProviders(_ context.Context, _ *minderv1.ListProvidersRequest, _ ...grpc.CallOption) (*minderv1.ListProvidersResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockProvidersService) GetProvider(_ context.Context, _ *minderv1.GetProviderRequest, _ ...grpc.CallOption) (*minderv1.GetProviderResponse, error) {
	return m.getResp, m.getErr
}

type mockProjectsService struct {
	minderv1.ProjectsServiceClient
	listResp      *minderv1.ListProjectsResponse
	listErr       error
	listChildResp *minderv1.ListChildProjectsResponse
	listChildErr  error
}

func (m *mockProjectsService) ListProjects(_ context.Context, _ *minderv1.ListProjectsRequest, _ ...grpc.CallOption) (*minderv1.ListProjectsResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockProjectsService) ListChildProjects(_ context.Context, _ *minderv1.ListChildProjectsRequest, _ ...grpc.CallOption) (*minderv1.ListChildProjectsResponse, error) {
	return m.listChildResp, m.listChildErr
}

type mockArtifactService struct {
	minderv1.ArtifactServiceClient
	listResp      *minderv1.ListArtifactsResponse
	listErr       error
	getByIDResp   *minderv1.GetArtifactByIdResponse
	getByIDErr    error
	getByNameResp *minderv1.GetArtifactByNameResponse
	getByNameErr  error
}

func (m *mockArtifactService) ListArtifacts(_ context.Context, _ *minderv1.ListArtifactsRequest, _ ...grpc.CallOption) (*minderv1.ListArtifactsResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockArtifactService) GetArtifactById(_ context.Context, _ *minderv1.GetArtifactByIdRequest, _ ...grpc.CallOption) (*minderv1.GetArtifactByIdResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockArtifactService) GetArtifactByName(_ context.Context, _ *minderv1.GetArtifactByNameRequest, _ ...grpc.CallOption) (*minderv1.GetArtifactByNameResponse, error) {
	return m.getByNameResp, m.getByNameErr
}

type mockEvalResultsService struct {
	minderv1.EvalResultsServiceClient
	listResp *minderv1.ListEvaluationHistoryResponse
	listErr  error
}

func (m *mockEvalResultsService) ListEvaluationHistory(_ context.Context, _ *minderv1.ListEvaluationHistoryRequest, _ ...grpc.CallOption) (*minderv1.ListEvaluationHistoryResponse, error) {
	return m.listResp, m.listErr
}
