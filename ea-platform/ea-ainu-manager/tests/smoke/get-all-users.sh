#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://api.ea.erulabs.local/ainu-manager/api/v1/users"

curl "$API_ENDPOINT"
echo ""

