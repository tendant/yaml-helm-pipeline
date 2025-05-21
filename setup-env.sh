#!/bin/bash

# Check if .env file exists
if [ ! -f .env ]; then
  echo "Error: .env file not found"
  echo "Please create a .env file based on .env.example"
  exit 1
fi

# Export environment variables from .env file
export $(grep -v '^#' .env | xargs)

# Verify that required environment variables are set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Error: GITHUB_TOKEN environment variable is not set in .env file"
  exit 1
fi

if [ -z "$REPO_OWNER" ]; then
  echo "Error: REPO_OWNER environment variable is not set in .env file"
  exit 1
fi

if [ -z "$REPO_NAME" ]; then
  echo "Error: REPO_NAME environment variable is not set in .env file"
  exit 1
fi

echo "Environment variables set successfully"
echo "GITHUB_TOKEN: ${GITHUB_TOKEN:0:5}..."
echo "REPO_OWNER: $REPO_OWNER"
echo "REPO_NAME: $REPO_NAME"
echo "PORT: ${PORT:-4000}"

# Print instructions
echo ""
echo "You can now run the application using one of the following commands:"
echo "  make dev         - Run in development mode"
echo "  make run         - Run the backend only"
echo "  make docker      - Build and run in Docker"
echo "  make docker-compose-up - Start with Docker Compose"
