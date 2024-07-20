package service

import (
	"context"
	"fmt"

	models "github.com/alexPavlikov/go-message/internal/domain"
	"github.com/alexPavlikov/go-message/internal/server/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) AddMessage(ctx context.Context, message *models.Message) error {
	if err := s.repo.InsertMessage(ctx, message); err != nil {
		return fmt.Errorf("add message failed: %w", err)
	}

	return nil
}

func (s *Service) GetMessage(ctx context.Context, id int) (message models.Message, err error) {
	message, err = s.repo.SelectMessage(ctx, id)
	if err != nil {
		return models.Message{}, fmt.Errorf("get message failed: %w", err)
	}

	return message, nil
}

func (s *Service) GetMessages(ctx context.Context, options map[string]interface{}) (messages []models.Message, err error) {
	var query string

	for k, v := range options {
		query += fmt.Sprintf("AND %s ILIKE %v ", k, v)
	}

	messages, err = s.repo.SelectMessages(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get messages failed: %w", err)
	}

	return messages, nil
}

func (s *Service) UpdateMessage(ctx context.Context, id int, options map[string]interface{}) error {
	var query string

	for k, v := range options {
		query += fmt.Sprintf("%s = %v ", k, v)
	}

	if err := s.repo.UpdateMessages(ctx, id, query); err != nil {
		return fmt.Errorf("update message failed: %w", err)
	}

	return nil
}

func (s *Service) AddMessageToKafka(ctx context.Context, message models.Message) error {
	if err := s.repo.StoreMessageToKafka(ctx, message); err != nil {
		return fmt.Errorf("failed add message to kafka: %w", err)
	}

	return nil
}

func (s *Service) GetMessageFromKafka() ([]models.Message, error) {
	msgs, err := s.repo.ReadMessageFromKafka()
	if err != nil {
		return nil, fmt.Errorf("failed get message from kafka: %w", err)
	}

	return msgs, nil
}
