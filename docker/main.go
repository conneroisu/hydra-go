package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Hidden      int    `json:"hidden"`
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
}

type User struct {
	Username string `json:"username"`
	FullName string `json:"fullname"`
}

func main() {
	// Health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})
		log.Println("Health check server starting on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Main API server
	mux := http.NewServeMux()

	// Projects endpoints
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/project/", handleProject)

	// Search endpoint
	mux.HandleFunc("/api/search", handleSearch)

	// Build endpoint
	mux.HandleFunc("/api/build/", handleBuild)

	// Login endpoint
	mux.HandleFunc("/login", handleLogin)

	// Root endpoint
	mux.HandleFunc("/", handleRoot)

	log.Println("Hydra API server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	projects := []Project{
		{
			Name:        "nixpkgs",
			DisplayName: "Nixpkgs",
			Description: "Nix packages collection",
			Enabled:     1,
			Hidden:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(projects)
}

func handleProject(w http.ResponseWriter, r *http.Request) {
	projectName := strings.TrimPrefix(r.URL.Path, "/api/project/")

	if projectName == "nixpkgs" {
		project := Project{
			Name:        "nixpkgs",
			DisplayName: "Nixpkgs",
			Description: "Nix packages collection",
			Enabled:     1,
			Hidden:      0,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	} else {
		http.NotFound(w, r)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("[]"))
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	buildID := strings.TrimPrefix(r.URL.Path, "/api/build/")

	if buildID == "1" {
		build := Build{
			ID:          1,
			Project:     "nixpkgs",
			Jobset:      "trunk",
			Job:         "hello",
			Timestamp:   1692000000,
			StartTime:   1692000000,
			StopTime:    1692000000,
			BuildStatus: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(build)
	} else {
		http.NotFound(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "admin" && password == "admin" {
		user := User{
			Username: "admin",
			FullName: "Admin",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		http.NotFound(w, r)

		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body><h1>Mock Hydra Server</h1><p>API available at /api/</p></body></html>`)
}
