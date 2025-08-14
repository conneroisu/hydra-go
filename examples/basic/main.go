// Package main demonstrates basic usage of the hydra-go client library.
// This example shows how to list projects, search, get project details,
// and retrieve build information using the flat API architecture.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/conneroisu/hydra-go"
)

func main() {
	// Create a client for the public Hydra instance.
	// For custom instances, use hydra.NewClient(&hydra.Config{BaseURL: "https://hydra.example.com"})
	client, err := hydra.NewDefaultClient()
	if err != nil {
		log.Fatalf("Failed to create Hydra client: %v", err)
	}

	ctx := context.Background()

	// Example 1: List all projects
	fmt.Println("=== Listing Projects ===")
	projects, err := client.ListProjects(ctx)
	if err != nil {
		log.Printf("Error listing projects: %v", err)
	} else {
		for i, project := range projects {
			if i >= 5 {
				fmt.Printf("... and %d more projects\n", len(projects)-5)

				break
			}
			fmt.Printf("Project: %s (%s)\n", project.Name, project.DisplayName)
		}
	}

	// Example 2: Search for nixpkgs
	fmt.Println("\n=== Searching for 'nixpkgs' ===")
	results, err := client.Search(ctx, "nixpkgs")
	if err != nil {
		log.Printf("Error searching: %v", err)
	} else {
		summary := hydra.GetSearchSummary("nixpkgs", results)
		fmt.Println(summary.Format())
	}

	// Example 3: Get a specific project (if it exists)
	projectName := "nixpkgs"
	fmt.Printf("\n=== Getting Project '%s' ===\n", projectName)
	project, err := client.GetProject(ctx, projectName)
	if err != nil {
		log.Printf("Error getting project: %v", err)
	} else {
		fmt.Printf("Project: %s\n", project.Name)
		fmt.Printf("Owner: %s\n", project.Owner)
		fmt.Printf("Description: %s\n", project.Description)
		fmt.Printf("Enabled: %v\n", project.Enabled)
		fmt.Printf("Jobsets: %v\n", project.Jobsets)
	}

	// Example 4: Get jobsets for a project
	if project != nil && len(project.Jobsets) > 0 {
		fmt.Printf("\n=== Getting Jobsets for '%s' ===\n", projectName)
		jobsetOverview, err := client.ListJobsets(ctx, projectName)
		if err != nil {
			log.Printf("Error getting jobsets: %v", err)
		} else {
			for i, js := range jobsetOverview {
				if i >= 3 {
					fmt.Printf("... and %d more jobsets\n", len(jobsetOverview)-3)

					break
				}
				fmt.Printf("Jobset: %s (checked: %d)\n", js.Name, js.LastCheckedTime)
			}
		}
	}

	// Example 5: Get a specific build (if provided as argument)
	if len(os.Args) > 1 {
		buildID := 0
		if _, err := fmt.Sscanf(os.Args[1], "%d", &buildID); err == nil {
			fmt.Printf("\n=== Getting Build #%d ===\n", buildID)
			build, err := client.GetBuild(ctx, buildID)
			if err != nil {
				log.Printf("Error getting build: %v", err)
			} else {
				fmt.Printf("Build: %s\n", build.NixName)
				fmt.Printf("Project: %s\n", build.Project)
				fmt.Printf("Jobset: %s\n", build.Jobset)
				fmt.Printf("Job: %s\n", build.Job)
				fmt.Printf("System: %s\n", build.System)
				fmt.Printf("Status: %s\n", build.GetBuildStatusString())
				if build.Finished {
					fmt.Printf("Duration: %v\n", build.GetDuration())
				}
			}
		}
	}
}
