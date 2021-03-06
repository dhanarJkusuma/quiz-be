package entity

import (
	"github.com/dgrijalva/jwt-go"
	socketio "github.com/googollee/go-socket.io"
)

type StandardResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`

	// for datatable only
	RecordsTotal    int64 `json:"recordsTotal,omitempty"`
	RecordsFiltered int64 `json:"recordsFiltered,omitempty"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpRequest struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type LogoutRequest struct {
	UserID int64 `json:"user_id"`
}

type AnswerSocket struct {
	SocketToken string `json:"token"`
	AnswerID    int64  `json:"answer_id"`
	UserID      int64  `json:"user_id"`
	Delta       int    `json:"delta"`
}

type QuestionSocket struct {
	SocketToken string         `json:"token"`
	Question    string         `json:"question"`
	Options     []OptionSocket `json:"options"`
	CountDown   int            `json:"count_down"`
}

type OptionSocket struct {
	Answer    string `json:"answer"`
	AnswerID  int64  `json:"answer_id"`
	IsCorrect bool   `json:"is_correct"`
}

type InitiateSocket struct {
	PlayerOneID int64  `json:"player_1_id"`
	PlayerTwoID int64  `json:"player_2_id"`
	PlayerOne   string `json:"player_1"`
	PlayerTwo   string `json:"player_2"`
}

type PlayerSocket struct {
	Player     int64         `json:"player"`
	Connection socketio.Conn `json:"connection"`
}

type ScoreSocket struct {
	PlayerID int64 `json:"player_id"`
	Score    int   `json:"score"`
}

type SummaryScoreSocket struct {
	PlayerOneID int64 `json:"player_1_id"`
	PlayerTwoID int64 `json:"player_2_id"`
	P1Score     int   `json:"player_1_score"`
	P2Score     int   `json:"player_2_score"`
}

type QuestionSocketClaims struct {
	jwt.Claims
	RoomID     string `json:"room_id"`
	GenerateAt string `json:"generate_at"`
	QuestionID int64  `json:"question_id"`
}

/* Entity for Admin Dashboard */
type BaseAdminData struct {
	BaseUrl       string      `json:"base_url"`
	ActiveSideBar string      `json:"active_sidebar"`
	ActiveMenu    string      `json:"active_menu"`
	AdminData     interface{} `json:"data"`
}

type QuestionStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type QuestionUpdateRequest struct {
	Question string `json:"question"`
}

type AnswerUpdateRequest struct {
	Answer  string `json:"answer"`
	Correct bool   `json:"correct"`
}
