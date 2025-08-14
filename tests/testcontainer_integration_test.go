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
	// Use the GHCR image that's built by CI
	defaultHydraImage = "ghcr.io/conneroisu/hydra-go/hydra-test:dev"
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
	// Use the GHCR image built by CI
	image := defaultHydraImage
	if t != nil {
		t.Logf("Using GHCR Hydra image: %s", image)
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{hydraPort, healthPort},
		ImagePlatform: "linux/amd64", // Force platform for consistency
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
	if err != nil {
		if t != nil {
			require.NoError(t, err)
		} else {
			panic(fmt.Sprintf("Failed to start Hydra container: %v", err))
		}
	}

	// Get the host and port
	host, err := hydraContainer.Host(ctx)
	if err != nil {
		if t != nil {
			require.NoError(t, err)
		} else {
			panic(fmt.Sprintf("Failed to get container host: %v", err))
		}
	}

	mappedPort, err := hydraContainer.MappedPort(ctx, hydraPort)
	if err != nil {
		if t != nil {
			require.NoError(t, err)
		} else {
			panic(fmt.Sprintf("Failed to get mapped port: %v", err))
		}
	}

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


func TestHydraContainerStartup(t *testing.T) {
	ctx := context.Background()

	container := StartHydraContainer(ctx, t)
	defer container.Terminate(ctx)

	t.Logf("Hydra container started at: %s", container.BaseURL)

	// Test that we can create a client and connect
	client, err := hydra.NewClientWithURL(container.BaseURL)
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

	container := StartHydraContainer(ctx, t)
	defer container.Terminate(ctx)

	client, err := hydra.NewClientWithURL(container.BaseURL)
	require.NoError(t, err)

	// Test authentication with the default admin user
	user, err := client.Login(ctx, "admin", "admin")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "Admin", user.FullName)
	assert.True(t, client.IsAuthenticated())

	t.Logf("Successfully authenticated as: %s", user.Username)

	// Test invalid credentials
	_, err = client.Login(ctx, "invalid", "wrong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestHydraContainerFullWorkflow(t *testing.T) {
	ctx := context.Background()

	container := StartHydraContainer(ctx, t)
	defer container.Terminate(ctx)

	client, err := hydra.NewClientWithURL(container.BaseURL)
	require.NoError(t, err)

	// Test authentication
	_, err = client.Login(ctx, "admin", "admin")
	require.NoError(t, err)
	assert.True(t, client.IsAuthenticated())

	// Test basic functionality without CRUD operations
	projects, err := client.ListProjects(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, projects)

	// Test individual project retrieval
	project, err := client.GetProject(ctx, "nixpkgs")
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "nixpkgs", project.Name)

	// Test logout
	client.Logout()
	assert.False(t, client.IsAuthenticated())

	t.Logf("Successfully completed full workflow test")
}

func TestHydraContainerConcurrentAccess(t *testing.T) {
	ctx := context.Background()

	container := StartHydraContainer(ctx, t)
	defer container.Terminate(ctx)

	// Create multiple clients to test concurrent access
	numClients := 5
	clients := make([]*hydra.Client, numClients)
	
	for i := 0; i < numClients; i++ {
		client, err := hydra.NewClientWithURL(container.BaseURL)
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