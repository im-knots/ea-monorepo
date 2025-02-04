#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://localhost:8083/api/v1/agents"

# Iterate through matching files in the payload directory
for file in "$PAYLOAD_DIR"/*create-agent*.json; do
    if [[ -f "$file" ]]; then
        curl -X POST "$API_ENDPOINT" \
            -H "Content-Type: application/json" \
            --data-binary @"$file"
    else
        echo "No matching files found in $PAYLOAD_DIR."
    fi
done
echo ""