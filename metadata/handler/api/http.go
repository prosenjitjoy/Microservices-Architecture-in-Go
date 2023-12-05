package api

import (
	"encoding/json"
	"errors"
	"log"
	"main/metadata/model"
	"main/metadata/repository"
	"main/metadata/service"
	"net/http"
)

// Handler defines a movie metadata HTTP handler.
type Handler struct {
	ctrl *service.MetadataService
}

// New creates a new movie metadata HTTP handler.
func New(ctrl *service.MetadataService) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		m, err := h.ctrl.GetMetadata(r.Context(), id)
		if err != nil && errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Repository get error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			log.Printf("Response encode error: %v\n", err)
		}
	case http.MethodPut:
		title := r.FormValue("title")
		if title == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		description := r.FormValue("description")
		if description == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		director := r.FormValue("director")
		if director == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := h.ctrl.PutMetadata(r.Context(), id, &model.Metadata{
			ID:          id,
			Title:       title,
			Description: description,
			Director:    director,
		})
		if err != nil {
			log.Printf("Repository put error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
