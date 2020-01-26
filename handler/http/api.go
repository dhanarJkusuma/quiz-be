package http

import (
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/config"
	"github.com/dhanarJkusuma/quiz/middleware"
	"github.com/dhanarJkusuma/quiz/quiz/usecase"
	"github.com/dhanarJkusuma/quiz/util"
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
	auth            *pager.Pager
	config          *config.Config
	templateHandler *util.TemplateHandler
	quizUC          usecase.QuizUseCase
}

type HandlerOptions struct {
	Config       *config.Config
	Auth         *pager.Pager
	QuizUC       usecase.QuizUseCase
	TemplatePath string
}

func NewHandler(opts *HandlerOptions) *Handler {
	templateHandler := util.NewTemplateHandler(opts.TemplatePath)
	return &Handler{
		opts.Auth,
		opts.Config,
		templateHandler,
		opts.QuizUC,
	}
}

func (h *Handler) Register(r *mux.Router) {
	r.HandleFunc("/api/quiz", h.insertQuestion).Methods(http.MethodPost)

	r.Handle("/api/user/login", middleware.HandleCORS(http.HandlerFunc(h.LoginHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/register", middleware.HandleCORS(http.HandlerFunc(h.RegisterHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/history", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.fetchUserHistory)))).Methods(http.MethodGet, http.MethodOptions)
	r.Handle("/api/user/logout", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.logoutHandler)))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/verify", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.verifyUser)))).Methods(http.MethodPost, http.MethodOptions)

	r.HandleFunc("/admin/dashboard", h.handleAdminDashboard).Methods(http.MethodGet)
	r.HandleFunc("/admin/quiz", h.handleQuizDashboard).Methods(http.MethodGet)
}
