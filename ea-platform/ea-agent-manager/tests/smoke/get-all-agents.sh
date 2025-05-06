#!/bin/bash

# Check if JWT_TOKEN is set
if [[ -z "$JWT_TOKEN" ]]; then
  echo "Error: JWT_TOKEN environment variable not set."
  exit 1
fi

# Extract the authenticated user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate the user ID
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "Error: Unable to extract user ID from JWT token."
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoint (user-specific agents)
API_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/agents"

# Fetch agents by authenticated user (creator)
echo "Fetching agents for user ID: $USER_ID"
curl "$API_ENDPOINT" -H "Authorization: Bearer $JWT_TOKEN"
echo ""
