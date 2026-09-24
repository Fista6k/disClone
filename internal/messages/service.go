package messages

import (
	"context"
	"time"
)

type MessageService struct {
	repo IMessageRepository
}

func NewMessageService(repo IMessageRepository) *MessageService {
	return &MessageService{
		repo: repo,
	}
}

func (s *MessageService) CreateMessage(ctx context.Context, authorId, recipientId int64, content string) (*Message, error) {
	message := &Message{
		AuthorID:    authorId,
		RecipientID: recipientId,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetMessagesByRecipientID(ctx context.Context, recipientID int64) ([]*Message, error) {
	return s.repo.GetMessagesByRecipientID(ctx, recipientID)
}
