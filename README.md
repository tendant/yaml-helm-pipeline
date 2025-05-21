# YAML Helm Pipeline

A web application that allows users to trigger Helm-based Kubernetes secret generation through a UI, instead of giving direct file access, and then commit the resulting YAML to GitHub.

## Features

- SolidJS frontend with a clean, responsive UI
- Golang backend API service with Chi router
- GitHub integration for repository access
- Helm CLI integration for templating
- Preview of keys (without values) before committing
- Branch selection for targeting specific environments
- Secure handling of secrets (values never shown in UI)

## Architecture

The application consists of the following components:

1. **Frontend UI (SolidJS)**
   - Branch selector
   - Preview of keys that will be changed
   - Commit form with message input

2. **Backend API (Golang with Chi Router)**
   - GitHub integration for repository access
   - Helm templating for generating Kubernetes secrets
   - Git operations for committing changes
   - Key extraction for previewing without values

3. **GitHub Integration**
   - Authentication via GitHub PAT (Personal Access Token)
   - Repository operations (clone, commit, push)

4. **Helm Integration**
   - Templating of Helm charts with values
   - Generation of Kubernetes secret YAML

## Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Helm CLI installed
- Git installed
- GitHub Personal Access Token with repo scope

## Environment Variables

The application uses [godotenv](https://github.com/joho/godotenv) to load environment variables from a `.env` file. You can create a `.env` file based on the provided `.env.example` file.

The following environment variables are required:

- `GITHUB_TOKEN`: GitHub Personal Access Token
- `REPO_OWNER`: GitHub repository owner
- `REPO_NAME`: GitHub repository name
- `PORT` (optional): Port for the server to listen on (default: 8080)

## Development Setup

### Backend

```bash
# Clone the repository
git clone https://github.com/yourusername/yaml-helm-pipeline.git
cd yaml-helm-pipeline

# Set up environment variables in .env file
cp .env.example .env
# Edit .env file with your values
# GITHUB_TOKEN=your_github_token
# REPO_OWNER=your_repo_owner
# REPO_NAME=your_repo_name

# Run the backend
go run cmd/server/main.go
```

### Frontend

```bash
# In a separate terminal
cd yaml-helm-pipeline/frontend

# Install dependencies
npm install

# Run the development server
npm run dev
```

The frontend will be available at http://localhost:3000 and will proxy API requests to the backend.

## Building and Running with Docker

```bash
# Build the Docker image
docker build -t yaml-helm-pipeline .

# Run the container with environment variables
docker run -p 8080:8080 \
  -e GITHUB_TOKEN=your_github_token \
  -e REPO_OWNER=your_repo_owner \
  -e REPO_NAME=your_repo_name \
  yaml-helm-pipeline

# Or run with .env file
docker run -p 8080:8080 \
  --env-file .env \
  yaml-helm-pipeline
```

The application will be available at http://localhost:8080.

## Repository Structure

The repository should have the following structure:

```
your-repo/
├── chart/           # Helm chart directory
│   ├── templates/   # Helm templates
│   └── ...
└── values/          # Values directory
    └── values.yaml  # Values file
```

## Usage

1. Open the application in your browser
2. Select a branch from the dropdown
3. Click "Preview Changes" to see the keys that will be included
4. Review the keys (values are not shown for security)
5. Enter a commit message and click "Commit Changes"
6. The changes will be committed to the selected branch

## License

This project is licensed under the MIT License - see the LICENSE file for details.
