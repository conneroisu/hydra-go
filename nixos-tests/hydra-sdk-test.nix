# NixOS VM test for Hydra Go SDK
{pkgs ? import <nixpkgs> {}}:
pkgs.nixosTest {
  name = "hydra-go-sdk-test";

  nodes = {
    # Mock Hydra server node
    hydraServer = {
      config,
      pkgs,
      ...
    }: {
      imports = [];

      # Basic system configuration
      virtualisation.memorySize = 2048;
      virtualisation.diskSize = 4096;

      # Network configuration
      networking = {
        firewall.allowedTCPPorts = [8080];
        interfaces.eth1 = {
          ipv4.addresses = [
            {
              address = "192.168.1.10";
              prefixLength = 24;
            }
          ];
        };
      };

      # Install required packages
      environment.systemPackages = with pkgs; [
        go_1_24
        git
        curl
        jq
        netcat-gnu
      ];

      # Mock Hydra server service
      systemd.services.mock-hydra-server = {
        description = "Mock Hydra Server";
        wantedBy = ["multi-user.target"];
        after = ["network.target"];

        serviceConfig = {
          Type = "simple";
          ExecStart = "${pkgs.go_1_24}/bin/go run /etc/mock-hydra-server/server.go";
          Restart = "always";
          RestartSec = 5;
          WorkingDirectory = "/etc/mock-hydra-server";
          Environment = [
            "PORT=8080"
            "HOST=0.0.0.0"
          ];
        };
      };

      # Copy mock server code
      environment.etc."mock-hydra-server/server.go".source = ./mock-server/server.go;
      environment.etc."mock-hydra-server/handlers.go".source = ./mock-server/handlers.go;
      environment.etc."mock-hydra-server/models.go".source = ./mock-server/models.go;
      environment.etc."mock-hydra-server/fixtures.json".source = ./fixtures/test-data.json;
    };

    # Test client node
    testClient = {
      config,
      pkgs,
      ...
    }: {
      imports = [];

      virtualisation.memorySize = 2048;
      virtualisation.diskSize = 4096;

      # Network configuration
      networking = {
        interfaces.eth1 = {
          ipv4.addresses = [
            {
              address = "192.168.1.20";
              prefixLength = 24;
            }
          ];
        };
      };

      # Install required packages
      environment.systemPackages = with pkgs; [
        go_1_24
        git
        curl
        gnumake
      ];

      # Copy SDK source code and tests
      environment.etc."hydra-go-sdk".source = ../..;
    };
  };

  testScript = ''
    import json
    import time

    # Start all nodes
    start_all()

    # Wait for network to be ready
    hydraServer.wait_for_unit("network.target")
    testClient.wait_for_unit("network.target")

    # Wait for mock Hydra server to start
    hydraServer.wait_for_unit("mock-hydra-server.service")
    hydraServer.wait_for_open_port(8080)

    # Verify server is responding
    hydraServer.succeed("curl -f http://localhost:8080/health")
    testClient.succeed("curl -f http://192.168.1.10:8080/health")

    print("Mock Hydra server is running")

    # Run SDK unit tests
    print("Running SDK unit tests...")
    testClient.succeed("""
      cd /etc/hydra-go-sdk
      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go test -v ./hydra/... -tags=integration
    """)

    # Run integration tests
    print("Running SDK integration tests...")
    testClient.succeed("""
      cd /etc/hydra-go-sdk
      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go test -v ./tests/... -tags=integration
    """)

    # Test authentication flow
    print("Testing authentication...")
    auth_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > auth_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()
        user, err := client.Login(ctx, "testuser", "testpass")
        if err != nil {
          log.Fatal(err)
        }

        if !client.IsAuthenticated() {
          log.Fatal("Client should be authenticated")
        }

        fmt.Printf("Authenticated as: %s\\n", user.Username)

        client.Logout()
        if client.IsAuthenticated() {
          log.Fatal("Client should not be authenticated after logout")
        }

        fmt.Println("Authentication test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run auth_test.go
    """)
    print(f"Authentication test output: {auth_test}")

    # Test project operations
    print("Testing project operations...")
    project_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > project_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()

        // List projects
        projects, err := client.ListProjects(ctx)
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Found %d projects\\n", len(projects))

        // Get specific project
        if len(projects) > 0 {
          project, err := client.GetProject(ctx, projects[0].Name)
          if err != nil {
            log.Fatal(err)
          }
          fmt.Printf("Got project: %s\\n", project.Name)
        }

        fmt.Println("Project operations test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run project_test.go
    """)
    print(f"Project test output: {project_test}")

    # Test jobset operations
    print("Testing jobset operations...")
    jobset_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > jobset_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()

        // Get jobset
        jobset, err := client.GetJobset(ctx, "nixpkgs", "trunk")
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Got jobset: %s/%s\\n", jobset.Project, jobset.Name)

        // Trigger evaluation
        response, err := client.TriggerEvaluation(ctx, "nixpkgs", "trunk")
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Triggered evaluation: %v\\n", response)

        fmt.Println("Jobset operations test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run jobset_test.go
    """)
    print(f"Jobset test output: {jobset_test}")

    # Test build operations
    print("Testing build operations...")
    build_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > build_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()

        // Get build
        build, err := client.GetBuild(ctx, 123456)
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Got build: %d - %s\\n", build.ID, build.NixName)

        if build.IsSuccess() {
          fmt.Println("Build succeeded")
        }

        fmt.Println("Build operations test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run build_test.go
    """)
    print(f"Build test output: {build_test}")

    # Test search operations
    print("Testing search operations...")
    search_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > search_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()

        // Search
        results, err := client.SearchAll(ctx, "hello")
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Search found %d projects, %d jobsets, %d builds\\n",
          len(results.Projects), len(results.Jobsets), len(results.Builds))

        fmt.Println("Search operations test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run search_test.go
    """)
    print(f"Search test output: {search_test}")

    # Test error handling
    print("Testing error handling...")
    error_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > error_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()

        // Test 404 error
        _, err = client.GetProject(ctx, "nonexistent")
        if err == nil {
          log.Fatal("Expected error for non-existent project")
        }
        fmt.Printf("Got expected error: %v\\n", err)

        // Test invalid build ID
        _, err = client.GetBuild(ctx, -1)
        if err == nil {
          log.Fatal("Expected error for invalid build ID")
        }
        fmt.Printf("Got expected error: %v\\n", err)

        fmt.Println("Error handling test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run error_test.go
    """)
    print(f"Error handling test output: {error_test}")

    # Test concurrent operations
    print("Testing concurrent operations...")
    concurrent_test = testClient.succeed("""
      cd /etc/hydra-go-sdk
      cat > concurrent_test.go << 'EOF'
      package main

      import (
        "context"
        "fmt"
        "log"
        "os"
        "sync"
        "time"

        "github.com/conneroisu/hydra-go/hydra"
      )

      func main() {
        client, err := hydra.NewClientWithURL(os.Getenv("HYDRA_TEST_URL"))
        if err != nil {
          log.Fatal(err)
        }

        ctx := context.Background()
        var wg sync.WaitGroup
        errors := make(chan error, 10)

        // Run 10 concurrent requests
        for i := 0; i < 10; i++ {
          wg.Add(1)
          go func(id int) {
            defer wg.Done()

            _, err := client.ListProjects(ctx)
            if err != nil {
              errors <- err
              return
            }
            fmt.Printf("Request %d completed\\n", id)
          }(i)
        }

        done := make(chan bool)
        go func() {
          wg.Wait()
          close(done)
        }()

        select {
        case <-done:
          fmt.Println("All concurrent requests completed successfully")
        case err := <-errors:
          log.Fatalf("Concurrent request failed: %v", err)
        case <-time.After(10 * time.Second):
          log.Fatal("Timeout waiting for concurrent requests")
        }

        fmt.Println("Concurrent operations test passed")
      }
      EOF

      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go run concurrent_test.go
    """)
    print(f"Concurrent operations test output: {concurrent_test}")

    # Run comprehensive test suite
    print("Running comprehensive test suite...")
    testClient.succeed("""
      cd /etc/hydra-go-sdk/tests
      export HYDRA_TEST_URL=http://192.168.1.10:8080
      go test -v -race -coverprofile=coverage.out ./...
      go tool cover -html=coverage.out -o coverage.html
      echo "Test coverage report generated"
    """)

    print("All tests passed successfully!")
  '';
}
