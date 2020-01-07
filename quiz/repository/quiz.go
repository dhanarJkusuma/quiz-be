package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

type dbQuizRepository struct {
	dbConn      *sql.DB
	cacheClient *redis.Client
}

var (
	ErrInvalidAnswer = errors.New("error invalid answer")
	ErrUserNotInGame = errors.New("error not in game")
	ErrUserInGame    = errors.New("error user in game")
)

const (
	InitialScore = 0
)

func NewQuizRepository(dbConn *sql.DB, cacheClient *redis.Client) QuizRepository {
	return &dbQuizRepository{
		dbConn,
		cacheClient,
	}
}

const (
	insertQuizQuery       = `INSERT INTO quiz(question) VALUES (?)`
	insertQuizAnswerQuery = `INSERT INTO answer(quiz_id, answer, correct_answer) VALUES (?,?,?)`
	insertTxnQuizQuery    = `INSERT INTO txn_quiz(user_id, start_time) VALUES (?,?)`

	fetchQuizAnswerQuery = `
	SELECT 
		q.id AS quiz_id, 
		q.question AS question,
		a.id AS answer_id,
		a.answer AS answer, 
		a.correct_answer AS correct_answer
	FROM answer a 
	JOIN (
		SELECT 
			id, 
			question 
		FROM quiz ORDER BY RAND() LIMIT ?
	) q ON a.quiz_id=q.id ORDER BY q.id`

	validateAnswerQuery = `
		SELECT 
			count(1) as count_data
		FROM quiz q
		JOIN answer a ON q.id = a.quiz_id
		WHERE a.correct_answer = 1 AND q.id = ? AND a.id = ?
	`

	insertUserScoreHistory = `
		INSERT INTO user_history(
			user_id_p1, 
			user_id_p2, 
			score_p1,
			score_p2
		) VALUES (?,?,?,?)
	`

	getUserHistoryQuery = `
		SELECT 
			h.id,
			h.user_id_p1,
			h.user_id_p2,
			u1.username AS username_p1,
			u2.username AS username_p2,
			h.score_p1,
			h.score_p2,
			h.created_at
		FROM user_history h
		JOIN rbac_user u1 ON h.user_id_p1 = u1.id 
		JOIN rbac_user u2 ON h.user_id_p2 = u2.id 
		WHERE 
			h.user_id_p1 = ? OR h.user_id_p2 = ?
		ORDER BY h.created_at DESC LIMIT ? OFFSET ? 
	`
)

func (db *dbQuizRepository) InsertQuiz(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error) {
	var result sql.Result
	var err error
	var dbTx *sql.Tx

	dbTx, err = db.dbConn.Begin()
	if err != nil {
		return nil, err
	}

	defer dbTx.Rollback()

	result, err = dbTx.ExecContext(ctx, insertQuizQuery, quiz.Question)
	if err != nil {
		return nil, err
	}

	quiz.ID, _ = result.LastInsertId()

	for i := range quiz.Answers {
		result, err = dbTx.ExecContext(
			ctx,
			insertQuizAnswerQuery,
			quiz.ID,
			quiz.Answers[i].Answer,
			quiz.Answers[i].IsCorrect,
		)
		if err != nil {
			return nil, err
		}
		quiz.Answers[i].ID, _ = result.LastInsertId()
	}
	err = dbTx.Commit()
	if err != nil {
		return nil, err
	}
	return quiz, nil
}

func (db *dbQuizRepository) GetQuiz(ctx context.Context, quizID int64) (*entity.Quiz, error) {

	return nil, nil
}

func (db *dbQuizRepository) FetchQuiz() []entity.Quiz {
	return nil
}

func (db *dbQuizRepository) RandomFetchQuiz(ctx context.Context, total int) ([]entity.Quiz, error) {
	rows, err := db.dbConn.QueryContext(ctx, fetchQuizAnswerQuery, total)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previousQuizID int64
	var previousQuiz *entity.Quiz
	result := make([]entity.Quiz, 0)

	for rows.Next() {
		var quizAnswer entity.QuizAnswerRaw
		err = rows.Scan(
			&quizAnswer.QuestionID,
			&quizAnswer.Question,
			&quizAnswer.AnswerID,
			&quizAnswer.Answer,
			&quizAnswer.IsCorrect,
		)
		if err != nil {
			return nil, err
		}

		if previousQuizID != quizAnswer.QuestionID {
			if previousQuiz != nil {
				result = append(result, *previousQuiz)
			}

			answers := []entity.QuizAnswer{
				{
					ID:        quizAnswer.AnswerID,
					QuizID:    quizAnswer.QuestionID,
					Answer:    quizAnswer.Answer,
					IsCorrect: quizAnswer.IsCorrect,
				},
			}
			previousQuiz = &entity.Quiz{
				ID:       quizAnswer.QuestionID,
				Question: quizAnswer.Question,
				Answers:  answers,
			}
			previousQuizID = previousQuiz.ID
		} else {
			if previousQuiz != nil {
				answers := append(previousQuiz.Answers, entity.QuizAnswer{
					ID:        quizAnswer.AnswerID,
					QuizID:    quizAnswer.QuestionID,
					Answer:    quizAnswer.Answer,
					IsCorrect: quizAnswer.IsCorrect,
				})
				previousQuiz.Answers = answers
			}
		}
	}

	if !rows.Next() && previousQuiz != nil {
		result = append(result, *previousQuiz)
	}

	return result, nil
}

func (db *dbQuizRepository) GetCachedScore(userId int64) (int, error) {
	keyUserID := strconv.FormatInt(userId, 10)
	score, err := db.cacheClient.Get(keyUserID).Int()
	if err != nil {
		return 0, err
	}
	return score, nil
}

func (db *dbQuizRepository) UpdateCachedScore(userId int64, score int) error {
	keyUserID := strconv.FormatInt(userId, 10)
	return db.cacheClient.Set(keyUserID, score, 30*time.Minute).Err()
}

func (db *dbQuizRepository) InsertTxnQuiz(ctx context.Context, userId int64, start time.Time) error {
	_, err := db.dbConn.ExecContext(ctx, insertTxnQuizQuery, userId, start)
	return err
}

func (db *dbQuizRepository) ValidateAnswer(ctx context.Context, userID, questionID, answerID int64, delta int) (*entity.ScoreData, error) {
	// get cached score
	score, err := db.GetCachedScore(userID)
	if err != nil {
		if err == redis.Nil {
			return nil, ErrUserNotInGame
		}
	}

	var countData entity.CountData
	row := db.dbConn.QueryRowContext(ctx, validateAnswerQuery, questionID, answerID)
	err = row.Scan(&countData.Count)
	if err != nil {
		return nil, err
	}

	if countData.Count == 0 {
		return nil, ErrInvalidAnswer
	}

	currentScore := score + delta
	err = db.UpdateCachedScore(userID, currentScore)
	if err != nil {
		return nil, err
	}

	return &entity.ScoreData{
		Score:      currentScore,
		DeltaScore: delta,
	}, nil
}

func (db *dbQuizRepository) SetUserInGame(ctx context.Context, userID int64, inGame bool) error {
	keyUserID := strconv.FormatInt(userID, 10)
	_, err := db.cacheClient.Get(keyUserID).Int()
	switch err {
	case redis.Nil:
		if inGame {
			return db.cacheClient.Set(keyUserID, InitialScore, 30*time.Minute).Err()
		} else {
			return ErrUserNotInGame
		}
	case nil:
		if inGame {
			return ErrUserInGame
		} else {
			return db.cacheClient.Del(keyUserID).Err()
		}

	default:
		return err
	}
}

func (db *dbQuizRepository) InsertUserScoreHistory(ctx context.Context, userIDP1, userIDP2 int64) (*entity.SummaryScoreData, error) {
	var err error
	var scoreP1, scoreP2 int
	scoreP1, err = db.cacheClient.Get(strconv.FormatInt(userIDP1, 10)).Int()
	if err != nil {
		return nil, err
	}

	scoreP2, err = db.cacheClient.Get(strconv.FormatInt(userIDP2, 10)).Int()
	if err != nil {
		return nil, err
	}

	_, err = db.dbConn.ExecContext(ctx, insertUserScoreHistory, userIDP1, userIDP2, scoreP1, scoreP2)
	if err != nil {
		return nil, err
	}
	return &entity.SummaryScoreData{
		PlayerOne: userIDP1,
		PlayerTwo: userIDP2,
		ScoreP1:   scoreP1,
		ScoreP2:   scoreP2,
	}, nil
}

func (db *dbQuizRepository) GetUserHistory(ctx context.Context, userIDP1, page, size int64) ([]entity.UserHistory, error) {
	var err error
	offset := page * size
	rows, err := db.dbConn.QueryContext(ctx, getUserHistoryQuery, userIDP1, userIDP1, size, offset)
	if err != nil {
		return nil, err
	}

	result := make([]entity.UserHistory, 0)
	for rows.Next() {
		var history entity.UserHistory
		err = rows.Scan(
			&history.ID,
			&history.UserIDP1,
			&history.UserIDP2,
			&history.UserNameP1,
			&history.UserNameP2,
			&history.ScoreP1,
			&history.ScoreP2,
			&history.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, history)
	}
	return result, nil
}
