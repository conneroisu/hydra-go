//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/conneroisu/hydra-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestURL returns the test server URL from environment or default
func getTestURL() string {
	if url := os.Getenv("HYDRA_TEST_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func TestIntegrationClientCreation(t *testing.T) {
	url := getTestURL()

	t.Run("create client with URL", func(t *testing.T) {
		client, err := hydra.NewClientWithURL(url)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, url, client.BaseURL())
	})

	t.Run("create client with config", func(t *testing.T) {
		cfg := &hydra.Config{
			BaseURL:   url,
			UserAgent: "test-client/1.0",
			Timeout:   30 * time.Second,
		}
		client, err := hydra.NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestIntegrationAuthentication(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("login with valid credentials", func(t *testing.T) {
		user, err := client.Login(ctx, "testuser", "testpass")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "Test User", user.FullName)
		assert.True(t, client.IsAuthenticated())
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		_, err := client.Login(ctx, "invalid", "wrong")
		assert.Error(t, err)
	})

	t.Run("logout", func(t *testing.T) {
		// First login
		_, err := client.Login(ctx, "testuser", "testpass")
		require.NoError(t, err)
		assert.True(t, client.IsAuthenticated())

		// Then logout
		client.Logout()
		assert.False(t, client.IsAuthenticated())
	})
}

func TestIntegrationProjects(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("list projects", func(t *testing.T) {
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, projects)

		// Check for expected projects
		projectNames := make(map[string]bool)
		for _, p := range projects {
			projectNames[p.Name] = true
		}
		assert.True(t, projectNames["nixpkgs"])
		assert.True(t, projectNames["hydra"])
	})

	t.Run("get specific project", func(t *testing.T) {
		project, err := client.GetProject(ctx, "nixpkgs")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "nixpkgs", project.Name)
		assert.Equal(t, "Nixpkgs", project.DisplayName)
		assert.True(t, project.Enabled)
	})

	t.Run("get non-existent project", func(t *testing.T) {
		_, err := client.GetProject(ctx, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("create and delete project", func(t *testing.T) {
		// Login first (required for create/delete)
		_, err := client.Login(ctx, "testuser", "testpass")
		require.NoError(t, err)

		// Create project
		req := &hydra.CreateProjectRequest{
			Name:        "test-create-project",
			DisplayName: "Test Create Project",
			Description: "Created by integration test",
			Owner:       "testuser",
			Enabled:     true,
			Visible:     true,
		}

		resp, err := client.CreateProject(ctx, "test-create-project", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify it exists
		project, err := client.GetProject(ctx, "test-create-project")
		assert.NoError(t, err)
		assert.Equal(t, "test-create-project", project.Name)

		// Delete project
		err = client.DeleteProject(ctx, "test-create-project")
		assert.NoError(t, err)

		// Verify it's gone
		_, err = client.GetProject(ctx, "test-create-project")
		assert.Error(t, err)
	})
}

func TestIntegrationJobsets(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("list jobsets", func(t *testing.T) {
		jobsets, err := client.ListJobsets(ctx, "nixpkgs")
		assert.NoError(t, err)
		assert.NotNil(t, jobsets)

		// Check for expected jobsets
		found := false
		for _, jobset := range jobsets {
			if jobset.Name == "trunk" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find 'trunk' jobset")
	})

	t.Run("get specific jobset", func(t *testing.T) {
		jobset, err := client.GetJobset(ctx, "nixpkgs", "trunk")
		assert.NoError(t, err)
		assert.NotNil(t, jobset)
		assert.Equal(t, "trunk", jobset.Name)
		assert.Equal(t, "nixpkgs", jobset.Project)
		assert.True(t, jobset.IsEnabled())
	})

	t.Run("get non-existent jobset", func(t *testing.T) {
		_, err := client.GetJobset(ctx, "nixpkgs", "nonexistent")
		assert.Error(t, err)
	})

	t.Run("get evaluations", func(t *testing.T) {
		evals, err := client.GetJobsetEvaluations(ctx, "nixpkgs", "trunk")
		assert.NoError(t, err)
		assert.NotNil(t, evals)
	})

	t.Run("trigger evaluation", func(t *testing.T) {
		resp, err := client.TriggerJobset(ctx, "nixpkgs", "trunk")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestIntegrationBuilds(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("get build", func(t *testing.T) {
		build, err := client.GetBuild(ctx, 123456)
		assert.NoError(t, err)
		assert.NotNil(t, build)
		assert.Equal(t, 123456, build.ID)
		assert.Equal(t, "hello-2.12.1", build.NixName)
		assert.True(t, build.Finished)
		assert.True(t, build.IsSuccess())
	})

	t.Run("get non-existent build", func(t *testing.T) {
		_, err := client.GetBuild(ctx, 999999999)
		assert.Error(t, err)
	})

	t.Run("get build constituents", func(t *testing.T) {
		constituents, err := client.GetBuildConstituents(ctx, 123456)
		assert.NoError(t, err)
		assert.NotNil(t, constituents)
	})

	t.Run("get build info", func(t *testing.T) {
		info, err := client.GetBuildInfo(ctx, 123456)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.NotNil(t, info.Build)
		assert.Equal(t, 123456, info.Build.ID)
	})

	t.Run("build status methods", func(t *testing.T) {
		// Test successful build
		build, err := client.GetBuild(ctx, 123456)
		require.NoError(t, err)

		assert.True(t, build.IsSuccess())
		assert.False(t, build.IsFailed())
		assert.Equal(t, "succeeded", build.GetBuildStatusString())

		// Test failed build
		failedBuild, err := client.GetBuild(ctx, 123460)
		require.NoError(t, err)

		assert.False(t, failedBuild.IsSuccess())
		assert.True(t, failedBuild.IsFailed())
		assert.Equal(t, "failed", failedBuild.GetBuildStatusString())

		// Test in-progress build
		inProgressBuild, err := client.GetBuild(ctx, 123459)
		require.NoError(t, err)

		assert.False(t, inProgressBuild.Finished)
		assert.Equal(t, "in progress", inProgressBuild.GetBuildStatusString())
	})
}

func TestIntegrationSearch(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("search all", func(t *testing.T) {
		results, err := client.Search(ctx, "hello")
		assert.NoError(t, err)
		assert.NotNil(t, results)

		// Should find at least the hello build
		assert.NotEmpty(t, results.Builds)

		foundHello := false
		for _, build := range results.Builds {
			if build.NixName == "hello-2.12.1" {
				foundHello = true
				break
			}
		}
		assert.True(t, foundHello)
	})

	t.Run("search projects", func(t *testing.T) {
		results, err := client.Search(ctx, "nix")
		assert.NoError(t, err)
		assert.NotNil(t, results)

		// Should find nixpkgs project
		foundNixpkgs := false
		for _, project := range results.Projects {
			if project.Name == "nixpkgs" {
				foundNixpkgs = true
				break
			}
		}
		assert.True(t, foundNixpkgs)
	})
}

func TestIntegrationQuickStart(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()
	quick := client.Quick()

	t.Run("get project with jobsets", func(t *testing.T) {
		project, jobsets, err := quick.GetProjectWithJobsets(ctx, "nixpkgs")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.NotNil(t, jobsets)
		assert.Equal(t, "nixpkgs", project.Name)
		assert.NotEmpty(t, jobsets)
	})

	t.Run("get latest build for job", func(t *testing.T) {
		build, err := quick.GetLatestBuildForJob(ctx, "nixpkgs", "trunk", "hello")
		if err == nil {
			assert.NotNil(t, build)
			assert.Equal(t, "hello", build.Job)
		}
		// May fail if no recent evaluations, which is OK for test
	})
}

func TestIntegrationConcurrency(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("concurrent project list", func(t *testing.T) {
		const numRequests = 10
		var wg sync.WaitGroup
		errors := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				projects, err := client.ListProjects(ctx)
				if err != nil {
					errors <- err
					return
				}

				if len(projects) == 0 {
					errors <- assert.AnError
				}
			}(i)
		}

		// Wait for all requests
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Fatalf("Concurrent request failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	})

	t.Run("concurrent mixed operations", func(t *testing.T) {
		const numOps = 5
		var wg sync.WaitGroup
		errors := make(chan error, numOps*3)

		// List projects concurrently
		for i := 0; i < numOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.ListProjects(ctx)
				if err != nil {
					errors <- err
				}
			}()
		}

		// Get specific projects concurrently
		for i := 0; i < numOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.GetProject(ctx, "nixpkgs")
				if err != nil {
					errors <- err
				}
			}()
		}

		// Search concurrently
		for i := 0; i < numOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.Search(ctx, "test")
				if err != nil {
					errors <- err
				}
			}()
		}

		// Wait for all operations
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Fatalf("Concurrent operation failed: %v", err)
		case <-time.After(15 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	})
}

func TestIntegrationErrorHandling(t *testing.T) {
	client, err := hydra.NewClientWithURL(getTestURL())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("handle 404 errors", func(t *testing.T) {
		_, err := client.GetProject(ctx, "definitely-does-not-exist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("handle invalid IDs", func(t *testing.T) {
		_, err := client.GetBuild(ctx, -1)
		assert.Error(t, err)
	})

	t.Run("handle network timeout", func(t *testing.T) {
		// Create client with very short timeout
		cfg := &hydra.Config{
			BaseURL: getTestURL(),
			Timeout: 1 * time.Nanosecond,
		}
		timeoutClient, err := hydra.NewClient(cfg)
		require.NoError(t, err)

		// This should timeout
		_, err = timeoutClient.ListProjects(ctx)
		assert.Error(t, err)
	})
}
