// Package main demonstrates comprehensive usage of the flat hydra-go API.
// This example shows how all functionality can be imported from a single package.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/conneroisu/hydra-go"
)

func main() {
	// Create client with custom configuration
	config := &hydra.Config{
		BaseURL: "https://hydra.nixos.org",
		Timeout: 30 * time.Second,
	}

	client, err := hydra.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// ===== Project Operations =====
	fmt.Println("=== Project Operations ===")

	// List all projects
	projects, err := client.ListProjects(ctx)
	if err != nil {
		log.Printf("Error listing projects: %v", err)
	} else {
		fmt.Printf("Found %d projects\n", len(projects))
		if len(projects) > 0 {
			fmt.Printf("First project: %s (%s)\n", projects[0].Name, projects[0].DisplayName)
		}
	}

	// Get specific project
	if len(projects) > 0 {
		project, err := client.GetProject(ctx, projects[0].Name)
		if err != nil {
			log.Printf("Error getting project: %v", err)
		} else {
			fmt.Printf("Project details - Owner: %s, Enabled: %v\n", project.Owner, project.Enabled)
		}
	}

	// ===== Search Operations =====
	fmt.Println("\n=== Search Operations ===")

	// Basic search
	searchResults, err := client.Search(ctx, "nixpkgs")
	if err != nil {
		log.Printf("Error searching: %v", err)
	} else {
		summary := hydra.GetSearchSummary("nixpkgs", searchResults)
		fmt.Println(summary.Format())
	}

	// Advanced search with options
	searchOpts := hydra.NewSearchOptions("hello")
	searchOpts.IncludeBuilds = true
	searchOpts.IncludeProjects = false
	searchOpts.IncludeJobsets = false
	searchOpts.IncludeDrvs = false

	filteredResults, err := client.SearchWithOptions(ctx, searchOpts)
	if err != nil {
		log.Printf("Error in filtered search: %v", err)
	} else {
		fmt.Printf("Found %d builds for 'hello'\n", len(filteredResults.Builds))
	}

	// Search specific types
	buildResults, err := client.SearchBuilds(ctx, "hello")
	if err != nil {
		log.Printf("Error searching builds: %v", err)
	} else {
		fmt.Printf("Build search found %d results\n", len(buildResults))
	}

	// ===== Jobset Operations =====
	fmt.Println("\n=== Jobset Operations ===")

	if len(projects) > 0 {
		projectName := projects[0].Name

		// List jobsets for project
		jobsets, err := client.ListJobsets(ctx, projectName)
		if err != nil {
			log.Printf("Error listing jobsets: %v", err)
		} else {
			fmt.Printf("Found %d jobsets for project %s\n", len(jobsets), projectName)

			if len(jobsets) > 0 {
				// Get specific jobset
				jobset, err := client.GetJobset(ctx, projectName, jobsets[0].Name)
				if err != nil {
					log.Printf("Error getting jobset: %v", err)
				} else {
					fmt.Printf("Jobset %s - Enabled: %v, Check interval: %d\n",
						jobset.Name, jobset.IsEnabled(), jobset.CheckInterval)
				}

				// Get evaluations
				evaluations, err := client.GetJobsetEvaluations(ctx, projectName, jobsets[0].Name)
				if err != nil {
					log.Printf("Error getting evaluations: %v", err)
				} else {
					fmt.Printf("Found %d evaluation pages\n", len(evaluations.Evals))
				}
			}
		}
	}

	// ===== Build Operations =====
	fmt.Println("\n=== Build Operations ===")

	// Get a specific build (if we have one from search)
	if len(buildResults) > 0 {
		buildID := buildResults[0].ID

		// Get build details
		build, err := client.GetBuild(ctx, buildID)
		if err != nil {
			log.Printf("Error getting build: %v", err)
		} else {
			fmt.Printf("Build %d - Status: %s, Job: %s\n",
				build.ID, build.GetBuildStatusString(), build.Job)

			if build.Finished {
				fmt.Printf("Duration: %v\n", build.GetDuration())
			}
		}

		// Get build info with all details
		buildInfo, err := client.GetBuildInfo(ctx, buildID)
		if err != nil {
			log.Printf("Error getting build info: %v", err)
		} else {
			fmt.Printf("Build info - Constituents: %d\n", len(buildInfo.Constituents))
			if buildInfo.Evaluation != nil {
				fmt.Printf("Evaluation ID: %d\n", buildInfo.Evaluation.ID)
			}
		}

		// Get constituents
		constituents, err := client.GetBuildConstituents(ctx, buildID)
		if err != nil {
			log.Printf("Error getting constituents: %v", err)
		} else {
			fmt.Printf("Found %d constituent builds\n", len(constituents))
		}
	}

	// ===== Utility Functions =====
	fmt.Println("\n=== Utility Functions ===")

	// Build statistics
	if len(buildResults) > 0 {
		stats := hydra.CalculateStatistics(buildResults)
		fmt.Printf("Build statistics - Total: %d, Success rate: %.1f%%\n",
			stats.Total, stats.GetSuccessRate())
	}

	// URL helpers
	if len(buildResults) > 0 {
		buildURL := hydra.GetBuildURL(client.BaseURL(), buildResults[0].ID)
		fmt.Printf("Build URL: %s\n", buildURL)
	}

	// Parse build ID
	if buildID, err := hydra.ParseBuildID("12345"); err == nil {
		fmt.Printf("Parsed build ID: %d\n", buildID)
	}

	// Flatten search results for display
	if searchResults != nil {
		items := hydra.FlattenSearchResults(client.BaseURL(), searchResults)
		fmt.Printf("Flattened search results: %d items\n", len(items))
		for i, item := range items {
			if i >= 3 { // Show first 3
				fmt.Printf("... and %d more items\n", len(items)-3)

				break
			}
			fmt.Printf("- %s: %s\n", item.Type, item.Name)
		}
	}

	// ===== Type Constants =====
	fmt.Println("\n=== Working with Constants ===")

	// Build status constants
	fmt.Printf("Build status constants available:\n")
	fmt.Printf("- Success: %d\n", hydra.BuildStatusSuccess)
	fmt.Printf("- Failed: %d\n", hydra.BuildStatusFailed)
	fmt.Printf("- Timed out: %d\n", hydra.BuildStatusTimedOut)

	// Jobset state constants
	fmt.Printf("Jobset state constants available:\n")
	fmt.Printf("- Enabled: %d\n", hydra.JobsetStateEnabled)
	fmt.Printf("- Disabled: %d\n", hydra.JobsetStateDisabled)
	fmt.Printf("- One shot: %d\n", hydra.JobsetStateOneShot)

	// Search type constants
	fmt.Printf("Search type constants available:\n")
	fmt.Printf("- Project: %s\n", hydra.SearchTypeProject)
	fmt.Printf("- Jobset: %s\n", hydra.SearchTypeJobset)
	fmt.Printf("- Build: %s\n", hydra.SearchTypeBuild)

	fmt.Println("\n=== Complete! All operations demonstrated with flat imports ===")
}
