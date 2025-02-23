# Email Validation API Design Document

## Section 1: Project Overview

### **Project Idea**
The Email Validation API is a lightweight, cost-effective, and reliable solution for validating email addresses in real-time. The primary goal is to provide indie hackers, small startups, and developers with a tool to ensure the quality and deliverability of email addresses. This service prevents invalid, disposable, or role-based email addresses from entering their systems, reducing bounce rates, improving email marketing efficiency, and maintaining sender reputation.

### **Target Population**
The API is specifically tailored for:
1. **Indie Hackers**: Solo developers or small teams creating side projects who need a budget-friendly solution.
2. **Small Startups**: Early-stage companies looking for simple yet effective email validation to integrate into their onboarding or marketing processes.
3. **Developers**: Developers building SaaS, web apps, or mobile applications who require an easy-to-integrate email validation service.
4. **Marketers**: Digital marketers cleaning email lists for improved deliverability and outreach success.

### **Promises**
* Extremely cheap
* Very fast
* Reliable
* JS + Python + REST APIs
* Pay as you go - refund the rest

### **Core Features**
1. **Syntax Validation**:
   - Checks if the email address is properly formatted.

2. **Domain Validation**:
   - Confirms if the domain exists and can receive emails.

3. **MX Record Validation**:
   - Ensures the domain has valid Mail Exchange (MX) records configured.

4. **Mailbox Validation**:
   - Verifies if the email address exists on the mail server.

5. **Disposable Email Detection**:
   - Detects if the email address belongs to a disposable email provider (e.g., 10minutemail.com).

6. **Role-Based Email Detection**:
   - Flags role-based email addresses like `admin@`, `support@`, or `sales@`.

7. **Real-Time API Calls**:
   - Fully synchronous API that returns validation results immediately without storing sensitive data.

8. **Pay-As-You-Go Pricing**:
   - Affordable, credit-based pricing with refunds for unused credits.

---

## Section 2: REST API Design

### **Base URL**
`https://api.myemailvalidator.com`

---

### **Endpoints**

#### **1. POST /validate**
**Description**: Validate a single email address.

**Request**:
```json
{
  "email": "example@email.com"
}
```

**Response (Success)**:
```json
{
  "email": "example@email.com",
  "validations": {
    "syntax": true,
    "domain_exists": true,
    "mx_records": true,
    "mailbox_exists": true,
    "is_disposable": false,
    "is_role_based": false
  },
  "score": 95,
  "message": "Email is valid and deliverable."
}
```

**Response (Failure)**:
```json
{
  "email": "example@email.com",
  "validations": {
    "syntax": false,
    "domain_exists": false,
    "mx_records": false,
    "mailbox_exists": false,
    "is_disposable": false,
    "is_role_based": false
  },
  "score": 0,
  "message": "Invalid email format."
}
```

---

#### **2. POST /validate/batch**
**Description**: Validate multiple email addresses in a single request.

**Request**:
```json
{
  "emails": [
    "example1@email.com",
    "example2@email.com"
  ]
}
```

**Response**:
```json
{
  "results": [
    {
      "email": "example1@email.com",
      "is_valid": true,
      "score": 95,
      "message": "Email is valid."
    },
    {
      "email": "example2@email.com",
      "is_valid": false,
      "score": 0,
      "message": "Invalid domain."
    }
  ]
}
```

---

#### **3. GET /status**
**Description**: Check the current status of the API.

**Response**:
```json
{
  "status": "healthy",
  "uptime": "120 days",
  "requests_handled": 15000,
  "average_response_time_ms": 25
}
```

---

#### **4. POST /typo-suggestions**
**Description**: Suggest corrections for typos in an email address.

**Request**:
```json
{
  "email": "user@gmial.com"
}
```

**Response**:
```json
{
  "email": "user@gmial.com",
  "suggestions": ["user@gmail.com"]
}
```

---

#### **5. GET /domains/disposable**
**Description**: Check if a domain belongs to a disposable email provider.

**Request**:
```json
{
  "domain": "10minutemail.com"
}
```

**Response**:
```json
{
  "domain": "10minutemail.com",
  "is_disposable": true
}
```

---

#### **6. GET /credits**
**Description**: Check the remaining credits for the API key.

**Response**:
```json
{
  "remaining_credits": 480,
  "total_credits": 1000
}
```

---

### **Key Considerations**
1. **Authentication**: Require an API key in the request headers for all endpoints except `/status`.
2. **Rate Limiting**: Implement rate limiting to prevent abuse (e.g., 100 requests per minute for free-tier users).
3. **Error Handling**: Provide clear error codes and messages for invalid requests or system errors.
4. **Performance Optimization**:
   - Cache DNS and MX lookups temporarily to reduce redundant queries.
   - Use a lightweight framework (e.g., FastAPI, Go Fiber) for efficient processing.

---

This design ensures the Email Validation API is easy to use, efficient, and tailored to the needs of indie hackers and small startups.


# Tech Stack
- Go
- Redis
- Docker

# Key Principles
- Simplicity
- Performance
- Reliability
- Cost-Effectiveness
- Testability
- Extensibility
- Low Latency
- High Throughput
- High Availability
- Maintainability


# Architecture
- Layered Architecture
- Separation of Concerns
- Single Responsibility
- YAGNI
- DRY
- KISS
- SOLID

# Deployment Strategy

## Platform Choice: Fly.io

### Rationale
- Global edge deployment for low-latency responses worldwide
- Generous free tier (3 shared-cpu-1x 256mb VMs)
- Simple deployment process
- Native Go support
- Built-in PostgreSQL and Redis support
- Automatic SSL certificate management
- Pay-as-you-grow pricing model aligns with our cost-effective goals

### Deployment Phases

#### 1. MVP Phase
- Utilize free tier for development and testing
- Deploy to multiple regions for global coverage
- Basic monitoring setup
- Estimated Cost: ~$5/month for production workload

#### 2. Growth Phase
- Scale horizontally with increased demand
- Add more regions based on user distribution
- Implement advanced monitoring
- Estimated Cost: Scales with usage, typically $20-50/month for moderate traffic

#### 3. Optimization Phase
- Fine-tune resource allocation
- Implement advanced caching strategies
- Optimize database queries and connections
- Monitor and adjust regions based on traffic patterns

### Infrastructure Components
1. **Application Servers**
   - Go API distributed across global regions
   - Automatic scaling based on load

2. **Caching Layer**
   - Redis for DNS and MX lookup caching
   - Distributed cache for improved performance

3. **Database**
   - Fly PostgreSQL for user data and API keys
   - Automatic backups and failover

4. **Load Balancing**
   - Built-in Fly.io load balancers
   - Automatic traffic distribution

5. **SSL/Security**
   - Automatic SSL certificate management
   - DDoS protection included

### Monitoring and Maintenance
1. **Performance Monitoring**
   - Built-in Fly.io metrics
   - Custom monitoring with Grafana Cloud (free tier)
   - Response time tracking per region

2. **Cost Monitoring**
   - Daily usage tracking
   - Cost optimization alerts
   - Resource utilization metrics

3. **Backup Strategy**
   - Automated database backups
   - Regular configuration backups
   - Disaster recovery planning

### Cost Control Measures
1. **Resource Optimization**
   - Aggressive caching for DNS lookups
   - Connection pooling for databases
   - Efficient memory usage in Go services

2. **Scaling Rules**
   - Scale based on actual usage metrics
   - Automatic scale-down during low traffic
   - Region-specific scaling policies

3. **Traffic Management**
   - Rate limiting per API key
   - Smart routing to nearest region
   - Cache hit ratio optimization

# MVP Monitoring Setup

## Core Metrics (Using Grafana Cloud Free Tier)

### 1. Application Health
- **Endpoint Health**
  - `/status` endpoint uptime
  - Response times (p95, p99)
  - Error rates by endpoint
  - Request volume

- **System Health**
  - CPU usage
  - Memory usage
  - Goroutine count
  - Garbage collection metrics

### 2. Business Metrics
- **API Usage**
  - Requests per API key
  - Daily active users
  - Credit consumption rate
  - Most used endpoints

- **Validation Results**
  - Validation success/failure rates
  - Average validation scores
  - Cache hit ratios
  - DNS lookup performance

### 3. Infrastructure Metrics
- **Database**
  - Connection pool status
  - Query performance
  - Error rates
  - Disk usage

- **Redis Cache**
  - Hit/miss ratios
  - Memory usage
  - Eviction rates
  - Connection status

## Implementation

### 1. Prometheus Integration
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "email_validator_requests_total",
            Help: "Total number of email validation requests",
        },
        []string{"endpoint", "status"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "email_validator_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"endpoint"},
    )
    
    validationScores = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "email_validator_scores",
            Help:    "Distribution of email validation scores",
            Buckets: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
        },
        []string{"validation_type"},
    )
)
```

### 2. Alert Rules
```yaml
groups:
  - name: email-validator-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(email_validator_requests_total{status="error"}[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High error rate detected
          
      - alert: SlowResponses
        expr: histogram_quantile(0.95, rate(email_validator_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: Slow response times detected
          
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes > 256 * 1024 * 1024
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High memory usage detected
```

### 3. Grafana Dashboards
- **Main Dashboard**: Overall system health and performance
- **Business Metrics**: Usage patterns and customer behavior
- **Infrastructure**: Detailed system metrics
- **Alerts**: Active and historical alerts

## Monitoring Workflow

### 1. Daily Checks
- Review error rates and response times
- Check credit consumption patterns
- Monitor system resource usage
- Verify cache performance

### 2. Weekly Analysis
- Review usage patterns
- Analyze performance trends
- Check resource optimization opportunities
- Update alert thresholds if needed

### 3. Monthly Review
- Generate usage reports
- Review and optimize costs
- Plan capacity adjustments
- Update monitoring rules

## Cost Estimation
- Grafana Cloud Free Tier
  - 10K series
  - 14-day metric retention
  - 3 users
  - Basic alerting
- Estimated Additional Costs: $0/month for MVP phase
