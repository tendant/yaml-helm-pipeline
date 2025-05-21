#!/bin/bash

# Check if .env file exists, if not create from .env.example
if [ ! -f .env ]; then
  if [ -f .env.example ]; then
    echo "No .env file found, creating from .env.example"
    cp .env.example .env
    echo "Please edit .env file with your values and run this script again"
    exit 1
  else
    echo "Error: Neither .env nor .env.example file found"
    exit 1
  fi
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
