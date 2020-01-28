package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	fetchQuizAdminQuery = `
		SELECT 
			id,
			question,
			active
		FROM quiz
		WHERE 
			question LIKE ?
		ORDER BY updated_at DESC 
		LIMIT ? OFFSET ?
	`
	fetchQuizAdminCountQuery = `
		SELECT 
			count(1) as count_data
		FROM quiz
		WHERE 
			question LIKE ?
	`
	fetchQuizAdminDetailQuery = `
		SELECT 
			q.id AS quiz_id, 
			q.question AS question,
			q.active AS active,
			a.id AS answer_id,
			a.answer AS answer, 
			a.correct_answer AS correct_answer
		FROM answer a 
		JOIN quiz q ON a.quiz_id = q.id 
		WHERE q.id = ?
	`

	fetchQuizAnswerQuery = `
		SELECT 
			q.id AS quiz_id, 
			q.question AS question,
			q.active AS active,
			a.id AS answer_id,
			a.answer AS answer, 
			a.correct_answer AS correct_answer
		FROM answer a 
		JOIN (
			SELECT 
				id, 
				question 
			FROM quiz ORDER BY RAND() LIMIT ?
		) q ON a.quiz_id=q.id ORDER BY q.id
	`
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

	updateQuestionStatusQuery = `
		UPDATE quiz
			SET active = ?
		WHERE id = ?
	`
	updateQuestionQuery = `
		UPDATE quiz
			SET question = ?
		WHERE id = ?
	`

	updateAnswerQuery = `
		UPDATE answer 
			SET answer = ?, correct_answer = ? 
		WHERE id = ?
	`

	deleteQuestionQuery = `
		DELETE FROM quiz WHERE id = ?
	`

	deleteAnswerQuery = `
		DELETE FROM answer WHERE id = ?
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
	rows, err := db.dbConn.QueryContext(ctx, fetchQuizAdminDetailQuery, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currentQuizID int64
	var currentQuiz *entity.Quiz
	result := make([]entity.Quiz, 0)

	for rows.Next() {
		var quizAnswer entity.QuizAnswerRaw
		err = rows.Scan(
			&quizAnswer.QuestionID,
			&quizAnswer.Question,
			&quizAnswer.Active,
			&quizAnswer.AnswerID,
			&quizAnswer.Answer,
			&quizAnswer.IsCorrect,
		)
		if err != nil {
			return nil, err
		}

		if currentQuizID != quizAnswer.QuestionID {
			if currentQuiz != nil {
				result = append(result, *currentQuiz)
			}

			answers := []entity.QuizAnswer{
				{
					ID:        quizAnswer.AnswerID,
					QuizID:    quizAnswer.QuestionID,
					Answer:    quizAnswer.Answer,
					IsCorrect: quizAnswer.IsCorrect,
				},
			}
			currentQuiz = &entity.Quiz{
				ID:       quizAnswer.QuestionID,
				Question: quizAnswer.Question,
				IsActive: quizAnswer.Active,
				Answers:  answers,
			}
			currentQuizID = currentQuiz.ID
		} else {
			if currentQuiz != nil {
				answers := append(currentQuiz.Answers, entity.QuizAnswer{
					ID:        quizAnswer.AnswerID,
					QuizID:    quizAnswer.QuestionID,
					Answer:    quizAnswer.Answer,
					IsCorrect: quizAnswer.IsCorrect,
				})
				currentQuiz.Answers = answers
			}
		}
	}

	return currentQuiz, nil
}

func (db *dbQuizRepository) FetchQuiz(ctx context.Context, search string, offset, size int64) ([]entity.QuizDashboard, int64, error) {
	var err error
	tx, err := db.dbConn.Begin()
	if err != nil {
		return nil, 0, err
	}
	search = fmt.Sprintf("%%%s%%", search)
	rows, err := tx.QueryContext(ctx, fetchQuizAdminQuery, search, size, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]entity.QuizDashboard, 0)
	idx := int64(1)
	for rows.Next() {
		var question entity.QuizDashboard
		err = rows.Scan(
			&question.ID,
			&question.Question,
			&question.IsActive,
		)
		if err != nil {
			return nil, 0, err
		}
		if question.IsActive {
			question.Status = "active"
		} else {
			question.Status = "deactivate"
		}

		question.No = idx
		result = append(result, question)
		idx++
	}

	var countData entity.CountData
	row := db.dbConn.QueryRowContext(ctx, fetchQuizAdminCountQuery, search)
	err = row.Scan(&countData.Count)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, 0, err
	}
	return result, countData.Count, nil
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
			&quizAnswer.Active,
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
				IsActive: quizAnswer.Active,
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

func (db *dbQuizRepository) GetCachedScore(roomID string, userId int64) (int, error) {
	keyUserID := strconv.FormatInt(userId, 10)
	cacheKey := fmt.Sprintf("%s:%s", roomID, keyUserID)
	score, err := db.cacheClient.Get(cacheKey).Int()
	if err != nil {
		return 0, err
	}
	return score, nil
}

func (db *dbQuizRepository) UpdateCachedScore(roomId string, userId int64, score int) error {
	keyUserID := strconv.FormatInt(userId, 10)
	cacheKey := fmt.Sprintf("%s:%s", roomId, keyUserID)
	return db.cacheClient.Set(cacheKey, score, 30*time.Minute).Err()
}

func (db *dbQuizRepository) InsertTxnQuiz(ctx context.Context, userId int64, start time.Time) error {
	_, err := db.dbConn.ExecContext(ctx, insertTxnQuizQuery, userId, start)
	return err
}

func (db *dbQuizRepository) ValidateAnswer(ctx context.Context, roomID string, userID, questionID, answerID int64, delta int) (*entity.ScoreData, error) {
	// get cached score
	score, err := db.GetCachedScore(roomID, userID)
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
	err = db.UpdateCachedScore(roomID, userID, currentScore)
	if err != nil {
		return nil, err
	}

	return &entity.ScoreData{
		Score:      currentScore,
		DeltaScore: delta,
	}, nil
}

func (db *dbQuizRepository) SetUserInGame(ctx context.Context, roomID string, userID int64, inGame bool) error {
	keyUserID := strconv.FormatInt(userID, 10)
	keyCache := fmt.Sprintf("%s:%s", roomID, keyUserID)

	_, err := db.cacheClient.Get(keyCache).Int()
	switch err {
	case redis.Nil:
		if inGame {
			return db.cacheClient.Set(keyCache, InitialScore, 30*time.Minute).Err()
		} else {
			return ErrUserNotInGame
		}
	case nil:
		return db.cacheClient.Del(keyUserID).Err()

	default:
		return err
	}
}

func (db *dbQuizRepository) InsertUserScoreHistory(ctx context.Context, roomID string, userIDP1, userIDP2 int64) (*entity.SummaryScoreData, error) {
	var err error
	var scoreP1, scoreP2 int
	cacheP1, cacheP2 := fmt.Sprintf("%s:%s", roomID, strconv.FormatInt(userIDP1, 10)), fmt.Sprintf("%s:%s", roomID, strconv.FormatInt(userIDP2, 10))
	scoreP1, err = db.cacheClient.Get(cacheP1).Int()
	if err != nil {
		return nil, err
	}

	scoreP2, err = db.cacheClient.Get(cacheP2).Int()
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

func (db *dbQuizRepository) SetQuestionStatus(ctx context.Context, questionId int64, enabled bool) error {
	_, err := db.dbConn.ExecContext(ctx, updateQuestionStatusQuery, enabled, questionId)
	return err
}

func (db *dbQuizRepository) UpdateQuestion(ctx context.Context, questionId int64, question string) error {
	_, err := db.dbConn.ExecContext(ctx, updateQuestionQuery, question, questionId)
	return err
}

func (db *dbQuizRepository) DeleteQuestion(ctx context.Context, questionID int64) error {
	_, err := db.dbConn.ExecContext(ctx, deleteQuestionQuery, questionID)
	return err
}

func (db *dbQuizRepository) UpdateAnswer(ctx context.Context, answerID int64, answer string, correct bool) error {
	_, err := db.dbConn.ExecContext(ctx, updateAnswerQuery, answer, correct, answerID)
	return err
}

func (db *dbQuizRepository) DeleteAnswer(ctx context.Context, answerID int64) error {
	_, err := db.dbConn.ExecContext(ctx, deleteAnswerQuery, answerID)
	return err
}
