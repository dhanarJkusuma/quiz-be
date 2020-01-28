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
	Config *config.Config
	Auth   *pager.Pager
	QuizUC usecase.QuizUseCase
}

func NewHandler(opts *HandlerOptions) *Handler {
	templateHandler := util.NewTemplateHandler(opts.Config.Quiz.TemplatePath)
	return &Handler{
		opts.Auth,
		opts.Config,
		templateHandler,
		opts.QuizUC,
	}
}

func (h *Handler) Register(r *mux.Router) {
	r.Handle("/api/user/login", middleware.HandleCORS(http.HandlerFunc(h.LoginHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/register", middleware.HandleCORS(http.HandlerFunc(h.RegisterHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/history", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.fetchUserHistory)))).Methods(http.MethodGet, http.MethodOptions)
	r.Handle("/api/user/logout", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.logoutHandler)))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/api/user/verify", middleware.HandleCORS(h.auth.Auth.ProtectRouteUsingToken(http.HandlerFunc(h.verifyUser)))).Methods(http.MethodPost, http.MethodOptions)

	// r.HandleFunc("/admin/dashboard", h.handleAdminDashboard).Methods(http.MethodGet)
	r.HandleFunc("/admin/login", h.handleAdminLogin).Methods(http.MethodGet)
	r.HandleFunc("/admin/login", h.handleAdminLoginPost).Methods(http.MethodPost)
	r.Handle("/admin/quiz", h.auth.Auth.ProtectRoute(h.auth.Auth.ProtectWithRBAC(http.HandlerFunc(h.handleQuizDashboard)))).Methods(http.MethodGet)

	r.HandleFunc("/api/admin/quiz", h.insertQuestion).Methods(http.MethodPost)
	r.HandleFunc("/api/admin/quiz", h.handleAjaxQuiz).Methods(http.MethodGet)
	r.HandleFunc("/api/admin/quiz/detail", h.handleAjaxDetailQuiz).Methods(http.MethodGet)
	r.HandleFunc("/api/admin/quiz/status", h.handleAjaxToggleQuiz).Methods(http.MethodPut)
	r.HandleFunc("/api/admin/quiz/update", h.handleAjaxUpdateQuestion).Methods(http.MethodPut)
	r.HandleFunc("/api/admin/quiz/delete", h.handleAjaxDeleteQuestion).Methods(http.MethodDelete)
	r.HandleFunc("/api/admin/answer/update", h.handleAjaxUpdateAnswer).Methods(http.MethodPut)
	r.HandleFunc("/api/admin/answer/delete", h.handleAjaxDeleteAnswer).Methods(http.MethodDelete)
}
