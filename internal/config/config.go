package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	v = viper.New()
)

type Config struct {
	Server   Server
	Database Database
	RDB      RDB
	S3       S3
	Twilio   Twilio
	Verify   Verify
	JWT      JWT
}

type Server struct {
	Port      string
	SwagAddr  string
	AWSRegion string
}

type Database struct {
	Url string
	Dsn string
}

type RDB struct {
	Addr string
}

type S3 struct {
	BucketName    string
	AvatarKeySalt string
}

type Twilio struct {
	SID     string
	AuthKey string
	Phone   string
}

type Verify struct {
	TTL time.Duration `mapstructure:"ttl"`
}

type JWT struct {
	Secret string

	AccessTokenTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_ttl"`
}

func InitConfig(folder string, name string) (*Config, error) {
	cfg := new(Config)

	v.AddConfigPath(folder)
	v.SetConfigName(name)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg, err := setEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
func setEnv(cfg *Config) (*Config, error) {
	if err := godotenv.Load("main.env"); err != nil {
		slog.Error("Error loading .env file")
		return nil, err
	}

	cfg.Server.Port = os.Getenv("SERVER_PORT")
	cfg.Server.SwagAddr = os.Getenv("SERVER_SWAG_ADDR")
	cfg.Server.AWSRegion = os.Getenv("AWS_DEFAULT_REGION")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	cfg.RDB.Addr = os.Getenv("REDIS_ADDR")

	cfg.Twilio.SID = os.Getenv("TWILIO_SID")
	cfg.Twilio.AuthKey = os.Getenv("TWILIO_TOKEN")
	cfg.Twilio.Phone = os.Getenv("TWILIO_PHONE")

	cfg.S3.BucketName = os.Getenv("S3_BUCKET_NAME")
	cfg.S3.AvatarKeySalt = os.Getenv("S3_AVATAR_SALT")

	postgresConn := generatDsn()

	cfg.Database.Url = postgresConn[0]
	cfg.Database.Dsn = postgresConn[1]

	return cfg, nil
}

func generatDsn() []string {
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, database, sslMode)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, database, port, sslMode)

	postgresConn := []string{url, dsn}

	return postgresConn
}
