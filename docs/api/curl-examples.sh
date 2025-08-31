#!/bin/bash

# Tenant Management API - cURL Examples
# =====================================
# This script demonstrates how to use the Tenant Management API endpoints using cURL
# 
# Prerequisites:
# - Set the BASE_URL environment variable
# - Set the JWT_TOKEN environment variable
#
# Usage:
#   export BASE_URL="https://api.example.com/v1"
#   export JWT_TOKEN="your-jwt-token-here"
#   bash curl-examples.sh

set -e  # Exit on error

# Configuration
BASE_URL=${BASE_URL:-"https://api.example.com/v1"}
JWT_TOKEN=${JWT_TOKEN:-"your-jwt-token-here"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print section headers
print_section() {
    echo -e "\n${BLUE}===========================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===========================================${NC}\n"
}

# Helper function to print step headers  
print_step() {
    echo -e "\n${GREEN}Step $1: $2${NC}"
    echo -e "${YELLOW}Command:${NC}"
}

# Helper function to execute curl with pretty output
execute_curl() {
    local description="$1"
    shift
    echo -e "\n${GREEN}$description${NC}"
    echo -e "${YELLOW}Executing:${NC} curl $*"
    echo -e "${BLUE}Response:${NC}"
    curl -s -w "\nHTTP Status: %{http_code}\n" "$@" | jq .
    echo ""
}

# Validate prerequisites
if [ "$JWT_TOKEN" == "your-jwt-token-here" ]; then
    echo -e "${RED}Error: Please set your JWT_TOKEN environment variable${NC}"
    echo "export JWT_TOKEN=\"your-actual-jwt-token\""
    exit 1
fi

echo -e "${GREEN}Tenant Management API - cURL Examples${NC}"
echo -e "Base URL: ${BASE_URL}"
echo -e "JWT Token: ${JWT_TOKEN:0:20}...\n"

# ==========================================
# 1. LIST TENANTS
# ==========================================
print_section "1. LIST TENANTS"

print_step "1.1" "List all tenants (default pagination)"
execute_curl "Get first page of tenants" \
    -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants"

print_step "1.2" "List tenants with pagination and filtering"
execute_curl "Get tenants with custom pagination and status filter" \
    -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants?page=1&limit=5&status=active"

print_step "1.3" "List tenants with name search"
execute_curl "Search tenants by name" \
    -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants?name=Test&limit=10"

# ==========================================
# 2. CREATE TENANT
# ==========================================
print_section "2. CREATE TENANT"

# Generate unique identifiers for testing
TIMESTAMP=$(date +%s)
TENANT_CODE="TEST$TIMESTAMP"
TENANT_NAME="Test Tenant $TIMESTAMP"

print_step "2.1" "Create a new tenant"
CREATE_RESPONSE=$(curl -s \
    -X POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"$TENANT_NAME\",
        \"code\": \"$TENANT_CODE\",
        \"config\": {
            \"maxUsers\": 50,
            \"features\": [\"basic\"]
        }
    }" \
    "$BASE_URL/tenants")

echo -e "${GREEN}Creating tenant: $TENANT_NAME${NC}"
echo -e "${BLUE}Response:${NC}"
echo "$CREATE_RESPONSE" | jq .

# Extract tenant ID for subsequent operations
TENANT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.tenant.id // empty')

if [ -z "$TENANT_ID" ] || [ "$TENANT_ID" == "null" ]; then
    echo -e "${RED}Failed to create tenant or extract tenant ID${NC}"
    echo -e "${YELLOW}Skipping tenant-specific operations${NC}"
    TENANT_ID="example-tenant-id"
else
    echo -e "${GREEN}Tenant created with ID: $TENANT_ID${NC}"
fi

print_step "2.2" "Create tenant with admin user"
execute_curl "Create tenant with admin user" \
    -X POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"Full Test Tenant $TIMESTAMP\",
        \"code\": \"FULL$TIMESTAMP\",
        \"config\": {
            \"maxUsers\": 100,
            \"features\": [\"basic\", \"premium\"],
            \"theme\": \"dark\",
            \"domain\": \"tenant$TIMESTAMP.example.com\"
        },
        \"adminUser\": {
            \"email\": \"admin$TIMESTAMP@test.com\",
            \"name\": \"Test Admin\",
            \"password\": \"SecurePassword123!\"
        }
    }" \
    "$BASE_URL/tenants"

# ==========================================
# 3. GET TENANT DETAILS
# ==========================================
print_section "3. GET TENANT DETAILS"

print_step "3.1" "Get specific tenant details"
execute_curl "Retrieve tenant details" \
    -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/$TENANT_ID"

# ==========================================
# 4. UPDATE TENANT
# ==========================================
print_section "4. UPDATE TENANT"

print_step "4.1" "Update tenant name"
execute_curl "Update tenant name only" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"Updated $TENANT_NAME\"
    }" \
    "$BASE_URL/tenants/$TENANT_ID"

print_step "4.2" "Update tenant configuration"
execute_curl "Update tenant configuration" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"config\": {
            \"maxUsers\": 200,
            \"features\": [\"basic\", \"premium\", \"analytics\"],
            \"theme\": \"light\"
        }
    }" \
    "$BASE_URL/tenants/$TENANT_ID"

print_step "4.3" "Batch update multiple fields"
execute_curl "Batch update name, status, and config" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"Fully Updated $TENANT_NAME\",
        \"status\": \"active\",
        \"config\": {
            \"maxUsers\": 300,
            \"features\": [\"basic\", \"premium\", \"analytics\", \"enterprise\"]
        }
    }" \
    "$BASE_URL/tenants/$TENANT_ID"

# ==========================================
# 5. TENANT STATUS MANAGEMENT
# ==========================================
print_section "5. TENANT STATUS MANAGEMENT"

print_step "5.1" "Activate tenant"
execute_curl "Activate tenant" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/$TENANT_ID/activate"

print_step "5.2" "Suspend tenant"
execute_curl "Suspend tenant" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/$TENANT_ID/suspend"

print_step "5.3" "Reactivate tenant"
execute_curl "Reactivate tenant" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/$TENANT_ID/activate"

print_step "5.4" "Disable tenant"
execute_curl "Disable tenant" \
    -X PUT \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/$TENANT_ID/disable"

# ==========================================
# 6. ERROR HANDLING EXAMPLES
# ==========================================
print_section "6. ERROR HANDLING EXAMPLES"

print_step "6.1" "Get non-existent tenant (404 error)"
execute_curl "Attempt to get non-existent tenant" \
    -X GET \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Accept: application/json" \
    "$BASE_URL/tenants/non-existent-uuid"

print_step "6.2" "Create tenant with invalid data (validation error)"
execute_curl "Attempt to create tenant with missing required fields" \
    -X POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"\",
        \"code\": \"\"
    }" \
    "$BASE_URL/tenants"

print_step "6.3" "Create tenant with duplicate code (conflict error)"
execute_curl "Attempt to create tenant with duplicate code" \
    -X POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{
        \"name\": \"Duplicate Test\",
        \"code\": \"$TENANT_CODE\"
    }" \
    "$BASE_URL/tenants"

# ==========================================
# 7. DELETE TENANT (Optional - Dangerous!)
# ==========================================
print_section "7. DELETE TENANT (OPTIONAL)"

echo -e "${YELLOW}Warning: This will permanently delete the test tenant and all associated data!${NC}"
read -p "Do you want to proceed with tenant deletion? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_step "7.1" "Delete tenant with confirmation"
    execute_curl "Delete tenant with proper confirmation" \
        -X DELETE \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d "{
            \"confirmation\": \"DELETE_TENANT_$TENANT_ID\",
            \"reason\": \"testing\",
            \"createBackup\": true
        }" \
        "$BASE_URL/tenants/$TENANT_ID"
else
    echo -e "${YELLOW}Skipping tenant deletion.${NC}"
    
    print_step "7.1" "Example: Delete tenant with wrong confirmation (security error)"
    execute_curl "Attempt deletion with incorrect confirmation" \
        -X DELETE \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d "{
            \"confirmation\": \"WRONG_CONFIRMATION\",
            \"reason\": \"testing\"
        }" \
        "$BASE_URL/tenants/$TENANT_ID"
fi

# ==========================================
# 8. RATE LIMITING DEMONSTRATION
# ==========================================
print_section "8. RATE LIMITING DEMONSTRATION"

print_step "8.1" "Make multiple rapid requests to trigger rate limiting"
echo -e "${YELLOW}Making 10 rapid requests to demonstrate rate limiting...${NC}"

for i in {1..10}; do
    echo -e "\nRequest $i:"
    curl -s -w "Status: %{http_code} | Rate-Limit-Remaining: %header{X-RateLimit-Remaining}\n" \
        -X GET \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Accept: application/json" \
        "$BASE_URL/tenants?limit=1" > /dev/null
    
    # Small delay to avoid overwhelming the server
    sleep 0.1
done

print_section "EXAMPLES COMPLETED"
echo -e "${GREEN}All API examples have been executed successfully!${NC}"
echo -e "${YELLOW}Note: Some operations may have failed if tenants don't exist or due to permissions.${NC}"
echo -e "${BLUE}Check the responses above for detailed information about each API call.${NC}"