# hydra-go

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/conneroisu/hydra-go)](https://goreportcard.com/report/github.com/conneroisu/hydra-go)

**hydra-go** is an idiomatic Go SDK for the [Nix Hydra](https://nixos.org/hydra/) continuous integration and build system. This SDK provides a clean, type-safe interface to interact with Hydra's REST API, supporting both public and private Hydra instances.

## 🚀 Flat Architecture Design

hydra-go uses a **flat architecture** where all client functionality is imported from a single package `github.com/conneroisu/hydra-go`. This design follows Go best practices for library ergonomics, eliminating the need for multiple import statements and providing a clean, discoverable API surface.

## Features

- ✨ **Full API Coverage**: Complete support for projects, jobsets, builds, evaluations, and search
- 🏗️ **Flat Architecture**: Single import for all functionality - no sub-package imports needed
- 🔐 **Authentication Support**: Login/logout with session management
- 🚀 **Idiomatic Go**: Type-safe interfaces with proper error handling
- 📦 **Zero Dependencies**: Uses only Go standard library plus testify for tests
- 🎯 **Context-Aware**: All operations support context.Context for cancellation and timeouts
- 🔧 **Configurable**: Flexible client configuration with sensible defaults
- 📖 **Well-Tested**: Comprehensive test suite with integration tests
- 🌐 **Multi-Instance**: Support for both public hydra.nixos.org and private instances

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/conneroisu/hydra-go"
)

func main() {
    // Create a client for the public Hydra instance
    client, err := hydra.NewDefaultClient()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // List all projects
    projects, err := client.ListProjects(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, project := range projects {
        fmt.Printf("Project: %s (%s)\n", project.Name, project.DisplayName)
    }
    
    // Search for builds
    results, err := client.Search(ctx, "nixpkgs")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d results\n", len(results.Builds))
}
```

## Installation

Add hydra-go to your Go module:

```bash
go get github.com/conneroisu/hydra-go
```

Or import directly in your Go code:

```go
import "github.com/conneroisu/hydra-go"
```

### Requirements

- Go 1.24 or later
- Internet access to reach Hydra instances

## Usage

### Basic Client Setup

```go
// Default client for hydra.nixos.org
client, err := hydra.NewDefaultClient()

// Custom Hydra instance
client, err := hydra.NewClientWithURL("https://my-hydra.example.com")

// Advanced configuration
config := &hydra.Config{
    BaseURL:    "https://hydra.nixos.org",
    Timeout:    time.Second * 30,
    UserAgent:  "MyApp/1.0",
    HTTPClient: &http.Client{...}, // optional custom HTTP client
}
client, err := hydra.NewClient(config)
```

### Authentication

```go
// Login with username/password
user, err := client.Login(ctx, "username", "password")
if err != nil {
    log.Fatal(err)
}

// Check authentication status
if client.IsAuthenticated() {
    fmt.Println("Authenticated!")
}

// Logout when done
defer client.Logout()
```

### Working with Projects

```go
// List all projects
projects, err := client.ListProjects(ctx)

// Get a specific project
project, err := client.GetProject(ctx, "nixpkgs")

// Create a project (requires authentication)
opts := hydra.NewCreateProjectOptions("my-project", "owner").
    WithDisplayName("My Project").
    WithDescription("A test project").
    WithEnabled(true)

resp, err := client.CreateProjectWithOptions(ctx, "my-project", opts)
```

### Working with Jobsets

```go
// List jobsets for a project
jobsets, err := client.ListJobsets(ctx, "nixpkgs")

// Get a specific jobset
jobset, err := client.GetJobset(ctx, "nixpkgs", "trunk")

// Get evaluations for a jobset
evaluations, err := client.GetJobsetEvaluations(ctx, "nixpkgs", "trunk")

// Create a jobset (requires authentication)
opts := hydra.NewCreateJobsetOptions("test-jobset", "my-project").
    WithDescription("Test jobset").
    WithNixExpression("nixpkgs", "release.nix").
    AddInput("nixpkgs", "git", "https://github.com/NixOS/nixpkgs.git", false)

resp, err := client.CreateJobsetWithOptions(ctx, "my-project", "test-jobset", opts)

// Trigger evaluation
pushResp, err := client.TriggerJobset(ctx, "my-project", "test-jobset")
```

### Working with Builds

```go
// Get a specific build
build, err := client.GetBuild(ctx, 12345)

// Get builds for an evaluation
builds, err := client.GetEvaluationBuilds(ctx, evalID)

// Check build status
if build.Finished {
    fmt.Printf("Build duration: %v\n", build.GetDuration())
    fmt.Printf("Build status: %s\n", build.GetBuildStatusString())
}
```

### Search Operations

```go
// Search all resource types
results, err := client.Search(ctx, "hello")

// Get search summary
summary := hydra.GetSearchSummary("hello", results)
fmt.Println(summary.Format())

// Advanced search with options
opts := hydra.NewSearchOptions("firefox")
results, err := client.SearchWithOptions(ctx, opts)
```

### QuickStart Helper

The SDK includes a QuickStart helper for common operations:

```go
quick := client.Quick()

// Get project with all its jobsets
project, jobsets, err := quick.GetProjectWithJobsets(ctx, "nixpkgs")

// Get latest build for a specific job
build, err := quick.GetLatestBuildForJob(ctx, "nixpkgs", "trunk", "hello")

// Trigger evaluation and wait for completion
eval, err := quick.WaitForJobsetEvaluation(ctx, "my-project", "my-jobset", time.Minute*5)
```

## API Documentation

### Core Types

The SDK provides comprehensive type definitions for all Hydra resources:

#### Project

```go
type Project struct {
    Name         string   `json:"name"`
    DisplayName  string   `json:"displayname"`
    Description  string   `json:"description"`
    Homepage     string   `json:"homepage"`
    Owner        string   `json:"owner"`
    Enabled      bool     `json:"enabled"`
    Hidden       bool     `json:"hidden"`
    Jobsets      []string `json:"jobsets"`
}
```

#### Jobset

```go
type Jobset struct {
    Project         string                 `json:"project"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    NixExpression   NixExpression         `json:"nixexpression"`
    Enabled         int                    `json:"enabled"`
    EnableEmail     bool                  `json:"enableemail"`
    EmailOverride   string                `json:"emailoverride"`
    Hidden          bool                  `json:"hidden"`
    SchedulingShares int                   `json:"schedulingshares"`
    CheckInterval   int                    `json:"checkinterval"`
    Inputs          map[string]JobsetInput `json:"inputs"`
    KeepEvaluations int                    `json:"keepnr"`
}
```

#### Build

```go
type Build struct {
    ID              int                    `json:"id"`
    Finished        bool                  `json:"finished"`
    Timestamp       int64                 `json:"timestamp"`
    Job             string                `json:"job"`
    Jobset          string                `json:"jobset"`
    Project         string                `json:"project"`
    System          string                `json:"system"`
    Priority        int                   `json:"priority"`
    Busy            int                   `json:"busy"`
    BuildStatus     *int                  `json:"buildstatus"`
    StartTime       *int64                `json:"starttime"`
    StopTime        *int64                `json:"stoptime"`
    NixName         string                `json:"nixname"`
    DrvPath         string                `json:"drvpath"`
    Outputs         map[string]BuildOutput `json:"buildoutputs"`
    BuildProducts   []BuildProduct         `json:"buildproducts"`
}
```

### Client Services

The main client provides access to specialized services:

#### Projects Service

```go
// List all projects
projects, err := client.Projects.List(ctx)

// Get specific project
project, err := client.Projects.Get(ctx, "nixpkgs")

// Create project with options
opts := projects.NewCreateOptions("name", "owner")
response, err := client.Projects.CreateWithOptions(ctx, "name", opts)

// Update project
response, err := client.Projects.Update(ctx, "name", updateOpts)

// Delete project
err := client.Projects.Delete(ctx, "name")
```

#### Jobsets Service

```go
// List jobsets for a project
jobsets, err := client.Jobsets.List(ctx, "nixpkgs")

// Get specific jobset
jobset, err := client.Jobsets.Get(ctx, "nixpkgs", "trunk")

// Get evaluations
evaluations, err := client.Jobsets.GetEvaluations(ctx, "nixpkgs", "trunk")

// Create jobset
opts := jobsets.NewJobsetOptions("name", "project")
response, err := client.Jobsets.CreateWithOptions(ctx, "project", "name", opts)

// Trigger evaluation
response, err := client.Jobsets.TriggerSingle(ctx, "project", "jobset")
```

#### Builds Service

```go
// Get specific build
build, err := client.Builds.Get(ctx, 12345)

// Get builds for evaluation
builds, err := client.Builds.GetEvaluationBuilds(ctx, evalID)

// Get build log (if available)
log, err := client.Builds.GetLog(ctx, buildID)
```

#### Search Service

```go
// Search across all resource types
results, err := client.Search.Search(ctx, "query")

// Get formatted search summary
summary := search.GetSearchSummary("query", results)
fmt.Println(summary.Format())
```

#### Auth Service

```go
// Login
user, err := client.Auth.Login(ctx, "username", "password")

// Check authentication
authenticated := client.Auth.IsAuthenticated()

// Get current user info
user, err := client.Auth.GetCurrentUser(ctx)

// Logout
client.Auth.Logout()
```

### Error Handling

The SDK provides structured error handling:

```go
build, err := client.GetBuild(ctx, 12345)
if err != nil {
    switch {
    case errors.Is(err, hydra.ErrNotFound):
        fmt.Println("Build not found")
    case errors.Is(err, hydra.ErrUnauthorized):
        fmt.Println("Authentication required")
    case errors.Is(err, hydra.ErrTimeout):
        fmt.Println("Request timed out")
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Examples

Complete examples are available in the [`examples/`](./examples/) directory:

- [`examples/basic/main.go`](./examples/basic/main.go) - Basic usage without authentication
- [`examples/authenticated/main.go`](./examples/authenticated/main.go) - Authenticated operations and project management

Run examples:

```bash
# Basic example
go run examples/basic/main.go

# Authenticated example (requires credentials)
HYDRA_USERNAME=user HYDRA_PASSWORD=pass go run examples/authenticated/main.go

# Basic example with specific build ID
go run examples/basic/main.go 12345
```

## Testing

The project includes comprehensive tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run integration tests (requires network access)
go test -tags=integration ./tests/

# Run tests against a real Hydra instance
HYDRA_URL=https://my-hydra.example.com go test ./tests/
```

### Test Scripts

The project provides several test scripts for different scenarios:

- `./run-tests.sh` - Run basic unit tests
- `./test-simple.sh` - Quick smoke tests
- `./test-with-docker.sh` - Tests with Docker-based Hydra instance
- `./test-against-nixos-hydra.sh` - Tests against the public Hydra instance
- `./verify-tests.sh` - Comprehensive test verification

## Development

### Project Structure

```
hydra-go/
├── hydra/                 # Main SDK package
│   ├── auth/             # Authentication service
│   ├── builds/           # Builds service  
│   ├── client/           # HTTP client
│   ├── jobsets/          # Jobsets service
│   ├── models/           # Type definitions
│   ├── projects/         # Projects service
│   ├── search/           # Search service
│   └── hydra.go          # Main client
├── examples/             # Usage examples
│   ├── basic/           # Basic usage
│   └── authenticated/   # Authenticated operations
├── tests/               # Test suite
├── nixos-tests/         # NixOS-specific tests
└── README.md
```

### Building

```bash
# Build the project
go build ./...

# Build examples
go build ./examples/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Style

This project follows Go best practices and the coding philosophy outlined in the project's CLAUDE.md:

- **Safety First**: Explicit error handling, bounded operations, type safety
- **Performance**: Efficient resource usage, predictable execution paths
- **Developer Experience**: Clear naming, logical organization, comprehensive documentation

### Dependencies

The project uses minimal external dependencies:
- **Go standard library** - Core functionality
- **github.com/stretchr/testify** - Testing utilities (test-only)

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the existing code style
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Reporting Issues

Please report bugs and feature requests using [GitHub Issues](https://github.com/conneroisu/hydra-go/issues).

When reporting issues, please include:
- Go version (`go version`)
- Operating system
- Hydra server version (if known)
- Minimal code example demonstrating the issue
- Error messages and stack traces

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Nix Hydra](https://nixos.org/hydra/) - The continuous integration and build system
- [NixOS Community](https://nixos.org/) - For maintaining the public Hydra instance
- Go community for excellent tooling and libraries

## Support

- **Documentation**: This README and inline code documentation
- **Examples**: See the [`examples/`](./examples/) directory
- **Issues**: [GitHub Issues](https://github.com/conneroisu/hydra-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/conneroisu/hydra-go/discussions)

---

Built with ❤️ for the Nix and Go communities.
