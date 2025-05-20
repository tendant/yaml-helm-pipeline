#!/bin/bash

echo "Installing dependencies for YAML Helm Pipeline..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed"
    echo "Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "Error: npm is not installed"
    echo "Please install npm (it usually comes with Node.js)"
    exit 1
fi

# Check if Helm is installed
if ! command -v helm &> /dev/null; then
    echo "Error: Helm is not installed"
    echo "Please install Helm from https://helm.sh/docs/intro/install/"
    exit 1
fi

# Check if Git is installed
if ! command -v git &> /dev/null; then
    echo "Error: Git is not installed"
    echo "Please install Git from https://git-scm.com/downloads"
    exit 1
fi

echo "All required tools are installed"

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
if [ $? -ne 0 ]; then
    echo "Error: Failed to download Go dependencies"
    exit 1
fi

# Install Node.js dependencies
echo "Installing Node.js dependencies..."
cd frontend && npm install && cd ..
if [ $? -ne 0 ]; then
    echo "Error: Failed to install Node.js dependencies"
    exit 1
fi

echo "All dependencies installed successfully"
echo ""
echo "You can now set up the environment variables using:"
echo "  source setup-env.sh"
echo ""
echo "And then run the application using one of the following commands:"
echo "  make dev         - Run in development mode"
echo "  make run         - Run the backend only"
echo "  make docker      - Build and run in Docker"
echo "  make docker-compose-up - Start with Docker Compose"
