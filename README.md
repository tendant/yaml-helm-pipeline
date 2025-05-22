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

## Configuration

The application supports multiple configuration methods:

### 1. YAML Configuration File

The application looks for a file named `config.yaml` in the application directory, or at the path specified by the `CONFIG_PATH` environment variable.

Example `config.yaml`:

```yaml
groups:
  - name: production
    values_repos:
      - owner: owner1
        repo: repo1
        path: values/production.yaml
        branch: main
      - owner: owner2
        repo: repo2
        path: secrets/production.yaml
        branch: main
    output_repo:
      owner: output-owner
      repo: repo
      path: k8s/production
      filename: secrets.yaml
      branch: main
  
  - name: staging
    values_repos:
      - owner: owner1
        repo: repo1
        path: values/staging.yaml
        branch: main
      - owner: owner2
        repo: repo2
        path: secrets/staging.yaml
        branch: main
    output_repo:
      owner: output-owner
      repo: repo
      path: k8s/staging
      filename: secrets.yaml
      branch: staging
```

### 2. JSON Environment Variable

If no configuration file is found, the application checks for a `CONFIG_GROUPS` environment variable containing a JSON array of configuration groups.

Example:
```
CONFIG_GROUPS=[{"name":"production","values_repos":[{"owner":"owner1","repo":"repo1","path":"values/production.yaml","branch":"main"},{"owner":"owner2","repo":"repo2","path":"secrets/production.yaml","branch":"main"}],"output_repo":{"owner":"output-owner","repo":"repo","path":"k8s/production","filename":"secrets.yaml","branch":"main"}},{"name":"staging","values_repos":[{"owner":"owner1","repo":"repo1","path":"values/staging.yaml","branch":"main"},{"owner":"owner2","repo":"repo2","path":"secrets/staging.yaml","branch":"main"}],"output_repo":{"owner":"output-owner","repo":"repo","path":"k8s/staging","filename":"secrets.yaml","branch":"staging"}}]
```

### 3. Prefixed Environment Variables

If the `CONFIG_GROUPS` variable is not set, the application looks for environment variables with the `CONFIG_GROUP_*` prefix.

Example:
```
CONFIG_GROUP_1_NAME=production
CONFIG_GROUP_1_VALUES_REPO_1=owner1/repo1:values/production.yaml:main
CONFIG_GROUP_1_VALUES_REPO_2=owner2/repo2:secrets/production.yaml:main
CONFIG_GROUP_1_OUTPUT_REPO=output-owner/repo:k8s/production/secrets.yaml:main

CONFIG_GROUP_2_NAME=staging
CONFIG_GROUP_2_VALUES_REPO_1=owner1/repo1:values/staging.yaml:main
CONFIG_GROUP_2_VALUES_REPO_2=owner2/repo2:secrets/staging.yaml:main
CONFIG_GROUP_2_OUTPUT_REPO=output-owner/repo:k8s/staging/secrets.yaml:staging
```

### Basic Environment Variables

The following environment variables are required:

- `GITHUB_TOKEN`: GitHub Personal Access Token
- `REPO_OWNER`: GitHub repository owner
- `REPO_NAME`: GitHub repository name
- `PORT` (optional): Port for the server to listen on (default: 4000)
- `HOST` (optional): Network interface to bind to (default: "0.0.0.0" - all interfaces)
  - Use "0.0.0.0" to bind to all network interfaces
  - Use "127.0.0.1" to bind to localhost only (for development)
  - Use a specific IP address to bind to a particular network interface
- `CONFIG_PATH` (optional): Path to the configuration file (default: "config.yaml")

### Health Check Endpoints

The application provides the following health check endpoints:

- `/healthz`: Basic health check that returns 200 OK if the server is running
- `/healthz/ready`: Readiness check that verifies all dependencies (GitHub API, Helm CLI) are available

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
docker run -p 4000:4000 \
  -e GITHUB_TOKEN=your_github_token \
  -e REPO_OWNER=your_repo_owner \
  -e REPO_NAME=your_repo_name \
  -e HOST=0.0.0.0 \
  yaml-helm-pipeline

# Or run with .env file
docker run -p 4000:4000 \
  --env-file .env \
  yaml-helm-pipeline
```

The application will be available at http://localhost:4000.

## Repository Structure

The repository should have the following structure:

```
your-repo/
├── Chart.yaml       # Helm chart metadata
├── templates/       # Helm templates directory
└── values/          # Values directory
    └── values.yaml  # Values file
```

This follows the standard Helm chart structure, with the repository root serving as the chart directory.

## Usage

1. Open the application in your browser
2. Select a branch from the dropdown
3. Click "Preview Changes" to see the keys that will be included
4. Review the keys (values are not shown for security)
5. Enter a commit message and click "Commit Changes"
6. The changes will be committed to the selected branch

## License

This project is licensed under the MIT License - see the LICENSE file for details.
