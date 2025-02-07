#!/bin/bash

# API Endpoint
API_ENDPOINT="http://agent-manager.ea.erulabs.local/api/v1/nodes"

# Get all nodes
ALL_NODES=$(curl -s "$API_ENDPOINT")

# Extract the first `_id` from the response
FIRST_NODE_ID=$(echo "$ALL_NODES" | jq -r '.[0].id')

# Check if an ID was found
if [ -z "$FIRST_NODE_ID" ] || [ "$FIRST_NODE_ID" == "null" ]; then
  echo "Error: No nodes found or unable to extract id"
  exit 1
fi

echo "First node id: $FIRST_NODE_ID"

# Fetch the specific node by _id
echo "Fetching node with id: $FIRST_NODE_ID"
SPECIFIC_NODE=$(curl -s "$API_ENDPOINT/$FIRST_NODE_ID")

# Output the specific node's details
echo "Specific Node Details:"
echo "$SPECIFIC_NODE"
