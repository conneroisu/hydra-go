package search

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/models"
)

// Service handles search operations.
type Service struct {
	client *client.Client
}

// NewService creates a new search service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// Search performs a search query.
func (s *Service) Search(ctx context.Context, query string) (*models.SearchResult, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// Trim whitespace
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	path := "/search?query=" + url.QueryEscape(query)
	var result models.SearchResult
	if err := s.client.DoRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("search failed for query '%s': %w", query, err)
	}

	return &result, nil
}

// SearchOptions provides advanced search options.
type SearchOptions struct {
	Query           string
	IncludeJobsets  bool
	IncludeProjects bool
	IncludeBuilds   bool
	IncludeDrvs     bool
}

// NewSearchOptions creates search options with defaults.
func NewSearchOptions(query string) *SearchOptions {
	return &SearchOptions{
		Query:           query,
		IncludeJobsets:  true,
		IncludeProjects: true,
		IncludeBuilds:   true,
		IncludeDrvs:     true,
	}
}

// SearchWithOptions performs a search with filtering options.
func (s *Service) SearchWithOptions(ctx context.Context, opts *SearchOptions) (*models.SearchResult, error) {
	if opts == nil {
		return nil, errors.New("search options are required")
	}

	// Perform the search
	result, err := s.Search(ctx, opts.Query)
	if err != nil {
		return nil, err
	}

	// Filter results based on options
	filtered := &models.SearchResult{}

	if opts.IncludeJobsets {
		filtered.Jobsets = result.Jobsets
	}
	if opts.IncludeProjects {
		filtered.Projects = result.Projects
	}
	if opts.IncludeBuilds {
		filtered.Builds = result.Builds
	}
	if opts.IncludeDrvs {
		filtered.BuildsDrv = result.BuildsDrv
	}

	return filtered, nil
}

// SearchProjects searches only for projects.
func (s *Service) SearchProjects(ctx context.Context, query string) ([]models.Project, error) {
	result, err := s.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Projects, nil
}

// SearchJobsets searches only for jobsets.
func (s *Service) SearchJobsets(ctx context.Context, query string) ([]models.Jobset, error) {
	result, err := s.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Jobsets, nil
}

// SearchBuilds searches only for builds.
func (s *Service) SearchBuilds(ctx context.Context, query string) ([]models.Build, error) {
	result, err := s.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Builds, nil
}

// SearchDerivations searches only for derivations.
func (s *Service) SearchDerivations(ctx context.Context, query string) ([]models.Build, error) {
	result, err := s.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.BuildsDrv, nil
}

// SearchSummary provides a summary of search results.
type SearchSummary struct {
	Query        string
	ProjectCount int
	JobsetCount  int
	BuildCount   int
	DrvCount     int
	TotalResults int
}

// GetSearchSummary returns a summary of search results.
func GetSearchSummary(query string, result *models.SearchResult) *SearchSummary {
	if result == nil {
		return &SearchSummary{Query: query}
	}

	summary := &SearchSummary{
		Query:        query,
		ProjectCount: len(result.Projects),
		JobsetCount:  len(result.Jobsets),
		BuildCount:   len(result.Builds),
		DrvCount:     len(result.BuildsDrv),
	}

	summary.TotalResults = summary.ProjectCount + summary.JobsetCount +
		summary.BuildCount + summary.DrvCount

	return summary
}

// HasResults returns true if the search returned any results.
func (s *SearchSummary) HasResults() bool {
	return s.TotalResults > 0
}

// Format formats the search summary as a string.
func (s *SearchSummary) Format() string {
	if !s.HasResults() {
		return fmt.Sprintf("No results found for '%s'", s.Query)
	}

	parts := []string{}
	if s.ProjectCount > 0 {
		parts = append(parts, fmt.Sprintf("%d project(s)", s.ProjectCount))
	}
	if s.JobsetCount > 0 {
		parts = append(parts, fmt.Sprintf("%d jobset(s)", s.JobsetCount))
	}
	if s.BuildCount > 0 {
		parts = append(parts, fmt.Sprintf("%d build(s)", s.BuildCount))
	}
	if s.DrvCount > 0 {
		parts = append(parts, fmt.Sprintf("%d derivation(s)", s.DrvCount))
	}

	return fmt.Sprintf("Found %s for query '%s'", strings.Join(parts, ", "), s.Query)
}

// SearchType represents the type of search result.
type SearchType string

const (
	SearchTypeProject SearchType = "project"
	SearchTypeJobset  SearchType = "jobset"
	SearchTypeBuild   SearchType = "build"
	SearchTypeDrv     SearchType = "derivation"
)

// SearchItem represents a unified search result item.
type SearchItem struct {
	Type        SearchType
	Name        string
	Description string
	URL         string
	Project     *models.Project
	Jobset      *models.Jobset
	Build       *models.Build
}

// FlattenSearchResults converts search results to a flat list of items.
func FlattenSearchResults(baseURL string, result *models.SearchResult) []SearchItem {
	if result == nil {
		return nil
	}

	// Preallocate with estimated capacity
	capacity := len(result.Projects) + len(result.Jobsets) + len(result.Builds) + len(result.BuildsDrv)
	items := make([]SearchItem, 0, capacity)

	// Add projects
	for _, project := range result.Projects {
		items = append(items, SearchItem{
			Type:        SearchTypeProject,
			Name:        project.Name,
			Description: project.Description,
			URL:         fmt.Sprintf("%s/project/%s", baseURL, project.Name),
			Project:     &project,
		})
	}

	// Add jobsets
	for _, jobset := range result.Jobsets {
		desc := ""
		if jobset.Description != nil {
			desc = *jobset.Description
		}
		items = append(items, SearchItem{
			Type:        SearchTypeJobset,
			Name:        fmt.Sprintf("%s:%s", jobset.Project, jobset.Name),
			Description: desc,
			URL:         fmt.Sprintf("%s/jobset/%s/%s", baseURL, jobset.Project, jobset.Name),
			Jobset:      &jobset,
		})
	}

	// Add builds
	for _, build := range result.Builds {
		items = append(items, SearchItem{
			Type:        SearchTypeBuild,
			Name:        fmt.Sprintf("%s #%d", build.Job, build.ID),
			Description: build.NixName,
			URL:         fmt.Sprintf("%s/build/%d", baseURL, build.ID),
			Build:       &build,
		})
	}

	// Add derivations
	for _, build := range result.BuildsDrv {
		items = append(items, SearchItem{
			Type:        SearchTypeDrv,
			Name:        build.DrvPath + " (drv)",
			Description: build.NixName,
			URL:         fmt.Sprintf("%s/build/%d", baseURL, build.ID),
			Build:       &build,
		})
	}

	return items
}
