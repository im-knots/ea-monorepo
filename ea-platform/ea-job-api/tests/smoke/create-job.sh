#!/bin/bash

# API Endpoints
AGENT_MANAGER_URL="http://localhost:8083/api/v1/agents"
JOB_ENGINE_URL="http://localhost:8084/api/v1/jobs"

echo "Fetching agent list from $AGENT_MANAGER_URL..."

# Fetch the list of agents
AGENT_RESPONSE=$(curl -s "$AGENT_MANAGER_URL")

# Check if the request was successful
if [[ $? -ne 0 ]]; then
    echo "Error: Failed to fetch agents from $AGENT_MANAGER_URL"
    exit 1
fi

# Extract the first agent ID using jq
FIRST_AGENT_ID=$(echo "$AGENT_RESPONSE" | jq -r '.[0]._id')

# Check if an agent ID was found
if [[ -z "$FIRST_AGENT_ID" || "$FIRST_AGENT_ID" == "null" ]]; then
    echo "Error: No agents found in the response."
    exit 1
fi

echo "First agent ID: $FIRST_AGENT_ID"

# Construct the JSON payload
PAYLOAD=$(jq -n --arg agentID "$FIRST_AGENT_ID" '{agentID: $agentID}')

echo "Submitting job request to $JOB_ENGINE_URL..."
echo "Payload: $PAYLOAD"

# Send the job creation request
RESPONSE=$(curl -s -X POST "$JOB_ENGINE_URL" \
    -H "Content-Type: application/json" \
    --data "$PAYLOAD")

# Output response
echo "Job submission response: $RESPONSE"
