package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/conneroisu/hydra-go"
)

func main() {
	// Check for credentials
	username := os.Getenv("HYDRA_USERNAME")
	password := os.Getenv("HYDRA_PASSWORD")
	hydraURL := os.Getenv("HYDRA_URL")

	if hydraURL == "" {
		hydraURL = "https://hydra.nixos.org"
	}

	// Create client
	client, err := hydra.NewClientWithURL(hydraURL)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Login if credentials are provided
	if username != "" && password != "" {
		fmt.Printf("=== Logging in as %s ===\n", username)
		user, err := client.Login(ctx, username, password)
		if err != nil {
			log.Printf("Login failed: %v", err)
		} else {
			fmt.Printf("Logged in successfully!\n")
			fmt.Printf("Full Name: %s\n", user.FullName)
			fmt.Printf("Email: %s\n", user.EmailAddress)
			fmt.Printf("Roles: %v\n", user.UserRoles)
			defer client.Logout()
		}
	}

	// Example: Create a project (requires authentication and permissions)
	if client.IsAuthenticated() {
		fmt.Println("\n=== Creating a Test Project ===")

		projectOpts := hydra.NewCreateProjectOptions("test-project", username).
			WithDisplayName("Test Project").
			WithDescription("A test project created via the Hydra Go SDK").
			WithHomepage("https://example.com").
			WithEnabled(true).
			WithVisible(true)

		resp, err := client.CreateProjectWithOptions(ctx, "test-project", projectOpts)
		if err != nil {
			log.Printf("Failed to create project: %v", err)
		} else {
			fmt.Printf("Project created successfully!\n")
			fmt.Printf("URI: %s\n", resp.URI)
			fmt.Printf("Redirect: %s\n", resp.Redirect)
		}

		// Example: Create a jobset
		fmt.Println("\n=== Creating a Test Jobset ===")

		jobsetOpts := hydra.NewCreateJobsetOptions("test-jobset", "test-project").
			WithDescription("Test jobset").
			WithNixExpression("nixpkgs", "release.nix").
			WithState(hydra.JobsetStateEnabled).
			WithScheduling(300, 100).
			AddInput("nixpkgs", "git", "https://github.com/NixOS/nixpkgs.git", false).
			AddInput("officialRelease", "boolean", "false", false)

		jobsetResp, err := client.CreateJobsetWithOptions(ctx, "test-project", "test-jobset", jobsetOpts)
		if err != nil {
			log.Printf("Failed to create jobset: %v", err)
		} else {
			fmt.Printf("Jobset created successfully!\n")
			fmt.Printf("URI: %s\n", jobsetResp.URI)
		}

		// Trigger evaluation
		fmt.Println("\n=== Triggering Evaluation ===")
		pushResp, err := client.TriggerJobset(ctx, "test-project", "test-jobset")
		if err != nil {
			log.Printf("Failed to trigger evaluation: %v", err)
		} else {
			fmt.Printf("Triggered jobsets: %v\n", pushResp.JobsetsTriggered)
		}
	} else {
		fmt.Println("\nNote: Set HYDRA_USERNAME and HYDRA_PASSWORD environment variables to test authenticated operations")
	}
}
