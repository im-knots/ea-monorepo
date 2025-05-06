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

# API Endpoint for retrieving authenticated user's data
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID"

# Fetch authenticated user's data
echo "Fetching user data..."
USER_DATA=$(curl -s "$API_ENDPOINT" -H "Authorization: Bearer $JWT_TOKEN")

# Output the retrieved user details
echo "User details:"
echo "$USER_DATA" | jq .
