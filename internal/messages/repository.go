package messages

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Fista6k/disClone/internal"
)

type IMessageRepository interface {
	CreateMessage(ctx context.Context, message *Message) error
	GetMessagesByRecipientID(ctx context.Context, recipientID int64) ([]*Message, error)
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
		INSERT INTO messages (author_id, recipiet_id, content, created_at)
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

func (r *MessageRepository) GetMessagesByRecipientID(ctx context.Context, recipientID int64) ([]*Message, error) {
	query := `
		SELECT id, author_id, content, recipient_id, created_at
		FROM messages
		WHERE recipient_id = $1;
	`
	rows, err := r.Storage.DB.QueryContext(ctx, query, recipientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Not found messages to this user")
		}

		return nil, err
	}

	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.AuthorID, &message.Content, &message.RecipientID, &message.CreatedAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	return messages, nil
}
