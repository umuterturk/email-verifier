# Email Validator API Overview

This document provides a comprehensive overview of the Email Validator API endpoints available through RapidAPI.

## Base URL
```
https://email-validator.p.rapidapi.com
```

## Authentication
All API requests require the following headers:
- `X-RapidAPI-Host`: email-validator.p.rapidapi.com
- `X-RapidAPI-Secret`: Your RapidAPI proxy secret

## Endpoints

### 1. Validate Single Email
Validates a single email address and returns detailed validation results.

**Endpoint:** `/validate`

**Methods:** GET, POST

#### GET Request
```
GET /validate?email=user@example.com
```

#### POST Request
```json
{
    "email": "user@example.com"
}
```

#### Response
```json
{
    "email": "user@example.com",
    "validations": {
        "syntax": true,
        "domain_exists": true,
        "mx_records": true,
        "is_disposable": false,
        "is_role_based": false
    },
    "score": 100,
    "status": "VALID"
}
```

#### Possible Status Values
- `VALID`: Email is valid and deliverable
- `PROBABLY_VALID`: Email is probably valid but has some issues
- `INVALID`: Email has significant validation issues
- `MISSING_EMAIL`: No email was provided
- `INVALID_FORMAT`: Invalid email format
- `INVALID_DOMAIN`: Domain does not exist
- `NO_MX_RECORDS`: Domain cannot receive emails
- `DISPOSABLE`: Disposable email address detected

### 2. Batch Email Validation
Validates multiple email addresses in a single request.

**Endpoint:** `/validate/batch`

**Methods:** GET, POST

#### GET Request
```
GET /validate/batch?email=user1@example.com&email=user2@example.com
```

#### POST Request
```json
{
    "emails": [
        "user1@example.com",
        "user2@example.com"
    ]
}
```

#### Response
```json
{
    "results": [
        {
            "email": "user1@example.com",
            "validations": {
                "syntax": true,
                "domain_exists": true,
                "mx_records": true,
                "is_disposable": false,
                "is_role_based": false
            },
            "score": 100,
            "status": "VALID"
        },
        {
            "email": "user2@example.com",
            "validations": {
                "syntax": true,
                "domain_exists": true,
                "mx_records": true,
                "is_disposable": false,
                "is_role_based": true
            },
            "score": 80,
            "status": "PROBABLY_VALID"
        }
    ]
}
```

### 3. Typo Suggestions
Get suggestions for possible typos in email addresses.

**Endpoint:** `/typo-suggestions`

**Methods:** GET, POST

#### GET Request
```
GET /typo-suggestions?email=user@gmial.com
```

#### POST Request
```json
{
    "email": "user@gmial.com"
}
```

#### Response
```json
{
    "email": "user@gmial.com",
    "suggestions": [
        "user@gmail.com",
        "user@mail.com"
    ]
}
```

### 4. API Status
Check the current status and performance metrics of the API.

**Endpoint:** `/status`

**Method:** GET

#### Response
```json
{
    "status": "operational",
    "uptime": "99.99%",
    "requests_handled": 1000000,
    "average_response_time_ms": 150.5
}
```

## Error Responses
When an error occurs, the API will return an error response with an appropriate HTTP status code:

```json
{
    "error": "Error message describing what went wrong"
}
```

Common HTTP status codes:
- `400 Bad Request`: Invalid input or missing parameters
- `401 Unauthorized`: Invalid or missing API key
- `403 Forbidden`: Exceeded API quota or unauthorized access
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server-side error

## Rate Limiting
The API implements rate limiting based on your RapidAPI subscription plan. When you exceed the rate limit, you'll receive a 429 status code.

## Best Practices
1. Always check the validation status and score to make decisions
2. Handle disposable email addresses according to your use case
3. Implement proper error handling for all possible status codes
4. Cache validation results when appropriate
5. Use batch validation for multiple email checks to improve performance

## Support
For API support and questions, please contact through RapidAPI's support channels or visit the documentation page on RapidAPI.com. 