#!/bin/bash

# Directory containing the payload files
PAYLOAD_DIR="smoke/payloads"

# API Endpoint
API_ENDPOINT="http://localhost:8084/api/v1/agents"

curl "$API_ENDPOINT"

