package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	data     *TestData
}

type Session struct {
	Username  string
	CreatedAt time.Time
}

type TestData struct {
	Projects  []Project      `json:"projects"`
	Jobsets   []Jobset       `json:"jobsets"`
	Builds    []Build        `json:"builds"`
	Users     []User         `json:"users"`
}

func NewServer() *Server {
	s := &Server{
		sessions: make(map[string]*Session),
		data:     loadTestData(),
	}
	return s
}

func loadTestData() *TestData {
	// Try to load from fixtures file first
	if fileData, err := os.ReadFile("../../fixtures/test-data.json"); err == nil {
		data := &TestData{}
		if err := json.Unmarshal(fileData, data); err == nil {
			return data
		}
	}
	
	// Load from fixtures file or use defaults
	data := &TestData{
		Projects: []Project{
			{
				Name:        "nixpkgs",
				DisplayName: "Nixpkgs",
				Owner:       "admin",
				Enabled:     true,
				Description: "Nix Packages collection",
			},
			{
				Name:        "hydra",
				DisplayName: "Hydra",
				Owner:       "admin",
				Enabled:     true,
				Description: "Continuous Integration System",
			},
		},
		Jobsets: []Jobset{
			{
				Name:        "trunk",
				Project:     "nixpkgs",
				Description: "Main development branch",
				Enabled:     1,
				Visible:     true,
			},
			{
				Name:        "staging",
				Project:     "nixpkgs",
				Description: "Staging branch",
				Enabled:     1,
				Visible:     true,
			},
		},
		Builds: []Build{
			{
				ID:          123456,
				NixName:     "hello-2.12.1",
				Finished:    true,
				BuildStatus: intPtr(0),
				Project:     "nixpkgs",
				Jobset:      "trunk",
				Job:         "hello",
				System:      "x86_64-linux",
				StartTime:   time.Now().Add(-1 * time.Hour).Unix(),
				StopTime:    time.Now().Unix(),
			},
			{
				ID:          123460,
				NixName:     "failed-build-1.0.0",
				Finished:    true,
				BuildStatus: intPtr(1),
				Project:     "test-project",
				Jobset:      "test-jobset",
				Job:         "failed-job",
				System:      "x86_64-linux",
				StartTime:   1700040000,
				StopTime:    1700040100,
			},
		},
		Users: []User{
			{
				Username:     "testuser",
				FullName:     "Test User",
				EmailAddress: "test@example.com",
				Password:     "testpass", // In production, this would be hashed
			},
		},
	}
	
	// Try to load from fixtures file
	if fileData, err := os.ReadFile("/etc/mock-hydra-server/fixtures.json"); err == nil {
		if err := json.Unmarshal(fileData, data); err != nil {
			log.Printf("Failed to parse fixtures: %v", err)
		}
	}
	
	return data
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Route requests
	switch {
	case r.URL.Path == "/health":
		s.handleHealth(w, r)
	case r.URL.Path == "/login" && r.Method == "POST":
		s.handleLogin(w, r)
	case r.URL.Path == "/" && r.Method == "GET":
		s.handleListProjects(w, r)
	case strings.HasPrefix(r.URL.Path, "/project/"):
		s.handleProject(w, r)
	case strings.HasPrefix(r.URL.Path, "/jobset/"):
		s.handleJobset(w, r)
	case strings.HasPrefix(r.URL.Path, "/build/"):
		s.handleBuild(w, r)
	case strings.HasPrefix(r.URL.Path, "/eval/"):
		s.handleEvaluation(w, r)
	case r.URL.Path == "/search":
		s.handleSearch(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		s.handleAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "mock-hydra-server",
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	
	// Check credentials
	var user *User
	for _, u := range s.data.Users {
		if u.Username == req.Username && u.Password == req.Password {
			user = &u
			break
		}
	}
	
	if user == nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	
	// Create session
	sessionID := generateSessionID()
	s.mu.Lock()
	s.sessions[sessionID] = &Session{
		Username:  user.Username,
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()
	
	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "hydra_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
	})
	
	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":     user.Username,
		"fullname":     user.FullName,
		"emailaddress": user.EmailAddress,
		"userroles":    []string{"user"},
	})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.data.Projects)
}

func (s *Server) handleProject(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	
	projectName := parts[2]
	
	switch r.Method {
	case "GET":
		for _, p := range s.data.Projects {
			if p.Name == projectName {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(p)
				return
			}
		}
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		
	case "PUT":
		// Create/update project
		var req Project
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}
		
		req.Name = projectName
		
		// Update or add project
		found := false
		for i, p := range s.data.Projects {
			if p.Name == projectName {
				s.data.Projects[i] = req
				found = true
				break
			}
		}
		if !found {
			s.data.Projects = append(s.data.Projects, req)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"redirect": "/project/" + projectName,
			"uri":      "/project/" + projectName,
			"name":     projectName,
		})
		
	case "DELETE":
		// Delete project
		for i, p := range s.data.Projects {
			if p.Name == projectName {
				s.data.Projects = append(s.data.Projects[:i], s.data.Projects[i+1:]...)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"redirect": "/",
				})
				return
			}
		}
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleJobset(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	
	projectName := parts[2]
	jobsetName := parts[3]
	
	// Handle evaluations endpoint
	if len(parts) > 4 && parts[4] == "evals" {
		s.handleEvaluations(w, r, projectName, jobsetName)
		return
	}
	
	switch r.Method {
	case "GET":
		for _, j := range s.data.Jobsets {
			if j.Project == projectName && j.Name == jobsetName {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(j)
				return
			}
		}
		http.Error(w, `{"error":"jobset not found"}`, http.StatusNotFound)
		
	case "PUT":
		// Create/update jobset
		var req Jobset
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}
		
		req.Name = jobsetName
		req.Project = projectName
		
		// Update or add jobset
		found := false
		for i, j := range s.data.Jobsets {
			if j.Project == projectName && j.Name == jobsetName {
				s.data.Jobsets[i] = req
				found = true
				break
			}
		}
		if !found {
			s.data.Jobsets = append(s.data.Jobsets, req)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"redirect": fmt.Sprintf("/jobset/%s/%s", projectName, jobsetName),
		})
		
	case "DELETE":
		// Delete jobset
		for i, j := range s.data.Jobsets {
			if j.Project == projectName && j.Name == jobsetName {
				s.data.Jobsets = append(s.data.Jobsets[:i], s.data.Jobsets[i+1:]...)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"redirect": fmt.Sprintf("/project/%s", projectName),
				})
				return
			}
		}
		http.Error(w, `{"error":"jobset not found"}`, http.StatusNotFound)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	
	buildID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, `{"error":"invalid build ID"}`, http.StatusBadRequest)
		return
	}
	
	// Handle constituents endpoint
	if len(parts) > 3 && parts[3] == "constituents" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Build{}) // Return empty array for now
		return
	}
	
	for _, b := range s.data.Builds {
		if b.ID == buildID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(b)
			return
		}
	}
	
	http.Error(w, `{"error":"build not found"}`, http.StatusNotFound)
}

func (s *Server) handleEvaluation(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	
	evalID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, `{"error":"invalid evaluation ID"}`, http.StatusBadRequest)
		return
	}
	
	// Handle builds endpoint
	if len(parts) > 3 && parts[3] == "builds" {
		w.Header().Set("Content-Type", "application/json")
		// Return builds for this evaluation
		builds := make(map[string][]Build)
		for _, b := range s.data.Builds {
			if contains(b.JobsetEvals, evalID) {
				builds[b.Job] = append(builds[b.Job], b)
			}
		}
		json.NewEncoder(w).Encode(builds)
		return
	}
	
	// Return mock evaluation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JobsetEval{
		ID:           evalID,
		Timestamp:    time.Now().Unix(),
		HasNewBuilds: true,
		Builds:       []int{123456},
	})
}

func (s *Server) handleEvaluations(w http.ResponseWriter, r *http.Request, projectName, jobsetName string) {
	w.Header().Set("Content-Type", "application/json")
	
	// Return mock evaluations
	evals := Evaluations{
		First:  "?page=1",
		Last:   "?page=1",
		Evals: []map[string]*JobsetEval{
			{
				"1": &JobsetEval{
					ID:           1,
					Timestamp:    time.Now().Unix(),
					HasNewBuilds: true,
					Builds:       []int{123456},
				},
			},
		},
	}
	
	json.NewEncoder(w).Encode(evals)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	
	results := SearchResult{
		Projects:  []Project{},
		Jobsets:   []Jobset{},
		Builds:    []Build{},
		BuildsDrv: []Build{},
	}
	
	// Simple search implementation
	queryLower := strings.ToLower(query)
	
	for _, p := range s.data.Projects {
		if strings.Contains(strings.ToLower(p.Name), queryLower) ||
		   strings.Contains(strings.ToLower(p.Description), queryLower) {
			results.Projects = append(results.Projects, p)
		}
	}
	
	for _, j := range s.data.Jobsets {
		if strings.Contains(strings.ToLower(j.Name), queryLower) ||
		   strings.Contains(strings.ToLower(j.Description), queryLower) {
			results.Jobsets = append(results.Jobsets, j)
		}
	}
	
	for _, b := range s.data.Builds {
		if strings.Contains(strings.ToLower(b.NixName), queryLower) ||
		   strings.Contains(strings.ToLower(b.Job), queryLower) {
			results.Builds = append(results.Builds, b)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	
	switch {
	case path == "push" && r.Method == "POST":
		// Handle trigger evaluation
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "triggered",
			"jobsets": r.URL.Query().Get("jobsets"),
		})
		
	case strings.HasPrefix(path, "jobsets"):
		// Handle jobsets listing - return as array for JobsetOverview
		project := r.URL.Query().Get("project")
		var jobsets []Jobset
		for _, j := range s.data.Jobsets {
			if j.Project == project {
				jobsets = append(jobsets, j)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobsets)
		
	default:
		http.NotFound(w, r)
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func intPtr(i int) *int {
	return &i
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	
	server := NewServer()
	addr := fmt.Sprintf("%s:%s", host, port)
	
	log.Printf("Mock Hydra server starting on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatal(err)
	}
}