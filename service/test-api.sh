#!/usr/bin/env bash
set -euo pipefail

# Configuration (override via env vars if desired)
BASE_URL=${BASE_URL:-http://localhost:3000}
USERNAME=${USERNAME:-alice}
PASSWORD=${PASSWORD:-secret}
NEW_NAME=${NEW_NAME:-alice_new}
RECIPIENT=${RECIPIENT:-bob}
ADD_MEMBER=${ADD_MEMBER:-charlie}
PHOTO_FILE=${PHOTO_FILE:-./example_image.png}
GROUP_PHOTO_FILE=${GROUP_PHOTO_FILE:-./example_image.png}
GROUP_NAME=${GROUP_NAME:-testgroup}

echo "⚙️  Base URL: $BASE_URL"
echo "👤  Login as: $USERNAME"

# 1) Login → get Bearer token from "identifier"
echo -e "\n=== 1. POST /session ==="
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/session" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\"}")
TOKEN_ID=$(echo "$LOGIN_RESP" | sed -E 's/.*"identifier":"([^"]+)".*/\1/')
TOKEN="Bearer $TOKEN_ID"
echo "Token: $TOKEN"

# 2) PATCH /user/name
echo -e "\n=== 2. PATCH /user/name ==="
curl -i -X PATCH "$BASE_URL/user/name" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"$NEW_NAME\"}"

# 3) PATCH /user/photo
echo -e "\n=== 3. PATCH /user/photo ==="
curl -i -X PATCH "$BASE_URL/user/photo" \
  -H "Authorization: $TOKEN" \
  -F "photo=@${PHOTO_FILE}"

# 4) GET /users
echo -e "\n=== 4. GET /users ==="
curl -i -X GET "$BASE_URL/users" \
  -H "Authorization: $TOKEN"

# 5) POST /conversations
echo -e "\n=== 5. POST /conversations ==="
CONV_CREATE=$(curl -s -X POST "$BASE_URL/conversations" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient":"'"$RECIPIENT"'",
    "initialMessage":{
      "contentType":"text",
      "text":"Hello, '"$RECIPIENT"'!"
    }
  }')
CONV_ID=$(echo "$CONV_CREATE" | sed -E 's/.*"ID":"?([^"]+)"?.*/\1/')
echo "Conversation ID: $CONV_ID"

# 6) GET /conversations?user=
echo -e "\n=== 6. GET /conversations?user=$USERNAME ==="
curl -i -X GET "$BASE_URL/conversations?user=$USERNAME" \
  -H "Authorization: $TOKEN"

# 7) GET /conversations/{id}
echo -e "\n=== 7. GET /conversations/$CONV_ID ==="
curl -i -X GET "$BASE_URL/conversations/$CONV_ID" \
  -H "Authorization: $TOKEN"

# 8) POST /messages
echo -e "\n=== 8. POST /messages ==="
MSG_RESP=$(curl -s -X POST "$BASE_URL/messages" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversationId":"'"$CONV_ID"'",
    "senderUsername":"'"$USERNAME"'",
    "contentType":"text",
    "text":"This is a test message"
  }')
MSG_ID=$(echo "$MSG_RESP" | sed -E 's/.*"id":"([^"]+)".*/\1/')
echo "Message ID: $MSG_ID"

# 9) GET /conversations/{id}/messages
echo -e "\n=== 9. GET /conversations/$CONV_ID/messages ==="
curl -i -X GET "$BASE_URL/conversations/$CONV_ID/messages" \
  -H "Authorization: $TOKEN"

# 10) POST /messages/{id}/forward
echo -e "\n=== 10. POST /messages/$MSG_ID/forward ==="
FWD_RESP=$(curl -s -X POST "$BASE_URL/messages/$MSG_ID/forward" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"conversationId\":\"$CONV_ID\"}")
FWD_ID=$(echo "$FWD_RESP" | sed -E 's/.*"id":"([^"]+)".*/\1/')
echo "Forwarded Message ID: $FWD_ID"

# 11) POST /messages/{id}/reply
echo -e "\n=== 11. POST /messages/$MSG_ID/reply ==="
REPL_RESP=$(curl -s -X POST "$BASE_URL/messages/$MSG_ID/reply" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Replying to your message"}')
REPL_ID=$(echo "$REPL_RESP" | sed -E 's/.*"id":"([^"]+)".*/\1/')
echo "Reply Message ID: $REPL_ID"

# 12) POST /messages/{id}/reaction
echo -e "\n=== 12. POST /messages/$MSG_ID/reaction ==="
REA_RESP=$(curl -s -X POST "$BASE_URL/messages/$MSG_ID/reaction" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"emoji\":\"👍\",\"username\":\"$USERNAME\"}")
REACTION_ID=$(echo "$REA_RESP" | sed -E 's/.*"id":"?([^"]+)"?.*/\1/')
echo "Reaction ID: $REACTION_ID"

# 13) DELETE /messages/{id}/reaction/{reactionId}
echo -e "\n=== 13. DELETE /messages/$MSG_ID/reaction/$REACTION_ID ==="
curl -i -X DELETE "$BASE_URL/messages/$MSG_ID/reaction/$REACTION_ID" \
  -H "Authorization: $TOKEN"

# 14) DELETE /messages/{id}
echo -e "\n=== 14. DELETE /messages/$MSG_ID ==="
curl -i -X DELETE "$BASE_URL/messages/$MSG_ID" \
  -H "Authorization: $TOKEN"

# 15) POST /groups
echo -e "\n=== 15. POST /groups ==="
curl -s -X POST "$BASE_URL/groups" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"'"$GROUP_NAME"'",
    "photoUrl":"",
    "members":["'"$USERNAME"'","'"$RECIPIENT"'"]
  }' | tee /dev/null

# 16) GET /groups
echo -e "\n=== 16. GET /groups ==="
curl -i -X GET "$BASE_URL/groups" \
  -H "Authorization: $TOKEN"

# 17) GET /groups/{name}
echo -e "\n=== 17. GET /groups/$GROUP_NAME ==="
curl -i -X GET "$BASE_URL/groups/$GROUP_NAME" \
  -H "Authorization: $TOKEN"

# 18) POST /groups/{name}/members
echo -e "\n=== 18. POST /groups/$GROUP_NAME/members ==="
curl -i -X POST "$BASE_URL/groups/$GROUP_NAME/members" \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADD_MEMBER\"}"

# 19) PATCH /groups/{name}/photo
echo -e "\n=== 19. PATCH /groups/$GROUP_NAME/photo ==="
curl -i -X PATCH "$BASE_URL/groups/$GROUP_NAME/photo" \
  -H "Authorization: $TOKEN" \
  -F "photo=@${GROUP_PHOTO_FILE}"

# 20) POST /groups/{name}/leave
echo -e "\n=== 20. POST /groups/$GROUP_NAME/leave ==="
curl -i -X POST "$BASE_URL/groups/$GROUP_NAME/leave" \
  -H "Authorization: $TOKEN"

# 21) GET /conversations/{id}/messages/status
echo -e "\n=== 21. GET /conversations/$CONV_ID/messages/status ==="
curl -i -X GET "$BASE_URL/conversations/$CONV_ID/messages/status" \
  -H "Authorization: $TOKEN"

echo -e "\n✅ All tests executed."
