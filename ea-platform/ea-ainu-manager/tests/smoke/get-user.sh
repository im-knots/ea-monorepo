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

# Fetch the specific user by _id
echo "Fetching user with _id: $FIRST_USER_ID"
SPECIFIC_USER=$(curl -s "$API_ENDPOINT/$FIRST_USER_ID")

# Output the specific user's details
echo "Specific User Details:"
echo "$SPECIFIC_USER"
