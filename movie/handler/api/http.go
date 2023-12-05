package api

import (
	"encoding/json"
	"errors"
	"log"
	"main/movie/service"
	"net/http"
)

// Handler defines a movie handler
type Handler struct {
	ctrl *service.MovieService
}

// New creates a new movie HTTP handler.
func New(ctrl *service.MovieService) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

// GetMovieDetails handles GET /movie requests.
func (h *Handler) GetMovieDetails(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	ctx := r.Context()
	details, err := h.ctrl.Get(ctx, id)
	if err != nil && errors.Is(err, service.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Repository got error:", err)
	}

	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Println("Response encode error:", err)
	}
}
