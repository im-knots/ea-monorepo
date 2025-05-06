#!/bin/bash

# Check if JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract user id ("sub") directly from JWT token
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Verify extraction
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT"
  exit 1
fi

echo "Authenticated user id (from JWT): $USER_ID"

# Directory containing payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID/devices"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*add-device*.json; do
    if [[ -f "$file" ]]; then
        echo "Submitting payload: $file"
        curl -X POST "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data-binary @"$file"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
