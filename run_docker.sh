#!/bin/bash

echo "Starting MLOps Pipeline with Docker Compose..."

# Build and start in detached mode
docker-compose up --build -d

echo "Deployment started!"
echo "--------------------------------------------------"
echo "Services will be available at:"
echo "- API (Go): http://localhost:8000/docs"
echo "- Main UI: http://localhost:8501"
echo "- Monitoring: http://localhost:8502"
echo "- llama.cpp: http://localhost:8080/v1"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3000"
echo "--------------------------------------------------"
echo "Use 'docker-compose logs -f api' to see API logs."
echo "Use 'docker-compose down' to stop the services."
echo ""
echo "Run smoke tests with: ./scripts/smoke.sh"
