package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Mongo    MongoConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Port        string `mapstructure:"APP_PORT"`
	Env         string `mapstructure:"APP_ENV"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	FrontendURL string `mapstructure:"FRONTEND_URL"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     string `mapstructure:"POSTGRES_PORT"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	DB       string `mapstructure:"POSTGRES_DB"`
	DSN      string `mapstructure:"POSTGRES_DSN"`
}

type MongoConfig struct {
	Host     string `mapstructure:"MONGO_HOST"`
	Port     string `mapstructure:"MONGO_PORT"`
	User     string `mapstructure:"MONGO_USER"`
	Password string `mapstructure:"MONGO_PASSWORD"`
	DB       string `mapstructure:"MONGO_DB"`
	URI      string `mapstructure:"MONGO_URI"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	Addr     string `mapstructure:"REDIS_ADDR"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"JWT_SECRET"`
	Expiration time.Duration `mapstructure:"JWT_EXPIRATION"`
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("../..")

	v.AutomaticEnv()

	// Read .env file (ignore error if file doesn't exist)
	_ = v.ReadInConfig()

	setDefaults(v)

	cfg := &Config{}

	cfg.App = AppConfig{
		Port:        v.GetString("APP_PORT"),
		Env:         v.GetString("APP_ENV"),
		LogLevel:    v.GetString("LOG_LEVEL"),
		FrontendURL: v.GetString("FRONTEND_URL"),
	}

	cfg.Postgres = PostgresConfig{
		Host:     v.GetString("POSTGRES_HOST"),
		Port:     v.GetString("POSTGRES_PORT"),
		User:     v.GetString("POSTGRES_USER"),
		Password: v.GetString("POSTGRES_PASSWORD"),
		DB:       v.GetString("POSTGRES_DB"),
		DSN:      v.GetString("POSTGRES_DSN"),
	}

	if cfg.Postgres.DSN == "" {
		cfg.Postgres.DSN = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Postgres.User, cfg.Postgres.Password,
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DB,
		)
	}

	cfg.Mongo = MongoConfig{
		Host:     v.GetString("MONGO_HOST"),
		Port:     v.GetString("MONGO_PORT"),
		User:     v.GetString("MONGO_USER"),
		Password: v.GetString("MONGO_PASSWORD"),
		DB:       v.GetString("MONGO_DB"),
		URI:      v.GetString("MONGO_URI"),
	}

	if cfg.Mongo.URI == "" {
		cfg.Mongo.URI = fmt.Sprintf(
			"mongodb://%s:%s@%s:%s/%s?authSource=admin",
			cfg.Mongo.User, cfg.Mongo.Password,
			cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.DB,
		)
	}

	cfg.Redis = RedisConfig{
		Host:     v.GetString("REDIS_HOST"),
		Port:     v.GetString("REDIS_PORT"),
		Password: v.GetString("REDIS_PASSWORD"),
		Addr:     v.GetString("REDIS_ADDR"),
	}

	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	}

	cfg.JWT = JWTConfig{
		Secret:     v.GetString("JWT_SECRET"),
		Expiration: v.GetDuration("JWT_EXPIRATION"),
	}

	if cfg.JWT.Expiration == 0 {
		cfg.JWT.Expiration = 24 * time.Hour
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("FRONTEND_URL", "http://localhost:3000")
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", "5432")
	v.SetDefault("POSTGRES_USER", "guimba")
	v.SetDefault("POSTGRES_PASSWORD", "guimba_secret")
	v.SetDefault("POSTGRES_DB", "guimba_db")
	v.SetDefault("MONGO_HOST", "localhost")
	v.SetDefault("MONGO_PORT", "27017")
	v.SetDefault("MONGO_USER", "guimba")
	v.SetDefault("MONGO_PASSWORD", "guimba_secret")
	v.SetDefault("MONGO_DB", "guimba_db")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6380")
	v.SetDefault("REDIS_PASSWORD", "guimba_secret")
	v.SetDefault("JWT_SECRET", "change-me-in-production")
}

func validate(cfg *Config) error {
	var errs []string

	if cfg.App.Port == "" {
		errs = append(errs, "APP_PORT is required")
	}
	if cfg.Postgres.DSN == "" {
		errs = append(errs, "POSTGRES_DSN is required")
	}
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "change-me-in-production" {
		if cfg.App.Env == "production" {
			errs = append(errs, "JWT_SECRET must be set in production")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
