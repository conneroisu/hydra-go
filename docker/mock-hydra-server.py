#!/usr/bin/env python3
import http.server
import socketserver
import json
import threading
import time
import urllib.parse

# In-memory storage for projects
projects_storage = {
    "nixpkgs": {"name": "nixpkgs", "displayname": "Nixpkgs", "enabled": True},
    "hydra": {"name": "hydra", "displayname": "Hydra", "enabled": True}
}

class MockHydraHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/" and "api" not in self.path:
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            projects_list = list(projects_storage.values())
            self.wfile.write(json.dumps(projects_list).encode())
        elif self.path.startswith("/api/v1/projects"):
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            projects_list = list(projects_storage.values())
            self.wfile.write(json.dumps(projects_list).encode())
        elif self.path.startswith("/project/"):
            project_name = self.path.split("/")[-1]
            if project_name in projects_storage:
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                project = projects_storage[project_name]
                self.wfile.write(json.dumps(project).encode())
            elif project_name == "definitely-does-not-exist":
                self.send_response(404)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "not found"}).encode())
            else:
                self.send_response(404)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "not found"}).encode())
        elif self.path.startswith("/api/jobsets"):
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            jobsets = [{"name": "trunk", "project": "nixpkgs", "enabled": 1}]
            self.wfile.write(json.dumps(jobsets).encode())
        elif self.path.startswith("/jobset/"):
            parts = self.path.split("/")
            if len(parts) >= 4 and parts[2] == "nixpkgs" and parts[3] == "trunk":
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                jobset = {"name": "trunk", "project": "nixpkgs", "enabled": 1}
                self.wfile.write(json.dumps(jobset).encode())
            else:
                self.send_response(404)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "not found"}).encode())
        elif self.path.startswith("/build/"):
            build_id = self.path.split("/")[-1]
            try:
                build_id_int = int(build_id)
                if build_id_int == 123456:
                    # Mock successful build
                    self.send_response(200)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    build = {
                        "id": 123456,
                        "nixname": "hello-2.12.1",
                        "finished": True,
                        "buildstatus": 0,  # Success
                        "job": "hello",
                        "project": "nixpkgs",
                        "jobset": "trunk"
                    }
                    self.wfile.write(json.dumps(build).encode())
                elif build_id_int == 123459:
                    # Mock in-progress build
                    self.send_response(200)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    build = {
                        "id": 123459,
                        "nixname": "hello-in-progress",
                        "finished": False,
                        "buildstatus": None,
                        "job": "hello",
                        "project": "nixpkgs",
                        "jobset": "trunk"
                    }
                    self.wfile.write(json.dumps(build).encode())
                elif build_id_int == 123460:
                    # Mock failed build
                    self.send_response(200)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    build = {
                        "id": 123460,
                        "nixname": "hello-failed",
                        "finished": True,
                        "buildstatus": 1,  # Failed
                        "job": "hello",
                        "project": "nixpkgs",
                        "jobset": "trunk"
                    }
                    self.wfile.write(json.dumps(build).encode())
                else:
                    self.send_response(404)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": "not found"}).encode())
            except ValueError:
                self.send_response(400)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "invalid build id"}).encode())
        elif "/constituents" in self.path:
            # Handle build constituents endpoint - /build/ID/constituents
            path_parts = self.path.split("/")
            if len(path_parts) >= 3 and path_parts[1] == "build":
                build_id = path_parts[2]
                try:
                    build_id_int = int(build_id)
                    if build_id_int in [123456, 123459, 123460]:
                        self.send_response(200)
                        self.send_header('Content-type', 'application/json')
                        self.end_headers()
                        constituents = [
                            {"id": 123457, "nixname": "dependency-1"},
                            {"id": 123458, "nixname": "dependency-2"}
                        ]
                        self.wfile.write(json.dumps(constituents).encode())
                    else:
                        self.send_response(404)
                        self.send_header('Content-type', 'application/json')
                        self.end_headers()
                        self.wfile.write(json.dumps({"error": "not found"}).encode())
                except ValueError:
                    self.send_response(400)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": "invalid build id"}).encode())
            else:
                self.send_response(404)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "not found"}).encode())
        elif self.path.startswith("/api/search") or self.path.startswith("/search"):
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            results = {
                "builds": [
                    {
                        "id": 123456,
                        "nixname": "hello-2.12.1",
                        "job": "hello",
                        "project": "nixpkgs",
                        "jobset": "trunk"
                    }
                ],
                "projects": [
                    {"name": "nixpkgs", "displayname": "Nixpkgs", "enabled": True}
                ]
            }
            self.wfile.write(json.dumps(results).encode())
        else:
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())
    
    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length) if content_length > 0 else b''
        
        if self.path == "/login":
            # Parse credentials from POST body
            try:
                if post_data:
                    # Handle form data or JSON
                    if b'username=' in post_data:
                        # Form data
                        data_str = post_data.decode('utf-8')
                        username = None
                        password = None
                        for param in data_str.split('&'):
                            if param.startswith('username='):
                                username = param.split('=')[1]
                            elif param.startswith('password='):
                                password = param.split('=')[1]
                    else:
                        # JSON data
                        import json as json_lib
                        data = json_lib.loads(post_data.decode('utf-8'))
                        username = data.get('username')
                        password = data.get('password')
                    
                    # Mock successful login for admin/admin
                    if username == 'admin' and password == 'admin':
                        self.send_response(200)
                        self.send_header('Content-type', 'application/json')
                        self.send_header('Set-Cookie', 'hydra_session=mock_session_token; Path=/; HttpOnly')
                        self.end_headers()
                        user = {
                            "username": "admin",
                            "fullname": "Admin",
                            "emailaddress": "admin@example.com",
                            "roles": ["admin"]
                        }
                        self.wfile.write(json.dumps(user).encode())
                    else:
                        self.send_response(401)
                        self.send_header('Content-type', 'application/json')
                        self.end_headers()
                        self.wfile.write(json.dumps({"error": "unauthorized"}).encode())
                else:
                    self.send_response(400)
                    self.send_header('Content-type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": "missing credentials"}).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        elif self.path.startswith("/project/"):
            # Handle project creation
            project_name = self.path.split("/")[-1]
            try:
                content_length = int(self.headers.get('Content-Length', 0))
                post_data = self.rfile.read(content_length) if content_length > 0 else b'{}'
                
                # Parse project data
                if post_data:
                    try:
                        project_data = json.loads(post_data.decode('utf-8'))
                    except:
                        # Handle form data
                        project_data = {"name": project_name, "displayname": project_name.title(), "enabled": True}
                else:
                    project_data = {"name": project_name, "displayname": project_name.title(), "enabled": True}
                
                # Store the project
                projects_storage[project_name] = {
                    "name": project_name,
                    "displayname": project_data.get("displayname", project_name.title()),
                    "enabled": project_data.get("enabled", True)
                }
                
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(projects_storage[project_name]).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        else:
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "created"}).encode())
    
    def do_PUT(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({"status": "updated"}).encode())
    
    def do_DELETE(self):
        if self.path.startswith("/project/"):
            project_name = self.path.split("/")[-1]
            if project_name in projects_storage:
                del projects_storage[project_name]
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"status": "deleted"}).encode())
            else:
                self.send_response(404)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": "not found"}).encode())
        else:
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "deleted"}).encode())

    def log_message(self, format, *args):
        pass  # Suppress logs

class HealthHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "healthy", "hydra": "running"}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass

def start_health_server():
    with socketserver.TCPServer(("", 8080), HealthHandler) as httpd:
        httpd.serve_forever()

def start_hydra_server():
    with socketserver.TCPServer(("", 3000), MockHydraHandler) as httpd:
        httpd.serve_forever()

if __name__ == "__main__":
    print("Starting mock Hydra server on port 3000...")
    print("Starting health check server on port 8080...")
    
    # Start health server in background thread
    health_thread = threading.Thread(target=start_health_server)
    health_thread.daemon = True
    health_thread.start()
    
    # Start main server (blocking)
    start_hydra_server()