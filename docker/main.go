package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
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
			Name:        "nixpkgs",
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
	case "nixpkgs":
		project := Project{
			Name:        "nixpkgs",
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
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("[]"))
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	buildID := strings.TrimPrefix(r.URL.Path, "/build/")

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

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	if loginReq.Username == "admin" && loginReq.Password == "admin" {
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
