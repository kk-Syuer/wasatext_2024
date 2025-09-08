package database

import (
	"context"
)

// DeliveryStatusRow represents one row from the delivery_status table.
type DeliveryStatusRow struct {
	MessageID string
	Recipient string
	Status    string
	UpdatedAt string
}

// GetDeliveryStatusForConversation returns the delivery status entries for all
// messages in the given conversation.
func (adb *AppDatabase) GetDeliveryStatusForConversation(
	ctx context.Context,
	conversationID string,
) ([]DeliveryStatusRow, error) {

	// We join delivery_status to messages to filter by conversation_id.
	const query = `
	SELECT ds.message_id,
	       ds.recipient,
	       ds.status,
	       ds.updated_at
	  FROM delivery_status ds
	  JOIN messages m
	    ON ds.message_id = m.id
	 WHERE m.conversation_id = ?
	`

	rows, err := adb.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DeliveryStatusRow
	for rows.Next() {
		var r DeliveryStatusRow
		if err := rows.Scan(&r.MessageID, &r.Recipient, &r.Status, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
