#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoints
AGENT_ENDPOINT="http://api.ea.erulabs.local/agent-manager/api/v1/nodes"
AINU_URL="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

# Fetch users from AINU manager
AINU_RESPONSE=$(curl -s "$AINU_URL")

# Extract the first user ID
FIRST_USER_ID=$(echo "$AINU_RESPONSE" | jq -r '.[0].id')

if [[ -z "$FIRST_USER_ID" || "$FIRST_USER_ID" == "null" ]]; then
    echo "Error: Unable to fetch a valid creator ID."
    exit 1
fi


# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*create-node*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        # Inject creatorID into the JSON payload
        MODIFIED_PAYLOAD=$(jq --arg creatorID "$FIRST_USER_ID" '.creator = $creatorID' "$file")

        echo "Posting payload with creatorID: $FIRST_USER_ID"

        # Send the modified payload to the API
        curl -X POST "$AGENT_ENDPOINT" \
            -H "Content-Type: application/json" \
            --data "$MODIFIED_PAYLOAD"
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""