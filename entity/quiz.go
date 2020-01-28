package entity

import "time"

type Quiz struct {
	ID       int64  `db:"id" json:"id"`
	Question string `db:"question" json:"question"`

	Answers  []QuizAnswer `json:"answers"`
	IsActive bool         `db:"active" json:"active"`
}

type QuizDashboard struct {
	Checkbox string `json:"checkbox"`
	No       int64  `json:"no"`
	ID       int64  `db:"id" json:"id"`
	Question string `db:"question" json:"question"`
	IsActive bool   `db:"active" json:"-"`
	Status   string `json:"status"`
}

type QuizAnswer struct {
	ID        int64  `db:"id" json:"answer_id"`
	QuizID    int64  `db:"quiz_id" json:"-"`
	Answer    string `db:"answer" json:"answer"`
	IsCorrect bool   `db:"correct_answer" json:"correct"`
}

type QuizAnswerRaw struct {
	QuestionID int64  `db:"question_id"`
	Question   string `db:"question"`
	Active     bool   `db:"active"`
	AnswerID   int64  `db:"answer_id"`
	Answer     string `db:"answer"`
	IsCorrect  bool   `db:"correct_answer"`
}

type UserHistory struct {
	ID         int64     `db:"id"`
	UserIDP1   int64     `db:"user_id_p1"`
	UserIDP2   int64     `db:"user_id_p2"`
	UserNameP1 string    `db:"username_p1"`
	UserNameP2 string    `db:"username_p2"`
	ScoreP1    int       `db:"score_p1"`
	ScoreP2    int       `db:"score_p2"`
	CreatedAt  time.Time `db:"created_at"`
}

type CountData struct {
	Count int64 `db:"count_data"`
}

type ScoreData struct {
	Score      int `json:"score"`
	DeltaScore int `json:"delta_score"`
}

type SummaryScoreData struct {
	PlayerOne int64 `json:"player_one"`
	PlayerTwo int64 `json:"player_two"`
	ScoreP1   int   `json:"score_player_one"`
	ScoreP2   int   `json:"score_player_two"`
}

type UserHistorySummary struct {
	Enemy      string `json:"enemy_name"`
	Score      int    `json:"score"`
	EnemyScore int    `json:"enemy_score"`
	Date       string `json:"date"`
}
