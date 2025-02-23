# Email Validator Service

[![Tests](https://github.com/umuterturk/email-verifier/actions/workflows/tests.yml/badge.svg)](https://github.com/umuterturk/email-verifier/actions/workflows/tests.yml)

A high-performance, cost-effective email validation service designed for indie hackers and small startups. The service validates email addresses in real-time, checking syntax, domain existence, MX records, and detecting disposable email providers.

## Features

- ✅ Syntax validation
- 🌐 Domain existence check
- 📨 MX record validation
- 🚫 Disposable email detection
- 👥 Role-based email detection
- ✍️ Typo suggestions
- 📊 Real-time monitoring
- 🔄 Batch processing support

## Tech Stack

- Go 1.21+
- Redis (caching, only for domains, not email addresses)
- Prometheus (metrics)
- Grafana (monitoring)
- Docker & Docker Compose

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git
- VSCode (recommended)

## Development Environment Setup

### 1. Install Go

#### macOS
```bash
# Using Homebrew
brew install go

# Verify installation
go version  # Should show go version 1.21 or higher
```

#### Linux (Ubuntu/Debian)
```bash
# Add Go repository
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update

# Install Go
sudo apt install golang-go

# Verify installation
go version
```

#### Windows
1. Download the installer from [Go Downloads](https://golang.org/dl/)
2. Run the installer
3. Open Command Prompt and verify:
```bash
go version
```

### 2. Configure Go Environment

Add these to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):
```bash
# Go environment variables
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

Reload your shell or run:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### 3. Install Docker

#### macOS
1. Download [Docker Desktop for Mac](https://www.docker.com/products/docker-desktop)
2. Install and start Docker Desktop
3. Verify installation:
```bash
docker --version
docker-compose --version
```

#### Linux
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose

# Start Docker
sudo systemctl start docker

# Add your user to docker group (optional)
sudo usermod -aG docker $USER

# Verify installation
docker --version
docker-compose --version
```

#### Windows
1. Download and install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)
2. Enable WSL 2 if prompted
3. Start Docker Desktop
4. Verify in Command Prompt:
```bash
docker --version
docker-compose --version
```

### 4. VSCode Setup

1. **Install VSCode**
   - Download from [Visual Studio Code](https://code.visualstudio.com/)
   - Install for your platform

2. **Install Required Extensions**
   ```
   # Essential Extensions
   - Go (by Go Team at Google)
   - Docker (by Microsoft)
   - GitLens (by GitKraken)
   - Remote Development (by Microsoft)
   
   # Recommended Extensions
   - Error Lens
   - Go Test Explorer
   - YAML
   - Markdown All in One
   ```

3. **Configure Go Extension**
   After installing the Go extension:
   1. Open Command Palette (Cmd/Ctrl + Shift + P)
   2. Type "Go: Install/Update Tools"
   3. Select all tools and click OK
   
   This will install:
   - gopls (Go language server)
   - dlv (debugger)
   - golangci-lint (linter)
   - goimports (code formatter)
   - and other essential Go tools

4. **VSCode Settings**
   Create or update `.vscode/settings.json`:
   ```json
   {
     "go.useLanguageServer": true,
     "go.lintTool": "golangci-lint",
     "go.lintFlags": ["--fast"],
     "editor.formatOnSave": true,
     "[go]": {
       "editor.defaultFormatter": "golang.go",
       "editor.codeActionsOnSave": {
         "source.organizeImports": true
       }
     }
   }
   ```

5. **Debugging Setup**
   Create `.vscode/launch.json`:
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Launch Email Validator",
         "type": "go",
         "request": "launch",
         "mode": "auto",
         "program": "${workspaceFolder}",
         "env": {},
         "args": []
       }
     ]
   }
   ```

### 5. Install Additional Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air for hot reload (optional)
go install github.com/cosmtrek/air@latest

# Install mockgen for testing
go install github.com/golang/mock/mockgen@latest
```

## Setup

1. **Clone the Repository**
```bash
git clone https://github.com/umuterturk/email-verifier.git
cd email-verifier
```

2. **Install Dependencies**
```bash
go mod tidy
```

3. **Start Monitoring Stack**
```bash
docker-compose up -d
```

4. **Build and Run the Service**
```bash
go build
./emailvalidator
```

The service will be available at:
- API: http://localhost:8080
- Metrics: http://localhost:8080/metrics
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

## Development

### Project Structure
```
.
├── cmd/                    # Command line tools
├── internal/              
│   ├── api/               # HTTP handlers
│   ├── middleware/        # HTTP middleware components
│   ├── model/            # Data models
│   ├── repository/       # Data access layer
│   └── service/          # Business logic
├── pkg/
│   ├── validator/        # Email validation logic
│   ├── monitoring/       # Metrics and monitoring
│   └── cache/           # Caching implementation
├── test/                 # Load and integration tests
└── config/               # Configuration files
```

### Running Tests

1. **Run All Tests (excluding load tests)**
```bash
go test ./... -v -skip "Load"
```

2. **Run Tests with Race Detection**
```bash
go test -race ./... -skip "Load"
```

Race detection is crucial for identifying potential data races in concurrent code. It's automatically run in CI/CD pipelines and should be run locally before submitting changes.

Common race conditions to watch for:
- Concurrent map access
- Shared variable access without proper synchronization
- Channel operations
- Goroutine lifecycle management

3. **Run Load Tests**
```bash
go test ./test/load_test.go -v
```

### Code Quality

1. **Run Linter**
```bash
golangci-lint run
```

2. **Format Code**
```bash
go fmt ./...
```

3. **Check for Common Mistakes**
```bash
go vet ./...
```

### Continuous Integration

The project uses GitHub Actions for CI/CD with the following checks:
- Unit tests with race detection
- Integration tests
- Acceptance tests
- Code linting
- Code formatting
- Security scanning
- Dependency updates

All these checks must pass before code can be merged into the main branch.

## Testing

The project includes several types of tests:

### API Testing Script
```bash
# Run the comprehensive API test suite
./test_api.sh
```

The `test_api.sh` script provides comprehensive testing of all API endpoints with:
- Single email validation (GET/POST)
- Batch email validation (GET/POST)
- Typo suggestions (GET/POST)
- Special cases (disposable emails, role-based emails)
- Error cases (invalid JSON, wrong content type)
- Status checks
- Performance metrics for each endpoint type

The script outputs:
- Detailed test results with colored output
- Success/failure status for each test
- Timing statistics for each endpoint type
- Overall performance metrics

### Unit Tests
```bash
# Run all unit tests
go test ./tests/unit/... -v

# Run with coverage
go test ./tests/unit/... -v -cover
```

Unit tests cover:
- Email validation logic
- Service layer functionality
- Validator components
- Cache behavior

### Integration Tests
```bash
# Run all integration tests
go test ./tests/integration/... -v
```

Integration tests cover:
- HTTP handlers
- API endpoints
- Request/response handling
- Error scenarios

### Acceptance Tests
```bash
# Run all acceptance tests
go test ./tests/acceptance/... -v
```

Acceptance tests cover:
- End-to-end email validation
- Concurrent request handling
- Error scenarios
- API behavior

### Performance Testing

When running performance tests, follow these best practices:

1. **Gradual Testing**
   Start with minimal load and increase gradually:
   ```bash
   # 1. Single concurrent request
   go test ./tests/acceptance -run TestAcceptanceConcurrentRequests -v -parallel 1
   
   # 2. If successful, increase parallel requests
   go test ./tests/acceptance -run TestAcceptanceConcurrentRequests -v -parallel 5
   
   # 3. Further increase if stable
   go test ./tests/acceptance -run TestAcceptanceConcurrentRequests -v -parallel 10
   ```

2. **Monitoring During Tests**
   Open these in separate terminals:
   ```bash
   # Watch service status
   watch -n1 curl -s http://localhost:8080/status
   
   # Monitor system resources
   top -p $(pgrep emailvalidator)
   
   # Watch Docker containers
   watch -n1 docker-compose ps
   ```

## API Endpoints

### 1. Validate Single Email
```bash
# Using POST
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'

# Using GET
curl "http://localhost:8080/validate?email=user@example.com"
```

### 2. Batch Validation
```bash
# Using POST
curl -X POST http://localhost:8080/validate/batch \
  -H "Content-Type: application/json" \
  -d '{"emails": ["user1@example.com", "user2@example.com"]}'

# Using GET
curl "http://localhost:8080/validate/batch?email=user1@example.com&email=user2@example.com"
```

### 3. Get Typo Suggestions
```bash
# Using POST
curl -X POST http://localhost:8080/typo-suggestions \
  -H "Content-Type: application/json" \
  -d '{"email": "user@gmial.com"}'

# Using GET
curl "http://localhost:8080/typo-suggestions?email=user@gmial.com"
```

### 4. Check API Status
```bash
curl http://localhost:8080/status
```

## Configuration

Create `.env` and configure the following environment variables:

```bash
# Redis configuration
REDIS_URL=redis://username:password@host:port

# Server configuration
PORT=8080

# Optional: Prometheus configuration
PROMETHEUS_ENABLED=true

```

The `test_api.sh` script will automatically load these environment variables if they are present in the `.env` file.