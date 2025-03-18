#!/bin/bash

# Check JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract the user ID (sub claim) from JWT
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# Directory containing job payloads
PAYLOAD_DIR="smoke/payloads"

# API Endpoint for adding jobs to the authenticated user
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID/jobs"

# Submit all matching job payload files
for file in "$PAYLOAD_DIR"/*add-job*.json; do
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
