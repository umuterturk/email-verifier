#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
USE_PROD=false
for arg in "$@"
do
    if [ "$arg" == "--prod" ]; then
        USE_PROD=true
    fi
done

# Array to track failing tests
declare -a failed_tests=()

# Set the Base URL based on environment
if [ "$USE_PROD" = true ]; then
    API_URL="https://rapid-email-verifier.fly.dev/api"
    echo -e "${BLUE}Using production API: ${API_URL}${NC}"
else
    API_URL="http://localhost:8080/api"
    echo -e "${BLUE}Using localhost API: ${API_URL}${NC}"
fi

# Connectivity check - Verify the API is accessible and can validate emails
echo -e "${BLUE}Checking API connectivity...${NC}"
CONNECTIVITY_TEST=$(curl -s -X POST "${API_URL}/validate" -H "Content-Type: application/json" -d '{"email":"test@gmail.com"}')

# Check if the request failed or returned an error
if [ $? -ne 0 ] || [ -z "$CONNECTIVITY_TEST" ] || [[ "$CONNECTIVITY_TEST" == *"error"* ]]; then
    echo -e "${RED}Error: Could not connect to the API or received an error response.${NC}"
    echo -e "${RED}Response: $CONNECTIVITY_TEST${NC}"
    echo -e "${RED}Please ensure the server is running and accessible.${NC}"
    exit 1
fi

echo -e "${GREEN}API connectivity check passed.${NC}"

# Set expected validation statuses based on real-world provider behavior
# These are constants and don't need to change based on environment
EXAMPLE_COM_STATUS="NO_MX_RECORDS"  # example.com typically doesn't have proper MX records
DISPOSABLE_STATUS="NO_MX_RECORDS"   # Disposable domains often have no MX or are blocked
ROLE_BASED_STATUS="NO_MX_RECORDS"   # Role-based emails at example.com will have no MX
GMAIL_DK_STATUS="NO_MX_RECORDS"     # gmail.dk has null MX records that should be detected
GMAIL_COM_STATUS="VALID"            # Gmail.com has valid MX records
YAHOO_COM_STATUS="VALID"            # Yahoo.com has valid MX records
OUTLOOK_COM_STATUS="VALID"          # Outlook.com has valid MX records
HOTMAIL_COM_STATUS="VALID"          # Hotmail.com has valid MX records

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi


# Initialize timing data arrays
declare -a endpoint_types=("single_validation" "batch_validation" "typo_suggestions" "special_cases" "error_cases" "status" "skip_secret" "alias_detection")
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
    local actual_status
    actual_status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$expected_status" ]; then
        if [ "$actual_status" = "$expected_status" ]; then
            echo -e "${GREEN}✓ $description - Status: $actual_status${NC}"
        else
            echo -e "${RED}✗ $description - Expected: $expected_status, Got: $actual_status${NC}"
            # Store the failure details
            failed_tests+=("${description} - Expected: ${expected_status}, Got: ${actual_status}")
        fi
    fi
}

# Function to check if response contains aliasOf field
check_alias() {
    local response=$1
    local expected_alias=$2
    local description=$3
    
    # Check if the response contains the aliasOf field
    if [ -n "$expected_alias" ]; then
        # Extract the aliasOf field from the JSON response if it exists
        local alias_field
        alias_field=$(echo "$response" | grep -o '"aliasOf":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$alias_field" ]; then
            if [ "$alias_field" = "$expected_alias" ]; then
                echo -e "${GREEN}✓ $description - Found aliasOf: $alias_field${NC}"
            else
                echo -e "${RED}✗ $description - Expected aliasOf: $expected_alias, Got: $alias_field${NC}"
                # Store the failure details
                failed_tests+=("${description} - Expected aliasOf: ${expected_alias}, Got: ${alias_field}")
            fi
        else
            if [ "$expected_alias" = "NONE" ]; then
                echo -e "${GREEN}✓ $description - No aliasOf field as expected${NC}"
            else
                echo -e "${RED}✗ $description - Expected aliasOf: $expected_alias, but no aliasOf field found${NC}"
                # Store the failure details
                failed_tests+=("${description} - Expected aliasOf: ${expected_alias}, but no aliasOf field found")
            fi
        fi
    fi
}

# Function to get endpoint type index
get_endpoint_index() {
    local type=$1
    for i in "${!endpoint_types[@]}"; do
        if [[ "${endpoint_types[$i]}" == "$type" ]]; then
            echo "$i"
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
    local expected_alias=$5
    
    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "${BLUE}Command: ${command}${NC}"
    echo -e "${BLUE}Response:${NC}"
    
    # Add timing parameters to curl and suppress progress
    local curl_cmd="${command} -s -w '\nTime: %{time_total}s\n'"
    
    # Capture the output and timing
    local output
    output=$(eval "$curl_cmd")
    local time
    time=$(echo "$output" | grep "Time:" | cut -d' ' -f2 | sed 's/s//')
    echo "$output"
    
    # Check response against expected status
    if [ -n "$expected_status" ]; then
        check_response "$output" "$expected_status" "$description"
    fi
    
    # Check response for aliasOf field
    if [ -n "$expected_alias" ]; then
        check_alias "$output" "$expected_alias" "$description"
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
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@example.com\"}' ${SKIP_SECRET_HEADER}" \
"single_validation" \
"${EXAMPLE_COM_STATUS}"

# Valid email - GET
test_endpoint "Valid email (GET)" \
"curl -X GET \"${API_URL}/validate?email=user@example.com\" ${SKIP_SECRET_HEADER}" \
"single_validation" \
"${EXAMPLE_COM_STATUS}"

# Invalid email format - POST
test_endpoint "Invalid email format (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"invalid-email\"}' ${SKIP_SECRET_HEADER}" \
"single_validation" \
"INVALID_FORMAT"

# Invalid email format - GET
test_endpoint "Invalid email format (GET)" \
"curl -X GET \"${API_URL}/validate?email=invalid-email\" ${SKIP_SECRET_HEADER}" \
"single_validation" \
"INVALID_FORMAT"

# Empty email - POST
test_endpoint "Empty email (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"\"}' ${SKIP_SECRET_HEADER}" \
"single_validation" \
"MISSING_EMAIL"

# Missing email parameter - GET
test_endpoint "Missing email parameter (GET)" \
"curl -X GET \"${API_URL}/validate\" ${SKIP_SECRET_HEADER}" \
"single_validation"

# 2. Batch Email Validation Tests
print_header "Batch Email Validation Tests"

# Multiple valid emails - POST
test_endpoint "Multiple valid emails (POST)" \
"curl -X POST \"${API_URL}/validate/batch\" -H \"Content-Type: application/json\" -d '{\"emails\":[\"user1@example.com\",\"user2@example.com\"]}' ${SKIP_SECRET_HEADER}" \
"batch_validation"

# Multiple valid emails - GET
test_endpoint "Multiple valid emails (GET)" \
"curl -X GET \"${API_URL}/validate/batch?email=user1@example.com&email=user2@example.com\" ${SKIP_SECRET_HEADER}" \
"batch_validation"

# Mixed valid and invalid emails - POST
test_endpoint "Mixed valid and invalid emails (POST)" \
"curl -X POST \"${API_URL}/validate/batch\" -H \"Content-Type: application/json\" -d '{\"emails\":[\"valid@example.com\",\"invalid-email\"]}' ${SKIP_SECRET_HEADER}" \
"batch_validation"

# Empty batch - POST
test_endpoint "Empty batch (POST)" \
"curl -X POST \"${API_URL}/validate/batch\" -H \"Content-Type: application/json\" -d '{\"emails\":[]}' ${SKIP_SECRET_HEADER}" \
"batch_validation"

# 3. Typo Suggestion Tests
print_header "Typo Suggestion Tests"

# Gmail typo - POST
test_endpoint "Gmail typo (POST)" \
"curl -X POST \"${API_URL}/typo-suggestions\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@gmial.com\"}' ${SKIP_SECRET_HEADER}" \
"typo_suggestions"

# Gmail typo - GET
test_endpoint "Gmail typo (GET)" \
"curl -X GET \"${API_URL}/typo-suggestions?email=user@gmial.com\" ${SKIP_SECRET_HEADER}" \
"typo_suggestions"

# Yahoo typo - POST
test_endpoint "Yahoo typo (POST)" \
"curl -X POST \"${API_URL}/typo-suggestions\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@yhaoo.com\"}' ${SKIP_SECRET_HEADER}" \
"typo_suggestions"

# Hotmail typo - GET
test_endpoint "Hotmail typo (GET)" \
"curl -X GET \"${API_URL}/typo-suggestions?email=user@hotmial.com\" ${SKIP_SECRET_HEADER}" \
"typo_suggestions"

# 4. Special Cases
print_header "Special Cases"

# Disposable email - POST
test_endpoint "Disposable email (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@mailnator.com\"}' ${SKIP_SECRET_HEADER}" \
"special_cases" \
"${DISPOSABLE_STATUS}"

# Role-based email - POST
test_endpoint "Role-based email (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"admin@example.com\"}' ${SKIP_SECRET_HEADER}" \
"special_cases" \
"${ROLE_BASED_STATUS}"

# Non-existent domain - POST
test_endpoint "Non-existent domain (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@nonexistentdomain123456.com\"}' ${SKIP_SECRET_HEADER}" \
"special_cases" \
"INVALID_DOMAIN"

# Domain with null MX record (gmail.dk) - POST
# Note: Score for NO_MX_RECORDS status is 40 instead of 60
test_endpoint "Domain with null MX record (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@gmail.dk\"}' ${SKIP_SECRET_HEADER}" \
"special_cases" \
"${GMAIL_DK_STATUS}"

# Domain with null MX record (gmail.dk) - GET
# Note: Score for NO_MX_RECORDS status is 40 instead of 60
test_endpoint "Domain with null MX record (GET)" \
"curl -X GET \"${API_URL}/validate?email=user@gmail.dk\" ${SKIP_SECRET_HEADER}" \
"special_cases" \
"${GMAIL_DK_STATUS}"

# Batch with mixed valid and null MX domains - POST
test_endpoint "Batch with mixed valid and null MX domains (POST)" \
"curl -X POST \"${API_URL}/validate/batch\" -H \"Content-Type: application/json\" -d '{\"emails\":[\"user@gmail.com\",\"user@gmail.dk\"]}' ${SKIP_SECRET_HEADER}" \
"special_cases"

# 5. Email Alias Detection Tests
print_header "Email Alias Detection Tests"

# Gmail dot alias - POST
test_endpoint "Gmail dot alias (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user.name@gmail.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${GMAIL_COM_STATUS}" \
"username@gmail.com"

# Gmail plus alias - POST
test_endpoint "Gmail plus alias (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"username+test@gmail.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${GMAIL_COM_STATUS}" \
"username@gmail.com"

# Gmail combined dots and plus - POST
test_endpoint "Gmail combined dots and plus (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user.name+test@gmail.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${GMAIL_COM_STATUS}" \
"username@gmail.com"

# Gmail combined dots and plus - GET (with proper URL encoding of +)
test_endpoint "Gmail combined dots and plus (GET)" \
"curl -X GET \"${API_URL}/validate?email=user.name%2Btest@gmail.com\" ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${GMAIL_COM_STATUS}" \
"username@gmail.com"

# Yahoo alias - POST
test_endpoint "Yahoo alias (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"username-test@yahoo.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${YAHOO_COM_STATUS}" \
"username@yahoo.com"

# Yahoo alias - GET
test_endpoint "Yahoo alias (GET)" \
"curl -X GET \"${API_URL}/validate?email=username-test@yahoo.com\" ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${YAHOO_COM_STATUS}" \
"username@yahoo.com"

# Outlook alias - POST
test_endpoint "Outlook alias (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"username+test@outlook.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${OUTLOOK_COM_STATUS}" \
"username@outlook.com"

# Hotmail alias - POST
test_endpoint "Hotmail alias (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"username+test@hotmail.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${HOTMAIL_COM_STATUS}" \
"username@hotmail.com"

# Non-alias email - POST
test_endpoint "Non-alias email (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"username@gmail.com\"}' ${SKIP_SECRET_HEADER}" \
"alias_detection" \
"${GMAIL_COM_STATUS}" \
"NONE"

# Batch with mixed aliases - POST
test_endpoint "Batch with mixed aliases (POST)" \
"curl -X POST \"${API_URL}/validate/batch\" -H \"Content-Type: application/json\" -d '{\"emails\":[\"user.name+test@gmail.com\",\"username-test@yahoo.com\",\"username+test@outlook.com\",\"normal@example.com\"]}' ${SKIP_SECRET_HEADER}" \
"alias_detection"

# 6. Error Cases
print_header "Error Cases"

# Invalid JSON - POST
test_endpoint "Invalid JSON (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d \"invalid json\" ${SKIP_SECRET_HEADER}" \
"error_cases"

# Wrong Content-Type - POST
test_endpoint "Wrong Content-Type (POST)" \
"curl -X POST \"${API_URL}/validate\" -H \"Content-Type: text/plain\" -d '{\"email\":\"user@example.com\"}' ${SKIP_SECRET_HEADER}" \
"error_cases"

# Method not allowed - PUT
test_endpoint "Method not allowed (PUT)" \
"curl -X PUT \"${API_URL}/validate\" -H \"Content-Type: application/json\" -d '{\"email\":\"user@example.com\"}' ${SKIP_SECRET_HEADER}" \
"error_cases"

# 7. Status Check
print_header "Status Check"

# Get service status
test_endpoint "Service Status" \
"curl \"${API_URL}/status\" ${SKIP_SECRET_HEADER}" \
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
    
    if [ "$count" -gt 0 ]; then
        avg_time=$(echo "scale=3; $total_time_for_type / $count" | bc)
        printf "%-25s %10d %15.3f %15.3f\n" \
            "$(echo "${endpoint_types[$i]}" | tr '_' ' ' | awk '{for(j=1;j<=NF;j++)sub(/./,toupper(substr($j,1,1)),$j)}1')" \
            "$count" "$total_time_for_type" "$avg_time"
        
        total_requests=$((total_requests + count))
        total_time=$(echo "$total_time + $total_time_for_type" | bc)
    fi
done

printf "%-60s\n" "================================================================="
if [ $total_requests -gt 0 ]; then
    avg_time=$(echo "scale=3; $total_time / $total_requests" | bc)
    printf "${GREEN}%-25s %10d %15.3f %15.3f${NC}\n" "Overall" "$total_requests" "$total_time" "$avg_time"
fi

# Print failing tests summary
print_header "Failing Tests Summary"

if [ ${#failed_tests[@]} -eq 0 ]; then
    echo -e "${GREEN}All tests passed successfully!${NC}"
else
    echo -e "${RED}Found ${#failed_tests[@]} failing tests:${NC}"
    for ((i=0; i<${#failed_tests[@]}; i++)); do
        echo -e "${RED}$(($i+1)). ${failed_tests[$i]}${NC}"
    done
    # Exit with non-zero status if there are failures
    exit 1
fi 