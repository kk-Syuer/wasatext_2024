package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// FindConversationByParticipants returns the ID of a conversation whose participants
// exactly match the given usernames (order doesn't matter), or empty string if none.
func (adb *AppDatabase) FindConversationByParticipants(
	ctx context.Context,
	participants []string,
) (string, error) {
	// We look in conversation_participants for conversation_id grouping by id
	// where the total count equals len(participants) and each participant is in the set.
	placeholders := strings.Repeat("?,", len(participants))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma

	query := fmt.Sprintf(`
	SELECT conversation_id
	  FROM conversation_participants
	 WHERE username IN (%s)
	 GROUP BY conversation_id
	HAVING 
	   COUNT(*) = ?              -- same number of participants
	   AND COUNT(*) = SUM(
	       CASE WHEN username IN (%s) THEN 1 ELSE 0 END
	   )
	LIMIT 1
	`, placeholders, placeholders)

	// build args: participants..., len, participants...
	args := make([]interface{}, 0, len(participants)*2+1)
	for _, u := range participants {
		args = append(args, u)
	}
	args = append(args, len(participants))
	for _, u := range participants {
		args = append(args, u)
	}

	var convID string
	err := adb.db.QueryRowContext(ctx, query, args...).Scan(&convID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // no existing conversation
		}
		return "", err
	}
	return convID, nil
}
