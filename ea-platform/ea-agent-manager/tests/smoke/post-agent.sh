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

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint for creating agents
AGENT_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/agents"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*create-agent*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        # Inject creatorID (authenticated user) into the JSON payload
        MODIFIED_PAYLOAD=$(jq --arg creatorID "$USER_ID" '.creator = $creatorID' "$file")

        echo "Posting payload with creatorID: $USER_ID"

        # Send the modified payload to the API
        curl -v -X POST "$AGENT_ENDPOINT" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data "$MODIFIED_PAYLOAD"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""
