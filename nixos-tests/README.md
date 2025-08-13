# NixOS VM Tests for Hydra Go SDK

This directory contains comprehensive NixOS VM tests for the Hydra Go SDK. These tests ensure that the SDK works correctly in a real environment with a mock Hydra server.

## Test Architecture

The test suite consists of:

1. **NixOS VM Configuration** (`hydra-sdk-test.nix`)
   - Sets up two VMs: a mock Hydra server and a test client
   - Configures networking between VMs
   - Runs comprehensive SDK tests

2. **Mock Hydra Server** (`mock-server/`)
   - Implements all Hydra API endpoints
   - Uses test fixtures for consistent data
   - Supports authentication and session management

3. **Test Fixtures** (`fixtures/`)
   - Pre-defined test data for projects, jobsets, builds, and users
   - Ensures consistent and predictable test results

4. **Integration Tests** (`../tests/`)
   - Comprehensive test coverage for all SDK features
   - Tests authentication, CRUD operations, search, and error handling
   - Includes concurrency and performance tests

## Running the Tests

### Quick Start

```bash
# Run the VM tests
./run-vm-test.sh
```

### Manual Execution

```bash
# Build the test
nix-build hydra-sdk-test.nix

# Run interactively
nix-build hydra-sdk-test.nix -A driverInteractive
./result/bin/nixos-test-driver
```

### Running Specific Tests

```bash
# Run only unit tests
cd ..
go test -v ./hydra/...

# Run only integration tests
go test -v -tags=integration ./tests/...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Coverage

The test suite covers:

### Core Functionality
- ✅ Client creation and configuration
- ✅ HTTP client with session management
- ✅ Error handling and retries
- ✅ Request/response serialization

### Authentication
- ✅ Login with credentials
- ✅ Session management
- ✅ Logout functionality
- ✅ Authentication state tracking

### Projects API
- ✅ List all projects
- ✅ Get specific project
- ✅ Create new project
- ✅ Update existing project
- ✅ Delete project

### Jobsets API
- ✅ List jobsets for project
- ✅ Get specific jobset
- ✅ Create/update jobset
- ✅ Delete jobset
- ✅ Trigger evaluation
- ✅ Get evaluations

### Builds API
- ✅ Get build by ID
- ✅ Get build constituents
- ✅ Get evaluation builds
- ✅ Wait for build completion
- ✅ Filter builds
- ✅ Calculate statistics

### Search API
- ✅ Search across all resources
- ✅ Search projects
- ✅ Search jobsets
- ✅ Search builds

### Advanced Features
- ✅ QuickStart helper methods
- ✅ Concurrent operations
- ✅ Error handling
- ✅ Network timeouts
- ✅ Session persistence

## Mock Server Endpoints

The mock server implements:

- `GET /health` - Health check
- `POST /login` - Authentication
- `GET /` - List projects
- `GET /project/:id` - Get project
- `PUT /project/:id` - Create/update project
- `DELETE /project/:id` - Delete project
- `GET /jobset/:project/:jobset` - Get jobset
- `PUT /jobset/:project/:jobset` - Create/update jobset
- `DELETE /jobset/:project/:jobset` - Delete jobset
- `GET /jobset/:project/:jobset/evals` - Get evaluations
- `GET /build/:id` - Get build
- `GET /build/:id/constituents` - Get build constituents
- `GET /eval/:id` - Get evaluation
- `GET /eval/:id/builds` - Get evaluation builds
- `GET /search` - Search
- `POST /api/push` - Trigger evaluation
- `GET /api/jobsets` - List jobsets

## Test Data

The fixture file contains:
- 3 projects (nixpkgs, hydra, test-project)
- 4 jobsets across different projects
- 5 builds with various statuses
- 2 users for authentication testing

## Extending the Tests

To add new tests:

1. **Add fixtures**: Update `fixtures/test-data.json` with new test data
2. **Update mock server**: Add new endpoints to `mock-server/server.go`
3. **Write tests**: Add test cases to `tests/integration_test.go`
4. **Run tests**: Execute `./run-vm-test.sh` to verify

## Troubleshooting

### Tests Failing

1. Check mock server is running:
   ```bash
   curl http://localhost:8080/health
   ```

2. Verify network connectivity in VM:
   ```bash
   nix-build hydra-sdk-test.nix -A driverInteractive
   ./result/bin/nixos-test-driver
   > hydraServer.succeed("curl http://localhost:8080/health")
   > testClient.succeed("curl http://192.168.1.10:8080/health")
   ```

3. Check test logs:
   ```bash
   journalctl -u mock-hydra-server
   ```

### Performance Issues

- Increase VM memory in `hydra-sdk-test.nix`
- Adjust timeout values in tests
- Run tests with `-parallel` flag

## CI Integration

To integrate with CI:

```yaml
# GitHub Actions example
- name: Run NixOS VM Tests
  run: |
    nix-build nixos-tests/hydra-sdk-test.nix
    ./result/bin/run-tests
```

## License

Same as the parent project - see LICENSE file.