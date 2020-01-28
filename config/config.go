package config

import "time"

type Config struct {
	BaseUrl      string      `json:"base_url"`
	Quiz         QuizConfig  `json:"quiz"`
	DbConnection DbConfig    `json:"database"`
	Redis        RedisConfig `json:"redis"`
	JWT          JWTConfig   `json:"jwt"`
}

type QuizConfig struct {
	NumberOfQuestion int           `json:"number_of_question"`
	ReadyCountDown   time.Duration `json:"ready_count_down"`
	CountDown        time.Duration `json:"count_down"`
	TemplatePath     string        `json:"template_path"`
}

type DbConfig struct {
	DbAddress  string `json:"address"`
	DbName     string `json:"name"`
	DbUser     string `json:"user"`
	DbPassword string `json:"password"`
}

type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

type JWTConfig struct {
	SecretKey string `json:"secret_key"`
}
