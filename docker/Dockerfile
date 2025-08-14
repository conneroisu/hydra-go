FROM nixos/nix:latest

# Set up environment
ENV HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra"
ENV PGDATA="/var/lib/postgresql/data"

# Install necessary packages including shadow utils for user management
RUN nix-channel --add https://nixos.org/channels/nixpkgs-unstable nixpkgs && \
    nix-channel --update && \
    nix-env -iA nixpkgs.hydra nixpkgs.postgresql nixpkgs.python3 nixpkgs.curl nixpkgs.bash nixpkgs.shadow

# Create necessary directories
RUN mkdir -p /var/lib/postgresql && \
    mkdir -p /var/lib/hydra

# Add startup script
COPY startup.sh /usr/local/bin/startup.sh
RUN chmod +x /usr/local/bin/startup.sh

# Add health check script
COPY health-check.py /usr/local/bin/health-check.py
RUN chmod +x /usr/local/bin/health-check.py

# Expose ports
EXPOSE 3000 8080

# Run the startup script
CMD ["/usr/local/bin/startup.sh"]