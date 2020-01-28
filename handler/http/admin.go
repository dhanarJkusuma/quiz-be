package http

import (
	"encoding/json"
	"fmt"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/dhanarJkusuma/quiz/util"
	"net/http"
	"strconv"
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

func (h *Handler) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	err := h.templateHandler.ServeTemplate(w, "login.gohtml", nil)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) handleAdminLoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/login.html")
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	_, err := h.auth.Auth.SignInWithCookie(w, pager.LoginParams{
		Identifier: email,
		Password:   password,
	})
	if err != nil {
		switch err {
		case pager.ErrInvalidUserLogin:
			//TODO::setSessionErr
			http.Redirect(w, r, "/admin/login", 302)
			return
		default:
			//TODO::setSessionErr
			http.Redirect(w, r, "/admin/login", 302)
		}
	}

	http.Redirect(w, r, "/admin/quiz", 302)
	return
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
	queryParams := r.URL.Query()
	searchParams := queryParams.Get("search[value]")
	//offset
	startRaw := queryParams.Get("start")
	//size
	lengthRaw := queryParams.Get("length")

	start, err := strconv.ParseInt(startRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid start params",
		})
		return
	}

	length, err := strconv.ParseInt(lengthRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid start params",
		})
		return
	}

	ctx := r.Context()
	questions, count, err := h.quizUC.FetchQuestionDashboard(
		ctx,
		searchParams,
		start,
		length)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}

	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Data:            questions,
		RecordsTotal:    count,
		RecordsFiltered: count,
	})
	return
}

func (h *Handler) handleAjaxDetailQuiz(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	questionIDraw := queryParams.Get("question_id")

	questionID, err := strconv.ParseInt(questionIDraw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `question_id`",
		})
		return
	}

	ctx := r.Context()
	question, err := h.quizUC.GetQuestionDetailDashboard(ctx, questionID)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	if question == nil {
		util.HandleJson(w, http.StatusNotFound, entity.StandardResponse{
			Message: "Question not found",
		})
		return
	}

	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Data: question,
	})
	return
}

func (h *Handler) handleAjaxToggleQuiz(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	questionIDraw := queryParams.Get("question_id")
	questionID, err := strconv.ParseInt(questionIDraw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `question_id`",
		})
		return
	}

	var request entity.QuestionStatusRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid request",
		})
		return
	}

	ctx := r.Context()
	err = h.quizUC.SetQuestionStatus(ctx, questionID, request.Enabled)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	var statusString string
	if request.Enabled {
		statusString = "enabled"
	} else {
		statusString = "disabled"
	}
	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Message: fmt.Sprintf("Question is successfully updated with status %s", statusString),
	})
	return
}

func (h *Handler) handleAjaxUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	questionIDraw := queryParams.Get("question_id")
	questionID, err := strconv.ParseInt(questionIDraw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `question_id`",
		})
		return
	}

	var request entity.QuestionUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid request",
		})
		return
	}

	ctx := r.Context()
	err = h.quizUC.SetQuestion(ctx, questionID, request.Question)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Message: "Question is successfully updated",
	})
	return
}

func (h *Handler) handleAjaxDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	questionIDRaw := queryParams.Get("question_id")
	questionID, err := strconv.ParseInt(questionIDRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `question_id`",
		})
		return
	}
	ctx := r.Context()
	err = h.quizUC.DeleteQuestion(ctx, questionID)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Message: "Question is successfully deleted",
	})
	return
}

func (h *Handler) handleAjaxUpdateAnswer(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	answerIDRaw := queryParams.Get("answer_id")
	answerID, err := strconv.ParseInt(answerIDRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `answer_id`",
		})
		return
	}

	var request entity.AnswerUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid request",
		})
		return
	}

	ctx := r.Context()
	err = h.quizUC.UpdateAnswer(ctx, answerID, request.Answer, request.Correct)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Message: "Answer is successfully updated",
	})
	return
}

func (h *Handler) handleAjaxDeleteAnswer(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	answerIDRaw := queryParams.Get("answer_id")
	answerID, err := strconv.ParseInt(answerIDRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "Invalid params `answer_id`",
		})
		return
	}
	ctx := r.Context()
	err = h.quizUC.DeleteAnswer(ctx, answerID)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "Internal Server Error",
		})
		return
	}
	util.HandleJson(w, http.StatusOK, entity.StandardResponse{
		Message: "Answer is successfully deleted",
	})
	return
}
