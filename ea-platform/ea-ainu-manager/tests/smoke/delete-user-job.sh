#!/bin/bash

# Verify JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract user ID from JWT sub claim
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoint to get user details
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID"

# Get user details
USER_DETAILS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$API_ENDPOINT")

# Extract the first JOB ID from the user's jobs list
FIRST_JOB_ID=$(echo "$USER_DETAILS" | jq -r '.jobs[0].id')

# Check if a Job ID was found
if [ -z "$FIRST_JOB_ID" ] || [ "$FIRST_JOB_ID" == "null" ]; then
  echo "Error: No jobs found or unable to extract job ID"
  exit 1
fi

echo "First job ID: $FIRST_JOB_ID"

# API Endpoint for deleting the job
DELETE_ENDPOINT="$API_ENDPOINT/jobs/$FIRST_JOB_ID"

# Delete the first job
echo "Deleting job at endpoint $DELETE_ENDPOINT"
curl -X DELETE "$DELETE_ENDPOINT" -H "Authorization: Bearer $JWT_TOKEN"
echo ""
