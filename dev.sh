#!/bin/bash

# Check if environment variables are set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Error: GITHUB_TOKEN environment variable is not set"
  exit 1
fi

if [ -z "$REPO_OWNER" ]; then
  echo "Error: REPO_OWNER environment variable is not set"
  exit 1
fi

if [ -z "$REPO_NAME" ]; then
  echo "Error: REPO_NAME environment variable is not set"
  exit 1
fi

# Start the backend in the background
echo "Starting backend server..."
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait for the backend to start
sleep 2

# Start the frontend
echo "Starting frontend development server..."
cd frontend && npm run dev

# When the frontend is stopped, also stop the backend
kill $BACKEND_PID
