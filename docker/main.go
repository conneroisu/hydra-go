package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	nixpkgsProject = "nixpkgs"
)

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Hidden      bool   `json:"hidden"`
}

type Build struct {
	ID          int    `json:"id"`
	Project     string `json:"project"`
	Jobset      string `json:"jobset"`
	Job         string `json:"job"`
	Timestamp   int64  `json:"timestamp"`
	StartTime   int64  `json:"starttime"`
	StopTime    int64  `json:"stoptime"`
	BuildStatus int    `json:"buildstatus"`
	NixName     string `json:"nixname"`
	Finished    bool   `json:"-"`
}

type Jobset struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Hidden      bool   `json:"hidden"`
}

type SearchResult struct {
	Jobsets   []Jobset  `json:"jobsets"`
	Projects  []Project `json:"projects"`
	Builds    []Build   `json:"builds"`
	BuildsDrv []Build   `json:"buildsdrv"`
}

type PushResponse struct {
	JobsetsTriggered []string `json:"jobsetsTriggered"`
}

type Evaluations struct {
	Evals []map[string]interface{} `json:"evals"`
}

type User struct {
	Username string `json:"username"`
	FullName string `json:"fullname"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// Health check server
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		server := &http.Server{
			Addr:         ":8080",
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		}

		log.Println("Health check server starting on :8080")
		log.Fatal(server.ListenAndServe())
	}()

	// Main API server
	mux := http.NewServeMux()

	// Projects endpoints (match real Hydra API paths)
	mux.HandleFunc("/project/", handleProject) // Individual project (must be before root)
	mux.HandleFunc("/", handleProjects)        // List projects

	// Jobset endpoints
	mux.HandleFunc("/jobset/", handleJobset)   // Individual jobset
	mux.HandleFunc("/jobsets/", handleJobsets) // List jobsets for project

	// Evaluation endpoints
	mux.HandleFunc("/eval/", handleEvaluations)

	// Trigger endpoint
	mux.HandleFunc("/push", handlePush)

	// Search endpoint
	mux.HandleFunc("/search", handleSearch)

	// Build endpoint
	mux.HandleFunc("/build/", handleBuild)

	// Login endpoint
	mux.HandleFunc("/login", handleLogin)

	server := &http.Server{
		Addr:         ":3000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Println("Hydra API server starting on :3000")
	log.Fatal(server.ListenAndServe())
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	projects := []Project{
		{
			Name:        nixpkgsProject,
			DisplayName: "Nixpkgs",
			Description: "Nix packages collection",
			Enabled:     true,
			Hidden:      false,
		},
		{
			Name:        "hydra",
			DisplayName: "Hydra",
			Description: "Hydra continuous integration system",
			Enabled:     true,
			Hidden:      false,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(projects)
}

func handleProject(w http.ResponseWriter, r *http.Request) {
	projectName := strings.TrimPrefix(r.URL.Path, "/project/")

	switch projectName {
	case nixpkgsProject:
		project := Project{
			Name:        nixpkgsProject,
			DisplayName: "Nixpkgs",
			Description: "Nix packages collection",
			Enabled:     true,
			Hidden:      false,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	case "hydra":
		project := Project{
			Name:        "hydra",
			DisplayName: "Hydra",
			Description: "Hydra continuous integration system",
			Enabled:     true,
			Hidden:      false,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	default:
		http.NotFound(w, r)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	result := SearchResult{
		Jobsets:   []Jobset{},
		Projects:  []Project{},
		Builds:    []Build{},
		BuildsDrv: []Build{},
	}

	// Mock search results based on query
	if query == "hello" || query == "" {
		result.Builds = append(result.Builds, Build{
			ID:          1,
			Project:     nixpkgsProject,
			Jobset:      "trunk",
			Job:         "hello",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    1692000000,
			BuildStatus: 0,
			NixName:     "hello-2.12.1",
			Finished:    true,
		})
	}

	if query == "nix" || query == "" {
		result.Projects = append(result.Projects, Project{
			Name:        nixpkgsProject,
			DisplayName: "Nixpkgs",
			Description: "Nix packages collection",
			Enabled:     true,
			Hidden:      false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	buildID := strings.TrimPrefix(r.URL.Path, "/build/")

	// Support multiple build IDs
	switch buildID {
	case "1":
		build := Build{
			ID:          1,
			Project:     nixpkgsProject,
			Jobset:      "trunk",
			Job:         "hello",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    1692000000,
			BuildStatus: 0,
			NixName:     "hello-2.12.1",
			Finished:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(build)
	case "123456":
		build := Build{
			ID:          123456,
			Project:     nixpkgsProject,
			Jobset:      "trunk",
			Job:         "hello",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    1692000000,
			BuildStatus: 0,
			NixName:     "hello-2.12.1",
			Finished:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(build)
	case "123459":
		// In-progress build
		build := Build{
			ID:          123459,
			Project:     nixpkgsProject,
			Jobset:      "trunk",
			Job:         "gcc",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    0, // Not finished yet
			BuildStatus: 0,
			NixName:     "gcc-11.3.0",
			Finished:    false, // In progress
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(build)
	case "123460":
		// Failed build
		build := Build{
			ID:          123460,
			Project:     nixpkgsProject,
			Jobset:      "trunk",
			Job:         "broken-package",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    1692000000,
			BuildStatus: 1, // Failed
			NixName:     "broken-package-1.0",
			Finished:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(build)
	default:
		http.NotFound(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	if loginReq.Username == "admin" && loginReq.Password == "admin" {
		// Set session cookie like real Hydra would
		http.SetCookie(w, &http.Cookie{
			Name:     "hydra_session", // Use underscore to match client expectation
			Value:    "mock-session-token-12345",
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to false for local testing
			MaxAge:   3600,  // 1 hour
		})

		user := User{
			Username: "admin",
			FullName: "Admin",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
}

func handleJobsets(w http.ResponseWriter, r *http.Request) {
	// Extract project name from /jobsets/PROJECT
	path := strings.TrimPrefix(r.URL.Path, "/jobsets/")
	project := strings.Split(path, "/")[0]

	if project == nixpkgsProject {
		jobsets := []Jobset{
			{
				Name:        "trunk",
				Project:     nixpkgsProject,
				Description: "Main development branch",
				Enabled:     true,
				Hidden:      false,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jobsets)
	} else {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Jobset{})
	}
}

func handleJobset(w http.ResponseWriter, r *http.Request) {
	// Extract project and jobset from /jobset/PROJECT/JOBSET
	path := strings.TrimPrefix(r.URL.Path, "/jobset/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 {
		project := parts[0]
		jobsetName := parts[1]

		if project == nixpkgsProject && jobsetName == "trunk" {
			jobset := Jobset{
				Name:        "trunk",
				Project:     nixpkgsProject,
				Description: "Main development branch",
				Enabled:     true,
				Hidden:      false,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(jobset)

			return
		}
	}

	http.NotFound(w, r)
}

func handleEvaluations(w http.ResponseWriter, r *http.Request) {
	// Mock evaluations response
	evaluations := Evaluations{
		Evals: []map[string]interface{}{
			{
				"id":        1,
				"project":   nixpkgsProject,
				"jobset":    "trunk",
				"timestamp": 1692000000,
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(evaluations)
}

func handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Mock push response
	response := PushResponse{
		JobsetsTriggered: []string{nixpkgsProject + ":trunk"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
