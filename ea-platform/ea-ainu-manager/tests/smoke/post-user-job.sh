#!/bin/bash

# API Endpoint
API_ENDPOINT="http://localhost:8085/api/v1/users"

# Get all users
echo "Fetching all users..."
ALL_USERS=$(curl -s "$API_ENDPOINT")

# Extract the first `_id` from the response
FIRST_USER_ID=$(echo "$ALL_USERS" | jq -r '.[0]._id')

# Check if an ID was found
if [ -z "$FIRST_USER_ID" ] || [ "$FIRST_USER_ID" == "null" ]; then
  echo "Error: No users found or unable to extract _id"
  exit 1
fi

echo "First user _id: $FIRST_USER_ID"

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://localhost:8085/api/v1/users/$FIRST_USER_ID/jobs"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*add-job*.json; do
    if [[ -f "$file" ]]; then
        curl -X POST "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            --data-binary @"$file"
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
