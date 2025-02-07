#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://ainu-manager.ea.erulabs.local/api/v1/users"

curl "$API_ENDPOINT"
echo ""

