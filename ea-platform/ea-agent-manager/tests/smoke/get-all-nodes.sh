#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://localhost:8083/api/v1/nodes"

curl "$API_ENDPOINT"
echo ""
