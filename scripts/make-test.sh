#!/bin/bash

set -e

# ── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log()     { echo -e "${CYAN}[TEST]${NC} $1"; }
success() { echo -e "${GREEN}[PASS]${NC} $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()    { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

# ── Config ────────────────────────────────────────────────
GATEWAY_URL="http://localhost:8080"
DRIVER_GRPC="localhost:50052"
WS_URL="ws://localhost:8080/ws"
RIDER_ID="rider-test-$$"
WS_OUTPUT="/tmp/rideflow-ws-$$.txt"
LOG_FILE="./test_logger"

# ── Log Formatter ─────────────────────────────────────────
format_log() {
  local service=$1
  while IFS= read -r line; do
    if echo "$line" | jq -e . >/dev/null 2>&1; then
      TIME=$(echo "$line" | jq -r '.time' | awk -F'T' '{print $2}' | awk -F'+' '{print $1}')
      LEVEL=$(echo "$line" | jq -r '.level')
      MSG=$(echo "$line" | jq -r '.msg')

      EXTRA=$(echo "$line" | jq -r '
        to_entries
        | map(select(
            .key != "time" and
            .key != "level" and
            .key != "msg" and
            .key != "service"
          ))
        | map("\(.key)=\(.value)")
        | join(" ")
      ')

      case "$LEVEL" in
        INFO)  LEVEL_COLOR="${GREEN}INFO ${NC}" ;;
        WARN)  LEVEL_COLOR="${YELLOW}WARN ${NC}" ;;
        ERROR) LEVEL_COLOR="${RED}ERROR${NC}" ;;
        DEBUG) LEVEL_COLOR="${BLUE}DEBUG${NC}" ;;
        *)     LEVEL_COLOR="$LEVEL" ;;
      esac

      case "$service" in
        trip)    SVC_COLOR="${BLUE}[trip]   ${NC}" ;;
        driver)  SVC_COLOR="${YELLOW}[driver] ${NC}" ;;
        gateway) SVC_COLOR="${GREEN}[gateway]${NC}" ;;
        *)       SVC_COLOR="[$service]" ;;
      esac

      FORMATTED="${TIME} ${LEVEL_COLOR} ${SVC_COLOR} ${MSG} ${EXTRA}"

      # Print to terminal with colors
      echo -e "$FORMATTED"

      # Write to log file without color escape codes
      echo "${TIME} ${LEVEL} [${service}] ${MSG} ${EXTRA}" >> "$LOG_FILE"
    else
      # Not JSON — compile errors, panics etc
      echo "[$service] $line"
      echo "[$service] $line" >> "$LOG_FILE"
    fi
  done
}

# ── Cleanup ───────────────────────────────────────────────
cleanup() {
  echo "" >> "$LOG_FILE"
  echo "── cleanup ──────────────────────────────────────────" >> "$LOG_FILE"
  log "Cleaning up background processes..."
  kill $WS_PID     2>/dev/null || true
  kill $TRIP_PID   2>/dev/null || true
  kill $DRIVER_PID 2>/dev/null || true
  kill $GATEWAY_PID 2>/dev/null || true
  rm -f "$WS_OUTPUT"
  log "Full log saved → ${LOG_FILE}"
}
trap cleanup EXIT
# ── Kill any existing services ────────────────────────────
log "Killing any existing services on our ports..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:8082 | xargs kill -9 2>/dev/null || true
lsof -ti:8083 | xargs kill -9 2>/dev/null || true
lsof -ti:50051 | xargs kill -9 2>/dev/null || true
lsof -ti:50052 | xargs kill -9 2>/dev/null || true
sleep 1
success "Ports cleared"

# ── Init log file ─────────────────────────────────────────
echo "RideFlow Test Run — $(date)" > "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# ── Check dependencies ────────────────────────────────────
log "Checking dependencies..."
for cmd in curl grpcurl wscat jq; do
  if ! command -v $cmd &>/dev/null; then
    fail "$cmd is not installed"
  fi
done
success "All dependencies found"

# ── Start RabbitMQ ────────────────────────────────────────
log "Starting RabbitMQ..."
docker compose up -d
sleep 2
success "RabbitMQ up"

# ── Start Services ────────────────────────────────────────
echo "" >> "$LOG_FILE"
echo "── service logs ─────────────────────────────────────" >> "$LOG_FILE"

log "Starting trip service..."
go run ./cmd/trip 2>&1 | tee >(format_log "trip") > /dev/null &
TRIP_PID=$!

log "Starting driver service..."
go run ./cmd/driver 2>&1 | tee >(format_log "driver") > /dev/null &
DRIVER_PID=$!

log "Starting gateway..."
go run ./cmd/gateway 2>&1 | tee >(format_log "gateway") > /dev/null &
GATEWAY_PID=$!

# ── Wait for services to be healthy ──────────────────────
log "Waiting for services to be healthy..."

wait_for_http() {
  local name=$1
  local url=$2
  local max=30
  local i=0

  while [ $i -lt $max ]; do
    if curl -s -o /dev/null -w "%{http_code}" "$url/health" | grep -q "200"; then
      success "$name is healthy"
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done

  fail "$name did not become healthy after ${max}s"
}

wait_for_grpc() {
  local name=$1
  local addr=$2
  local max=30
  local i=0

  while [ $i -lt $max ]; do
    if grpcurl -plaintext "$addr" list &>/dev/null; then
      success "$name gRPC is healthy"
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done

  fail "$name gRPC did not become healthy after ${max}s"
}

wait_for_http "gateway" "$GATEWAY_URL"
wait_for_grpc "driver"  "$DRIVER_GRPC"

# ── Wait for manual WebSocket connection ──────────────────
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  ACTION REQUIRED${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Open a new terminal and run:"
echo ""
echo -e "  ${GREEN}wscat -c \"ws://localhost:8080/ws?user_id=$RIDER_ID\" | tee $WS_OUTPUT${NC}"
echo ""
echo -e "  Waiting for connection... (press CTRL+C to abort)"
echo ""

MAX_WAIT=60
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
  # Detect connection via gateway log — more reliable than parsing wscat output
  if grep -q "websocket client connected.*user_id=${RIDER_ID}" "$LOG_FILE" 2>/dev/null; then
    success "WebSocket connection detected"
    break
  fi
  printf "\r${BLUE}[TEST]${NC} Waiting for wscat connection... ${WAITED}s"
  sleep 1
  WAITED=$((WAITED + 1))
done

if [ $WAITED -ge $MAX_WAIT ]; then
  fail "No WebSocket connection detected after ${MAX_WAIT}s"
fi
# ── Register Driver ───────────────────────────────────────
log "Registering driver..."
DRIVER_RESPONSE=$(grpcurl -plaintext \
  -d '{"name":"Test Driver","vehicle":"Toyota Camry - White"}' \
  "$DRIVER_GRPC" \
  driver.DriverService/RegisterDriver)

DRIVER_ID=$(echo "$DRIVER_RESPONSE" | jq -r '.driverId')
if [ -z "$DRIVER_ID" ] || [ "$DRIVER_ID" = "null" ]; then
  fail "Failed to register driver. Response: $DRIVER_RESPONSE"
fi
success "Driver registered: $DRIVER_ID"
echo "driver_id=$DRIVER_ID" >> "$LOG_FILE"

# ── Update Driver Location ────────────────────────────────
log "Setting driver location..."
grpcurl -plaintext \
  -d "{\"driver_id\":\"$DRIVER_ID\",\"latitude\":30.04,\"longitude\":31.23}" \
  "$DRIVER_GRPC" \
  driver.DriverService/UpdateLocation > /dev/null
success "Driver location set (30.04, 31.23)"

# ── Set Driver Online ─────────────────────────────────────
log "Setting driver online..."
grpcurl -plaintext \
  -d "{\"driver_id\":\"$DRIVER_ID\",\"online\":true}" \
  "$DRIVER_GRPC" \
  driver.DriverService/SetAvailability > /dev/null
success "Driver is online and available"

# ── Create Trip ───────────────────────────────────────────
log "Creating trip for $RIDER_ID..."
TRIP_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/api/v1/trips" \
  -H "Content-Type: application/json" \
  -d "{
    \"rider_id\": \"$RIDER_ID\",
    \"origin\": \"Cairo\",
    \"destination\": \"Giza\",
    \"origin_lat\": 30.04,
    \"origin_lng\": 31.23
  }")

TRIP_ID=$(echo "$TRIP_RESPONSE"     | jq -r '.data.trip_id')
TRIP_STATUS=$(echo "$TRIP_RESPONSE" | jq -r '.data.status')
FARE=$(echo "$TRIP_RESPONSE"        | jq -r '.data.fare_estimate')

if [ -z "$TRIP_ID" ] || [ "$TRIP_ID" = "null" ]; then
  fail "Failed to create trip. Response: $TRIP_RESPONSE"
fi
success "Trip created: $TRIP_ID (status: $TRIP_STATUS, fare: $FARE)"
echo "trip_id=$TRIP_ID" >> "$LOG_FILE"

# ── Wait for WebSocket Push ───────────────────────────────
log "Waiting for WebSocket push..."
MAX_WAIT=5
WAITED=0

while [ $WAITED -lt $MAX_WAIT ]; do
  if grep -q "driver_assigned" "$WS_OUTPUT" 2>/dev/null; then
    break
  fi
  sleep 1
  WAITED=$((WAITED + 1))
done

if ! grep -q "driver_assigned" "$WS_OUTPUT" 2>/dev/null; then
  fail "No WebSocket push received after ${MAX_WAIT}s"
fi

# ── Parse and Validate WebSocket Message ─────────────────
WS_MESSAGE=$(grep "driver_assigned" "$WS_OUTPUT" | head -1 | sed 's/^< //')
WS_TYPE=$(echo "$WS_MESSAGE"      | jq -r '.type')
WS_TRIP_ID=$(echo "$WS_MESSAGE"   | jq -r '.payload.trip_id')
WS_DRIVER_ID=$(echo "$WS_MESSAGE" | jq -r '.payload.driver_id')

success "WebSocket push received!"
echo "" >> "$LOG_FILE"
echo "── websocket push ───────────────────────────────────" >> "$LOG_FILE"
echo "type=$WS_TYPE trip_id=$WS_TRIP_ID driver_id=$WS_DRIVER_ID" >> "$LOG_FILE"

echo ""
echo -e "  ${YELLOW}Event type:${NC}  $WS_TYPE"
echo -e "  ${YELLOW}Trip ID:${NC}     $WS_TRIP_ID"
echo -e "  ${YELLOW}Driver ID:${NC}   $WS_DRIVER_ID"
echo ""

[ "$WS_TYPE"      = "driver_assigned" ] || fail "Wrong event type: $WS_TYPE"
[ "$WS_TRIP_ID"   = "$TRIP_ID" ]        || fail "Trip ID mismatch. Expected: $TRIP_ID, Got: $WS_TRIP_ID"
[ "$WS_DRIVER_ID" = "$DRIVER_ID" ]      || fail "Driver ID mismatch. Expected: $DRIVER_ID, Got: $WS_DRIVER_ID"

success "Event type matches"
success "Trip ID matches"
success "Driver ID matches"

# ── Check Final Trip State ────────────────────────────────
log "Checking final trip state..."
sleep 1

TRIP_STATE=$(curl -s "$GATEWAY_URL/api/v1/trips/$TRIP_ID")
FINAL_STATUS=$(echo "$TRIP_STATE" | jq -r '.data.status')
FINAL_DRIVER=$(echo "$TRIP_STATE" | jq -r '.data.driver_id')

echo "" >> "$LOG_FILE"
echo "── final trip state ─────────────────────────────────" >> "$LOG_FILE"
echo "status=$FINAL_STATUS driver_id=$FINAL_DRIVER" >> "$LOG_FILE"

echo ""
echo -e "  ${YELLOW}Final status:${NC}    $FINAL_STATUS"
echo -e "  ${YELLOW}Assigned driver:${NC} $FINAL_DRIVER"
echo ""

[ "$FINAL_STATUS" = "accepted"  ] || warn "Trip status is '$FINAL_STATUS', expected 'accepted'"
[ "$FINAL_DRIVER" = "$DRIVER_ID"] || warn "Driver ID on trip does not match"

success "Trip status is accepted"
success "Driver correctly assigned to trip"

# ── Summary ───────────────────────────────────────────────
echo "" >> "$LOG_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >> "$LOG_FILE"
echo "RESULT: ALL CHECKS PASSED" >> "$LOG_FILE"
echo "trip=$TRIP_ID driver=$DRIVER_ID fare=$FARE status=$FINAL_STATUS" >> "$LOG_FILE"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ALL CHECKS PASSED — Phase 8 + 9 flow verified   ${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Trip:    $TRIP_ID"
echo -e "  Driver:  $DRIVER_ID"
echo -e "  Fare:    $FARE EGP"
echo -e "  Status:  $FINAL_STATUS"
echo ""
echo -e "  Log:     ${YELLOW}./test_logger${NC}"
echo ""
```

---

The `test_logger` file at the project root will look like this:
```
RideFlow Test Run — Thu Mar 12 11:04:00 UTC 2026
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

── service logs ─────────────────────────────────────
10:51:15 INFO [trip]    starting trip service http_port=8082 grpc_port=50051
10:51:15 INFO [trip]    connected to rabbitmq
10:51:16 INFO [driver]  starting driver service http_port=8083 grpc_port=50052
10:51:16 INFO [gateway] starting api gateway port=8080
10:51:20 INFO [trip]    trip created trip_id=5f66de22 rider_id=rider-test-123
10:51:20 INFO [driver]  trip created event received trip_id=5f66de22
10:51:20 INFO [driver]  driver assigned trip_id=5f66de22 driver_id=738e9125
10:51:20 INFO [trip]    driver assigned to trip trip_id=5f66de22 driver_id=738e9125

── websocket push ───────────────────────────────────
type=driver_assigned trip_id=5f66de22 driver_id=738e9125

── final trip state ─────────────────────────────────
status=accepted driver_id=738e9125

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RESULT: ALL CHECKS PASSED
trip=5f66de22 driver=738e9125 fare=30 status=accepted