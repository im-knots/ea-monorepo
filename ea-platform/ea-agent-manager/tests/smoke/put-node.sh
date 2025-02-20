#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
NODE_ENDPOINT="http://agent-manager.ea.erulabs.local/api/v1/nodes"

# Get all nodes
ALL_NODES=$(curl -s "$NODE_ENDPOINT")

# Extract the first `id` from the response
FIRST_NODE_ID=$(echo "$ALL_NODES" | jq -r '.[0].id')
CREATOR_ID=$(echo "$ALL_NODES" | jq -r '.[0].creator')

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*update-node*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        MODIFIED_PAYLOAD=$(jq --arg creatorID "$CREATOR_ID" --arg nodeID "$FIRST_NODE_ID" \
            '.creator = $creatorID | .id = $nodeID' "$file")

        # Send the modified payload to the API
        curl -X PUT "$NODE_ENDPOINT/$FIRST_NODE_ID" \
            -H "Content-Type: application/json" \
            --data "$MODIFIED_PAYLOAD"
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""


