package repository

import (
	"context"
	"github.com/dhanarJkusuma/quiz/entity"
	"time"
)

type QuizRepository interface {
	InsertQuiz(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error)
	GetQuiz(ctx context.Context, quizID int64) (*entity.Quiz, error)
	FetchQuiz(ctx context.Context, search string, start, size int64) ([]entity.QuizDashboard, int64, error)
	RandomFetchQuiz(ctx context.Context, total int) ([]entity.Quiz, error)
	ValidateAnswer(ctx context.Context, roomID string, userID, questionID, answerID int64, delta int) (*entity.ScoreData, error)

	GetCachedScore(roomID string, userId int64) (int, error)
	UpdateCachedScore(roomID string, userId int64, score int) error

	InsertTxnQuiz(ctx context.Context, userId int64, start time.Time) error
	SetUserInGame(ctx context.Context, roomID string, userID int64, inGame bool) error
	InsertUserScoreHistory(ctx context.Context, roomID string, userIDP1, userIDP2 int64) (*entity.SummaryScoreData, error)
	GetUserHistory(ctx context.Context, userIDP1, page, size int64) ([]entity.UserHistory, error)

	SetQuestionStatus(ctx context.Context, questionId int64, enabled bool) error
	UpdateQuestion(ctx context.Context, questionId int64, question string) error
	DeleteQuestion(ctx context.Context, questionID int64) error
	UpdateAnswer(ctx context.Context, answerID int64, answer string, correct bool) error
	DeleteAnswer(ctx context.Context, answerID int64) error
}
