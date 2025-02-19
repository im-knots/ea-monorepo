#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
AGENT_ENDPOINT="http://agent-manager.ea.erulabs.local/api/v1/agents"

# Get all agents
ALL_AGENTS=$(curl -s "$AGENT_ENDPOINT")

# Extract the first `id` from the response
FIRST_AGENT_ID=$(echo "$ALL_AGENTS" | jq -r '.[0].id')
CREATOR_ID=$(echo "$ALL_AGENTS" | jq -r '.[0].creator')

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*update-agent*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        # Inject creatorID and agentID into the JSON payload
        MODIFIED_PAYLOAD=$(jq --arg creatorID "$CREATOR_ID" --arg agentID "$FIRST_AGENT_ID" \
            '.creator = $creatorID | .id = $agentID' "$file")

        # Send the modified payload to the API
        curl -X PUT "$AGENT_ENDPOINT/$FIRST_AGENT_ID" \
            -H "Content-Type: application/json" \
            --data "$MODIFIED_PAYLOAD"
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""


