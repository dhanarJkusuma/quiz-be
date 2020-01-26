package http

import (
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/dhanarJkusuma/quiz/util"
	"net/http"
)

func (h *Handler) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := entity.BaseAdminData{
		BaseUrl:    h.config.BaseUrl,
		ActiveMenu: "main-dashboard-menu",
	}
	err := h.templateHandler.ServeTemplate(w, "index.gohtml", data)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) handleQuizDashboard(w http.ResponseWriter, r *http.Request) {
	data := entity.BaseAdminData{
		BaseUrl:    h.config.BaseUrl,
		ActiveMenu: "question-menu",
	}
	err := h.templateHandler.ServeTemplate(w, "quiz.gohtml", data)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) handleAjaxQuiz(w http.ResponseWriter, r *http.Request) {

}