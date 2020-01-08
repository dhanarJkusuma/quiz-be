package http

import (
	"encoding/json"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/dhanarJkusuma/quiz/util"
	"net/http"
)

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defaultResponse := entity.StandardResponse{}
	var request entity.SignInRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		defaultResponse.Message = "invalid request body"
		util.HandleJson(w, http.StatusBadRequest, defaultResponse)
		return
	}
	user, token, err := h.auth.Auth.SignIn(pager.LoginParams{
		Identifier: request.Email,
		Password:   request.Password,
	})
	if err != nil {
		switch err {
		case pager.ErrInvalidUserLogin:
			defaultResponse.Message = "invalid user"
		case pager.ErrInvalidPasswordLogin:
			defaultResponse.Message = "invalid password"
		case pager.ErrUserNotActive:
			defaultResponse.Message = "user is not active"
		default:
			defaultResponse.Message = "internal server error"
			util.HandleJson(w, http.StatusInternalServerError, defaultResponse)
			return
		}
		util.HandleJson(w, http.StatusUnauthorized, defaultResponse)
		return
	}
	response := struct {
		User  *pager.User `json:"user"`
		Token string      `json:"token"`
	}{
		User:  user,
		Token: token,
	}

	util.HandleJson(w, http.StatusOK, response)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	defaultResponse := entity.StandardResponse{}
	var request entity.SignUpRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		defaultResponse.Message = "invalid request body"
		util.HandleJson(w, http.StatusBadRequest, defaultResponse)
		return
	}

	if request.Password != request.PasswordConfirmation {
		defaultResponse.Message = "password confirmation is invalid"
		util.HandleJson(w, http.StatusBadRequest, defaultResponse)
		return
	}

	err = h.auth.Auth.Register(&pager.User{
		Username: request.Username,
		Password: request.Password,
		Email:    request.Email,
		Active:   true,
	})
	if err != nil {
		defaultResponse.Message = "internal server error"
		util.HandleJson(w, http.StatusInternalServerError, defaultResponse)
		return
	}

	defaultResponse.Message = "user registered successfully"
	util.HandleJson(w, http.StatusOK, defaultResponse)
}

func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	defaultResponse := entity.StandardResponse{}
	var request entity.SignUpRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		defaultResponse.Message = "invalid request body"
		util.HandleJson(w, http.StatusBadRequest, defaultResponse)
		return
	}

	user := pager.GetUserLogin(r)
	if user == nil {
		defaultResponse.Message = "user not found"
		util.HandleJson(w, http.StatusUnauthorized, defaultResponse)
		return
	}

	err = h.auth.Auth.Logout(r)
	if err != nil {
		defaultResponse.Message = "internal server error"
		util.HandleJson(w, http.StatusInternalServerError, defaultResponse)
		return
	}

	defaultResponse.Message = "user logout successfully"
	util.HandleJson(w, http.StatusOK, defaultResponse)
}

func (h *Handler) verifyUser(w http.ResponseWriter, r *http.Request) {
	defaultResponse := entity.StandardResponse{}

	user := pager.GetUserLogin(r)
	if user == nil {
		defaultResponse.Message = "user not found"
		util.HandleJson(w, http.StatusUnauthorized, defaultResponse)
		return
	}

	util.HandleJson(w, http.StatusOK, user)
}