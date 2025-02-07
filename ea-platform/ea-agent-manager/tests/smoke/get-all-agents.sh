#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://agent-manager.ea.erulabs.local/api/v1/agents"

curl "$API_ENDPOINT"
echo ""


