package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conneroisu/hydra-go/hydra"
	"github.com/conneroisu/hydra-go/hydra/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *hydra.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &hydra.Config{
				BaseURL: "https://hydra.nixos.org",
			},
			wantErr: false,
		},
		{
			name: "custom config",
			config: &hydra.Config{
				BaseURL:   "https://custom.hydra.instance",
				UserAgent: "test-client/1.0",
				Timeout:   60,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := hydra.NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.Auth)
				assert.NotNil(t, client.Projects)
				assert.NotNil(t, client.Jobsets)
				assert.NotNil(t, client.Builds)
				assert.NotNil(t, client.Search)
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"name":"test-project","displayname":"Test Project","owner":"test"}]`))
		case "/project/test-project":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"test-project","displayname":"Test Project","owner":"test","enabled":true}`))
		case "/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"projects":[],"jobsets":[],"builds":[],"buildsdrv":[]}`))
		case "/build/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":123,"nixname":"test-build","finished":true,"buildstatus":0}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	client, err := hydra.NewClientWithURL(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("ListProjects", func(t *testing.T) {
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, "test-project", projects[0].Name)
	})

	t.Run("GetProject", func(t *testing.T) {
		project, err := client.GetProject(ctx, "test-project")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "test-project", project.Name)
		assert.True(t, project.Enabled)
	})

	t.Run("SearchAll", func(t *testing.T) {
		results, err := client.SearchAll(ctx, "test")
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results.Projects)
	})

	t.Run("GetBuild", func(t *testing.T) {
		build, err := client.GetBuild(ctx, 123)
		assert.NoError(t, err)
		assert.NotNil(t, build)
		assert.Equal(t, 123, build.ID)
		assert.True(t, build.Finished)
		assert.True(t, build.IsSuccess())
	})
}

func TestAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" && r.Method == http.MethodPost {
			// Set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "hydra_session",
				Value: "test-session",
				Path:  "/",
			})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"username":"testuser","fullname":"Test User","emailaddress":"test@example.com","userroles":["user"]}`))
		} else {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		}
	}))
	defer server.Close()

	client, err := hydra.NewClientWithURL(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Login", func(t *testing.T) {
		user, err := client.Login(ctx, "testuser", "testpass")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "Test User", user.FullName)
		assert.True(t, client.IsAuthenticated())
	})

	t.Run("Logout", func(t *testing.T) {
		client.Logout()
		assert.False(t, client.IsAuthenticated())
	})
}

func TestBuildHelpers(t *testing.T) {
	build := &models.Build{
		ID:        123,
		StartTime: 1000,
		StopTime:  1300,
		Timestamp: 900,
		Finished:  true,
		NixName:   "test-package-1.0",
	}

	t.Run("BuildStatus", func(t *testing.T) {
		// Test successful build
		status := models.BuildStatusSuccess
		build.BuildStatus = &status
		assert.True(t, build.IsSuccess())
		assert.False(t, build.IsFailed())
		assert.Equal(t, "succeeded", build.GetBuildStatusString())

		// Test failed build
		status = models.BuildStatusFailed
		build.BuildStatus = &status
		assert.False(t, build.IsSuccess())
		assert.True(t, build.IsFailed())
		assert.Equal(t, "failed", build.GetBuildStatusString())

		// Test in progress
		build.Finished = false
		build.BuildStatus = nil
		assert.Equal(t, "in progress", build.GetBuildStatusString())
	})

	t.Run("BuildTimes", func(t *testing.T) {
		duration := build.GetDuration()
		assert.Equal(t, 300, int(duration.Seconds()))

		startTime := build.GetStartTime()
		assert.Equal(t, int64(1000), startTime.Unix())

		stopTime := build.GetStopTime()
		assert.Equal(t, int64(1300), stopTime.Unix())
	})
}

func TestJobsetHelpers(t *testing.T) {
	jobset := &models.Jobset{
		Name:    "test-jobset",
		Project: "test-project",
		Enabled: 1,
	}

	t.Run("JobsetState", func(t *testing.T) {
		assert.True(t, jobset.IsEnabled())
		assert.Equal(t, models.JobsetStateEnabled, jobset.GetState())

		jobset.SetState(models.JobsetStateDisabled)
		assert.False(t, jobset.IsEnabled())
		assert.Equal(t, 0, jobset.Enabled)

		jobset.SetState(models.JobsetStateOneShot)
		assert.Equal(t, 2, jobset.Enabled)
	})
}
