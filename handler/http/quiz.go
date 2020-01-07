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

func (h *Handler) insertQuestion(w http.ResponseWriter, r *http.Request) {
	var request entity.Quiz
	var err error
	response := entity.StandardResponse{}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.Message = "failed to parse request body"
		util.HandleJson(w, http.StatusBadRequest, response)
		return
	}

	ctx := r.Context()
	data, err := h.quizUC.DoInsertQuiz(ctx, &request)
	if err != nil {
		fmt.Println("error while do insert query, err =", err.Error())
		response.Message = "internal server error"
		util.HandleJson(w, http.StatusInternalServerError, response)
		return
	}

	response.Data = data
	util.HandleJson(w, http.StatusOK, response)
	return
}

func (h *Handler) fetchUserHistory(w http.ResponseWriter, r *http.Request) {
	user := pager.GetUserLogin(r)
	query := r.URL.Query()
	pageRaw := query.Get("page")
	sizeRaw := query.Get("size")
	if pageRaw == "" || sizeRaw == "" {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "query params `page` and `size` are required.",
		})
		return
	}

	page, err := strconv.ParseInt(pageRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "invalid params `page`.",
		})
		return
	}

	size, err := strconv.ParseInt(sizeRaw, 10, 64)
	if err != nil {
		util.HandleJson(w, http.StatusBadRequest, entity.StandardResponse{
			Message: "invalid params `page`.",
		})
		return
	}

	ctx := r.Context()
	result, err := h.quizUC.GetUserHistory(ctx, user.ID, page, size)
	if err != nil {
		util.HandleJson(w, http.StatusInternalServerError, entity.StandardResponse{
			Message: "internal server error",
		})
		return
	}
	util.HandleJson(w, http.StatusOK, result)
	return
}
