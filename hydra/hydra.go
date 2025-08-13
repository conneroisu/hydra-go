// Package hydra provides a Go client for the Nix Hydra build service API.
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
	"github.com/conneroisu/hydra-go/hydra/models"
	"github.com/conneroisu/hydra-go/hydra/projects"
	"github.com/conneroisu/hydra-go/hydra/search"
)

// Client is the main Hydra API client.
type Client struct {
	// Base client for HTTP operations
	client *client.Client

	// Services
	Auth     *auth.Service
	Projects *projects.Service
	Jobsets  *jobsets.Service
	Builds   *builds.Service
	Search   *search.Service
}

// Config represents client configuration.
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
		Auth:     auth.NewService(baseClient),
		Projects: projects.NewService(baseClient),
		Jobsets:  jobsets.NewService(baseClient),
		Builds:   builds.NewService(baseClient),
		Search:   search.NewService(baseClient),
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

// Login authenticates with the Hydra instance.
func (c *Client) Login(ctx context.Context, username, password string) (*models.User, error) {
	return c.Auth.Login(ctx, username, password)
}

// Logout clears the current session.
func (c *Client) Logout() {
	c.Auth.Logout()
}

// IsAuthenticated returns true if the client is authenticated.
func (c *Client) IsAuthenticated() bool {
	return c.Auth.IsAuthenticated()
}

// BaseURL returns the base URL of the Hydra instance.
func (c *Client) BaseURL() string {
	return c.client.BaseURL()
}

// SetBaseURL updates the base URL of the client.
func (c *Client) SetBaseURL(baseURL string) error {
	return c.client.SetBaseURL(baseURL)
}

// GetProject is a convenience method to get a project.
func (c *Client) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	return c.Projects.Get(ctx, projectID)
}

// ListProjects is a convenience method to list all projects.
func (c *Client) ListProjects(ctx context.Context) ([]models.Project, error) {
	return c.Projects.List(ctx)
}

// GetJobset is a convenience method to get a jobset.
func (c *Client) GetJobset(ctx context.Context, projectID, jobsetID string) (*models.Jobset, error) {
	return c.Jobsets.Get(ctx, projectID, jobsetID)
}

// GetBuild is a convenience method to get a build.
func (c *Client) GetBuild(ctx context.Context, buildID int) (*models.Build, error) {
	return c.Builds.Get(ctx, buildID)
}

// SearchAll is a convenience method to search across all resource types.
func (c *Client) SearchAll(ctx context.Context, query string) (*models.SearchResult, error) {
	return c.Search.Search(ctx, query)
}

// TriggerEvaluation is a convenience method to trigger jobset evaluation.
func (c *Client) TriggerEvaluation(ctx context.Context, projectID, jobsetID string) (*models.PushResponse, error) {
	return c.Jobsets.TriggerSingle(ctx, projectID, jobsetID)
}

// QuickStart provides a quick way to get started with common operations.
type QuickStart struct {
	client *Client
}

// Quick returns a QuickStart helper for common operations.
func (c *Client) Quick() *QuickStart {
	return &QuickStart{client: c}
}

// GetProjectWithJobsets gets a project and all its jobsets.
func (q *QuickStart) GetProjectWithJobsets(ctx context.Context, projectID string) (*models.Project, models.JobsetOverview, error) {
	project, err := q.client.Projects.Get(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	jobsets, err := q.client.Jobsets.List(ctx, projectID)
	if err != nil {
		return project, nil, err
	}

	return project, jobsets, nil
}

// GetLatestBuildForJob gets the latest build for a specific job.
func (q *QuickStart) GetLatestBuildForJob(ctx context.Context, projectID, jobsetID, jobName string) (*models.Build, error) {
	// Get evaluations for the jobset
	evals, err := q.client.Jobsets.GetEvaluations(ctx, projectID, jobsetID)
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
			builds, err := q.client.Builds.GetEvaluationBuilds(ctx, eval.ID)
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

// WaitForJobsetEvaluation triggers and waits for a jobset evaluation to complete.
func (q *QuickStart) WaitForJobsetEvaluation(ctx context.Context, projectID, jobsetID string, timeout time.Duration) (*models.JobsetEval, error) {
	// Trigger evaluation
	_, err := q.client.Jobsets.TriggerSingle(ctx, projectID, jobsetID)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger evaluation: %w", err)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for completion
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastEvalID int
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for evaluation")
		case <-ticker.C:
			// Get latest evaluation
			evals, err := q.client.Jobsets.GetEvaluations(ctx, projectID, jobsetID)
			if err != nil {
				continue
			}

			if len(evals.Evals) > 0 {
				for _, evalMap := range evals.Evals {
					for _, eval := range evalMap {
						if eval != nil && eval.ID > lastEvalID {
							// Check if evaluation is complete
							if len(eval.Builds) > 0 {
								return eval, nil
							}
							lastEvalID = eval.ID
						}
					}
				}
			}
		}
	}
}
