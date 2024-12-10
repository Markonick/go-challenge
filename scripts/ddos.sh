#!/bin/bash

# URL of the webhook endpoint
WEBHOOK_URL="http://localhost:8080/notifications"

# JSON payload to send
PAYLOAD='{
    "id": "evt_ddos",
    "type": "test.event",
    "project": "test",
    "data": {"id": "123"}
}'

# Number of requests to send
REQUEST_COUNT=50

echo "Starting DDoS test with $REQUEST_COUNT requests..."

for i in $(seq 1 $REQUEST_COUNT); do
    curl -X POST -H "Content-Type: application/json" -d "$PAYLOAD" "$WEBHOOK_URL" &
done

echo "DDoS test completed. Check logs for rate limiting behavior."