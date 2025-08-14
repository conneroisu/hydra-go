// Package hydra provides a comprehensive Go client for the Nix Hydra build service API.
//
// This package implements a flat architecture design where all client functionality
// is accessible through a single import, following Go best practices for library ergonomics.
// Users can perform all Hydra operations - authentication, project management, jobset
// control, build monitoring, and searching - through the unified Client interface.
//
// Example usage:
//
//	client, err := hydra.NewDefaultClient()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	projects, err := client.ListProjects(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// For custom configurations:
//
//	client, err := hydra.NewClient(&hydra.Config{
//		BaseURL: "https://hydra.example.com",
//		Timeout: 30 * time.Second,
//	})
package hydra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/conneroisu/hydra-go/hydra/auth"
	"github.com/conneroisu/hydra-go/hydra/builds"
	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/jobsets"
	"github.com/conneroisu/hydra-go/hydra/projects"
	"github.com/conneroisu/hydra-go/hydra/search"
)

// Client is the main Hydra API client that provides access to all Hydra
// operations.
//
// It encapsulates HTTP communication, authentication state, and
// service-specific functionality through a unified interface.
type Client struct {
	// Base client for HTTP operations
	client *client.Client

	// Services
	auth     *auth.Service
	projects *projects.Service
	jobsets  *jobsets.Service
	builds   *builds.Service
	search   *search.Service
}

// Config represents client configuration options.
//
// All fields are optional; sensible defaults will be applied for omitted
// values.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	Timeout    time.Duration
}

// NewClient creates a new Hydra client with the given configuration.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("configuration is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://hydra.nixos.org"
	}

	// Build client options
	var opts []client.Option
	if cfg.HTTPClient != nil {
		opts = append(opts, client.WithHTTPClient(cfg.HTTPClient))
	}
	if cfg.UserAgent != "" {
		opts = append(opts, client.WithUserAgent(cfg.UserAgent))
	}
	if cfg.Timeout > 0 {
		opts = append(opts, client.WithTimeout(cfg.Timeout))
	}

	// Create base client
	baseClient, err := client.NewClient(cfg.BaseURL, opts...)
	if err != nil {
		return nil, err
	}

	// Create services
	c := &Client{
		client:   baseClient,
		auth:     auth.NewService(baseClient),
		projects: projects.NewService(baseClient),
		jobsets:  jobsets.NewService(baseClient),
		builds:   builds.NewService(baseClient),
		search:   search.NewService(baseClient),
	}

	return c, nil
}

// NewDefaultClient creates a client with default settings for hydra.nixos.org.
func NewDefaultClient() (*Client, error) {
	return NewClient(&Config{
		BaseURL: "https://hydra.nixos.org",
	})
}

// NewClientWithURL creates a client for a specific Hydra instance.
func NewClientWithURL(baseURL string) (*Client, error) {
	return NewClient(&Config{
		BaseURL: baseURL,
	})
}

// BaseURL returns the base URL of the Hydra instance.
func (c *Client) BaseURL() string {
	return c.client.BaseURL()
}

// SetBaseURL updates the base URL of the client.
func (c *Client) SetBaseURL(baseURL string) error {
	return c.client.SetBaseURL(baseURL)
}

// ===== Authentication Methods =====

// Login authenticates with the Hydra instance.
func (c *Client) Login(ctx context.Context, username, password string) (*User, error) {
	return c.auth.Login(ctx, username, password)
}

// Logout clears the current session.
func (c *Client) Logout() {
	c.auth.Logout()
}

// IsAuthenticated returns true if the client is authenticated.
func (c *Client) IsAuthenticated() bool {
	return c.auth.IsAuthenticated()
}

// GetCurrentUser returns the username of the authenticated user.
func (c *Client) GetCurrentUser() string {
	return c.auth.GetCurrentUser()
}

// ===== Project Methods =====

// ListProjects retrieves all projects.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	return c.projects.List(ctx)
}

// GetProject retrieves a project by ID.
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	return c.projects.Get(ctx, projectID)
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, projectID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	return c.projects.Create(ctx, projectID, req)
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, projectID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	return c.projects.Update(ctx, projectID, req)
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	return c.projects.Delete(ctx, projectID)
}

// CreateProjectWithOptions creates a project using options builder pattern.
func (c *Client) CreateProjectWithOptions(ctx context.Context, projectID string, opts *CreateProjectOptions) (*ProjectResponse, error) {
	return c.projects.CreateWithOptions(ctx, projectID, opts)
}

// ===== Jobset Methods =====

// ListJobsets retrieves all jobsets for a project.
func (c *Client) ListJobsets(ctx context.Context, projectID string) (JobsetOverview, error) {
	return c.jobsets.List(ctx, projectID)
}

// GetJobset retrieves a jobset by project and jobset ID.
func (c *Client) GetJobset(ctx context.Context, projectID, jobsetID string) (*Jobset, error) {
	return c.jobsets.Get(ctx, projectID, jobsetID)
}

// CreateJobset creates a new jobset.
func (c *Client) CreateJobset(ctx context.Context, projectID, jobsetID string, jobset *Jobset) (*ProjectResponse, error) {
	return c.jobsets.Create(ctx, projectID, jobsetID, jobset)
}

// UpdateJobset updates an existing jobset.
func (c *Client) UpdateJobset(ctx context.Context, projectID, jobsetID string, jobset *Jobset) (*ProjectResponse, error) {
	return c.jobsets.Update(ctx, projectID, jobsetID, jobset)
}

// DeleteJobset deletes a jobset.
func (c *Client) DeleteJobset(ctx context.Context, projectID, jobsetID string) error {
	return c.jobsets.Delete(ctx, projectID, jobsetID)
}

// GetJobsetEvaluations retrieves all evaluations for a jobset.
func (c *Client) GetJobsetEvaluations(ctx context.Context, projectID, jobsetID string) (*Evaluations, error) {
	return c.jobsets.GetEvaluations(ctx, projectID, jobsetID)
}

// TriggerJobsets triggers evaluation of multiple jobsets.
func (c *Client) TriggerJobsets(ctx context.Context, jobsets ...string) (*PushResponse, error) {
	return c.jobsets.Trigger(ctx, jobsets...)
}

// TriggerJobset triggers evaluation of a single jobset.
func (c *Client) TriggerJobset(ctx context.Context, projectID, jobsetID string) (*PushResponse, error) {
	return c.jobsets.TriggerSingle(ctx, projectID, jobsetID)
}

// GetJobsetShieldData generates data for a shields.io badge.
func (c *Client) GetJobsetShieldData(ctx context.Context, projectID, jobsetID, jobID string) (*ShieldData, error) {
	return c.jobsets.GetShieldData(ctx, projectID, jobsetID, jobID)
}

// CreateJobsetWithOptions creates a jobset using options builder pattern.
func (c *Client) CreateJobsetWithOptions(ctx context.Context, projectID, jobsetID string, opts *CreateJobsetOptions) (*ProjectResponse, error) {
	return c.jobsets.CreateWithOptions(ctx, projectID, jobsetID, opts)
}

// ===== Build Methods =====

// GetBuild retrieves a build by ID.
func (c *Client) GetBuild(ctx context.Context, buildID int) (*Build, error) {
	return c.builds.Get(ctx, buildID)
}

// GetBuildConstituents retrieves a build's constituent jobs.
func (c *Client) GetBuildConstituents(ctx context.Context, buildID int) ([]Build, error) {
	return c.builds.GetConstituents(ctx, buildID)
}

// GetEvaluation retrieves an evaluation by ID.
func (c *Client) GetEvaluation(ctx context.Context, evalID int) (*JobsetEval, error) {
	return c.builds.GetEvaluation(ctx, evalID)
}

// GetEvaluationBuilds retrieves all builds for an evaluation.
func (c *Client) GetEvaluationBuilds(ctx context.Context, evalID int) (JobsetEvalBuilds, error) {
	return c.builds.GetEvaluationBuilds(ctx, evalID)
}

// GetBuildInfo retrieves comprehensive information about a build.
func (c *Client) GetBuildInfo(ctx context.Context, buildID int) (*BuildInfo, error) {
	return c.builds.GetBuildInfo(ctx, buildID)
}

// WaitForBuild waits for a build to complete.
func (c *Client) WaitForBuild(ctx context.Context, buildID int, pollInterval int) (*Build, error) {
	return c.builds.WaitForBuild(ctx, buildID, pollInterval)
}

// ===== Search Methods =====

// Search performs a search query across all resource types.
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	return c.search.Search(ctx, query)
}

// SearchWithOptions performs a search with filtering options.
func (c *Client) SearchWithOptions(ctx context.Context, opts *SearchOptions) (*SearchResult, error) {
	return c.search.SearchWithOptions(ctx, opts)
}

// SearchProjects searches only for projects.
func (c *Client) SearchProjects(ctx context.Context, query string) ([]Project, error) {
	return c.search.SearchProjects(ctx, query)
}

// SearchJobsets searches only for jobsets.
func (c *Client) SearchJobsets(ctx context.Context, query string) ([]Jobset, error) {
	return c.search.SearchJobsets(ctx, query)
}

// SearchBuilds searches only for builds.
func (c *Client) SearchBuilds(ctx context.Context, query string) ([]Build, error) {
	return c.search.SearchBuilds(ctx, query)
}

// SearchDerivations searches only for derivations.
func (c *Client) SearchDerivations(ctx context.Context, query string) ([]Build, error) {
	return c.search.SearchDerivations(ctx, query)
}

// ===== Convenience Methods (backward compatibility) =====

// SearchAll is an alias for Search for backward compatibility.
func (c *Client) SearchAll(ctx context.Context, query string) (*SearchResult, error) {
	return c.Search(ctx, query)
}

// TriggerEvaluation is an alias for TriggerJobset for backward compatibility.
func (c *Client) TriggerEvaluation(ctx context.Context, projectID, jobsetID string) (*PushResponse, error) {
	return c.TriggerJobset(ctx, projectID, jobsetID)
}

// ===== Helper Functions =====

// NewCreateProjectOptions creates project options with required fields.
func NewCreateProjectOptions(name, owner string) *CreateProjectOptions {
	return projects.NewCreateOptions(name, owner)
}

// NewCreateJobsetOptions creates jobset options with defaults.
func NewCreateJobsetOptions(name, project string) *CreateJobsetOptions {
	return jobsets.NewJobsetOptions(name, project)
}

// NewSearchOptions creates search options with defaults.
func NewSearchOptions(query string) *SearchOptions {
	return search.NewSearchOptions(query)
}

// ParseBuildID parses a build ID from string.
func ParseBuildID(s string) (int, error) {
	return builds.ParseBuildID(s)
}

// GetBuildURL returns the web URL for a build.
func GetBuildURL(baseURL string, buildID int) string {
	return builds.GetBuildURL(baseURL, buildID)
}

// GetEvaluationURL returns the web URL for an evaluation.
func GetEvaluationURL(baseURL string, evalID int) string {
	return builds.GetEvaluationURL(baseURL, evalID)
}

// FilterBuilds filters builds from evaluation results.
func FilterBuilds(buildsSlice []Build, filter *BuildFilter) []Build {
	return builds.FilterBuilds(buildsSlice, filter)
}

// CalculateStatistics calculates statistics for a set of builds.
func CalculateStatistics(buildsSlice []Build) *BuildStatistics {
	return builds.CalculateStatistics(buildsSlice)
}

// GetSearchSummary returns a summary of search results.
func GetSearchSummary(query string, result *SearchResult) *SearchSummary {
	return search.GetSearchSummary(query, result)
}

// FlattenSearchResults converts search results to a flat list of items.
func FlattenSearchResults(baseURL string, result *SearchResult) []SearchItem {
	return search.FlattenSearchResults(baseURL, result)
}

// ===== QuickStart Helper Methods =====

// QuickStart provides a quick way to get started with common operations.
type QuickStart struct {
	client *Client
}

// Quick returns a QuickStart helper for common operations.
func (c *Client) Quick() *QuickStart {
	return &QuickStart{client: c}
}

// GetProjectWithJobsets gets a project and all its jobsets.
func (q *QuickStart) GetProjectWithJobsets(ctx context.Context, projectID string) (*Project, JobsetOverview, error) {
	project, err := q.client.GetProject(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	jobsets, err := q.client.ListJobsets(ctx, projectID)
	if err != nil {
		return project, nil, err
	}

	return project, jobsets, nil
}

// GetLatestBuildForJob gets the latest build for a specific job.
func (q *QuickStart) GetLatestBuildForJob(ctx context.Context, projectID, jobsetID, jobName string) (*Build, error) {
	// Get evaluations for the jobset
	evals, err := q.client.GetJobsetEvaluations(ctx, projectID, jobsetID)
	if err != nil {
		return nil, err
	}

	// Look for the job in recent evaluations
	for _, evalMap := range evals.Evals {
		for _, eval := range evalMap {
			if eval == nil {
				continue
			}
			// Get builds for this evaluation
			builds, err := q.client.GetEvaluationBuilds(ctx, eval.ID)
			if err != nil {
				continue
			}

			// Look for the specific job
			for _, buildMap := range builds {
				for _, build := range buildMap {
					if build != nil && build.Job == jobName {
						return build, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no build found for job %s", jobName)
}
