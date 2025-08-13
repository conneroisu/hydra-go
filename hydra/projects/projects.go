// Package projects provides project operations for Hydra.
package projects

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/models"
)

// Service handles project operations.
type Service struct {
	client *client.Client
}

// NewService creates a new projects service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// List retrieves all projects.
func (s *Service) List(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project
	if err := s.client.DoRequest(ctx, "GET", "/", nil, &projects); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

// Get retrieves a project by ID.
func (s *Service) Get(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	path := "/project/" + url.PathEscape(projectID)
	var project models.Project
	if err := s.client.DoRequest(ctx, "GET", path, nil, &project); err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectID, err)
	}

	return &project, nil
}

// Create creates a new project.
func (s *Service) Create(ctx context.Context, projectID string, req *models.CreateProjectRequest) (*models.ProjectResponse, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if req == nil {
		return nil, errors.New("project request is required")
	}
	if req.Name == "" {
		return nil, errors.New("project name is required")
	}
	if req.Owner == "" {
		return nil, errors.New("project owner is required")
	}

	path := "/project/" + url.PathEscape(projectID)
	var response models.ProjectResponse
	if err := s.client.DoRequest(ctx, "PUT", path, req, &response); err != nil {
		return nil, fmt.Errorf("failed to create project %s: %w", projectID, err)
	}

	return &response, nil
}

// Update updates an existing project.
func (s *Service) Update(ctx context.Context, projectID string, req *models.CreateProjectRequest) (*models.ProjectResponse, error) {
	// Same endpoint as Create but will return 200 instead of 201 if project exists
	return s.Create(ctx, projectID, req)
}

// Delete deletes a project.
func (s *Service) Delete(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}

	path := "/project/" + url.PathEscape(projectID)
	var response models.ProjectResponse
	if err := s.client.DoRequest(ctx, "DELETE", path, nil, &response); err != nil {
		return fmt.Errorf("failed to delete project %s: %w", projectID, err)
	}

	return nil
}

// CreateOptions represents options for creating a project with builder pattern.
type CreateOptions struct {
	Name                    string
	DisplayName             string
	Description             string
	Homepage                string
	Owner                   string
	Enabled                 bool
	EnableDynamicRunCommand bool
	Visible                 bool
	Declarative             *models.DeclarativeInput
}

// NewCreateOptions creates a new CreateOptions with required fields.
func NewCreateOptions(name, owner string) *CreateOptions {
	return &CreateOptions{
		Name:    name,
		Owner:   owner,
		Enabled: true,
		Visible: true,
	}
}

// WithDisplayName sets the display name.
func (o *CreateOptions) WithDisplayName(displayName string) *CreateOptions {
	o.DisplayName = displayName

	return o
}

// WithDescription sets the description.
func (o *CreateOptions) WithDescription(description string) *CreateOptions {
	o.Description = description

	return o
}

// WithHomepage sets the homepage.
func (o *CreateOptions) WithHomepage(homepage string) *CreateOptions {
	o.Homepage = homepage

	return o
}

// WithEnabled sets whether the project is enabled.
func (o *CreateOptions) WithEnabled(enabled bool) *CreateOptions {
	o.Enabled = enabled

	return o
}

// WithDynamicRunCommand enables dynamic run command.
func (o *CreateOptions) WithDynamicRunCommand(enabled bool) *CreateOptions {
	o.EnableDynamicRunCommand = enabled

	return o
}

// WithVisible sets whether the project is visible.
func (o *CreateOptions) WithVisible(visible bool) *CreateOptions {
	o.Visible = visible

	return o
}

// WithDeclarative sets the declarative configuration.
func (o *CreateOptions) WithDeclarative(declarative *models.DeclarativeInput) *CreateOptions {
	o.Declarative = declarative

	return o
}

// Build converts CreateOptions to CreateProjectRequest.
func (o *CreateOptions) Build() *models.CreateProjectRequest {
	return &models.CreateProjectRequest{
		Name:                    o.Name,
		DisplayName:             o.DisplayName,
		Description:             o.Description,
		Homepage:                o.Homepage,
		Owner:                   o.Owner,
		Enabled:                 o.Enabled,
		EnableDynamicRunCommand: o.EnableDynamicRunCommand,
		Visible:                 o.Visible,
		Declarative:             o.Declarative,
	}
}

// CreateWithOptions creates a project using options.
func (s *Service) CreateWithOptions(ctx context.Context, projectID string, opts *CreateOptions) (*models.ProjectResponse, error) {
	if opts == nil {
		return nil, errors.New("options are required")
	}

	return s.Create(ctx, projectID, opts.Build())
}
