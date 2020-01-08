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

	r.Handle("/api/user/login", middleware.HandleCORS(http.HandlerFunc(h.LoginHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/register", middleware.HandleCORS(http.HandlerFunc(h.RegisterHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/history", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.fetchUserHistory)))).Methods(http.MethodGet, http.MethodOptions)
	r.Handle("/api/user/logout", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.logoutHandler)))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/verify", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.verifyUser)))).Methods(http.MethodPost, http.MethodOptions)
}
