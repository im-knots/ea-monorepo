#!/bin/bash

# Verify JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
  echo "Error: JWT_TOKEN environment variable not set"
  exit 1
fi

# Extract user ID from JWT sub claim
USER_ID=$(echo "$JWT_TOKEN" | cut -d '.' -f2 | base64 -d 2>/dev/null | jq -r '.sub')

# Validate extracted user ID
if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
  echo "Error: Unable to extract user ID from JWT token"
  exit 1
fi

echo "Authenticated user ID: $USER_ID"

# API Endpoint for user details
API_ENDPOINT="http://api.erulabs.local/ainu-manager/api/v1/users/$USER_ID"

# Get user details
USER_DETAILS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$API_ENDPOINT")

# Extract first device ID from user's compute_devices
FIRST_DEVICE_ID=$(echo "$USER_DETAILS" | jq -r '.compute_devices[0].id')

# Check if device ID was found
if [ -z "$FIRST_DEVICE_ID" ] || [ "$FIRST_DEVICE_ID" == "null" ]; then
  echo "Error: No devices found or unable to extract device ID"
  exit 1
fi

echo "First device ID: $FIRST_DEVICE_ID"

# API Endpoint for deleting the device
DELETE_ENDPOINT="$API_ENDPOINT/devices/$FIRST_DEVICE_ID"

# Delete the device
echo "Deleting device at endpoint $DELETE_ENDPOINT"
curl -X DELETE "$DELETE_ENDPOINT" -H "Authorization: Bearer $JWT_TOKEN"
echo ""
