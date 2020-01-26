package usecase

import (
	"context"
	"database/sql"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/dhanarJkusuma/quiz/quiz/repository"
	"github.com/go-redis/redis"
)

type QuizOptions struct {
	DbConn      *sql.DB
	CacheClient *redis.Client
}

type quizUseCase struct {
	quizRepo repository.QuizRepository
}

func NewQuizUseCase(opts *QuizOptions) QuizUseCase {
	quizRepository := repository.NewQuizRepository(
		opts.DbConn,
		opts.CacheClient,
	)
	return &quizUseCase{
		quizRepo: quizRepository,
	}
}

func (qu *quizUseCase) DoInsertQuiz(ctx context.Context, data *entity.Quiz) (*entity.Quiz, error) {
	return qu.quizRepo.InsertQuiz(ctx, data)
}

func (qu *quizUseCase) AnswerQuiz(ctx context.Context, quizID, answerID int64) error {
	return nil
}

func (qu *quizUseCase) DoInitQuiz(ctx context.Context, roomID string, userID int64) error {
	return qu.quizRepo.UpdateCachedScore(roomID, userID, 0)
}

func (qu *quizUseCase) GetRandomQuiz(ctx context.Context, total int) ([]entity.Quiz, error) {
	return qu.quizRepo.RandomFetchQuiz(ctx, total)
}

func (qu *quizUseCase) ValidateAnswer(ctx context.Context, roomID string, userID, questionID, answerID int64, delta int) (*entity.ScoreData, error) {
	return qu.quizRepo.ValidateAnswer(ctx, roomID, userID, questionID, answerID, delta)
}

func (qu *quizUseCase) GetUserEnemy(ctx context.Context, userID int64) (*pager.User, error) {
	return pager.FindUser(map[string]interface{}{
		"id": userID,
	}, nil)
}

func (qu *quizUseCase) SetUserInGame(ctx context.Context, roomID string, userID int64, inGame bool) error {
	return qu.quizRepo.SetUserInGame(ctx, roomID, userID, inGame)
}

func (qu *quizUseCase) InsertUserScoreHistory(ctx context.Context, roomID string, userIDP1, userIDP2 int64) (*entity.SummaryScoreData, error) {
	return qu.quizRepo.InsertUserScoreHistory(ctx, roomID, userIDP1, userIDP2)
}

func (qu *quizUseCase) GetUserHistory(ctx context.Context, userID, page, size int64) ([]entity.UserHistorySummary, error) {
	result := make([]entity.UserHistorySummary, 0)
	userHistory, err := qu.quizRepo.GetUserHistory(ctx, userID, page, size)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
		return nil, err
	}
	for i := range userHistory {
		var enemy string
		var score, enemyScore int
		if userHistory[i].UserIDP1 == userID {
			enemy = userHistory[i].UserNameP2
			enemyScore = userHistory[i].ScoreP2
			score = userHistory[i].ScoreP1

		} else {
			enemy = userHistory[i].UserNameP1
			enemyScore = userHistory[i].ScoreP1
			score = userHistory[i].ScoreP2
		}

		date := userHistory[i].CreatedAt.Format("02 Jan 06 15:04")

		result = append(result, entity.UserHistorySummary{
			Enemy:      enemy,
			EnemyScore: enemyScore,
			Score:      score,
			Date:       date,
		})
	}
	return result, nil
}
