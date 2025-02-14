# Email Validator Service

[![Tests](https://github.com/yourusername/email-verifier/actions/workflows/tests.yml/badge.svg)](https://github.com/yourusername/email-verifier/actions/workflows/tests.yml)

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
git clone https://github.com/yourusername/email-verifier.git
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
│   ├── model/             # Data models
│   └── service/           # Business logic
├── pkg/
│   ├── validator/         # Email validation logic
│   └── monitoring/        # Metrics and monitoring
├── test/                  # Load and integration tests
└── config/                # Configuration files
```

### Making Changes

1. **Run Tests**
```bash
go test ./...
```

2. **Run Linter**
```bash
golangci-lint run
```

3. **Format Code**
```bash
go fmt ./...
```

## Testing

### Unit Tests
```bash
go test -v ./... -skip "TestLoadGeneration"
```
### Unit Tests with Coverage
```bash
go test -v ./... -skip "TestLoadGeneration" -cover
```

### Load Testing

⚠️ **Important**: Before running load tests, ensure:
1. The email validator service is running (`./emailvalidator`)
2. The monitoring stack is up (`docker-compose up -d`)
3. You can access the service at http://localhost:8080/status

#### Running Load Tests

1. **Start Prerequisites** (in separate terminal windows)
```bash
# Terminal 1: Start monitoring stack
docker-compose up -d

# Terminal 2: Build and start the service
go build
./emailvalidator

# Terminal 3: Verify service is running
curl http://localhost:8080/status
```

2. **Run Load Tests** (in a new terminal)

Quick Local Test (30s):
```bash
go test ./test -v -run TestLoadGeneration
```

Custom Load Test:
```bash
go test ./test -v -run TestLoadGeneration \
  -duration 2m \
  -concurrent-users 5 \
  -short=false
```

#### Troubleshooting Load Test Errors

1. **Connection Refused Errors**
   ```
   dial tcp [::1]:8080: connect: connection refused
   ```
   Even if you think the service is running, verify:

   a. **Check Service Status**
   ```bash
   # Check if the process is actually running
   ps aux | grep emailvalidator
   
   # Verify the port is actually listening
   netstat -an | grep 8080
   # or
   lsof -i :8080
   ```

   b. **Test Service Connectivity**
   ```bash
   # Try different ways to connect
   curl http://localhost:8080/status
   curl http://127.0.0.1:8080/status
   wget http://localhost:8080/status
   ```

   c. **Check Service Logs**
   ```bash
   # If running in foreground, check the terminal
   # If running in background, find and check logs:
   ps aux | grep emailvalidator
   ```

   d. **Common Solutions**:
   - If service shows running but won't connect:
     ```bash
     # Kill and restart the service
     pkill emailvalidator
     ./emailvalidator
     ```
   - If localhost isn't resolving:
     ```bash
     # Add to /etc/hosts if missing:
     127.0.0.1 localhost
     ::1 localhost
     ```
   - Try explicit IP:
     ```bash
     # Modify test/load_test.go to use
     "http://127.0.0.1:8080" instead of "http://localhost:8080"
     ```

2. **Continuous Connection Errors**
   If you're getting repeated connection errors:

   a. **Check Service Stability**
   ```bash
   # Run with verbose logging
   ./emailvalidator -v
   
   # Monitor service stability
   watch -n1 "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/status"
   ```

   b. **Test Service Load**
   ```bash
   # Start with minimal load
   go test ./test -v -run TestLoadGeneration -concurrent-users 1 -duration 10s
   
   # If successful, gradually increase load
   ```

3. **Docker Network Issues**
   If using Docker and still having issues:
   ```bash
   # Check Docker network
   docker network ls
   docker network inspect bridge
   
   # Ensure host.docker.internal is working
   docker run --rm alpine ping host.docker.internal
   
   # Check Prometheus connectivity
   curl http://localhost:9090/api/v1/targets
   ```

4. **Quick Verification Script**
   Create a file `verify.sh`:
   ```bash
   #!/bin/bash
   echo "Checking service status..."
   curl -s http://localhost:8080/status || echo "Service not responding"
   
   echo -e "\nChecking Docker containers..."
   docker-compose ps
   
   echo -e "\nChecking port usage..."
   lsof -i :8080
   
   echo -e "\nChecking Prometheus targets..."
   curl -s http://localhost:9090/api/v1/targets
   ```
   Run it:
   ```bash
   chmod +x verify.sh
   ./verify.sh
   ```

5. **Debug Mode**
   Run the service with debug logging:
   ```bash
   # Set environment variable for debug mode
   export DEBUG=true
   
   # Run service
   ./emailvalidator
   ```

#### Load Test Best Practices

1. **Gradual Testing**
   Start with minimal load and increase gradually:
   ```bash
   # 1. Single user, short duration
   go test ./test -v -run TestLoadGeneration -concurrent-users 1 -duration 10s
   
   # 2. If successful, increase duration
   go test ./test -v -run TestLoadGeneration -concurrent-users 1 -duration 30s
   
   # 3. If successful, increase users
   go test ./test -v -run TestLoadGeneration -concurrent-users 2 -duration 30s
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

### Monitoring Tests

1. **Access Grafana**
   - URL: http://localhost:3000
   - Username: admin
   - Password: admin

2. **Import Dashboard**
   - Go to "+" -> "Import"
   - Upload `grafana-dashboard.json`

3. **View Metrics**
   - Request rates
   - Response times
   - Success/failure rates
   - Cache performance
   - System metrics

## CI/CD

### GitHub Actions

The project includes automated workflows for:

1. **Load Testing**
   - Runs weekly (Sunday midnight)
   - Can be triggered manually
   - Customizable duration and concurrent users
   - Generates detailed reports

To run manual load test:
1. Go to GitHub Actions
2. Select "Load Testing"
3. Click "Run workflow"
4. Configure:
   - Test duration (seconds)
   - Number of concurrent users

### Test Reports

Load test results are available as:
1. GitHub Actions logs
2. Downloadable artifacts
3. PR comments (when run on pull requests)

## API Endpoints

### 1. Validate Single Email
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

### 2. Batch Validation
```bash
curl -X POST http://localhost:8080/validate/batch \
  -H "Content-Type: application/json" \
  -d '{"emails": ["user1@example.com", "user2@example.com"]}'
```

### 3. Get Typo Suggestions
```bash
curl -X POST http://localhost:8080/typo-suggestions \
  -H "Content-Type: application/json" \
  -d '{"email": "user@gmial.com"}'
```

### 4. Check API Status
```bash
curl http://localhost:8080/status
```

## Monitoring

### Key Metrics

1. **Application Health**
   - Endpoint uptime
   - Response times (p95, p99)
   - Error rates
   - Request volume

2. **Business Metrics**
   - Validation success rates
   - Cache hit ratios
   - Most used endpoints
   - Average validation scores

3. **System Metrics**
   - CPU usage
   - Memory usage
   - Goroutine count
   - GC metrics

## Troubleshooting

1. **Service Won't Start**
   - Check port 8080 is free
   - Verify Go version (`go version`)
   - Check error logs

2. **Monitoring Issues**
   - Ensure Docker is running
   - Check container logs (`docker-compose logs`)
   - Verify Prometheus targets

3. **Load Test Failures**
   - Verify service is running
   - Check system resources
   - Review test logs

## Contributing

1. Fork the repository
2. Create your feature branch
3. Run tests and linter
4. Submit a pull request

## License

MIT License - see LICENSE file for details

### 4. Install golangci-lint

golangci-lint is used for code quality checks in this project. It runs multiple linters concurrently and has integrations with popular editors.

#### macOS
```bash
# Using Homebrew
brew install golangci-lint

# Verify installation
golangci-lint --version
```

#### Linux and Windows
```bash
# Binary installation
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Verify installation
golangci-lint --version
```

## Code Quality

### Running the Linter

The project uses golangci-lint for code quality checks. To run the linter:

```bash
# Run all linters
golangci-lint run

# Run specific linters
golangci-lint run --disable-all -E errcheck,gosimple,govet,ineffassign,staticcheck,typecheck,unused

# Run linters with auto-fix
golangci-lint run --fix
```

### Common Linting Issues and Solutions

1. **G107: Potential HTTP request made with variable url**
   ```go
   // Incorrect
   resp, err := http.Get(url)

   // Correct
   req, err := http.NewRequest(http.MethodGet, url, nil)
   if err != nil {
       return err
   }
   resp, err := http.DefaultClient.Do(req)
   ```

2. **ineffectual assignment to err**
   ```go
   // Incorrect
   req, err := http.NewRequest("GET", url, nil)
   resp, err = client.Do(req)  // err from NewRequest is lost

   // Correct
   req, reqErr := http.NewRequest("GET", url, nil)
   if reqErr != nil {
       return reqErr
   }
   resp, err = client.Do(req)
   ```

3. **unused variable/parameter**
   ```go
   // Incorrect
   func process(ctx context.Context, data string) error {
       return nil  // ctx is unused
   }

   // Correct
   func process(_ context.Context, data string) error {
       return nil
   }
   ```

### CI/CD Integration

The linter is integrated into our CI/CD pipeline in `.github/workflows/tests.yml`:

```yaml
- name: Run linter
  uses: golangci/golangci-lint-action@v4
  with:
    version: latest
    args: --timeout=5m
    skip-cache: false
    skip-pkg-cache: false
    skip-build-cache: false
    only-new-issues: true
```

### Pre-commit Hook Setup

1. Create `.git/hooks/pre-commit`:
```bash
#!/bin/sh
# Run golangci-lint before commit
golangci-lint run

# Check the exit code
if [ $? -ne 0 ]; then
    echo "Linting failed! Please fix the issues before committing."
    exit 1
fi
```

2. Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

### Project-specific Linting Rules

Our `.golangci.yml` enforces:

1. **Enabled Linters**
   - `errcheck`: Find unchecked errors
   - `gosimple`: Simplify code
   - `govet`: Report suspicious code
   - `ineffassign`: Detect ineffectual assignments
   - `staticcheck`: State of the art checks
   - `typecheck`: Type-checking
   - `unused`: Find unused code

2. **Custom Rules**
   ```yaml
   linters-settings:
     govet:
       check-shadowing: true
     errcheck:
       check-type-assertions: true
     gosimple:
       checks: ["all"]
     staticcheck:
       checks: ["all"]
   ```

3. **Ignored Issues**
   - `EXC0001`: Complexity checks for test files
   - `ST1000`: Package comment style

### Editor Integration

#### VS Code
```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.lintOnSave": "package"
}
```

#### GoLand
1. Go to Settings → Go → Go Linter
2. Select 'golangci-lint'
3. Set "--fast" in "Arguments"

#### Vim/Neovim
Add to your config:
```vim
let g:go_metalinter_command = "golangci-lint"
let g:go_metalinter_autosave = 1
```

### Troubleshooting Linter Issues

1. **Linter is too slow**
   ```bash
   # Use fast mode
   golangci-lint run --fast

   # Run only on changed files
   golangci-lint run --new-from-rev=HEAD~1
   ```

2. **Memory issues**
   ```bash
   # Limit memory usage
   GOGC=50 golangci-lint run
   ```

3. **Cache issues**
   ```bash
   # Clear cache
   golangci-lint cache clean
   ```

To modify linter settings, edit the `.golangci.yml` file in the project root. 