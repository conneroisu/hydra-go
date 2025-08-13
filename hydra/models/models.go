package models

import (
	"encoding/json"
	"time"
)

// Error represents an API error response.
type Error struct {
	Error string `json:"error"`
}

// LoginRequest represents the login request payload.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User represents the authenticated user information.
type User struct {
	Username     string   `json:"username"`
	FullName     string   `json:"fullname"`
	EmailAddress string   `json:"emailaddress"`
	UserRoles    []string `json:"userroles"`
}

// Project represents a Hydra project.
type Project struct {
	Owner                   string            `json:"owner"`
	Name                    string            `json:"name"`
	DisplayName             string            `json:"displayname"`
	Description             string            `json:"description"`
	Homepage                string            `json:"homepage"`
	Hidden                  bool              `json:"hidden"`
	Enabled                 bool              `json:"enabled"`
	EnableDynamicRunCommand bool              `json:"enable_dynamic_run_command"`
	Declarative             *DeclarativeInput `json:"declarative,omitempty"`
	Jobsets                 []string          `json:"jobsets,omitempty"`
}

// DeclarativeInput represents declarative configuration for a project.
type DeclarativeInput struct {
	File  string `json:"file"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CreateProjectRequest represents the request to create a project.
type CreateProjectRequest struct {
	Name                    string            `json:"name"`
	DisplayName             string            `json:"displayname"`
	Description             string            `json:"description"`
	Homepage                string            `json:"homepage"`
	Owner                   string            `json:"owner"`
	Enabled                 bool              `json:"enabled"`
	EnableDynamicRunCommand bool              `json:"enable_dynamic_run_command"`
	Visible                 bool              `json:"visible"`
	Declarative             *DeclarativeInput `json:"declarative,omitempty"`
}

// ProjectResponse represents the response from project creation.
type ProjectResponse struct {
	URI      string `json:"uri,omitempty"`
	Name     string `json:"name,omitempty"`
	Redirect string `json:"redirect"`
	Type     string `json:"type,omitempty"`
}

// JobsetInput represents a jobset input configuration.
type JobsetInput struct {
	Name             string `json:"name"`
	Value            string `json:"value"`
	Type             string `json:"type"`
	EmailResponsible bool   `json:"emailresponsible"`
}

// Jobset represents a Hydra jobset.
type Jobset struct {
	Name                    string                 `json:"name"`
	Project                 string                 `json:"project"`
	Description             *string                `json:"description"`
	NixExprInput            *string                `json:"nixexprinput"`
	NixExprPath             *string                `json:"nixexprpath"`
	ErrorMsg                *string                `json:"errormsg"`
	ErrorTime               *int64                 `json:"errortime"`
	LastCheckedTime         *int64                 `json:"lastcheckedtime"`
	TriggerTime             *int64                 `json:"triggertime"`
	Enabled                 int                    `json:"enabled"` // 0=disabled, 1=enabled, 2=one-shot, 3=one-at-a-time
	EnableEmail             bool                   `json:"enableemail"`
	EnableDynamicRunCommand bool                   `json:"enable_dynamic_run_command"`
	Visible                 bool                   `json:"visible"`
	EmailOverride           string                 `json:"emailoverride"`
	KeepNr                  int                    `json:"keepnr"`
	CheckInterval           int                    `json:"checkinterval"`
	SchedulingShares        int                    `json:"schedulingshares"`
	FetchErrorMsg           *string                `json:"fetcherrormsg"`
	StartTime               *int64                 `json:"startime"`
	Type                    int                    `json:"type"`
	Flake                   *string                `json:"flake"`
	Inputs                  map[string]JobsetInput `json:"inputs,omitempty"`
}

// JobsetOverviewItem represents a single jobset in the overview.
type JobsetOverviewItem struct {
	Name            string  `json:"name"`
	Project         string  `json:"project"`
	NrTotal         int     `json:"nrtotal"`
	CheckInterval   int     `json:"checkinterval"`
	HasErrorMsg     bool    `json:"haserrormsg"`
	NrScheduled     int     `json:"nrscheduled"`
	NrFailed        int     `json:"nrfailed"`
	ErrorTime       int64   `json:"errortime"`
	FetchErrorMsg   *string `json:"fetcherrormsg"`
	StartTime       *int64  `json:"starttime"`
	LastCheckedTime int64   `json:"lastcheckedtime"`
	TriggerTime     *int64  `json:"triggertime"`
}

// JobsetOverview is a list of jobset overview items.
type JobsetOverview []JobsetOverviewItem

// JobsetEvalInput represents an evaluation input.
type JobsetEvalInput struct {
	URI        *string     `json:"uri"`
	Type       string      `json:"type"`
	Revision   *string     `json:"revision"`
	Value      interface{} `json:"value"`      // Can be bool, string, or []string
	Dependency *string     `json:"dependency"` // Deprecated
}

// JobsetEval represents a jobset evaluation.
type JobsetEval struct {
	ID               int                        `json:"id"`
	Timestamp        int64                      `json:"timestamp"`
	CheckoutTime     int                        `json:"checkouttime"`
	EvalTime         int                        `json:"evaltime"`
	HasNewBuilds     bool                       `json:"hasnewbuilds"`
	Flake            *string                    `json:"flake"`
	Builds           []int                      `json:"builds"`
	JobsetEvalInputs map[string]JobsetEvalInput `json:"jobsetevalinputs"`
}

// Evaluations represents a paginated list of evaluations.
type Evaluations struct {
	First string                   `json:"first"`
	Next  string                   `json:"next"`
	Last  string                   `json:"last"`
	Evals []map[string]*JobsetEval `json:"evals"`
}

// BuildProduct represents a build product.
type BuildProduct struct {
	FileSize    *int64  `json:"filesize"`
	DefaultPath string  `json:"defaultpath"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	SubType     string  `json:"subtype"`
	SHA256Hash  *string `json:"sha256hash"`
}

// BuildOutput represents a build output.
type BuildOutput struct {
	Path string `json:"path"`
}

// BuildMetric represents a build metric.
type BuildMetric struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// BuildStatus represents the status of a build.
type BuildStatus int

const (
	BuildStatusSuccess                 BuildStatus = 0
	BuildStatusFailed                  BuildStatus = 1
	BuildStatusDependencyFailed        BuildStatus = 2
	BuildStatusAborted                 BuildStatus = 3
	BuildStatusCanceledByUser          BuildStatus = 4
	BuildStatusFailedWithOutput        BuildStatus = 6
	BuildStatusTimedOut                BuildStatus = 7
	BuildStatusAborted2                BuildStatus = 9
	BuildStatusLogSizeLimitExceeded    BuildStatus = 10
	BuildStatusOutputSizeLimitExceeded BuildStatus = 11
)

// Build represents a Hydra build.
type Build struct {
	ID            int                     `json:"id"`
	StartTime     int64                   `json:"starttime"`
	StopTime      int64                   `json:"stoptime"`
	Timestamp     int64                   `json:"timestamp"`
	JobsetEvals   []int                   `json:"jobsetevals"`
	Finished      bool                    `json:"-"` // Manually unmarshaled
	NixName       string                  `json:"nixname"`
	BuildStatus   *BuildStatus            `json:"buildstatus"` // Only null if finished is false
	Jobset        string                  `json:"jobset"`
	Priority      int                     `json:"priority"`
	Job           string                  `json:"job"`
	DrvPath       string                  `json:"drvpath"`
	System        string                  `json:"system"`
	Project       string                  `json:"project"`
	BuildProducts map[string]BuildProduct `json:"buildproducts"`
	BuildOutputs  map[string]BuildOutput  `json:"buildoutputs"`
	BuildMetrics  map[string]BuildMetric  `json:"buildmetrics"`
}

// JobsetEvalBuilds represents builds for an evaluation.
type JobsetEvalBuilds []map[string]*Build

// SearchResult represents search results.
type SearchResult struct {
	Jobsets   []Jobset  `json:"jobsets"`
	Projects  []Project `json:"projects"`
	Builds    []Build   `json:"builds"`
	BuildsDrv []Build   `json:"buildsdrv"`
}

// PushRequest represents a request to trigger jobsets.
type PushRequest struct {
	Jobsets string `json:"jobsets"` // Format: "project:jobset"
}

// PushResponse represents the response from triggering jobsets.
type PushResponse struct {
	JobsetsTriggered []string `json:"jobsetsTriggered"`
}

// ShieldData represents data for shields.io badge.
type ShieldData struct {
	Color         string `json:"color"`
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
}

// Helper methods

// IsSuccess returns true if the build was successful.
func (b *Build) IsSuccess() bool {
	return b.BuildStatus != nil && *b.BuildStatus == BuildStatusSuccess
}

// IsFailed returns true if the build failed.
func (b *Build) IsFailed() bool {
	if b.BuildStatus == nil {
		return false
	}

	return *b.BuildStatus != BuildStatusSuccess
}

// GetBuildStatusString returns a human-readable status string.
func (b *Build) GetBuildStatusString() string {
	if !b.Finished {
		return "in progress"
	}
	if b.BuildStatus == nil {
		return "unknown"
	}

	switch *b.BuildStatus {
	case BuildStatusSuccess:
		return "succeeded"
	case BuildStatusFailed:
		return "failed"
	case BuildStatusDependencyFailed:
		return "dependency failed"
	case BuildStatusAborted, BuildStatusAborted2:
		return "aborted"
	case BuildStatusCanceledByUser:
		return "canceled by user"
	case BuildStatusFailedWithOutput:
		return "failed with output"
	case BuildStatusTimedOut:
		return "timed out"
	case BuildStatusLogSizeLimitExceeded:
		return "log size limit exceeded"
	case BuildStatusOutputSizeLimitExceeded:
		return "output size limit exceeded"
	default:
		return "failed"
	}
}

// GetStartTime returns the start time as a time.Time.
func (b *Build) GetStartTime() time.Time {
	return time.Unix(b.StartTime, 0)
}

// GetStopTime returns the stop time as a time.Time.
func (b *Build) GetStopTime() time.Time {
	return time.Unix(b.StopTime, 0)
}

// GetTimestamp returns the creation timestamp as a time.Time.
func (b *Build) GetTimestamp() time.Time {
	return time.Unix(b.Timestamp, 0)
}

// GetDuration returns the build duration.
func (b *Build) GetDuration() time.Duration {
	return time.Duration(b.StopTime-b.StartTime) * time.Second
}

// JobsetState represents the enabled state of a jobset.
type JobsetState int

const (
	JobsetStateDisabled  JobsetState = 0
	JobsetStateEnabled   JobsetState = 1
	JobsetStateOneShot   JobsetState = 2
	JobsetStateOneAtTime JobsetState = 3
)

// GetState returns the jobset state.
func (j *Jobset) GetState() JobsetState {
	return JobsetState(j.Enabled)
}

// IsEnabled returns true if the jobset is enabled.
func (j *Jobset) IsEnabled() bool {
	return j.Enabled == int(JobsetStateEnabled)
}

// SetState sets the jobset state.
func (j *Jobset) SetState(state JobsetState) {
	j.Enabled = int(state)
}

// UnmarshalJSON custom unmarshaler to handle null values properly.
func (bs *BuildStatus) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var status int
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	*bs = BuildStatus(status)

	return nil
}

// MarshalJSON custom marshaler.
func (bs BuildStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(bs))
}

// UnmarshalJSON custom unmarshaler for Build to handle finished field as both bool and number.
func (b *Build) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid recursion
	type Alias Build
	aux := &struct {
		Finished interface{} `json:"finished"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle finished field which can be bool or number (0/1)
	if aux.Finished != nil {
		switch v := aux.Finished.(type) {
		case bool:
			b.Finished = v
		case float64:
			b.Finished = v != 0
		case int:
			b.Finished = v != 0
		default:
			// Default to false if type is unexpected
			b.Finished = false
		}
	}

	return nil
}

// MarshalJSON custom marshaler for Build to ensure finished is always a bool.
func (b Build) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid recursion
	type Alias Build

	return json.Marshal(&struct {
		Finished bool `json:"finished"`
		*Alias
	}{
		Finished: b.Finished,
		Alias:    (*Alias)(&b),
	})
}
