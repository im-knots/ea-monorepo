#!/bin/bash

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

# Get all users
ALL_USERS=$(curl -s "$API_ENDPOINT")

# Extract the first `id` from the response
FIRST_USER_ID=$(echo "$ALL_USERS" | jq -r '.[0].id')

# Check if an ID was found
if [ -z "$FIRST_USER_ID" ] || [ "$FIRST_USER_ID" == "null" ]; then
  echo "Error: No users found or unable to extract id"
  exit 1
fi

echo "First user _id: $FIRST_USER_ID"

# Get user details
USER_DETAILS=$(curl -s "$API_ENDPOINT/$FIRST_USER_ID")

# Extract the first JOB ID from the user's jobs list
FIRST_JOB_ID=$(echo "$USER_DETAILS" | jq -r '.jobs[0].id')

# Check if a JOB ID was found
if [ -z "$FIRST_JOB_ID" ] || [ "$FIRST_JOB_ID" == "null" ]; then
  echo "Error: No Jobs found for user $FIRST_USER_ID or unable to extract Job ID"
  exit 1
fi

echo "First Job ID: $FIRST_JOB_ID"

# API Endpoint for deleting the job
DELETE_ENDPOINT="$API_ENDPOINT/$FIRST_USER_ID/jobs/$FIRST_JOB_ID"

# Delete the first job
echo "Deleting job at endpoing $DELETE_ENDPOINT"
curl -X DELETE "$DELETE_ENDPOINT" 
echo ""

