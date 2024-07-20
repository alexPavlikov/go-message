package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alexPavlikov/go-message/internal/server/locations"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Router struct {
	handler *locations.Handler
}

func New(handler *locations.Handler) *Router {
	return &Router{
		handler: handler,
	}
}

func (r *Router) Build() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/v1/message/add", handlerWrapper(r.handler.MessageAddHandler))
	router.Get("/v1/message/get", handlerWrapper(r.handler.MessageGetHandler))

	return router
}

type wrapperFunc[Input, Output any] func(r *http.Request, data Input) (Output, error)

func handlerWrapper[Input, Output any](fn wrapperFunc[Input, Output]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data Input

		if r.Method != http.MethodGet {
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&data); err != nil {
				slog.ErrorContext(r.Context(), "failed to decode request data to handler", "error", err)
				http.Error(w, "bad request"+err.Error(), http.StatusBadRequest)
				return
			}

		}

		response, err := fn(r, data)
		if err != nil {
			slog.ErrorContext(r.Context(), "can`t handler request", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err = json.NewEncoder(w).Encode(&response); err != nil {
			slog.ErrorContext(r.Context(), "faield to encode response", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}
