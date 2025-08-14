package jobsets

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/models"
)

// Service handles jobset operations.
type Service struct {
	client *client.Client
}

// NewService creates a new jobsets service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// List retrieves all jobsets for a project.
func (s *Service) List(ctx context.Context, projectID string) (models.JobsetOverview, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	path := "/api/jobsets?project=" + url.QueryEscape(projectID)
	var overview models.JobsetOverview
	if err := s.client.DoRequest(ctx, "GET", path, nil, &overview); err != nil {
		return nil, fmt.Errorf("failed to list jobsets for project %s: %w", projectID, err)
	}

	return overview, nil
}

// Get retrieves a jobset by project and jobset ID.
func (s *Service) Get(ctx context.Context, projectID, jobsetID string) (*models.Jobset, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if jobsetID == "" {
		return nil, errors.New("jobset ID is required")
	}

	path := fmt.Sprintf("/jobset/%s/%s", url.PathEscape(projectID), url.PathEscape(jobsetID))
	var jobset models.Jobset
	if err := s.client.DoRequest(ctx, "GET", path, nil, &jobset); err != nil {
		return nil, fmt.Errorf("failed to get jobset %s/%s: %w", projectID, jobsetID, err)
	}

	return &jobset, nil
}

// Create creates a new jobset.
func (s *Service) Create(ctx context.Context, projectID, jobsetID string, jobset *models.Jobset) (*models.ProjectResponse, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if jobsetID == "" {
		return nil, errors.New("jobset ID is required")
	}
	if jobset == nil {
		return nil, errors.New("jobset is required")
	}

	path := fmt.Sprintf("/jobset/%s/%s", url.PathEscape(projectID), url.PathEscape(jobsetID))
	var response models.ProjectResponse
	if err := s.client.DoRequest(ctx, "PUT", path, jobset, &response); err != nil {
		return nil, fmt.Errorf("failed to create jobset %s/%s: %w", projectID, jobsetID, err)
	}

	return &response, nil
}

// Update updates an existing jobset.
func (s *Service) Update(ctx context.Context, projectID, jobsetID string, jobset *models.Jobset) (*models.ProjectResponse, error) {
	// Same endpoint as Create but will return 200 instead of 201 if jobset exists
	return s.Create(ctx, projectID, jobsetID, jobset)
}

// Delete deletes a jobset.
func (s *Service) Delete(ctx context.Context, projectID, jobsetID string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}
	if jobsetID == "" {
		return errors.New("jobset ID is required")
	}

	path := fmt.Sprintf("/jobset/%s/%s", url.PathEscape(projectID), url.PathEscape(jobsetID))
	var response models.ProjectResponse
	if err := s.client.DoRequest(ctx, "DELETE", path, nil, &response); err != nil {
		return fmt.Errorf("failed to delete jobset %s/%s: %w", projectID, jobsetID, err)
	}

	return nil
}

// GetEvaluations retrieves all evaluations for a jobset.
func (s *Service) GetEvaluations(ctx context.Context, projectID, jobsetID string) (*models.Evaluations, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if jobsetID == "" {
		return nil, errors.New("jobset ID is required")
	}

	path := fmt.Sprintf("/jobset/%s/%s/evals", url.PathEscape(projectID), url.PathEscape(jobsetID))
	var evals models.Evaluations
	if err := s.client.DoRequest(ctx, "GET", path, nil, &evals); err != nil {
		return nil, fmt.Errorf("failed to get evaluations for jobset %s/%s: %w", projectID, jobsetID, err)
	}

	return &evals, nil
}

// Trigger triggers evaluation of jobsets.
func (s *Service) Trigger(ctx context.Context, jobsets ...string) (*models.PushResponse, error) {
	if len(jobsets) == 0 {
		return nil, errors.New("at least one jobset is required")
	}

	// Validate jobset format (should be "project:jobset")
	for _, js := range jobsets {
		if !isValidJobsetFormat(js) {
			return nil, fmt.Errorf("invalid jobset format '%s', expected 'project:jobset'", js)
		}
	}

	// Join jobsets with comma
	jobsetsStr := ""
	for i, js := range jobsets {
		if i > 0 {
			jobsetsStr += ","
		}
		jobsetsStr += js
	}

	path := "/api/push?jobsets=" + url.QueryEscape(jobsetsStr)
	var response models.PushResponse
	if err := s.client.DoRequest(ctx, "POST", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to trigger jobsets: %w", err)
	}

	return &response, nil
}

// TriggerSingle triggers evaluation of a single jobset.
func (s *Service) TriggerSingle(ctx context.Context, projectID, jobsetID string) (*models.PushResponse, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if jobsetID == "" {
		return nil, errors.New("jobset ID is required")
	}

	jobsetStr := fmt.Sprintf("%s:%s", projectID, jobsetID)

	return s.Trigger(ctx, jobsetStr)
}

// GetShieldData generates data for a shields.io badge.
func (s *Service) GetShieldData(ctx context.Context, projectID, jobsetID, jobID string) (*models.ShieldData, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}
	if jobsetID == "" {
		return nil, errors.New("jobset ID is required")
	}
	if jobID == "" {
		return nil, errors.New("job ID is required")
	}

	path := fmt.Sprintf("/job/%s/%s/%s/shield",
		url.PathEscape(projectID),
		url.PathEscape(jobsetID),
		url.PathEscape(jobID))

	var shield models.ShieldData
	if err := s.client.DoRequest(ctx, "GET", path, nil, &shield); err != nil {
		return nil, fmt.Errorf("failed to get shield data for %s/%s/%s: %w", projectID, jobsetID, jobID, err)
	}

	return &shield, nil
}

// isValidJobsetFormat checks if the jobset string is in "project:jobset" format.
func isValidJobsetFormat(jobset string) bool {
	parts := splitJobsetString(jobset)

	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// splitJobsetString splits a "project:jobset" string.
func splitJobsetString(jobset string) []string {
	// Only split on the first colon to handle project/jobset names with colons
	idx := indexOf(jobset, ':')
	if idx == -1 {
		return []string{jobset}
	}

	return []string{jobset[:idx], jobset[idx+1:]}
}

func indexOf(s string, char rune) int {
	for i, c := range s {
		if c == char {
			return i
		}
	}

	return -1
}

// JobsetOptions represents options for creating/updating a jobset.
type JobsetOptions struct {
	Name                    string
	Project                 string
	Description             string
	NixExprInput            string
	NixExprPath             string
	Enabled                 models.JobsetState
	EnableEmail             bool
	EnableDynamicRunCommand bool
	Visible                 bool
	EmailOverride           string
	KeepNr                  int
	CheckInterval           int
	SchedulingShares        int
	Flake                   string
	Inputs                  map[string]models.JobsetInput
}

// NewJobsetOptions creates new jobset options with defaults.
func NewJobsetOptions(name, project string) *JobsetOptions {
	return &JobsetOptions{
		Name:             name,
		Project:          project,
		Enabled:          models.JobsetStateEnabled,
		Visible:          true,
		KeepNr:           3,
		CheckInterval:    300,
		SchedulingShares: 100,
		Inputs:           make(map[string]models.JobsetInput),
	}
}

// WithDescription sets the description.
func (o *JobsetOptions) WithDescription(description string) *JobsetOptions {
	o.Description = description

	return o
}

// WithNixExpression sets the nix expression input and path.
func (o *JobsetOptions) WithNixExpression(input, path string) *JobsetOptions {
	o.NixExprInput = input
	o.NixExprPath = path

	return o
}

// WithFlake sets the flake URI.
func (o *JobsetOptions) WithFlake(flake string) *JobsetOptions {
	o.Flake = flake

	return o
}

// WithState sets the jobset state.
func (o *JobsetOptions) WithState(state models.JobsetState) *JobsetOptions {
	o.Enabled = state

	return o
}

// WithEmail configures email settings.
func (o *JobsetOptions) WithEmail(enable bool, override string) *JobsetOptions {
	o.EnableEmail = enable
	o.EmailOverride = override

	return o
}

// WithScheduling sets scheduling parameters.
func (o *JobsetOptions) WithScheduling(checkInterval, schedulingShares int) *JobsetOptions {
	o.CheckInterval = checkInterval
	o.SchedulingShares = schedulingShares

	return o
}

// WithKeepNr sets the number of evaluations to keep.
func (o *JobsetOptions) WithKeepNr(keepNr int) *JobsetOptions {
	o.KeepNr = keepNr

	return o
}

// AddInput adds an input to the jobset.
func (o *JobsetOptions) AddInput(name, inputType, value string, emailResponsible bool) *JobsetOptions {
	o.Inputs[name] = models.JobsetInput{
		Name:             name,
		Type:             inputType,
		Value:            value,
		EmailResponsible: emailResponsible,
	}

	return o
}

// Build converts options to a Jobset model.
func (o *JobsetOptions) Build() *models.Jobset {
	var description *string
	if o.Description != "" {
		description = &o.Description
	}

	var nixExprInput *string
	if o.NixExprInput != "" {
		nixExprInput = &o.NixExprInput
	}

	var nixExprPath *string
	if o.NixExprPath != "" {
		nixExprPath = &o.NixExprPath
	}

	var flake *string
	if o.Flake != "" {
		flake = &o.Flake
	}

	return &models.Jobset{
		Name:                    o.Name,
		Project:                 o.Project,
		Description:             description,
		NixExprInput:            nixExprInput,
		NixExprPath:             nixExprPath,
		Enabled:                 int(o.Enabled),
		EnableEmail:             o.EnableEmail,
		EnableDynamicRunCommand: o.EnableDynamicRunCommand,
		Visible:                 o.Visible,
		EmailOverride:           o.EmailOverride,
		KeepNr:                  o.KeepNr,
		CheckInterval:           o.CheckInterval,
		SchedulingShares:        o.SchedulingShares,
		Flake:                   flake,
		Inputs:                  o.Inputs,
	}
}

// CreateWithOptions creates a jobset using options.
func (s *Service) CreateWithOptions(ctx context.Context, projectID, jobsetID string, opts *JobsetOptions) (*models.ProjectResponse, error) {
	if opts == nil {
		return nil, errors.New("options are required")
	}

	return s.Create(ctx, projectID, jobsetID, opts.Build())
}
