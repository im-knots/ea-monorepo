#!/bin/bash

# API Endpoints
AGENT_MANAGER_URL="http://api.ea.erulabs.local/agent-manager/api/v1/agents"
AINU_MANAGER_URL="http://api.ea.erulabs.local/ainu-manager/api/v1/users"
JOB_ENGINE_URL="http://api.ea.erulabs.local/job-api/api/v1/jobs"


# Fetch the list of agents
AGENT_RESPONSE=$(curl -s "$AGENT_MANAGER_URL")
AINU_RESPONSE=$(curl -s "$AINU_MANAGER_URL")

# Extract the first agent ID using jq
FIRST_AGENT_ID=$(echo "$AGENT_RESPONSE" | jq -r '.[0].id')
FIRST_USER_ID=$(echo "$AINU_RESPONSE" | jq -r '.[0].id')

# Check if an agent ID was found
if [[ -z "$FIRST_AGENT_ID" || "$FIRST_AGENT_ID" == "null" ]]; then
    echo "Error: No agents found in the response."
    exit 1
fi

# Check if an User ID was found
if [[ -z "$FIRST_USER_ID" || "$FIRST_USER_ID" == "null" ]]; then
    echo "Error: No users found in the response."
    exit 1
fi

echo "First agent ID: $FIRST_AGENT_ID"
echo "First user ID: $FIRST_USER_ID"

# Construct the JSON payload
PAYLOAD=$(jq -n --arg agentID "$FIRST_AGENT_ID" --arg userID "$FIRST_USER_ID" '{agent_id: $agentID, user_id: $userID}')

echo "Submitting job request to $JOB_ENGINE_URL..."
echo "Payload: $PAYLOAD"

# Send the job creation request
RESPONSE=$(curl -s -X POST "$JOB_ENGINE_URL" \
    -H "Content-Type: application/json" \
    --data "$PAYLOAD")

# Output response
echo "Job submission response: $RESPONSE"
