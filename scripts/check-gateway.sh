#!/usr/bin/env bash

set -u

BASE_URL="${1:-http://localhost:8080}"
SLEEP_SECONDS=3

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

pause_between_checks() {
	echo
	echo "waiting ${SLEEP_SECONDS}s..."
	sleep "${SLEEP_SECONDS}"
	echo
}

print_section() {
	echo "============================================================"
	echo "$1"
	echo "============================================================"
}

run_json_check() {
	local title="$1"
	local method="$2"
	local path="$3"
	local body="${4:-}"

	print_section "$title"
	echo "${method} ${BASE_URL}${path}"

	if [[ -n "$body" ]]; then
		curl -sS -X "$method" \
			-H "Content-Type: application/json" \
			-d "$body" \
			"${BASE_URL}${path}" | jq
	else
		curl -sS -X "$method" "${BASE_URL}${path}" | jq
	fi

	pause_between_checks
}

run_header_check() {
	local title="$1"
	local method="$2"
	local path="$3"
	local header_pattern="$4"

	print_section "$title"
	echo "${method} ${BASE_URL}${path}"
	curl -si -X "$method" "${BASE_URL}${path}" | grep -i "$header_pattern"

	pause_between_checks
}

require_command curl
require_command jq
require_command grep

run_json_check "Health Check" "GET" "/health"
run_json_check "Valid Trip Preview" "POST" "/api/v1/trips/preview" '{"origin":"Cairo","destination":"Giza"}'
run_json_check "Missing Origin Should Return 400 VALIDATION_ERROR" "POST" "/api/v1/trips/preview" '{"destination":"Giza"}'
run_json_check "Same Origin And Destination Should Return 400 VALIDATION_ERROR" "POST" "/api/v1/trips/preview" '{"origin":"Cairo","destination":"Cairo"}'
run_json_check "Invalid JSON Should Return 400 INVALID_REQUEST" "POST" "/api/v1/trips/preview" 'not-json'
run_json_check "Unknown Route Should Return 404 NOT_FOUND" "GET" "/api/v1/unknown"
run_json_check "Wrong Method Should Return 405 METHOD_NOT_ALLOWED" "GET" "/api/v1/trips/preview"
run_header_check "Health Check Should Include X-Request-ID Header" "GET" "/health" "x-request-id"

print_section "Gateway Smoke Checks Finished"
