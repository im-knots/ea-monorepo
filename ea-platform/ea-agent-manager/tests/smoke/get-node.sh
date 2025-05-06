#!/bin/bash

# Ensure JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
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

# API Endpoint (user-specific nodes)
NODE_ENDPOINT="http://api.erulabs.local/agent-manager/api/v1/nodes"

# Fetch the user's nodes
USER_NODES=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$NODE_ENDPOINT")

# Extract the first node ID from user's nodes list
FIRST_NODE_ID=$(echo "$USER_NODES" | jq -r '.[0].id')

# Validate node ID
if [[ -z "$FIRST_NODE_ID" || "$FIRST_NODE_ID" == "null" ]]; then
  echo "Error: No nodes found or unable to extract node ID."
  exit 1
fi

echo "First node ID: $FIRST_NODE_ID"

# Fetch and display details for the specific node
echo "Fetching node with ID: $FIRST_NODE_ID"
SPECIFIC_NODE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$NODE_ENDPOINT/$FIRST_NODE_ID")

echo "Specific Node Details:"
echo "$SPECIFIC_NODE"
