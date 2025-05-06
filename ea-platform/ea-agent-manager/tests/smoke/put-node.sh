#!/bin/bash

# Verify JWT_TOKEN environment variable is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract the user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# Directory containing payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint for nodes
NODE_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/nodes"

# Fetch user's nodes
USER_NODES=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$NODE_ENDPOINT")

# Extract the first node ID from user's nodes list
FIRST_NODE_ID=$(echo "$USER_NODES" | jq -r '.[0].id')

# Check if node ID was found
if [[ -z "$FIRST_NODE_ID" || "$FIRST_NODE_ID" == "null" ]]; then
  echo "Error: No nodes found or unable to extract node ID."
  exit 1
fi

echo "Authenticated user ID: $USER_ID"
echo "First node ID: $FIRST_NODE_ID"

# Iterate through matching payload files
for file in "$PAYLOAD_DIR"/*update-node*.json; do
    if [[ -f "$file" ]]; then
        echo "Processing file: $file"

        MODIFIED_PAYLOAD=$(jq --arg creatorID "$USER_ID" --arg nodeID "$FIRST_NODE_ID" \
            '.creator = $creatorID | .id = $nodeID' "$file")

        echo "Posting payload with creatorID: $USER_ID and nodeID: $FIRST_NODE_ID"

        # Send the modified payload to the API
        curl -X PUT "$NODE_ENDPOINT/$FIRST_NODE_ID" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data "$MODIFIED_PAYLOAD"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""
