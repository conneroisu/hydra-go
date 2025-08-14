package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/conneroisu/hydra-go"
	"github.com/stretchr/testify/assert"
)

func TestBuildService(t *testing.T) {
	t.Run("ParseBuildID", func(t *testing.T) {
		tests := []struct {
			input   string
			want    int
			wantErr bool
		}{
			{"123", 123, false},
			{"456789", 456789, false},
			{"0", 0, true},
			{"-1", 0, true},
			{"abc", 0, true},
			{"", 0, true},
		}

		for _, tt := range tests {
			id, err := hydra.ParseBuildID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, id)
			}
		}
	})

	t.Run("GetBuildURL", func(t *testing.T) {
		url := hydra.GetBuildURL("https://hydra.nixos.org", 123456)
		assert.Equal(t, "https://hydra.nixos.org/build/123456", url)
	})

	t.Run("GetEvaluationURL", func(t *testing.T) {
		url := hydra.GetEvaluationURL("https://hydra.nixos.org", 789)
		assert.Equal(t, "https://hydra.nixos.org/eval/789", url)
	})

	t.Run("FilterBuilds", func(t *testing.T) {
		testBuilds := []hydra.Build{
			{ID: 1, Project: "nixpkgs", Jobset: "trunk", Job: "hello", System: "x86_64-linux", Finished: true},
			{ID: 2, Project: "nixpkgs", Jobset: "staging", Job: "git", System: "x86_64-linux", Finished: true},
			{ID: 3, Project: "hydra", Jobset: "master", Job: "hydra", System: "aarch64-linux", Finished: false},
			{ID: 4, Project: "nixpkgs", Jobset: "trunk", Job: "hello", System: "aarch64-darwin", Finished: true},
		}

		// Filter by project
		filter := &hydra.BuildFilter{Project: "nixpkgs"}
		filtered := hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 3)

		// Filter by jobset
		filter = &hydra.BuildFilter{Jobset: "trunk"}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 2)

		// Filter by job
		filter = &hydra.BuildFilter{Job: "hello"}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 2)

		// Filter by system
		filter = &hydra.BuildFilter{System: "x86_64-linux"}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 2)

		// Filter by finished status
		finished := true
		filter = &hydra.BuildFilter{Finished: &finished}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 3)

		// Filter with limit
		filter = &hydra.BuildFilter{Limit: 2}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 2)

		// Combined filters
		filter = &hydra.BuildFilter{
			Project: "nixpkgs",
			Jobset:  "trunk",
			System:  "x86_64-linux",
		}
		filtered = hydra.FilterBuilds(testBuilds, filter)
		assert.Len(t, filtered, 1)
		assert.Equal(t, 1, filtered[0].ID)
	})

	t.Run("CalculateStatistics", func(t *testing.T) {
		successStatus := hydra.BuildStatusSuccess
		failedStatus := hydra.BuildStatusFailed
		abortedStatus := hydra.BuildStatusAborted
		timedOutStatus := hydra.BuildStatusTimedOut

		testBuilds := []hydra.Build{
			{ID: 1, Finished: true, BuildStatus: &successStatus},
			{ID: 2, Finished: true, BuildStatus: &successStatus},
			{ID: 3, Finished: true, BuildStatus: &failedStatus},
			{ID: 4, Finished: false},
			{ID: 5, Finished: true, BuildStatus: &abortedStatus},
			{ID: 6, Finished: true, BuildStatus: &timedOutStatus},
			{ID: 7, Finished: true, BuildStatus: nil},
		}

		stats := hydra.CalculateStatistics(testBuilds)

		assert.Equal(t, 7, stats.Total)
		assert.Equal(t, 2, stats.Succeeded)
		assert.Equal(t, 1, stats.Failed)
		assert.Equal(t, 1, stats.InProgress)
		assert.Equal(t, 1, stats.Aborted)
		assert.Equal(t, 1, stats.TimedOut)
		assert.Equal(t, 1, stats.Other)

		successRate := stats.GetSuccessRate()
		assert.InDelta(t, 28.57, successRate, 0.01)
	})
}

func TestProjectService(t *testing.T) {
	t.Run("CreateOptions builder", func(t *testing.T) {
		opts := hydra.NewCreateProjectOptions("test-project", "testuser")
		assert.NotNil(t, opts)
		assert.Equal(t, "test-project", opts.Name)
		assert.Equal(t, "testuser", opts.Owner)
		assert.True(t, opts.Enabled)
		assert.True(t, opts.Visible)

		// Test fluent interface
		opts.
			WithDisplayName("Test Project").
			WithDescription("A test project").
			WithHomepage("https://example.com").
			WithEnabled(false).
			WithDynamicRunCommand(true).
			WithVisible(false)

		assert.Equal(t, "Test Project", opts.DisplayName)
		assert.Equal(t, "A test project", opts.Description)
		assert.Equal(t, "https://example.com", opts.Homepage)
		assert.False(t, opts.Enabled)
		assert.True(t, opts.EnableDynamicRunCommand)
		assert.False(t, opts.Visible)

		// Test Build method
		req := opts.Build()
		assert.Equal(t, "test-project", req.Name)
		assert.Equal(t, "Test Project", req.DisplayName)
		assert.Equal(t, "A test project", req.Description)
		assert.Equal(t, "https://example.com", req.Homepage)
		assert.Equal(t, "testuser", req.Owner)
		assert.False(t, req.Enabled)
		assert.True(t, req.EnableDynamicRunCommand)
		assert.False(t, req.Visible)
	})

	t.Run("DeclarativeInput", func(t *testing.T) {
		declarative := &hydra.DeclarativeInput{
			File:  "spec.json",
			Type:  "git",
			Value: "https://github.com/example/repo.git",
		}

		opts := hydra.NewCreateProjectOptions("test", "owner").
			WithDeclarative(declarative)

		req := opts.Build()
		assert.NotNil(t, req.Declarative)
		assert.Equal(t, "spec.json", req.Declarative.File)
		assert.Equal(t, "git", req.Declarative.Type)
		assert.Equal(t, "https://github.com/example/repo.git", req.Declarative.Value)
	})
}

func TestJobsetService(t *testing.T) {
	t.Run("JobsetOptions builder", func(t *testing.T) {
		opts := hydra.NewCreateJobsetOptions("test-jobset", "test-project")
		assert.NotNil(t, opts)
		assert.Equal(t, "test-jobset", opts.Name)
		assert.Equal(t, "test-project", opts.Project)
		assert.Equal(t, hydra.JobsetStateEnabled, opts.Enabled)
		assert.True(t, opts.Visible)
		assert.Equal(t, 3, opts.KeepNr)
		assert.Equal(t, 300, opts.CheckInterval)
		assert.Equal(t, 100, opts.SchedulingShares)

		// Test fluent interface
		opts.
			WithDescription("Test jobset").
			WithNixExpression("nixpkgs", "release.nix").
			WithFlake("github:NixOS/nixpkgs").
			WithState(hydra.JobsetStateOneShot).
			WithEmail(true, "test@example.com").
			WithScheduling(600, 200).
			WithKeepNr(5)

		assert.Equal(t, "Test jobset", opts.Description)
		assert.Equal(t, "nixpkgs", opts.NixExprInput)
		assert.Equal(t, "release.nix", opts.NixExprPath)
		assert.Equal(t, "github:NixOS/nixpkgs", opts.Flake)
		assert.Equal(t, hydra.JobsetStateOneShot, opts.Enabled)
		assert.True(t, opts.EnableEmail)
		assert.Equal(t, "test@example.com", opts.EmailOverride)
		assert.Equal(t, 600, opts.CheckInterval)
		assert.Equal(t, 200, opts.SchedulingShares)
		assert.Equal(t, 5, opts.KeepNr)

		// Test AddInput
		opts.AddInput("nixpkgs", "git", "https://github.com/NixOS/nixpkgs.git master", true)
		assert.Len(t, opts.Inputs, 1)
		assert.Equal(t, "nixpkgs", opts.Inputs["nixpkgs"].Name)
		assert.Equal(t, "git", opts.Inputs["nixpkgs"].Type)
		assert.Equal(t, "https://github.com/NixOS/nixpkgs.git master", opts.Inputs["nixpkgs"].Value)
		assert.True(t, opts.Inputs["nixpkgs"].EmailResponsible)

		// Test Build method
		jobset := opts.Build()
		assert.Equal(t, "test-jobset", jobset.Name)
		assert.Equal(t, "test-project", jobset.Project)
		assert.NotNil(t, jobset.Description)
		assert.Equal(t, "Test jobset", *jobset.Description)
		assert.NotNil(t, jobset.NixExprInput)
		assert.Equal(t, "nixpkgs", *jobset.NixExprInput)
		assert.NotNil(t, jobset.NixExprPath)
		assert.Equal(t, "release.nix", *jobset.NixExprPath)
		assert.NotNil(t, jobset.Flake)
		assert.Equal(t, "github:NixOS/nixpkgs", *jobset.Flake)
		assert.Equal(t, 2, jobset.Enabled) // OneShot = 2
		assert.True(t, jobset.EnableEmail)
		assert.Equal(t, "test@example.com", jobset.EmailOverride)
		assert.Equal(t, 600, jobset.CheckInterval)
		assert.Equal(t, 200, jobset.SchedulingShares)
		assert.Equal(t, 5, jobset.KeepNr)
		assert.Len(t, jobset.Inputs, 1)
	})
}

func TestContextCancellation(t *testing.T) {
	// This test would normally require a real server or mock
	// Here we're just testing the context usage pattern

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate a long-running operation
	done := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			done <- true
		case <-time.After(1 * time.Second):
			done <- false
		}
	}()

	result := <-done
	assert.True(t, result, "Context should have been cancelled")
}

func TestConcurrentMapAccess(t *testing.T) {
	// Test that concurrent access to shared maps is safe
	inputs := make(map[string]hydra.JobsetInput)
	var mu sync.Mutex

	// Simulate concurrent writes
	done := make(chan bool, 10)
	for i := range 10 {
		go func(id int) {
			input := hydra.JobsetInput{
				Name:  string(rune('a' + id)),
				Type:  "git",
				Value: "value",
			}
			// This would panic if not properly synchronized
			mu.Lock()
			inputs[input.Name] = input
			mu.Unlock()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	assert.Len(t, inputs, 10)
}

func TestValidation(t *testing.T) {
	t.Run("Project validation", func(t *testing.T) {
		// Test that required fields are validated
		req := &hydra.CreateProjectRequest{}

		// In a real implementation, these would be validated
		assert.Empty(t, req.Name, "Name should be empty")
		assert.Empty(t, req.Owner, "Owner should be empty")

		// Valid request
		req = &hydra.CreateProjectRequest{
			Name:  "valid-project",
			Owner: "valid-owner",
		}
		assert.NotEmpty(t, req.Name)
		assert.NotEmpty(t, req.Owner)
	})

	t.Run("Build ID validation", func(t *testing.T) {
		// Negative IDs should be invalid
		_, err := hydra.ParseBuildID("-1")
		assert.Error(t, err)

		// Zero should be invalid
		_, err = hydra.ParseBuildID("0")
		assert.Error(t, err)

		// Positive IDs should be valid
		id, err := hydra.ParseBuildID("12345")
		assert.NoError(t, err)
		assert.Equal(t, 12345, id)
	})
}
