#!/usr/bin/env python3
import http.server
import socketserver
import json
from urllib.request import urlopen
from urllib.error import URLError

class HealthHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            try:
                response = urlopen("http://localhost:3000/", timeout=5)
                if response.status == 200:
                    self.send_response(200)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({
                        "status": "healthy", 
                        "hydra": "running",
                        "port": 3000
                    }).encode())
                else:
                    raise Exception("Hydra not responding")
            except Exception as e:
                self.send_response(503)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({
                    "status": "unhealthy", 
                    "error": str(e)
                }).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # Suppress logging to keep output clean
        pass

if __name__ == "__main__":
    with socketserver.TCPServer(("", 8080), HealthHandler) as httpd:
        print("Health check server starting on port 8080...")
        httpd.serve_forever()