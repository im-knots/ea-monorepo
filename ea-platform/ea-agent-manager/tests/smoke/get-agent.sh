#!/bin/bash

# API Endpoint
API_ENDPOINT="http://localhost:8083/api/v1/agents"

# Get all agents
echo "Fetching all agents..."
ALL_AGENTS=$(curl -s "$API_ENDPOINT")

# Extract the first `_id` from the response
FIRST_AGENT_ID=$(echo "$ALL_AGENTS" | jq -r '.[0]._id')

# Check if an ID was found
if [ -z "$FIRST_AGENT_ID" ] || [ "$FIRST_AGENT_ID" == "null" ]; then
  echo "Error: No agents found or unable to extract _id"
  exit 1
fi

echo "First agent _id: $FIRST_AGENT_ID"

# Fetch the specific agent by _id
echo "Fetching agent with _id: $FIRST_AGENT_ID"
SPECIFIC_AGENT=$(curl -s "$API_ENDPOINT/$FIRST_AGENT_ID")

# Output the specific agent's details
echo "Specific Agent Details:"
echo "$SPECIFIC_AGENT"
