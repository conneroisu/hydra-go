{
  description = "Hydra Go - Nix Builder Go SDK";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    treefmt-nix.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    treefmt-nix,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [
          (final: prev: {
            # Add your overlays here
            # Example:
            # my-overlay = final: prev: {
            #   my-package = prev.callPackage ./my-package { };
            # };
            final.buildGoModule = prev.buildGo124Module;
          })
        ];
      };

      rooted = exec:
        builtins.concatStringsSep "\n"
        [
          ''            # shellcheck disable=SC2034
                      REPO_ROOT="$(git rev-parse --show-toplevel)"''
          exec
        ];

      scripts = {
        dx = {
          exec = rooted ''$EDITOR "$REPO_ROOT"/flake.nix'';
          description = "Edit flake.nix";
        };
        gx = {
          exec = rooted ''$EDITOR "$REPO_ROOT"/go.mod'';
          description = "Edit go.mod";
        };
        lint = {
          exec = rooted ''
            golangci-lint run --fix
          '';
          description = "Lint";
        };
        tests = {
          exec = rooted ''
            set -euo pipefail

            echo "========================================="
            echo "Running Comprehensive Test Suite"
            echo "========================================="
            echo ""

            # Colors for output
            GREEN='\033[0;32m'
            RED='\033[0;31m'
            BLUE='\033[0;34m'
            NC='\033[0m' # No Color

            FAILED_TESTS=""

            # Phase 1: Go Unit Tests
            echo -e "$BLUE"Phase 1: Running Go unit tests..."$NC"
            echo "-----------------------------------"
            if go test -v ./...; then
              echo -e "$GREEN"+ Go unit tests passed"$NC"
            else
              echo -e "$RED"- Go unit tests failed"$NC"
              FAILED_TESTS="$FAILED_TESTS go-unit"
            fi
            echo ""

            # Phase 2: Go Integration Tests
            echo -e "$BLUE"Phase 2: Running Go integration tests..."$NC"
            echo "----------------------------------------"
            if go test -v -tags=integration ./tests/...; then
              echo -e "$GREEN"+ Go integration tests passed"$NC"
            else
              echo -e "$RED"- Go integration tests failed"$NC"
              FAILED_TESTS="$FAILED_TESTS go-integration"
            fi
            echo ""

            # Phase 3: NixOS VM Tests
            echo -e "$BLUE"Phase 3: Running NixOS VM tests..."$NC"
            echo "----------------------------------"

            # List of NixOS tests to run in order of complexity
            NIXOS_TESTS="ultra-simple-test.nix minimal-hydra-test.nix basic-hydra-test.nix working-hydra-test.nix hydra-sdk-test.nix real-sdk-test.nix final-test.nix"

            NIXOS_FAILED=""

            for test_file in $NIXOS_TESTS; do
              echo "Running $test_file..."
              if nix-build "nixos-tests/$test_file" -o "result-$test_file" --quiet; then
                echo -e "$GREEN"+ "$test_file" passed"$NC"
                rm -f "result-$test_file"
              else
                echo -e "$RED"- "$test_file" failed"$NC"
                NIXOS_FAILED="$NIXOS_FAILED $test_file"
                FAILED_TESTS="$FAILED_TESTS nixos-$test_file"
              fi
            done

            if [ -z "$NIXOS_FAILED" ]; then
              echo -e "$GREEN"+ All NixOS tests passed"$NC"
            else
              echo -e "$RED"- NixOS tests failed:"$NIXOS_FAILED""$NC"
            fi
            echo ""

            # Final Summary
            echo "========================================="
            echo "Test Suite Summary"
            echo "========================================="

            if [ -z "$FAILED_TESTS" ]; then
              echo -e "$GREEN"ALL TESTS PASSED!"$NC"
              echo ""
              echo "* Go unit tests"
              echo "* Go integration tests"
              echo "* NixOS VM tests"
              echo ""
              echo "The Hydra Go SDK has been comprehensively tested and is ready for use!"
              exit 0
            else
              echo -e "$RED"SOME TESTS FAILED"$NC"
              echo ""
              echo "Failed test suites:$FAILED_TESTS"
              echo ""
              echo "Run individual test phases for debugging:"
              echo "  • Go tests only: go test -v ./..."
              echo "  • Integration tests: go test -v -tags=integration ./tests/..."
              echo "  • Specific NixOS test: nix-build nixos-tests/<test-name>.nix"
              exit 1
            fi
          '';
          description = "Run comprehensive tests (Go + NixOS VM)";
        };
        tests-integration = {
          exec = rooted ''
            echo "Running integration tests with testcontainers..."
            go test -v -tags=integration ./tests/...
          '';
          description = "Run integration tests with testcontainers";
        };
        build-hydra-image = {
          exec = rooted ''
            echo "Building Hydra Docker image..."
            docker build -t ghcr.io/conneroisu/hydra-go/hydra-test:latest "$REPO_ROOT"/docker/
          '';
          description = "Build Hydra Docker image locally";
        };
        test-with-container = {
          exec = rooted ''
            echo "Building Hydra image and running tests..."
            docker build -t ghcr.io/conneroisu/hydra-go/hydra-test:latest "$REPO_ROOT"/docker/
            go test -v -tags=integration ./tests/...
          '';
          description = "Build container and run integration tests";
        };
      };

      scriptPackages =
        pkgs.lib.mapAttrs
        (
          name: script:
            pkgs.writeShellApplication {
              inherit name;
              text = script.exec;
              runtimeInputs = script.deps or [];
            }
        )
        scripts;

      treefmtModule = {
        projectRootFile = "flake.nix";
        programs = {
          alejandra.enable = true; # Nix formatter
        };
      };
    in {
      devShells.default = pkgs.mkShell {
        name = "dev";

        # Available packages on https://search.nixos.org/packages
        packages = with pkgs;
          [
            alejandra # Nix
            nixd
            statix
            deadnix

            go_1_24 # Go Tools
            air
            golangci-lint
            gopls
            revive
            golines
            golangci-lint-langserver
            gomarkdoc
            gotests
            gotools
            reftools
            pprof
            graphviz
            goreleaser
            cobra-cli
          ]
          ++ builtins.attrValues scriptPackages;
      };

      packages = {
        default = pkgs.buildGoModule {
          pname = "my-go-project";
          version = "0.0.1";
          src = self;
          vendorHash = null;
          meta = with pkgs.lib; {
            description = "My Go project";
            homepage = "https://github.com/conneroisu/my-go-project";
            license = licenses.asl20;
            maintainers = with maintainers; [connerohnesorge];
          };
        };
      };

      formatter = treefmt-nix.lib.mkWrapper pkgs treefmtModule;
    });
}
