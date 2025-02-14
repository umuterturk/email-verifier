#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL
API_URL="https://fast-email-verifier.fly.dev"

# Initialize timing data arrays
declare -a endpoint_types=("single_validation" "batch_validation" "typo_suggestions" "special_cases" "error_cases" "status")
declare -a times=()
declare -a counts=()

# Initialize arrays with zeros
for i in "${!endpoint_types[@]}"; do
    times[$i]=0
    counts[$i]=0
done

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

# Function to check response against expected status
check_response() {
    local response=$1
    local expected_status=$2
    local description=$3
    
    # Extract the status field from the JSON response
    local actual_status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$expected_status" ]; then
        if [ "$actual_status" = "$expected_status" ]; then
            echo -e "${GREEN}✓ $description - Status: $actual_status${NC}"
        else
            echo -e "${RED}✗ $description - Expected: $expected_status, Got: $actual_status${NC}"
        fi
    fi
}

# Function to get endpoint type index
get_endpoint_index() {
    local type=$1
    for i in "${!endpoint_types[@]}"; do
        if [[ "${endpoint_types[$i]}" == "$type" ]]; then
            echo $i
            return
        fi
    done
}

# Function to test and print response
test_endpoint() {
    local description=$1
    local command=$2
    local endpoint_type=$3
    local expected_status=$4
    
    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "${BLUE}Command: ${command}${NC}"
    echo -e "${BLUE}Response:${NC}"
    
    # Add timing parameters to curl and suppress progress
    local curl_cmd="${command} -s -w '\nTime: %{time_total}s\n'"
    
    # Capture the output and timing
    local output
    output=$(eval $curl_cmd)
    local time=$(echo "$output" | grep "Time:" | cut -d' ' -f2 | sed 's/s//')
    echo "$output"
    
    # Check response against expected status
    if [ -n "$expected_status" ]; then
        check_response "$output" "$expected_status" "$description"
    fi
    
    # Store timing data
    local idx=$(get_endpoint_index "$endpoint_type")
    times[$idx]=$(echo "${times[$idx]} + $time" | bc)
    counts[$idx]=$((counts[$idx] + 1))
}

# 1. Single Email Validation Tests
print_header "Single Email Validation Tests"

# Valid email - POST
test_endpoint "Valid email (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"user@example.com\"}"' \
"single_validation" \
"VALID"

# Valid email - GET
test_endpoint "Valid email (GET)" \
"curl -X GET \"${API_URL}/validate?email=user@example.com\"" \
"single_validation" \
"VALID"

# Invalid email format - POST
test_endpoint "Invalid email format (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"invalid-email\"}"' \
"single_validation" \
"INVALID_FORMAT"

# Invalid email format - GET
test_endpoint "Invalid email format (GET)" \
"curl -X GET \"${API_URL}/validate?email=invalid-email\"" \
"single_validation" \
"INVALID_FORMAT"

# Empty email - POST
test_endpoint "Empty email (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"\"}"' \
"single_validation" \
"MISSING_EMAIL"

# Missing email parameter - GET
test_endpoint "Missing email parameter (GET)" \
"curl -X GET \"${API_URL}/validate\"" \
"single_validation"

# 2. Batch Email Validation Tests
print_header "Batch Email Validation Tests"

# Multiple valid emails - POST
test_endpoint "Multiple valid emails (POST)" \
'curl -X POST "${API_URL}/validate/batch" -H "Content-Type: application/json" -d "{\"emails\":[\"user1@example.com\",\"user2@example.com\"]}"' \
"batch_validation"

# Multiple valid emails - GET
test_endpoint "Multiple valid emails (GET)" \
"curl -X GET \"${API_URL}/validate/batch?email=user1@example.com&email=user2@example.com\"" \
"batch_validation"

# Mixed valid and invalid emails - POST
test_endpoint "Mixed valid and invalid emails (POST)" \
'curl -X POST "${API_URL}/validate/batch" -H "Content-Type: application/json" -d "{\"emails\":[\"valid@example.com\",\"invalid-email\"]}"' \
"batch_validation"

# Empty batch - POST
test_endpoint "Empty batch (POST)" \
'curl -X POST "${API_URL}/validate/batch" -H "Content-Type: application/json" -d "{\"emails\":[]}"' \
"batch_validation"

# 3. Typo Suggestion Tests
print_header "Typo Suggestion Tests"

# Gmail typo - POST
test_endpoint "Gmail typo (POST)" \
'curl -X POST "${API_URL}/typo-suggestions" -H "Content-Type: application/json" -d "{\"email\":\"user@gmial.com\"}"' \
"typo_suggestions"

# Gmail typo - GET
test_endpoint "Gmail typo (GET)" \
"curl -X GET \"${API_URL}/typo-suggestions?email=user@gmial.com\"" \
"typo_suggestions"

# Yahoo typo - POST
test_endpoint "Yahoo typo (POST)" \
'curl -X POST "${API_URL}/typo-suggestions" -H "Content-Type: application/json" -d "{\"email\":\"user@yhaoo.com\"}"' \
"typo_suggestions"

# Hotmail typo - GET
test_endpoint "Hotmail typo (GET)" \
"curl -X GET \"${API_URL}/typo-suggestions?email=user@hotmial.com\"" \
"typo_suggestions"

# 4. Special Cases
print_header "Special Cases"

# Disposable email - POST
test_endpoint "Disposable email (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"user@tempmail.com\"}"' \
"special_cases" \
"DISPOSABLE"

# Role-based email - POST
test_endpoint "Role-based email (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"admin@example.com\"}"' \
"special_cases" \
"VALID"

# Non-existent domain - POST
test_endpoint "Non-existent domain (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"user@nonexistentdomain123456.com\"}"' \
"special_cases" \
"INVALID_DOMAIN"

# 5. Error Cases
print_header "Error Cases"

# Invalid JSON - POST
test_endpoint "Invalid JSON (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d "invalid json"' \
"error_cases"

# Wrong Content-Type - POST
test_endpoint "Wrong Content-Type (POST)" \
'curl -X POST "${API_URL}/validate" -H "Content-Type: text/plain" -d "{\"email\":\"user@example.com\"}"' \
"error_cases"

# Method not allowed - PUT
test_endpoint "Method not allowed (PUT)" \
'curl -X PUT "${API_URL}/validate" -H "Content-Type: application/json" -d "{\"email\":\"user@example.com\"}"' \
"error_cases"

# 6. Status Check
print_header "Status Check"

# Get service status
test_endpoint "Service Status" \
"curl ${API_URL}/status" \
"status"

# Print timing statistics
print_header "Timing Statistics"

printf "${YELLOW}%-25s %10s %15s %15s${NC}\n" "Endpoint Type" "Requests" "Total Time(s)" "Avg Time(s)"
printf "%-60s\n" "================================================================="

total_requests=0
total_time=0

for i in "${!endpoint_types[@]}"; do
    count=${counts[$i]}
    total_time_for_type=${times[$i]}
    
    if [ $count -gt 0 ]; then
        avg_time=$(echo "scale=3; $total_time_for_type / $count" | bc)
        printf "%-25s %10d %15.3f %15.3f\n" \
            "$(echo ${endpoint_types[$i]} | tr '_' ' ' | awk '{for(j=1;j<=NF;j++)sub(/./,toupper(substr($j,1,1)),$j)}1')" \
            $count $total_time_for_type $avg_time
        
        total_requests=$((total_requests + count))
        total_time=$(echo "$total_time + $total_time_for_type" | bc)
    fi
done

printf "%-60s\n" "================================================================="
if [ $total_requests -gt 0 ]; then
    avg_time=$(echo "scale=3; $total_time / $total_requests" | bc)
    printf "${GREEN}%-25s %10d %15.3f %15.3f${NC}\n" "Overall" $total_requests $total_time $avg_time
fi 