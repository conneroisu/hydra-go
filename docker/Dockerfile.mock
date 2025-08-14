FROM python:3.9-alpine

# Copy mock server script
COPY mock-hydra-server.py /app/server.py
COPY health-check.py /app/health-check.py

WORKDIR /app

EXPOSE 3000 8080

CMD ["python", "server.py"]