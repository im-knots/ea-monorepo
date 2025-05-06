#!/bin/bash

# Verify JWT_TOKEN is set
if [[ -z "$JWT_TOKEN" ]]; then
  echo "Error: JWT_TOKEN environment variable not set."
  exit 1
fi

# Extract authenticated user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate user ID
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "Error: Unable to extract user ID from JWT token."
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoints (user-specific)
AGENT_MANAGER_URL="http://api.erulabs.local/agent-manager/api/v1/agents"
JOB_ENGINE_URL="http://api.erulabs.local/job-api/api/v1/jobs"

# Fetch agents for authenticated user
AGENT_RESPONSE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$AGENT_MANAGER_URL")

# Extract first agent ID
FIRST_AGENT_ID=$(echo "$AGENT_RESPONSE" | jq -r '.[0].id')

# Check if an agent ID was found
if [[ -z "$FIRST_AGENT_ID" || "$FIRST_AGENT_ID" == "null" ]]; then
    echo "Error: No agents found for user $USER_ID."
    exit 1
fi

echo "First agent ID: $FIRST_AGENT_ID"

# Construct JSON payload for job creation
PAYLOAD=$(jq -n --arg agentID "$FIRST_AGENT_ID" --arg userID "$USER_ID" '{agent_id: $agentID, user_id: $userID}')

echo "Submitting job request to $JOB_ENGINE_URL..."
echo "Payload: $PAYLOAD"

# Submit job creation request with Authorization header
RESPONSE=$(curl -s -X POST "$JOB_ENGINE_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    --data "$PAYLOAD")

# Output response
echo "Job submission response: $RESPONSE"
