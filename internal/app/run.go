package app

import (
	"context"

	"os"

	"time"

	"github.com/Woodfyn/chat-api-backend-go/internal/config"
	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/Woodfyn/chat-api-backend-go/internal/repository/psql"
	"github.com/Woodfyn/chat-api-backend-go/internal/repository/rdb"
	repoS3 "github.com/Woodfyn/chat-api-backend-go/internal/repository/s3"
	"github.com/Woodfyn/chat-api-backend-go/internal/repository/twilio"
	"github.com/Woodfyn/chat-api-backend-go/internal/service"
	"github.com/Woodfyn/chat-api-backend-go/internal/transport"
	"github.com/Woodfyn/chat-api-backend-go/internal/transport/rest"
	"github.com/Woodfyn/chat-api-backend-go/internal/transport/websocket"
	"github.com/Woodfyn/chat-api-backend-go/pkg/encoder"
	"github.com/Woodfyn/chat-api-backend-go/pkg/server"
	"github.com/Woodfyn/chat-api-backend-go/pkg/signaler"
	"github.com/Woodfyn/chat-api-backend-go/pkg/token"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/go-redis/redis"
	tw "github.com/twilio/twilio-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	configFolder = "configs"
	configName   = "prod"

	logFileName = "app.log"
)

var (
	log    = logrus.New()
	appCtx = context.Background()
)

func init() {
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetFormatter(new(logrus.JSONFormatter))
	log.SetOutput(logFile)
	log.SetLevel(logrus.InfoLevel)
}

func Run() {
	cfg, err := config.InitConfig(configFolder, configName)
	if err != nil {
		log.Error("Error when initializing config: ", err)
		panic(err)
	}

	log.Info("Config: ", cfg)

	// init aws s3
	awsCfg, err := awsCfg.LoadDefaultConfig(appCtx, awsCfg.WithRegion(cfg.Server.AWSRegion))
	if err != nil {
		log.Fatal(err)
	}

	storageS3 := s3.NewFromConfig(awsCfg)
	presignS3 := s3.NewPresignClient(storageS3)

	// init db
	db, err := gorm.Open(postgres.Open(cfg.Database.Dsn), &gorm.Config{})
	if err != nil {
		log.Error("Error when connecting to the database: ", err)
		panic(err)
	}

	// make migration
	if err := core.AutoMigrate(db); err != nil {
		log.Error("Error when migrating database: ", err)
		panic(err)
	}

	// init token manager
	manager, err := token.NewManager(cfg.JWT.Secret)
	if err != nil {
		log.Error("Error when creating token manager: ", err)
		panic(err)
	}

	// init redis
	rdbClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RDB.Addr,
		Password:     "",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	if err := rdbClient.Ping().Err(); err != nil {
		log.Error("Error when connecting to the redis: ", err)
		panic(err)
	}

	defer rdbClient.Close()

	// init twilio
	twilioClient := tw.NewRestClientWithParams(tw.ClientParams{
		Username: cfg.Twilio.SID,
		Password: cfg.Twilio.AuthKey,
	})

	// init dependencies
	handler := rest.NewHandler(rest.Deps{
		Auth: service.NewAuth(psql.NewAuth(db, log),
			rdb.NewVerife(rdbClient, cfg.Verify.TTL, log),
			twilio.NewVerify(twilioClient, cfg.Twilio.Phone, cfg.Twilio.SID, log),
			manager, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, log),
		Profile: service.NewProfile(psql.NewProfile(db, log),
			repoS3.NewProfile(storageS3, presignS3, cfg.S3.BucketName, log),
			cfg.S3.AvatarKeySalt, log),
		WebSocket: service.NewWebSocket(psql.NewWebSocket(db, log),
			log),

		Encoder: encoder.New(cfg.Server.EncodeSecret),

		WebSocketHandler: websocket.NewWebSocketHandler(encoder.New(cfg.Server.EncodeSecret)),

		Log: log,
	})

	// init api
	api := transport.NewApi(handler, cfg.Server.SwagAddr)

	// start server
	srv := new(server.Server)

	go func() {
		if err := srv.Run(cfg.Server.Port, api.InitApi()); err != nil {
			log.Error("Error when running server: ", err)
			panic(err)
		}
	}()

	log.WithFields(logrus.Fields{"server started:": time.Now().Format(time.DateTime)}).Info()

	// graceful shutdown
	signaler.Wait()

	log.WithFields(logrus.Fields{"server stoped:": time.Now().Format(time.DateTime)}).Info()

	// stop server
	srv.Shutdown(appCtx)
}
