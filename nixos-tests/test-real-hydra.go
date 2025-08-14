// Package main is a test program for the Hydra Go SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/conneroisu/hydra-go"
)

func main() {
	hydraURL := os.Getenv("HYDRA_URL")
	if hydraURL == "" {
		hydraURL = "http://localhost:3000"
	}

	fmt.Printf("========================================\n")
	fmt.Printf("Testing Hydra Go SDK with Real Hydra\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Hydra URL: %s\n\n", hydraURL)

	// Create client
	client, err := hydra.NewClientWithURL(hydraURL)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	testsPassed := 0
	testsFailed := 0

	// Test 1: Check Hydra connectivity
	fmt.Println("Test 1: Checking Hydra connectivity...")
	projects, err := client.ListProjects(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Hydra: %v\n", err)
		testsFailed++
	} else {
		fmt.Printf("✅ Connected to Hydra successfully\n")
		fmt.Printf("   Found %d existing projects\n", len(projects))
		testsPassed++
	}

	// Test 2: Authentication (optional - may not work on all Hydra setups)
	fmt.Println("\nTest 2: Testing authentication...")
	user, err := client.Login(ctx, "admin", "admin")
	if err != nil {
		fmt.Printf("⚠️  Authentication not available or failed: %v\n", err)
		fmt.Println("   (This is OK - many Hydra instances don't require auth for read operations)")
	} else {
		fmt.Printf("✅ Authenticated as: %s (%s)\n", user.Username, user.FullName)
		testsPassed++
	}

	// Test 3: Get specific project (if any exist)
	fmt.Println("\nTest 3: Getting project details...")
	if len(projects) > 0 {
		var project *hydra.Project
		project, err = client.GetProject(ctx, projects[0].Name)
		if err != nil {
			fmt.Printf("❌ Failed to get project: %v\n", err)
			testsFailed++
		} else {
			fmt.Printf("✅ Retrieved project: %s\n", project.Name)
			fmt.Printf("   Display Name: %s\n", project.DisplayName)
			fmt.Printf("   Owner: %s\n", project.Owner)
			fmt.Printf("   Enabled: %v\n", project.Enabled)
			testsPassed++
		}

		// Test 4: List jobsets for the project
		fmt.Println("\nTest 4: Listing jobsets...")
		jobsets, err := client.ListJobsets(ctx, projects[0].Name)
		if err != nil {
			fmt.Printf("❌ Failed to list jobsets: %v\n", err)
			testsFailed++
		} else {
			fmt.Printf("✅ Found %d jobsets\n", len(jobsets))
			for i, js := range jobsets {
				if i < 3 { // Show first 3
					fmt.Printf("   - %s\n", js.Name)
				}
			}
			testsPassed++
		}

		// Test 5: Get specific jobset (if any exist)
		if len(jobsets) > 0 {
			fmt.Println("\nTest 5: Getting jobset details...")
			jobset, err := client.GetJobset(ctx, projects[0].Name, jobsets[0].Name)
			if err != nil {
				fmt.Printf("❌ Failed to get jobset: %v\n", err)
				testsFailed++
			} else {
				fmt.Printf("✅ Retrieved jobset: %s/%s\n", jobset.Project, jobset.Name)
				if jobset.Description != nil {
					fmt.Printf("   Description: %s\n", *jobset.Description)
				}
				fmt.Printf("   Enabled: %v\n", jobset.IsEnabled())
				testsPassed++
			}
		}
	} else {
		fmt.Println("\n⚠️  No projects found - skipping project-specific tests")
	}

	// Test 6: Search functionality
	fmt.Println("\nTest 6: Testing search...")
	searchResults, err := client.Search(ctx, "nixos")
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		testsFailed++
	} else {
		fmt.Printf("✅ Search completed\n")
		fmt.Printf("   Projects found: %d\n", len(searchResults.Projects))
		fmt.Printf("   Jobsets found: %d\n", len(searchResults.Jobsets))
		fmt.Printf("   Builds found: %d\n", len(searchResults.Builds))
		testsPassed++
	}

	// Test 7: Create project (if authenticated)
	if client.IsAuthenticated() {
		fmt.Println("\nTest 7: Creating test project...")
		projectReq := &hydra.CreateProjectRequest{
			Name:        "sdk-test-" + strconv.FormatInt(time.Now().Unix(), 10),
			DisplayName: "SDK Test Project",
			Description: "Created by Hydra Go SDK test",
			Owner:       "admin",
			Enabled:     true,
			Visible:     true,
		}

		resp, err := client.CreateProject(ctx, projectReq.Name, projectReq)
		if err != nil {
			fmt.Printf("❌ Failed to create project: %v\n", err)
			testsFailed++
		} else {
			fmt.Printf("✅ Created project: %s\n", resp.Name)
			testsPassed++

			// Clean up - delete the test project
			fmt.Println("\nTest 8: Cleaning up test project...")
			err = client.DeleteProject(ctx, projectReq.Name)
			if err != nil {
				fmt.Printf("❌ Failed to delete project: %v\n", err)
				testsFailed++
			} else {
				fmt.Printf("✅ Deleted test project\n")
				testsPassed++
			}
		}
	} else {
		fmt.Println("\n⚠️  Not authenticated - skipping write operations")
	}

	// Test 9: Get recent builds (if available)
	fmt.Println("\nTest 9: Checking for recent builds...")
	// This would normally require knowing a valid build ID
	// For now, we'll just verify the API endpoint works
	_, err = client.GetBuild(ctx, 1)
	if err != nil {
		fmt.Printf("⚠️  No build with ID 1 found (this is normal)\n")
	} else {
		fmt.Printf("✅ Build API endpoint is working\n")
		testsPassed++
	}

	// Summary
	fmt.Printf("\n========================================\n")
	fmt.Printf("Test Summary\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Tests Passed: %d\n", testsPassed)
	fmt.Printf("Tests Failed: %d\n", testsFailed)

	if testsFailed == 0 && testsPassed > 0 {
		fmt.Printf("\n✅ All tests passed! SDK is working correctly with Hydra.\n")
		os.Exit(0)
	} else if testsPassed > 0 {
		fmt.Printf("\n⚠️  Some tests passed. SDK partially works with Hydra.\n")
		os.Exit(1)
	}
	fmt.Printf("\n❌ No tests passed. Check Hydra connectivity.\n")
	os.Exit(1)
}
