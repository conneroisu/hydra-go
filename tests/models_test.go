package tests

import (
	"testing"
	"time"

	"github.com/conneroisu/hydra-go/hydra/models"
	"github.com/stretchr/testify/assert"
)

func TestBuildStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status models.BuildStatus
		want   int
	}{
		{"Success", models.BuildStatusSuccess, 0},
		{"Failed", models.BuildStatusFailed, 1},
		{"DependencyFailed", models.BuildStatusDependencyFailed, 2},
		{"Aborted", models.BuildStatusAborted, 3},
		{"Aborted2", models.BuildStatusAborted2, 9},
		{"CanceledByUser", models.BuildStatusCanceledByUser, 4},
		{"FailedWithOutput", models.BuildStatusFailedWithOutput, 6},
		{"TimedOut", models.BuildStatusTimedOut, 7},
		{"LogSizeLimitExceeded", models.BuildStatusLogSizeLimitExceeded, 10},
		{"OutputSizeLimitExceeded", models.BuildStatusOutputSizeLimitExceeded, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, int(tt.status))
		})
	}
}

func TestBuildMethods(t *testing.T) {
	t.Run("IsSuccess", func(t *testing.T) {
		successStatus := models.BuildStatusSuccess
		failedStatus := models.BuildStatusFailed

		build := &models.Build{
			BuildStatus: &successStatus,
			Finished:    true,
		}
		assert.True(t, build.IsSuccess())

		build.BuildStatus = &failedStatus
		assert.False(t, build.IsSuccess())

		build.BuildStatus = nil
		assert.False(t, build.IsSuccess())
	})

	t.Run("IsFailed", func(t *testing.T) {
		failedStatus := models.BuildStatusFailed
		depFailedStatus := models.BuildStatusDependencyFailed
		successStatus := models.BuildStatusSuccess

		build := &models.Build{
			BuildStatus: &failedStatus,
			Finished:    true,
		}
		assert.True(t, build.IsFailed())

		build.BuildStatus = &depFailedStatus
		assert.True(t, build.IsFailed())

		build.BuildStatus = &successStatus
		assert.False(t, build.IsFailed())
	})

	t.Run("GetBuildStatusString", func(t *testing.T) {
		tests := []struct {
			name     string
			build    *models.Build
			expected string
		}{
			{
				name: "succeeded",
				build: &models.Build{
					BuildStatus: func() *models.BuildStatus {
						s := models.BuildStatusSuccess
						return &s
					}(),
					Finished: true,
				},
				expected: "succeeded",
			},
			{
				name: "failed",
				build: &models.Build{
					BuildStatus: func() *models.BuildStatus {
						s := models.BuildStatusFailed
						return &s
					}(),
					Finished: true,
				},
				expected: "failed",
			},
			{
				name: "in progress",
				build: &models.Build{
					BuildStatus: nil,
					Finished:    false,
				},
				expected: "in progress",
			},
			{
				name: "unknown",
				build: &models.Build{
					BuildStatus: nil,
					Finished:    true,
				},
				expected: "unknown",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.build.GetBuildStatusString())
			})
		}
	})

	t.Run("Time functions", func(t *testing.T) {
		now := time.Now().Unix()
		build := &models.Build{
			StartTime: now - 300,
			StopTime:  now,
			Timestamp: now - 400,
		}

		// Test GetDuration
		duration := build.GetDuration()
		assert.Equal(t, 5*time.Minute, duration)

		// Test GetStartTime
		startTime := build.GetStartTime()
		assert.Equal(t, now-300, startTime.Unix())

		// Test GetStopTime
		stopTime := build.GetStopTime()
		assert.Equal(t, now, stopTime.Unix())

		// Test GetTimestamp
		timestamp := build.GetTimestamp()
		assert.Equal(t, now-400, timestamp.Unix())

		// Test with zero times
		emptyBuild := &models.Build{}
		assert.Equal(t, time.Duration(0), emptyBuild.GetDuration())
		assert.Equal(t, time.Unix(0, 0), emptyBuild.GetStartTime())
	})
}

func TestJobsetMethods(t *testing.T) {
	t.Run("JobsetState", func(t *testing.T) {
		assert.Equal(t, 0, int(models.JobsetStateDisabled))
		assert.Equal(t, 1, int(models.JobsetStateEnabled))
		assert.Equal(t, 2, int(models.JobsetStateOneShot))
		assert.Equal(t, 3, int(models.JobsetStateOneAtTime))
	})

	t.Run("IsEnabled", func(t *testing.T) {
		jobset := &models.Jobset{Enabled: 0}
		assert.False(t, jobset.IsEnabled())

		jobset.Enabled = 1
		assert.True(t, jobset.IsEnabled())

		jobset.Enabled = 2
		assert.False(t, jobset.IsEnabled()) // OneShot is not "enabled" in the traditional sense

		jobset.Enabled = 3
		assert.False(t, jobset.IsEnabled()) // OneAtATime is not "enabled" in the traditional sense
	})

	t.Run("GetState and SetState", func(t *testing.T) {
		jobset := &models.Jobset{Enabled: 0}
		assert.Equal(t, models.JobsetStateDisabled, jobset.GetState())

		jobset.SetState(models.JobsetStateEnabled)
		assert.Equal(t, 1, jobset.Enabled)
		assert.Equal(t, models.JobsetStateEnabled, jobset.GetState())

		jobset.SetState(models.JobsetStateOneShot)
		assert.Equal(t, 2, jobset.Enabled)
		assert.Equal(t, models.JobsetStateOneShot, jobset.GetState())

		jobset.SetState(models.JobsetStateOneAtTime)
		assert.Equal(t, 3, jobset.Enabled)
		assert.Equal(t, models.JobsetStateOneAtTime, jobset.GetState())
	})
}

func TestProjectMethods(t *testing.T) {
	t.Run("Project validation", func(t *testing.T) {
		project := &models.Project{
			Name:        "test-project",
			DisplayName: "Test Project",
			Owner:       "testuser",
			Enabled:     true,
			Hidden:      false,
		}

		assert.Equal(t, "test-project", project.Name)
		assert.Equal(t, "Test Project", project.DisplayName)
		assert.Equal(t, "testuser", project.Owner)
		assert.True(t, project.Enabled)
		assert.False(t, project.Hidden)
	})

	t.Run("CreateProjectRequest", func(t *testing.T) {
		req := &models.CreateProjectRequest{
			Name:                    "new-project",
			DisplayName:             "New Project",
			Description:             "A new test project",
			Homepage:                "https://example.com",
			Owner:                   "owner",
			Enabled:                 true,
			EnableDynamicRunCommand: false,
			Visible:                 true,
		}

		assert.Equal(t, "new-project", req.Name)
		assert.Equal(t, "New Project", req.DisplayName)
		assert.Equal(t, "A new test project", req.Description)
		assert.Equal(t, "https://example.com", req.Homepage)
		assert.Equal(t, "owner", req.Owner)
		assert.True(t, req.Enabled)
		assert.False(t, req.EnableDynamicRunCommand)
		assert.True(t, req.Visible)
	})
}

func TestSearchResult(t *testing.T) {
	result := &models.SearchResult{
		Projects: []models.Project{
			{Name: "project1"},
			{Name: "project2"},
		},
		Jobsets: []models.Jobset{
			{Name: "jobset1"},
		},
		Builds: []models.Build{
			{ID: 1, NixName: "build1"},
			{ID: 2, NixName: "build2"},
			{ID: 3, NixName: "build3"},
		},
		BuildsDrv: []models.Build{},
	}

	assert.Len(t, result.Projects, 2)
	assert.Len(t, result.Jobsets, 1)
	assert.Len(t, result.Builds, 3)
	assert.Empty(t, result.BuildsDrv)
}

func TestUser(t *testing.T) {
	user := &models.User{
		Username:     "testuser",
		FullName:     "Test User",
		EmailAddress: "test@example.com",
		UserRoles:    []string{"user", "admin"},
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, "test@example.com", user.EmailAddress)
	assert.Contains(t, user.UserRoles, "user")
	assert.Contains(t, user.UserRoles, "admin")
}

func TestJobsetInput(t *testing.T) {
	input := models.JobsetInput{
		Name:             "nixpkgs",
		Type:             "git",
		Value:            "https://github.com/NixOS/nixpkgs.git master",
		EmailResponsible: true,
	}

	assert.Equal(t, "nixpkgs", input.Name)
	assert.Equal(t, "git", input.Type)
	assert.Equal(t, "https://github.com/NixOS/nixpkgs.git master", input.Value)
	assert.True(t, input.EmailResponsible)
}

func TestEvaluations(t *testing.T) {
	eval := &models.JobsetEval{
		ID:           123,
		Timestamp:    time.Now().Unix(),
		HasNewBuilds: true,
		Builds:       []int{1, 2, 3},
	}

	assert.Equal(t, 123, eval.ID)
	assert.True(t, eval.HasNewBuilds)
	assert.Len(t, eval.Builds, 3)

	evals := &models.Evaluations{
		First: "?page=1",
		Last:  "?page=10",
		Evals: []map[string]*models.JobsetEval{
			{"1": eval},
		},
	}

	assert.Equal(t, "?page=1", evals.First)
	assert.Equal(t, "?page=10", evals.Last)
	assert.Len(t, evals.Evals, 1)
}
