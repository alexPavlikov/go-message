package kafka

import (
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
)

// Функция получения producer
func GetProducer(address ...string) (producer sarama.SyncProducer, err error) {
	producer, err = sarama.NewSyncProducer(address, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating producer: %w", err)
	}

	slog.Info("get kafka producer complited")

	return producer, nil
}

// Функция получения consumer
func GetConsumer(address ...string) (sarama.Consumer, func() error, error) {

	client, err := sarama.NewClient(address, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating client kafka: %w", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, consumer.Close, fmt.Errorf("error creating consumer: %w", err)
	}

	slog.Info("get kafka consumer complited")

	return consumer, consumer.Close, nil
}
