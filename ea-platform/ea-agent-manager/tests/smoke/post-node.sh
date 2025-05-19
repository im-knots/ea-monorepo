#!/bin/bash

# Check if JWT_TOKEN environment variable is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# Directory containing payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint (user-specific nodes)
AGENT_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/nodes"

# Iterate through matching payload files
for file in "$PAYLOAD_DIR"/*create-node*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        # Inject creatorID (authenticated user ID) into payload
        MODIFIED_PAYLOAD=$(jq --arg creatorID "$USER_ID" '.creator = $creatorID' "$file")

        echo "Posting payload with creatorID: $USER_ID"

        # Post payload to API endpoint
        curl -X POST "$AGENT_ENDPOINT" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data "$MODIFIED_PAYLOAD"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""
