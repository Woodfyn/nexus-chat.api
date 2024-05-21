package app

import (
	"context"

	"os"

	"os/signal"
	"syscall"
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
	"github.com/Woodfyn/chat-api-backend-go/pkg/server"
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

	// file, err := os.Open("test.jpg")
	// if err != nil {
	// 	log.Error("Error when opening file: ", err)
	// 	panic(err)
	// }

	// defer file.Close()

	// result, err := presignS3.PresignGetObject(appCtx, &s3.GetObjectInput{
	// 	Bucket: aws.String("nexus-chat-bucket/avatar-develop/"),
	// 	Key:    aws.String("qwefmivn230n_1_5"),
	// })
	// if err != nil {
	// 	log.Error("Error when uploading file: ", err)
	// 	panic(err)
	// }

	// resp, err := http.Get("https://s3.eu-west-3.amazonaws.com/nexus-chat-bucket/avatars-develop/qwefmivn230n_1_5.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIA5FTY65RMU6UYMEHE%2F20240515%2Feu-west-3%2Fs3%2Faws4_request&X-Amz-Date=20240515T122504Z&X-Amz-Expires=60&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=86bbdb33ab04d1fa647475eb5add82572abd849a840bb948432adc350431c86b")
	// if err != nil {
	// 	log.Println("Error sending HTTP request:", err)
	// 	return
	// }
	// defer resp.Body.Close()

	// // Create a new file to save the photo
	// outFile, err := os.Create("downloaded_photo.jpg")
	// if err != nil {
	// 	log.Println("Error creating file:", err)
	// 	return
	// }

	// defer outFile.Close()

	// // Copy the response body to the file
	// _, err = io.Copy(outFile, resp.Body)
	// if err != nil {
	// 	log.Println("Error copying response body to file:", err)
	// 	return
	// }

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

		Log: log,
	})

	api := transport.NewApi(handler, cfg.Server.SwagAddr)

	srv := new(server.Server)

	// start server
	go func() {
		if err := srv.Run(cfg.Server.Port, api.InitApi()); err != nil {
			log.Error("Error when running server: ", err)
			panic(err)
		}
	}()

	log.WithFields(logrus.Fields{"server started:": time.Now().Format(time.DateTime)}).Info()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	log.WithFields(logrus.Fields{"server stoped:": time.Now().Format(time.DateTime)}).Info()

	srv.Shutdown(appCtx)
}
