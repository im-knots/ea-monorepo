#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/agent-manager/api/v1/nodes"

# TEST GET ALL NODES
curl "$API_ENDPOINT"
echo ""

# TEST GET ALL NODES BY CREATORID
AINU_URL="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

# Fetch users from AINU manager
AINU_RESPONSE=$(curl -s "$AINU_URL")

# Extract the first user ID
FIRST_USER_ID=$(echo "$AINU_RESPONSE" | jq -r '.[0].id')

if [[ -z "$FIRST_USER_ID" || "$FIRST_USER_ID" == "null" ]]; then
    echo "Error: Unable to fetch a valid creator ID."
    exit 1
fi
curl "$API_ENDPOINT?creator_id=$FIRST_USER_ID"
echo ""