#!/bin/bash

# Check if SVIX_AUTH_TOKEN is set
if [ -z "$SVIX_AUTH_TOKEN" ]; then
    echo "Error: SVIX_AUTH_TOKEN environment variable is not set"
    echo "Usage: SVIX_AUTH_TOKEN=<your-token> ./delete-svix-apps.sh"
    exit 1
fi

# Show current server information
echo "Current Svix configuration:"
echo "Server URL: ${SVIX_SERVER_URL:-'default'}"
svix version

# List all applications and extract their IDs
echo "Fetching all applications..."
app_ids=$(svix application list | jq -r '.data[].id')

if [ -z "$app_ids" ]; then
    echo "No applications found"
    exit 0
fi

# Delete each application with automatic yes
for app_id in $app_ids; do
    echo "Deleting application: $app_id"
    yes | svix application delete "$app_id"
done

echo "All Svix applications deleted successfully"