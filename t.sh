#!/bin/bash


# This script is for testing the impossible travel feature by simulating login attempts from given geo-coordinates baesd on
# forged CF headers, and expecting impossible_travel flag to true if the supposed speed is higher than the configured threshold

##################
# IMPORTANT NOTE #
##################
########################################################################################################################
# This script IS/SHOULD-NOT be considered as part of the PR itself, it should be rather replaced by an integration test
# It is included here just to give an idea how the feature was quickly tested during the development phase
########################################################################################################################
usage() {
    echo "Missing Email"
    echo "Usage ./test_impossible_travel.sh <email> <password> <LAT> <LON>"
    echo "Ex (Munich) ./test_impossible_travel.sh test@test.com ChangeMe12_34 48.1475 11.5645"
    echo "Ex (NYC) ./test_impossible_travel.sh test@test.com ChangeMe12_34 40.7128 -74.0060"
    echo "Ex (Tokyo) ./test_impossible_travel.sh test@test.com ChangeMe12_34 35.6895 -139.6917"
    exit 1
}

EMAIL=${1}
PASSWORD=${2}
LAT=${3}
LON=${4}

KRATOS_URL="http://127.0.0.1:4433"
COOKIE_FILE="session_cookies.txt"
if [[ -z "$PASSWORD" || -z "$EMAIL" || -z "$LAT" || -z "$LON" ]]; then
  usage
fi

echo "Starting Login Flow ..."
FLOW_ID=$(curl -s -c "${COOKIE_FILE}" "${KRATOS_URL}/self-service/login/api" | jq -r .id)

if [[ "$FLOW_ID" = "null" ]]; then
    echo "ERROR: Failed to retrieve Flow ID. Check if Kratos is running and accessible."
    exit 1
fi
echo "Flow ID: $FLOW_ID"

echo "Simulating Login from geo-coordinates (${LAT}, ${LON})..."
impossible_travel=$(curl -s -X POST \
  -b "${COOKIE_FILE}" -c "${COOKIE_FILE}" \
  -H "Content-Type: application/json" \
  -H "Cf-Iplatitude: ${LAT}" \
  -H "Cf-Iplongitude: ${LON}" \
  "${KRATOS_URL}/self-service/login?flow=${FLOW_ID}" \
  -d "{\"method\":\"password\",\"identifier\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}" | jq .session.impossible_travel)

if [ $? -eq 0 ]; then
  echo "Login Successful!"
fi

echo "impossible_travel: $impossible_travel"
