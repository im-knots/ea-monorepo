#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*create-user*.json; do
    if [[ -f "$file" ]]; then
        curl -X POST "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            --data-binary @"$file"
        echo ""
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done