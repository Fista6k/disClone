package messages

import (
	"context"
	"time"
)

type MessageService struct {
	messageRepo IMessageRepository
}

func NewMessageService(repo IMessageRepository) *MessageService {
	return &MessageService{
		messageRepo: repo,
	}
}

func (s *MessageService) CreateMessage(ctx context.Context, authorId, recipientId int64, content string) (*Message, error) {
	message := &Message{
		AuthorID:    authorId,
		RecipientID: recipientId,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	err := s.messageRepo.CreateMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetConversation(ctx context.Context, authorID, recipientID int64) ([]*Message, error) {
	return s.messageRepo.GetConversation(ctx, authorID, recipientID)
}
