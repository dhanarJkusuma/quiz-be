package usecase

import (
	"context"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/entity"
)

type QuizUseCase interface {
	DoInsertQuiz(ctx context.Context, data *entity.Quiz) (*entity.Quiz, error)
	AnswerQuiz(ctx context.Context, quizID, answerID int64) error
	DoInitQuiz(ctx context.Context, roomID string, userID int64) error

	GetUserEnemy(ctx context.Context, userID int64) (*pager.User, error)
	GetRandomQuiz(ctx context.Context, total int) ([]entity.Quiz, error)
	ValidateAnswer(ctx context.Context, roomID string, userID, questionID, answerID int64, delta int) (*entity.ScoreData, error)

	SetUserInGame(ctx context.Context, roomID string, userID int64, inGame bool) error
	InsertUserScoreHistory(ctx context.Context, roomID string, userIDP1, userIDP2 int64) (*entity.SummaryScoreData, error)
	GetUserHistory(ctx context.Context, userID, page, size int64) ([]entity.UserHistorySummary, error)

	FetchQuestionDashboard(ctx context.Context, query string, start, length int64) ([]entity.QuizDashboard, int64, error)
	GetQuestionDetailDashboard(ctx context.Context, questionID int64) (*entity.Quiz, error)
	SetQuestionStatus(ctx context.Context, questionID int64, enabled bool) error
	SetQuestion(ctx context.Context, questionID int64, question string) error
	DeleteQuestion(ctx context.Context, questionID int64) error

	UpdateAnswer(ctx context.Context, answerID int64, answer string, correct bool) error
	DeleteAnswer(ctx context.Context, answerID int64) error
}
