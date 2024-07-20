package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/alexPavlikov/go-message/internal/config"
	models "github.com/alexPavlikov/go-message/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB       *pgxpool.Pool
	Cfg      *config.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
}

func New(db *pgxpool.Pool, producer sarama.SyncProducer, cfg *config.Config, consumer sarama.Consumer) *Repository {
	return &Repository{
		DB:       db,
		Cfg:      cfg,
		Producer: producer,
		Consumer: consumer,
	}
}

func (r *Repository) InsertMessage(ctx context.Context, message *models.Message) error {
	query := `
	INSERT INTO public."message" (chat_id, user_id, text, time, db, kafka, deleted) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	row := r.DB.QueryRow(ctx, query, message.ChatID, message.UserID, message.Text, message.Time, true, false, false)

	if err := row.Scan(&message.ID); err != nil {
		return fmt.Errorf("failed insert message: %w", err)
	}

	return nil
}

func (r *Repository) SelectMessage(ctx context.Context, id int) (message models.Message, err error) {
	query := `SELECT id, chat_id, user_id, text, time, db, kafka FROM public."message" WHERE id = $1 AND deleted = false`

	row := r.DB.QueryRow(ctx, query, id)

	err = row.Scan(&message.ID, &message.ChatID, &message.UserID, &message.Text, &message.Time, &message.DB, &message.Kafka)
	if err != nil {
		return models.Message{}, fmt.Errorf("failed select message by id: %d - error: %w", id, err)
	}

	return message, nil
}

func (r *Repository) SelectMessages(ctx context.Context, options string) (messages []models.Message, err error) {
	query := `SELECT id, chat_id, user_id, text, time, db, kafka FROM public."message" WHERE deleted = false ` + options

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed select messages: %w", err)
	}

	for rows.Next() {
		var message models.Message
		err = rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.Text, &message.Time, &message.DB, &message.Kafka)
		if err != nil {
			return nil, fmt.Errorf("failed select messages: %w", err)
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *Repository) UpdateMessages(ctx context.Context, id int, options string) error {
	query := `UPDATE public."message" SET ` + options

	query += "WHERE id = $1 AND deleted = false"

	_ = r.DB.QueryRow(ctx, query, id)

	return nil
}

func (r *Repository) StoreMessageToKafka(ctx context.Context, message models.Message) error {

	jsonMsg, err := json.Marshal(&message)
	if err != nil {
		return fmt.Errorf("kafka producer send message: %w", err)
	}

	var msg = sarama.ProducerMessage{
		Topic:     r.Cfg.Topic,
		Key:       sarama.StringEncoder(message.ID),
		Value:     sarama.ByteEncoder(jsonMsg),
		Headers:   []sarama.RecordHeader{},
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Now(),
	}

	if _, _, err := r.Producer.SendMessage(&msg); err != nil {
		return fmt.Errorf("kafka producer send message error: %w", err)
	}

	return nil
}

func (r *Repository) ReadMessageFromKafka() ([]models.Message, error) {

	partition, err := r.Consumer.Partitions(r.Cfg.Topic)
	if err != nil {
		return nil, err
	}

	partitionConsumer, err := r.Consumer.ConsumePartition(r.Cfg.Topic, partition[0], sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}
	defer partitionConsumer.Close()

	var msgs []models.Message

	for msg := range partitionConsumer.Messages() {
		slog.Info("Consumed message: [%s], offset: [%d]", "value", msg.Value, "offset", msg.Offset)
		var message models.Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			return nil, fmt.Errorf("failed unmarshal result message from kafka: %w", err)
		}
		msgs = append(msgs, message)
	}

	return msgs, nil
}
