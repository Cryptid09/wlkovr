#!/usr/bin/env bash
# ==============================================================================
# Script: test_viasocket_webhook.sh
# Purpose: Test and benchmark viasocket WhatsApp webhook ingestion endpoint
# Expected Response: HTTP 200 with JSON payload in < 2 seconds (< 50ms typical)
# ==============================================================================

set -e

HOST="${1:-http://localhost:8080}"
ENDPOINT="${HOST}/api/v1/webhooks/viasocket"

echo "=========================================================="
echo "🚀 Testing viasocket Webhook Ingestion on ${ENDPOINT}"
echo "=========================================================="

PAYLOAD='{
  "event_id": "test-wa-'$(date +%s)'",
  "provider": "WhatsApp",
  "sender": "+91-9826012345",
  "body": "Emergency: Main drinking water line mixed with sewage near street 4 community clinic in Chandan Nagar",
  "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
  "metadata": {
    "flow_id": "viasocket-flow-civic-01",
    "sender_name": "Ramesh Sharma",
    "location_hint": "Chandan Nagar Street 4",
    "ward_id": "indore-ward-02"
  }
}'

START_TIME=$(date +%s%N 2>/dev/null || python3 -c 'import time; print(int(time.time()*1e9))')

RESPONSE=$(curl -s -w "\n%{http_code}\n%{time_total}" -X POST "${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "${PAYLOAD}")

HTTP_CODE=$(echo "${RESPONSE}" | tail -n 2 | head -n 1)
TIME_TOTAL=$(echo "${RESPONSE}" | tail -n 1)
BODY=$(echo "${RESPONSE}" | head -n -2)

echo "Response Body: ${BODY}"
echo "HTTP Status Code: ${HTTP_CODE}"
echo "Total Round-trip Time: ${TIME_TOTAL}s"

if [ "${HTTP_CODE}" -eq 200 ]; then
  echo "✅ TEST PASSED: Webhook processed successfully!"
  # Verify latency < 2 seconds
  IS_FAST=$(awk -v t="${TIME_TOTAL}" 'BEGIN {print (t < 2.0) ? "YES" : "NO"}')
  if [ "${IS_FAST}" == "YES" ]; then
    echo "⚡ Latency Benchmark: PASSED (< 2.0s requirement met: ${TIME_TOTAL}s)"
  else
    echo "⚠️ Warning: Response time exceeded 2 seconds (${TIME_TOTAL}s)"
  fi
else
  echo "❌ TEST FAILED: Expected HTTP 200, got ${HTTP_CODE}"
  exit 1
fi
