//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/conneroisu/hydra-go"
)

const (
	// Default image to use - CI will push to GHCR, locally we can build
	defaultHydraImage = "ghcr.io/conneroisu/hydra-go/hydra-test:latest"
	hydraPort         = "3000/tcp"
	healthPort        = "8080/tcp"
)

// HydraContainer wraps testcontainer functionality for Hydra
type HydraContainer struct {
	Container testcontainers.Container
	BaseURL   string
}

// StartHydraContainer starts a Hydra container and waits for it to be healthy
func StartHydraContainer(ctx context.Context, t *testing.T) *HydraContainer {
	// Try to use local image first, fallback to GHCR
	image := defaultHydraImage
	if localImage := getLocalHydraImage(); localImage != "" {
		image = localImage
		t.Logf("Using local Hydra image: %s", image)
	} else {
		t.Logf("Using GHCR Hydra image: %s", image)
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{hydraPort, healthPort},
		WaitingFor: wait.ForAll(
			// Wait for health check endpoint to return 200
			wait.ForHTTP("/health").WithPort(healthPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}).WithStartupTimeout(120*time.Second),
			// Wait for Hydra main endpoint to be accessible
			wait.ForHTTP("/").WithPort(hydraPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}).WithStartupTimeout(120*time.Second),
		),
		Networks: []string{"hydra-test-network"},
	}

	hydraContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the host and port
	host, err := hydraContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := hydraContainer.MappedPort(ctx, hydraPort)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	return &HydraContainer{
		Container: hydraContainer,
		BaseURL:   baseURL,
	}
}

// Terminate cleans up the container
func (h *HydraContainer) Terminate(ctx context.Context) error {
	return h.Container.Terminate(ctx)
}

// getLocalHydraImage checks if we have a locally built image to use for testing
func getLocalHydraImage() string {
	// This could check for locally built images or return empty to use GHCR
	// For now, return empty to always use GHCR
	return ""
}

func TestHydraContainerStartup(t *testing.T) {
	ctx := context.Background()

	hydra := StartHydraContainer(ctx, t)
	defer hydra.Terminate(ctx)

	t.Logf("Hydra container started at: %s", hydra.BaseURL)

	// Test that we can create a client and connect
	client, err := hydra.NewClientWithURL(hydra.BaseURL)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test basic connectivity
	projects, err := client.ListProjects(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, projects)
	
	t.Logf("Successfully connected to Hydra container, found %d projects", len(projects))
}

func TestHydraContainerAuthentication(t *testing.T) {
	ctx := context.Background()

	hydra := StartHydraContainer(ctx, t)
	defer hydra.Terminate(ctx)

	client, err := hydra.NewClientWithURL(hydra.BaseURL)
	require.NoError(t, err)

	// Test authentication with the default admin user created in the container
	user, err := client.Login(ctx, "admin", "admin")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "Admin", user.FullName)
	assert.True(t, client.IsAuthenticated())

	t.Logf("Successfully authenticated as: %s", user.Username)
}

func TestHydraContainerFullWorkflow(t *testing.T) {
	ctx := context.Background()

	hydra := StartHydraContainer(ctx, t)
	defer hydra.Terminate(ctx)

	client, err := hydra.NewClientWithURL(hydra.BaseURL)
	require.NoError(t, err)

	// Authenticate first
	_, err = client.Login(ctx, "admin", "admin")
	require.NoError(t, err)

	// Create a test project
	projectName := fmt.Sprintf("test-project-%d", time.Now().Unix())
	createReq := &hydra.CreateProjectRequest{
		Name:        projectName,
		DisplayName: "Test Project",
		Description: "Created by testcontainer integration test",
		Owner:       "admin",
		Enabled:     true,
		Visible:     true,
	}

	resp, err := client.CreateProject(ctx, projectName, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, projectName, resp.Name)

	t.Logf("Created project: %s", projectName)

	// Verify the project exists
	project, err := client.GetProject(ctx, projectName)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, "Test Project", project.DisplayName)
	assert.True(t, project.Enabled)

	// List projects and make sure ours is there
	projects, err := client.ListProjects(ctx)
	assert.NoError(t, err)
	
	found := false
	for _, p := range projects {
		if p.Name == projectName {
			found = true
			break
		}
	}
	assert.True(t, found, "Created project should be in project list")

	// Clean up - delete the project
	err = client.DeleteProject(ctx, projectName)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = client.GetProject(ctx, projectName)
	assert.Error(t, err, "Project should not exist after deletion")

	t.Logf("Successfully completed full workflow test")
}

func TestHydraContainerConcurrentAccess(t *testing.T) {
	ctx := context.Background()

	hydra := StartHydraContainer(ctx, t)
	defer hydra.Terminate(ctx)

	// Create multiple clients to test concurrent access
	numClients := 5
	clients := make([]*hydra.Client, numClients)
	
	for i := 0; i < numClients; i++ {
		client, err := hydra.NewClientWithURL(hydra.BaseURL)
		require.NoError(t, err)
		clients[i] = client
	}

	// Test concurrent project listing
	done := make(chan bool, numClients)
	errors := make(chan error, numClients)

	for i, client := range clients {
		go func(idx int, c *hydra.Client) {
			projects, err := c.ListProjects(ctx)
			if err != nil {
				errors <- err
				return
			}
			
			// Verify we got some response
			if projects == nil {
				errors <- fmt.Errorf("client %d got nil projects", idx)
				return
			}
			
			done <- true
		}(i, client)
	}

	// Wait for all goroutines to complete
	completed := 0
	for completed < numClients {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent access error: %v", err)
		case <-done:
			completed++
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}

	t.Logf("Successfully completed concurrent access test with %d clients", numClients)
}

// NewClientWithURL creates a new Hydra client for the container
func (h *HydraContainer) NewClientWithURL(baseURL string) (*hydra.Client, error) {
	return hydra.NewClientWithURL(baseURL)
}