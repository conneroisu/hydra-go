// Package hydra type definitions and re-exports.
// This file provides a flat API surface by re-exporting all public types
// from internal packages, enabling users to access everything through
// a single import: github.com/conneroisu/hydra-go
package hydra

// Type re-exports for flat API access.
// All public types from internal packages are made available at the root level
// to provide a consistent and discoverable API surface.

import (
	"github.com/conneroisu/hydra-go/hydra/builds"
	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/jobsets"
	"github.com/conneroisu/hydra-go/hydra/models"
	"github.com/conneroisu/hydra-go/hydra/projects"
	"github.com/conneroisu/hydra-go/hydra/search"
)

// ===== Core Models =====

// Error represents an API error response.
type Error = models.Error

// LoginRequest represents the login request payload.
type LoginRequest = models.LoginRequest

// User represents the authenticated user information.
type User = models.User

// ===== Project Types =====

// Project represents a Hydra project.
type Project = models.Project

// DeclarativeInput represents declarative configuration for a project.
type DeclarativeInput = models.DeclarativeInput

// CreateProjectRequest represents the request to create a project.
type CreateProjectRequest = models.CreateProjectRequest

// ProjectResponse represents the response from project creation.
type ProjectResponse = models.ProjectResponse

// CreateProjectOptions represents options for creating a project with builder pattern.
type CreateProjectOptions = projects.CreateOptions

// ===== Jobset Types =====

// JobsetInput represents a jobset input configuration.
type JobsetInput = models.JobsetInput

// Jobset represents a Hydra jobset.
type Jobset = models.Jobset

// JobsetOverviewItem represents a single jobset in the overview.
type JobsetOverviewItem = models.JobsetOverviewItem

// JobsetOverview is a list of jobset overview items.
type JobsetOverview = models.JobsetOverview

// JobsetEvalInput represents an evaluation input.
type JobsetEvalInput = models.JobsetEvalInput

// JobsetEval represents a jobset evaluation.
type JobsetEval = models.JobsetEval

// Evaluations represents a paginated list of evaluations.
type Evaluations = models.Evaluations

// JobsetState represents the enabled state of a jobset.
type JobsetState = models.JobsetState

// JobsetEvalBuilds represents builds for an evaluation.
type JobsetEvalBuilds = models.JobsetEvalBuilds

// CreateJobsetOptions represents options for creating/updating a jobset.
type CreateJobsetOptions = jobsets.JobsetOptions

// Jobset state constants.
const (
	JobsetStateDisabled  = models.JobsetStateDisabled
	JobsetStateEnabled   = models.JobsetStateEnabled
	JobsetStateOneShot   = models.JobsetStateOneShot
	JobsetStateOneAtTime = models.JobsetStateOneAtTime
)

// ===== Build Types =====

// BuildProduct represents a build product.
type BuildProduct = models.BuildProduct

// BuildOutput represents a build output.
type BuildOutput = models.BuildOutput

// BuildMetric represents a build metric.
type BuildMetric = models.BuildMetric

// BuildStatus represents the status of a build.
type BuildStatus = models.BuildStatus

// Build represents a Hydra build.
type Build = models.Build

// BuildInfo provides detailed information about a build.
type BuildInfo = builds.BuildInfo

// BuildFilter represents filters for querying builds.
type BuildFilter = builds.BuildFilter

// BuildStatistics provides statistics about a set of builds.
type BuildStatistics = builds.BuildStatistics

// Build status constants.
const (
	BuildStatusSuccess                 = models.BuildStatusSuccess
	BuildStatusFailed                  = models.BuildStatusFailed
	BuildStatusDependencyFailed        = models.BuildStatusDependencyFailed
	BuildStatusAborted                 = models.BuildStatusAborted
	BuildStatusCanceledByUser          = models.BuildStatusCanceledByUser
	BuildStatusFailedWithOutput        = models.BuildStatusFailedWithOutput
	BuildStatusTimedOut                = models.BuildStatusTimedOut
	BuildStatusAborted2                = models.BuildStatusAborted2
	BuildStatusLogSizeLimitExceeded    = models.BuildStatusLogSizeLimitExceeded
	BuildStatusOutputSizeLimitExceeded = models.BuildStatusOutputSizeLimitExceeded
)

// ===== Search Types =====

// SearchResult represents search results.
type SearchResult = models.SearchResult

// SearchOptions provides advanced search options.
type SearchOptions = search.SearchOptions

// SearchSummary provides a summary of search results.
type SearchSummary = search.SearchSummary

// SearchType represents the type of search result.
type SearchType = search.SearchType

// SearchItem represents a unified search result item.
type SearchItem = search.SearchItem

// Search type constants.
const (
	SearchTypeProject = search.SearchTypeProject
	SearchTypeJobset  = search.SearchTypeJobset
	SearchTypeBuild   = search.SearchTypeBuild
	SearchTypeDrv     = search.SearchTypeDrv
)

// ===== Other Types =====

// PushRequest represents a request to trigger jobsets.
type PushRequest = models.PushRequest

// PushResponse represents the response from triggering jobsets.
type PushResponse = models.PushResponse

// ShieldData represents data for shields.io badge.
type ShieldData = models.ShieldData

// APIError represents an API error.
type APIError = client.APIError

// Option represents a client configuration option.
type Option = client.Option

// Client option functions.
var (
	WithHTTPClient = client.WithHTTPClient
	WithUserAgent  = client.WithUserAgent
	WithTimeout    = client.WithTimeout
)
