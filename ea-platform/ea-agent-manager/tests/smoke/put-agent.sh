#!/bin/bash

# Ensure JWT_TOKEN is set
if [[ -z "$JWT_TOKEN" ]]; then
  echo "Error: JWT_TOKEN environment variable not set."
  exit 1
fi

# Extract user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "Error: Unable to extract user ID from JWT token."
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoint for fetching the user's agents
AGENT_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/agents"

# Get user's agents
USER_AGENTS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$AGENT_ENDPOINT")

# Extract the first agent ID from user's agents list
FIRST_AGENT_ID=$(echo "$USER_AGENTS" | jq -r '.[0].id')

# Check if agent ID was found
if [[ -z "$FIRST_AGENT_ID" || "$FIRST_AGENT_ID" == "null" ]]; then
  echo "Error: No agents found for user $USER_ID or unable to extract agent ID."
  exit 1
fi

echo "First agent ID: $FIRST_AGENT_ID"

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# Iterate through matching payload files
for file in "$PAYLOAD_DIR"/*update-agent*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        # Inject creatorID (authenticated user ID) and agent ID into payload
        MODIFIED_PAYLOAD=$(jq --arg creatorID "$USER_ID" --arg agentID "$FIRST_AGENT_ID" \
            '.creator = $creatorID | .id = $agentID' "$file")

        # Send modified payload to API endpoint
        curl -X PUT "$AGENT_ENDPOINT/$FIRST_AGENT_ID" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data "$MODIFIED_PAYLOAD"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""
