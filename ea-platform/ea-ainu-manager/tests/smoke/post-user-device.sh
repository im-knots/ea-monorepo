#!/bin/bash

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

# Get all users
ALL_USERS=$(curl -s "$API_ENDPOINT")

# Extract the first `id` from the response
FIRST_USER_ID=$(echo "$ALL_USERS" | jq -r '.[0].id')

# Check if an ID was found
if [ -z "$FIRST_USER_ID" ] || [ "$FIRST_USER_ID" == "null" ]; then
  echo "Error: No users found or unable to extract id"
  exit 1
fi

echo "First user id: $FIRST_USER_ID"

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/ainu-manager/api/v1/users/$FIRST_USER_ID/devices"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*add-device*.json; do
    if [[ -f "$file" ]]; then
        curl -X POST "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            --data-binary @"$file"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
