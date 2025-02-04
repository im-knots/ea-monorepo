#!/bin/bash

# API Endpoint
API_ENDPOINT="http://localhost:8085/api/v1/users"

# Get all users
echo "Fetching all users..."
ALL_USERS=$(curl -s "$API_ENDPOINT")

# Extract the first `id` from the response
FIRST_USER_ID=$(echo "$ALL_USERS" | jq -r '.[0].id')

# Check if an ID was found
if [ -z "$FIRST_USER_ID" ] || [ "$FIRST_USER_ID" == "null" ]; then
  echo "Error: No users found or unable to extract id"
  exit 1
fi

echo "First user _id: $FIRST_USER_ID"

# Get user details
USER_DETAILS=$(curl -s "$API_ENDPOINT/$FIRST_USER_ID")

# Extract the first device ID from the user's compute_devices list
FIRST_DEVICE_ID=$(echo "$USER_DETAILS" | jq -r '.compute_devices[0].id')

# Check if a device ID was found
if [ -z "$FIRST_DEVICE_ID" ] || [ "$FIRST_DEVICE_ID" == "null" ]; then
  echo "Error: No devices found for user $FIRST_USER_ID or unable to extract device ID"
  exit 1
fi

echo "First device ID: $FIRST_DEVICE_ID"

# API Endpoint for deleting the device
DELETE_ENDPOINT="$API_ENDPOINT/$FIRST_USER_ID/devices/$FIRST_DEVICE_ID"

# Delete the first device
echo "Deleting device at endpoing $DELETE_ENDPOINT"
curl -X DELETE "$DELETE_ENDPOINT" 

