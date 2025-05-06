#!/bin/bash

# Verify JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract the user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoint to fetch user's agents
API_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/users/$USER_ID/agents"

# Get user's agents
USER_AGENTS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$API_ENDPOINT")

# Extract the first agent ID from user's agents list
FIRST_AGENT_ID=$(echo "$USER_AGENTS" | jq -r '.[0].id')

# Check if an agent ID was found
if [ -z "$FIRST_AGENT_ID" ] || [ "$FIRST_AGENT_ID" == "null" ]; then
  echo "Error: No agents found or unable to extract agent ID"
  exit 1
fi

echo "First agent ID: $FIRST_AGENT_ID"

# Fetch specific agent details
SPECIFIC_AGENT=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$API_ENDPOINT/$FIRST_AGENT_ID")

# Output agent details
echo "Specific Agent Details:"
echo "$SPECIFIC_AGENT" | jq .
echo ""
