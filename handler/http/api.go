package http

import (
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/middleware"
	"github.com/dhanarJkusuma/quiz/quiz/usecase"
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
	auth   *pager.Pager
	quizUC usecase.QuizUseCase
}

func NewHandler(auth *pager.Pager, quc usecase.QuizUseCase) *Handler {
	return &Handler{
		auth,
		quc,
	}
}

func (h *Handler) Register(r *mux.Router) {
	r.HandleFunc("/api/quiz", h.insertQuestion).Methods(http.MethodPost)

	r.Handle("/api/user/history", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.fetchUserHistory)))).Methods(http.MethodGet, http.MethodOptions)
}
