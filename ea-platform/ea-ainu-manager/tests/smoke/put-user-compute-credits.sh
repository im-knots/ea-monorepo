#!/bin/bash

# Ensure JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract the user ID from JWT (sub claim)
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# Directory containing the compute credits payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint for updating compute credits for authenticated user
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID/credits"

# Iterate and submit each compute credits payload file
for file in "$PAYLOAD_DIR"/*update-compute-credits*.json; do
    if [[ -f "$file" ]]; then
        echo "Submitting payload: $file"
        curl -X PUT "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            --data-binary @"$file"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
