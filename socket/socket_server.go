package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/config"
	"github.com/dhanarJkusuma/quiz/entity"
	"github.com/dhanarJkusuma/quiz/quiz/usecase"
	"github.com/dhanarJkusuma/quiz/util"
	socketio "github.com/googollee/go-socket.io"
	uuid "github.com/satori/go.uuid"

	//"log"
	"net/http"
	"sync"
	"time"
)

const (
	TAG_DOQUIZ_INIT             = "do_quiz:initiation"
	TAG_DOQUIZ_COUNTER          = "do_quiz:counter"
	TAG_DOQUIZ_PUBLISH_QUESTION = "do_quiz:publish_question"
	TAG_DOQUIZ_COUNT_DOWN       = "do_quiz:count_down"
	TAG_DOQUIZ_NEXT_QUESTION    = "do_quiz:next_question"
	TAG_DOQUIZ_FINISH_QUESTION  = "do_quiz:finish_question"
	TAG_DOQUIZ_UPDATE_SCORE     = "do_quiz:update_score"
)

// SocketCrossOriginServer represent socket connection
type SocketCrossOriginServer struct {
	config             *config.Config
	answerValidator    *util.JWTValidator
	IoServer           *socketio.Server
	Auth               *pager.Auth
	QuizUC             usecase.QuizUseCase
	mutex              sync.RWMutex
	userPoolConnection []entity.PlayerSocket
}

func New(cfg *config.Config, server *socketio.Server, auth *pager.Auth, quc usecase.QuizUseCase) *SocketCrossOriginServer {
	jwtValidator := util.NewValidator(jwt.SigningMethodHS256, cfg.JWT.SecretKey)
	s := &SocketCrossOriginServer{
		config:          cfg,
		answerValidator: jwtValidator,
		IoServer:        server,
		QuizUC:          quc,
		Auth:            auth,
	}
	s.initPool()
	return s
}

func (s *SocketCrossOriginServer) initPool() {
	s.mutex.Lock()
	s.userPoolConnection = make([]entity.PlayerSocket, 0)
	s.mutex.Unlock()
}

func (s *SocketCrossOriginServer) addUserPool(player entity.PlayerSocket) {
	s.mutex.Lock()
	s.userPoolConnection = append(s.userPoolConnection, player)
	s.mutex.Unlock()
}

func (s *SocketCrossOriginServer) DoMatchMaking() {
	var roomID string
	var playerOne, playerTwo entity.PlayerSocket
	for {
		time.Sleep(1 * time.Second)

		s.mutex.RLock()
		fmt.Println(s.userPoolConnection)
		lenUsersPool := len(s.userPoolConnection)
		iteration := lenUsersPool / 2
		for k := 0; k < iteration; k++ {
			playerOne = s.userPoolConnection[k]
			playerTwo = s.userPoolConnection[iteration+k]

			roomID = uuid.NewV4().String()

			// join room
			playerOne.Connection.Join(roomID)
			playerTwo.Connection.Join(roomID)

			// broadcast question
			fmt.Println("do matchmaking")
			go s.DoQuiz(roomID, playerOne, playerTwo)
		}
		s.mutex.RUnlock()

		if lenUsersPool > 0 && lenUsersPool%2 != 0 {
			// set remaining user to user pool
			remainingPlayer := s.userPoolConnection[len(s.userPoolConnection)-1]

			s.initPool()
			s.addUserPool(remainingPlayer)
		} else {
			s.initPool()
		}
	}
}

// DoQuiz func to broadcast question to user login
func (s *SocketCrossOriginServer) DoQuiz(roomID string, playerOne, playerTwo entity.PlayerSocket) {
	var err error
	ctx := context.Background()

	err = s.QuizUC.SetUserInGame(ctx, playerOne.Player, true)
	if err != nil {
		return
	}

	err = s.QuizUC.SetUserInGame(ctx, playerTwo.Player, true)
	if err != nil {
		return
	}

	questions, err := s.QuizUC.GetRandomQuiz(ctx, s.config.Quiz.NumberOfQuestion)
	if err != nil {
		// TODO :: Handle error here, notify the users that the quiz should be terminated
		return
	}

	p1, err := s.QuizUC.GetUserEnemy(ctx, playerOne.Player)
	if err != nil || p1 == nil {
		return
	}

	p2, err := s.QuizUC.GetUserEnemy(ctx, playerTwo.Player)
	if err != nil || p2 == nil {
		return
	}

	// notify the users to ready before quiz is started
	initiateData := entity.InitiateSocket{
		PlayerOneID: p1.ID,
		PlayerTwoID: p2.ID,
		PlayerOne:   p1.Username,
		PlayerTwo:   p2.Username,
	}
	initiateMsg, _ := json.Marshal(initiateData)
	s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_INIT, string(initiateMsg))
	setInterval(s.config.Quiz.ReadyCountDown*time.Second, func(count int64) {
		if count > 0 {
			s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_COUNTER, count)
		} else {
			s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_COUNTER, "Go !")
		}
	})

	for i := range questions {
		question := questions[i]

		generateAt := time.Now().Format(time.RFC3339)
		claims := entity.QuestionSocketClaims{
			RoomID:     roomID,
			QuestionID: question.ID,
			GenerateAt: generateAt,
		}
		jwtToken, err := util.
			NewBuilder(
				jwt.SigningMethodHS256,
				s.config.JWT.SecretKey,
			).
			SetClaims(claims).
			GenerateToken()
		if err != nil {
			return
		}

		questionMsg := entity.QuestionSocket{
			Question:    question.Question,
			SocketToken: *jwtToken,
			CountDown:   int(s.config.Quiz.CountDown),
		}
		options := make([]entity.OptionSocket, 0)
		for j := range question.Answers {
			options = append(options, entity.OptionSocket{
				AnswerID:  question.Answers[j].ID,
				Answer:    question.Answers[j].Answer,
				IsCorrect: question.Answers[j].IsCorrect,
			})
		}
		questionMsg.Options = options
		mb, _ := json.Marshal(questionMsg)
		s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_PUBLISH_QUESTION, string(mb))
		//time.Sleep(1 * time.Second)
		setInterval((s.config.Quiz.CountDown-1)*time.Second, func(count int64) {
			if count >= 0 {
				s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_COUNT_DOWN, count)
			}
		})
		s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_NEXT_QUESTION, string(mb))
		time.Sleep(1 * time.Second)
	}
	// calculate score
	scoreSummary, err := s.QuizUC.InsertUserScoreHistory(ctx, playerOne.Player, playerTwo.Player)
	if err != nil {
		return
	}

	// calculate score
	scoreSummarySocketMsg := entity.SummaryScoreSocket{
		PlayerOneID: scoreSummary.PlayerOne,
		PlayerTwoID: scoreSummary.PlayerTwo,
		P1Score:     scoreSummary.ScoreP1,
		P2Score:     scoreSummary.ScoreP2,
	}
	scoreSummaryMsg, _ := json.Marshal(scoreSummarySocketMsg)
	s.IoServer.BroadcastToRoom(roomID, TAG_DOQUIZ_FINISH_QUESTION, string(scoreSummaryMsg))

	err = s.QuizUC.SetUserInGame(ctx, playerOne.Player, false)
	if err != nil {
		return
	}

	err = s.QuizUC.SetUserInGame(ctx, playerTwo.Player, false)
	if err != nil {
		return
	}

	// leave all rooms
	s.IoServer.ClearRoom(roomID)

}

func (s *SocketCrossOriginServer) InitSocket() {
	s.IoServer.OnConnect("/", func(s socketio.Conn) error {
		// s.SetContext("")

		fmt.Println("connected:", s.ID())
		s.Emit("onConnected", "you are in the lobby")
		return nil
	})

	// socket for init lobby connection
	s.IoServer.OnEvent("/", "init", func(conn socketio.Conn, msg string) {

		user, err := s.Auth.GetUserByToken(msg)
		if err != nil {
			conn.Emit("init", "have "+msg)
			return
		}
		go s.addUserPool(entity.PlayerSocket{
			Player:     user.ID,
			Connection: conn,
		})
		conn.Emit("onInitQuiz", "you have successfully init the quiz")
	})

	// socket for answer the question
	s.IoServer.OnEvent("/", "answer", func(conn socketio.Conn, msg string) {
		var requestData entity.AnswerSocket
		err := json.Unmarshal([]byte(msg), &requestData)
		if err != nil {
			return
		}

		claims, err := s.answerValidator.ValidateToken(requestData.SocketToken)
		if err != nil {

			return
		}

		generatedAtRaw := claims["generate_at"].(string)
		generatedAt, err := time.Parse(time.RFC3339, generatedAtRaw)
		if err != nil {
			return
		}

		deltaTime := time.Now().Sub(generatedAt)
		if deltaTime >= (15 * time.Second) {
			fmt.Println("expired")
			return
		}

		ctx := context.Background()
		questionID := claims["question_id"].(float64)
		scoreData, err := s.QuizUC.ValidateAnswer(
			ctx,
			requestData.UserID,
			int64(questionID),
			requestData.AnswerID,
			requestData.Delta,
		)
		if err != nil {
			return
		}

		roomID := claims["room_id"].(string)
		msgScore := entity.ScoreSocket{
			PlayerID: requestData.UserID,
			Score:    scoreData.Score,
		}
		msgUpdateScore, _ := json.Marshal(msgScore)
		s.IoServer.BroadcastToRoom(
			roomID,
			TAG_DOQUIZ_UPDATE_SCORE,
			string(msgUpdateScore),
		)
	})

	s.IoServer.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})
	s.IoServer.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	s.IoServer.OnError("/", func(e error) {
		fmt.Println("meet error:", e)
	})
	s.IoServer.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
}

func (s *SocketCrossOriginServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		rw.Header().Set("Access-Control-Allow-Methods", "POST, PUT, PATCH, GET, DELETE, OPTIONS")
		rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
	if req.Method == "OPTIONS" {
		return
	}

	s.IoServer.ServeHTTP(rw, req)
}

func setInterval(maxDuration time.Duration, f func(int64)) {
	count := int64(maxDuration.Seconds())
	for range time.Tick(1 * time.Second) {
		if count >= 0 {
			f(count)
			count--
		} else {
			break
		}
	}
}
