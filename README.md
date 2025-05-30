# Email Validator Service

[![Tests](https://github.com/umuterturk/email-verifier/actions/workflows/tests.yml/badge.svg)](https://github.com/umuterturk/email-verifier/actions/workflows/tests.yml)
[![Build and Publish Docker Image](https://github.com/umuterturk/email-verifier/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/umuterturk/email-verifier/actions/workflows/docker-publish.yml)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-Support-yellow.svg)](https://www.buymeacoffee.com/codeonbrew)
[![Patreon](https://img.shields.io/badge/Patreon-Support-f96854.svg)](https://www.patreon.com/codeonbrew)

A high-performance, cost-effective email validation service designed for indie hackers and small startups. The service validates email addresses in real-time, checking syntax, domain existence, MX records, and detecting disposable email providers. The main focus is on precision instead of recall, meaning instead of edge cases the focus is on having the biggest coverage.

🌐 **Website**: [https://rapid-email-verifier.fly.dev/](https://rapid-email-verifier.fly.dev/)

🚀 **API**
[https://rapid-email-verifier.fly.dev/api/validate](https://rapid-email-verifier.fly.dev/api/validate?email=user@example.com)

This is a completely free and open source email validation API that never stores your data. Built to support solopreneurs and the developer community. Features include:
- Zero data storage - your emails are never saved!
- GDPR, CCPA, and PIPEDA compliant
- No authentication required
- No usage limits
- Quick response times
- Batch validation up to 100 emails

## Features

- ✅ Syntax validation
- 🌐 Domain existence check
- 📨 MX record validation
- 🚫 Disposable email detection
- 👥 Role-based email detection
- 🔍 Email alias detection for major providers (Gmail, Yahoo, Outlook/Hotmail)
- ✍️ Typo suggestions
- 📊 Real-time monitoring
- 🔄 Batch processing support with domain-level optimizations

## Validation Examples

> **Note on Validation Approach**: While email RFCs (5321, 5322) define extensive rules for valid email formats, many of these rules are not practically enforced by modern email servers. Our validation takes a practical approach, focusing on rules that are actually enforced by major email providers. For example, we don't support quoted strings in local parts (`"john doe"@example.com`) even though they're technically valid according to RFC, as they're rarely supported in practice and often cause issues.

### Valid Email Formats
```json
// Standard email
{
  "email": "user@example.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID"
}

// Email with plus addressing
{
  "email": "user+tag@example.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID"
}

// International email (Unicode)
{
  "email": "用户@例子.广告",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID"
}

// Hindi characters
{
  "email": "अजय@डाटा.भारत",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID"
}
```

### Invalid Email Formats
```json
// Missing @ symbol
{
  "email": "invalid-email",
  "validations": {
    "syntax": false
  },
  "status": "INVALID_FORMAT"
}

// Double dots in local part
{
  "email": "john..doe@example.com",
  "validations": {
    "syntax": false
  },
  "status": "INVALID_FORMAT"
}

// Spaces in quotes (not supported)
{
  "email": "\"john doe\"@example.com",
  "validations": {
    "syntax": false
  },
  "status": "INVALID_FORMAT"
}

// Multiple @ symbols
{
  "email": "user@domain@example.com",
  "validations": {
    "syntax": false
  },
  "status": "INVALID_FORMAT"
}
```

### Special Cases
```json
// Disposable email detection
{
  "email": "user@tempmail.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true,
    "is_disposable": true
  },
  "status": "DISPOSABLE"
}

// Role-based email detection
{
  "email": "admin@company.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true,
    "is_role_based": true
  },
  "status": "PROBABLY_VALID"
}

// Email alias detection
{
  "email": "user.name+test@gmail.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID",
  "aliasOf": "username@gmail.com"
}
```

### Batch Validation
```json
// Request
POST /api/validate/batch
{
  "emails": [
    "user@example.com",
    "invalid-email",
    "user@nonexistent.com",
    "admin@company.com"
  ]
}

// Response
{
  "results": [
    {
      "email": "user@example.com",
      "validations": {
        "syntax": true,
        "domain_exists": true,
        "mx_records": true
      },
      "status": "VALID"
    },
    {
      "email": "invalid-email",
      "validations": {
        "syntax": false
      },
      "status": "INVALID_FORMAT"
    },
    {
      "email": "user@nonexistent.com",
      "validations": {
        "syntax": true,
        "domain_exists": false
      },
      "status": "INVALID_DOMAIN"
    },
    {
      "email": "admin@company.com",
      "validations": {
        "syntax": true,
        "domain_exists": true,
        "mx_records": true,
        "is_role_based": true
      },
      "status": "PROBABLY_VALID"
    }
  ]
}
```

## Email Alias Detection

The service can detect email aliases for major email providers and identify the canonical form of the email address.

### Supported Providers and Alias Formats

| Provider | Alias Format | Example | Canonical Form |
|----------|--------------|---------|----------------|
| Gmail | Dots in local part | `user.name@gmail.com` | `username@gmail.com` |
| Gmail | Plus addressing | `username+test@gmail.com` | `username@gmail.com` |
| Yahoo | Hyphen addressing | `username-test@yahoo.com` | `username@yahoo.com` |
| Outlook/Hotmail | Plus addressing | `username+test@outlook.com` | `username@outlook.com` |

### How It Works

When an email is validated, the service checks if it's an alias for a supported provider. If it is, the `aliasOf` field in the response will contain the canonical form of the email address.

```json
{
  "email": "user.name+test@gmail.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true
  },
  "status": "VALID",
  "aliasOf": "username@gmail.com"
}
```

## Batch Processing Optimizations

The service optimizes batch email validation by grouping emails by domain to avoid redundant domain checks. This significantly reduces network calls and resource usage:

- Emails are grouped by domain before validation
- Domain validations (existence, MX, disposable) are performed once per unique domain
- The cached domain results are applied to all emails with the same domain
- Original order and response accuracy are preserved

This optimization is particularly effective for large batches with common domains, reducing domain checks from O(n) to O(unique domains).

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
│   ├── model/             # Data models
│   ├── repository/        # Data access layer
│   └── service/           # Business logic
│       ├── email_service.go           # Core email validation service
│       ├── batch_validation_service.go # Batch processing with domain optimizations
│       ├── domain_validation_service.go # Domain validation with concurrency
│       └── interfaces.go              # Service interfaces for DI
├── pkg/
│   ├── validator/        # Email validation logic
│   │   ├── email.go      # Email syntax and validation
│   │   ├── domain.go     # Domain validation 
│   │   ├── role.go       # Role-based email detection
│   │   ├── disposable.go # Disposable email detection
│   │   └── alias_detector.go # Email alias detection
│   ├── monitoring/       # Metrics and monitoring
│   └── cache/            # Caching implementation
├── test/                 # Unit, integration and acceptance tests
└── config/               # Configuration files
```

### Service Architecture

The service follows a layered architecture with dependency injection for testability:

1. **API Layer**: HTTP handlers for REST endpoints
   - Validates request data
   - Translates to/from service layer models
   - Implements rate limiting and authentication

2. **Service Layer**: Core business logic
   - `EmailService`: Main validation service
   - `BatchValidationService`: Optimized batch processing with domain grouping
   - `DomainValidationService`: Concurrent domain validation
   - Uses interfaces for dependency injection

3. **Validation Layer**: Email validation logic
   - Syntax validation
   - Domain validation
   - Role-based detection
   - Disposable email detection
   - Email alias detection

4. **Repository Layer**: Data access
   - Cache implementation
   - External API clients

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

## Disclaimers

### Accuracy Disclaimer
The results provided by this service are based on best-effort validation and should be treated as recommendations rather than absolute truth. Several factors can affect the accuracy of results:
- Domain DNS records may change
- Temporary DNS resolution issues
- Network connectivity problems
- Email server configuration changes
- Rate limiting by email servers
- Syntax validation differences between email providers

### Legal Disclaimer
This service is provided "as is" without any warranties or guarantees of any kind, either express or implied. The validation results are for informational purposes only and should not be considered legally binding or definitive. We expressly disclaim any liability for damages of any kind arising from the use of this service or its results. Users are solely responsible for verifying the accuracy of email addresses through additional means before using them for any purpose.

## Email Existence Verification Limitations

It's important to understand that this service (like any email validation service) **cannot definitively verify if an email address actually exists** or if it will accept mail. For example, an address like `blablablabla@gmail.com` may or may not exist, and there is no reliable way to determine this with certainty.

Modern email providers have implemented extensive security measures specifically to prevent email existence checking:

- **Catch-all policies**: Many domains use catch-all configurations that accept mail for any address at their domain
- **Anti-harvesting measures**: Major providers like Gmail, Yahoo, and Outlook deliberately return positive responses for all syntactically valid addresses to prevent email harvesting
- **Anti-spam protections**: Email servers often employ throttling, temporary blocks, and other anti-spam techniques against automated verification attempts
- **SMTP blocking**: Most major email providers block SMTP-based verification techniques to protect user privacy
- **False positives/negatives**: Even when servers respond, they may provide misleading responses to protect privacy

While our service can validate syntax, verify domain existence, and check MX records, these checks only confirm that the domain *could* receive email, not that a specific address exists or is actively monitored.

For the most reliable verification, consider using a double opt-in process where users must confirm their email by clicking a link sent to that address.

## Using Docker

### Docker Hub Image

You can use the pre-built Docker image directly from Docker Hub:

```bash
# Pull the latest image
docker pull umutert/emailvalidator:latest

# Run the container
docker run -p 8080:8080 umutert/emailvalidator:latest
```

For detailed instructions on setting up automated Docker image deployment, see [Docker Deployment Guide](docs/docker-deployment.md).

### Docker Compose

To run the full stack with Redis, Prometheus, and Grafana:

```bash
# Pull and start all services
docker-compose up -d
```

This will start the email validation service along with:
- Redis for caching
- Prometheus for metrics collection
- Grafana for visualization (accessible at http://localhost:3000)

### Environment Variables

The Docker image supports the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | The port on which the service will listen |
| PROMETHEUS_ENABLED | false | Enable Prometheus metrics |
| REDIS_URL | | Redis connection URL (format: redis://host:port) |