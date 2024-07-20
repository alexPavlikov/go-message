package locations

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	models "github.com/alexPavlikov/go-message/internal/domain"
	"github.com/alexPavlikov/go-message/internal/server/service"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

type messageRequest struct {
	Id     int    `json:"id"`
	ChatID int    `json:"chat_id"`
	UserID int    `json:"user_id"`
	Text   string `json:"text"`
	Time   time.Duration
}

type messageResponse struct {
	Id     int       `json:"id"`
	ChatID int       `json:"chat_id"`
	UserID int       `json:"user_id"`
	Text   string    `json:"text"`
	Time   time.Time `json:"timestamp"`
}

type emptyRequest struct{}

type emptyResponse struct{}

func (h *Handler) MessageAddHandler(r *http.Request, message messageRequest) (emptyResponse, error) {

	ctx := r.Context()

	var msg = models.Message{
		ChatID: message.ChatID,
		UserID: message.UserID,
		Text:   message.Text,
		Time:   time.Now(),
		DB:     true,
	}

	if err := h.service.AddMessage(ctx, &msg); err != nil {
		slog.ErrorContext(ctx, "POST MessageAddHandler", "error", err)
		return emptyResponse{}, fmt.Errorf("POST MessageAddHandler error: %w", err)
	}

	if err := h.service.AddMessageToKafka(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "POST MessageAddHandler", "error", err)
		return emptyResponse{}, fmt.Errorf("POST MessageAddHandler error: %w", err)
	}

	return emptyResponse{}, nil
}

func (h *Handler) MessageGetHandler(r *http.Request, er emptyRequest) ([]messageResponse, error) {

	ctx := r.Context()

	var msgRep = make([]messageResponse, 0)

	var mu sync.Mutex

	var ch = make(chan models.Message)

	go func() {

		mu.Lock()

		msgs, err := h.service.GetMessageFromKafka()
		if err != nil {
			slog.ErrorContext(ctx, "POST MessageAddHandler", "error", err)
		}

		for _, v := range msgs {
			ch <- v
		}

		mu.Unlock()
	}()

	go func() {

		for v := range ch {
			mu.Lock()

			var msg = messageResponse{
				Id:     v.ChatID,
				ChatID: v.ChatID,
				UserID: v.UserID,
				Text:   v.Text,
				Time:   v.Time,
			}

			msgRep = append(msgRep, msg)
			mu.Unlock()
		}

	}()

	fmt.Println(msgRep)

	return msgRep, nil
}
