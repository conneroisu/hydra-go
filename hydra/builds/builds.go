package builds

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/models"
)

// Service handles build operations.
type Service struct {
	client *client.Client
}

// NewService creates a new builds service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// Get retrieves a build by ID.
func (s *Service) Get(ctx context.Context, buildID int) (*models.Build, error) {
	if buildID <= 0 {
		return nil, fmt.Errorf("invalid build ID: %d", buildID)
	}

	path := fmt.Sprintf("/build/%d", buildID)
	var build models.Build
	if err := s.client.DoRequest(ctx, "GET", path, nil, &build); err != nil {
		return nil, fmt.Errorf("failed to get build %d: %w", buildID, err)
	}

	return &build, nil
}

// GetConstituents retrieves a build's constituent jobs.
func (s *Service) GetConstituents(ctx context.Context, buildID int) ([]models.Build, error) {
	if buildID <= 0 {
		return nil, fmt.Errorf("invalid build ID: %d", buildID)
	}

	path := fmt.Sprintf("/build/%d/constituents", buildID)
	var builds []models.Build
	if err := s.client.DoRequest(ctx, "GET", path, nil, &builds); err != nil {
		return nil, fmt.Errorf("failed to get constituents for build %d: %w", buildID, err)
	}

	return builds, nil
}

// GetEvaluation retrieves an evaluation by ID.
func (s *Service) GetEvaluation(ctx context.Context, evalID int) (*models.JobsetEval, error) {
	if evalID <= 0 {
		return nil, fmt.Errorf("invalid evaluation ID: %d", evalID)
	}

	path := fmt.Sprintf("/eval/%d", evalID)
	var eval models.JobsetEval
	if err := s.client.DoRequest(ctx, "GET", path, nil, &eval); err != nil {
		return nil, fmt.Errorf("failed to get evaluation %d: %w", evalID, err)
	}

	return &eval, nil
}

// GetEvaluationBuilds retrieves all builds for an evaluation.
func (s *Service) GetEvaluationBuilds(ctx context.Context, evalID int) (models.JobsetEvalBuilds, error) {
	if evalID <= 0 {
		return nil, fmt.Errorf("invalid evaluation ID: %d", evalID)
	}

	path := fmt.Sprintf("/eval/%d/builds", evalID)
	var builds models.JobsetEvalBuilds
	if err := s.client.DoRequest(ctx, "GET", path, nil, &builds); err != nil {
		return nil, fmt.Errorf("failed to get builds for evaluation %d: %w", evalID, err)
	}

	return builds, nil
}

// BuildInfo provides detailed information about a build.
type BuildInfo struct {
	Build        *models.Build
	Constituents []models.Build
	Evaluation   *models.JobsetEval
}

// GetBuildInfo retrieves comprehensive information about a build.
func (s *Service) GetBuildInfo(ctx context.Context, buildID int) (*BuildInfo, error) {
	// Get the build
	build, err := s.Get(ctx, buildID)
	if err != nil {
		return nil, err
	}

	info := &BuildInfo{
		Build: build,
	}

	// Get constituents if this is an aggregate build
	constituents, err := s.GetConstituents(ctx, buildID)
	if err == nil {
		info.Constituents = constituents
	}

	// Get evaluation if available
	if len(build.JobsetEvals) > 0 {
		eval, err := s.GetEvaluation(ctx, build.JobsetEvals[0])
		if err == nil {
			info.Evaluation = eval
		}
	}

	return info, nil
}

// WaitForBuild waits for a build to complete.
func (s *Service) WaitForBuild(ctx context.Context, buildID int, pollInterval int) (*models.Build, error) {
	if buildID <= 0 {
		return nil, fmt.Errorf("invalid build ID: %d", buildID)
	}
	if pollInterval <= 0 {
		pollInterval = 5 // Default to 5 seconds
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			build, err := s.Get(ctx, buildID)
			if err != nil {
				return nil, err
			}

			if build.Finished {
				return build, nil
			}

			// Wait before next poll
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(pollInterval) * time.Second):
				// Continue polling
			}
		}
	}
}

// BuildFilter represents filters for querying builds.
type BuildFilter struct {
	Project     string
	Jobset      string
	Job         string
	System      string
	Finished    *bool
	BuildStatus *models.BuildStatus
	Limit       int
}

// FilterBuilds filters builds from evaluation results.
func FilterBuilds(builds []models.Build, filter *BuildFilter) []models.Build {
	if filter == nil {
		return builds
	}

	filtered := make([]models.Build, 0, len(builds))
	count := 0

	for _, build := range builds {
		// Apply filters
		if filter.Project != "" && build.Project != filter.Project {
			continue
		}
		if filter.Jobset != "" && build.Jobset != filter.Jobset {
			continue
		}
		if filter.Job != "" && build.Job != filter.Job {
			continue
		}
		if filter.System != "" && build.System != filter.System {
			continue
		}
		if filter.Finished != nil && build.Finished != *filter.Finished {
			continue
		}
		if filter.BuildStatus != nil && build.BuildStatus != nil && *build.BuildStatus != *filter.BuildStatus {
			continue
		}

		filtered = append(filtered, build)
		count++

		// Apply limit
		if filter.Limit > 0 && count >= filter.Limit {
			break
		}
	}

	return filtered
}

// ParseBuildID parses a build ID from string.
func ParseBuildID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid build ID '%s': %w", s, err)
	}
	if id <= 0 {
		return 0, fmt.Errorf("build ID must be positive: %d", id)
	}

	return id, nil
}

// GetBuildURL returns the web URL for a build.
func GetBuildURL(baseURL string, buildID int) string {
	return fmt.Sprintf("%s/build/%d", baseURL, buildID)
}

// GetEvaluationURL returns the web URL for an evaluation.
func GetEvaluationURL(baseURL string, evalID int) string {
	return fmt.Sprintf("%s/eval/%d", baseURL, evalID)
}

// BuildStatistics provides statistics about a set of builds.
type BuildStatistics struct {
	Total      int
	Succeeded  int
	Failed     int
	InProgress int
	Aborted    int
	TimedOut   int
	Other      int
}

// CalculateStatistics calculates statistics for a set of builds.
func CalculateStatistics(builds []models.Build) *BuildStatistics {
	stats := &BuildStatistics{
		Total: len(builds),
	}

	for _, build := range builds {
		if !build.Finished {
			stats.InProgress++

			continue
		}

		if build.BuildStatus == nil {
			stats.Other++

			continue
		}

		switch *build.BuildStatus {
		case models.BuildStatusSuccess:
			stats.Succeeded++
		case models.BuildStatusFailed, models.BuildStatusDependencyFailed, models.BuildStatusFailedWithOutput:
			stats.Failed++
		case models.BuildStatusAborted, models.BuildStatusAborted2, models.BuildStatusCanceledByUser:
			stats.Aborted++
		case models.BuildStatusTimedOut:
			stats.TimedOut++
		case models.BuildStatusLogSizeLimitExceeded, models.BuildStatusOutputSizeLimitExceeded:
			stats.Failed++ // Count size limit exceeded as failures
		default:
			stats.Other++
		}
	}

	return stats
}

// GetSuccessRate returns the success rate as a percentage.
func (s *BuildStatistics) GetSuccessRate() float64 {
	if s.Total == 0 {
		return 0
	}

	return float64(s.Succeeded) / float64(s.Total) * 100
}
