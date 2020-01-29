package main

import (
	"encoding/json"
	"fmt"
	"github.com/dhanarJkusuma/pager"
	"github.com/dhanarJkusuma/quiz/config"
	"github.com/dhanarJkusuma/quiz/quiz/usecase"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"time"

	"net/http"

	httpHandler "github.com/dhanarJkusuma/quiz/handler/http"

	socketModule "github.com/dhanarJkusuma/quiz/socket"
	socketio "github.com/googollee/go-socket.io"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func readFile(filePath string) (content []byte) {
	config, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func main() {

	// read config
	var mainCfg config.Config
	err := json.Unmarshal(readFile("./env.json"), &mainCfg)
	if err != nil {
		panic(fmt.Sprintf("failed to read config ./env.json, err = %v", err))
	}

	// init db conn
	dbConf := mainCfg.DbConnection
	dbConn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbConf.DbUser, dbConf.DbPassword, dbConf.DbAddress, dbConf.DbName)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// init redis
	cacheClient := redis.NewClient(&redis.Options{
		Addr:     mainCfg.Redis.Address,
		Password: mainCfg.Redis.Password,
	})

	// init auth
	auth := generateAuth(&authOptions{
		db:          db,
		redisClient: cacheClient,
		schema:      dbConf.DbName,
		origin:      mainCfg.BaseUrl,
	})
	err = auth.Migration.InitDBMigration()
	if err != nil {
		panic(err.Error())
	}

	err = auth.Migration.Run(&AdminRoleMigration{
		auth: auth.Auth,
	})
	if err != nil && err != pager.ErrMigrationAlreadyExist {
		panic(err.Error())
	}

	//init modules
	quc := usecase.NewQuizUseCase(&usecase.QuizOptions{
		DbConn:      db,
		CacheClient: cacheClient,
	})

	// init socket server
	server, err := socketio.NewServer(nil)
	if err != nil {
		panic(err.Error())
	}
	s := socketModule.New(
		&mainCfg,
		server,
		auth.Auth,
		quc)

	s.InitSocket()

	go s.DoMatchMaking()
	go s.IoServer.Serve()
	defer s.IoServer.Close()

	// handle router
	r := mux.NewRouter()
	r.Handle("/socket.io/", s)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./templates/"))))

	handlerAPI := httpHandler.NewHandler(&httpHandler.HandlerOptions{
		Config: &mainCfg,
		Auth:   auth,
		QuizUC: quc,
	})
	handlerAPI.Register(r)

	fmt.Println("Your server running on port :8000")
	log.Fatal(http.ListenAndServe(":8000", r))

}

type authOptions struct {
	redisClient *redis.Client
	db          *sql.DB
	schema      string
	origin      string
}

func generateAuth(options *authOptions) *pager.Pager {
	return pager.NewPager(&pager.Options{
		Dialect:      pager.MYSQLDialect,
		CacheClient:  options.redisClient,
		DbConnection: options.db,
		SchemaName:   options.schema,
		Session: pager.SessionOptions{
			Origin:           options.origin,
			LoginMethod:      pager.LoginEmail,
			ExpiredInSeconds: int64(24 * time.Hour),
			SessionName:      "_Quiz_App",
		},
	}).BuildPager()
}
