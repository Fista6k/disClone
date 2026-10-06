package messages

import (
	"context"
	"log/slog"

	"github.com/Fista6k/disClone/internal"
)

type IMessageRepository interface {
	CreateMessage(ctx context.Context, message *Message) error
	GetConversation(ctx context.Context, authorID, recipientID int64) ([]*Message, error)
}

type MessageRepository struct {
	Storage *internal.Storage
}

func NewMessageRepo(storage *internal.Storage) *MessageRepository {
	return &MessageRepository{
		Storage: storage,
	}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (author_id, recipient_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.Storage.DB.QueryRowContext(
		ctx,
		query,
		message.AuthorID,
		message.RecipientID,
		message.Content,
		message.CreatedAt,
	).Scan(&message.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *MessageRepository) GetConversation(ctx context.Context, authorID, recipientID int64) ([]*Message, error) {
	query := `
		SELECT id, author_id, content, recipient_id, created_at
		FROM messages
		WHERE (author_id = $1 AND recipient_id = $2) OR (author_id = $2 AND recipient_id = $1)
		ORDER BY created_at ASC;
	`
	rows, err := r.Storage.DB.QueryContext(ctx, query, authorID, recipientID)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	var messages []*Message
	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.AuthorID, &message.Content, &message.RecipientID, &message.CreatedAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
