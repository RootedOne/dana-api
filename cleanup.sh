#!/bin/bash

echo "🚀 Starting CineSearch Pro full cleanup..."

# Stop and remove containers, networks, and images
docker-compose down --rmi all --volumes --remove-orphans

# Remove specific project image if still present
if [[ "$(docker images -q cinesearch-pro 2> /dev/null)" != "" ]]; then
  docker rmi cinesearch-pro
fi

# Optional: Clear build cache to save space
docker builder prune -f

echo "✅ Cleanup complete. System is fresh."
