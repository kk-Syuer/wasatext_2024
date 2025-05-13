#!/usr/bin/env bash
set -euo pipefail

HOST=localhost:3000

echo "1) Log in as alice → get a session token"
TOKEN=$(curl -s -X POST $HOST/session \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice"}' \
  | jq -r .identifier)

echo "   → token = $TOKEN"
echo

echo "2) Try updating alice’s name *without* auth (should 401)"
curl -i -X PATCH $HOST/user/name \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","name":"Alice Cooper"}'

echo
echo "3) Now update alice’s name *with* auth (should 200 + JSON echo)"
curl -i -X PATCH $HOST/user/name \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","name":"Alice Cooper"}'

echo
echo "4) Fetch alice’s profile to verify the change"
curl -i -X GET $HOST/users/alice \
  -H "Authorization: Bearer $TOKEN"
echo
