package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"task-tracker/internal/email/cache"
	kafka2 "task-tracker/internal/email/transport/kafka"
	"task-tracker/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accountpb "task-tracker/gen/internal/account"
	"task-tracker/internal/email/config"
	"task-tracker/internal/email/mailer"
	transportgrpc "task-tracker/internal/email/transport/grpc"
	"task-tracker/internal/email/usecase"
	pkgcache "task-tracker/pkg/cache"
	pkgkafka "task-tracker/pkg/kafka"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatalf("load config: %v", err)
	}

	mailerClient, err := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, cfg.SMTPUseTLS, cfg.Timeout)
	if err != nil {
		logger.Log.Fatalf("init smtp: %v", err)
	}

	redisClient, err := pkgcache.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.Timeout)
	if err != nil {
		logger.Log.Fatalf("init redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Log.Infof("close redis: %v", err)
		}
	}()

	dedupe := cache.NewRedisDedupe(redisAdapter{client: redisClient})
	service := usecase.NewService(mailerClient, dedupe, cfg.DedupeTTL)
	consumer := kafka2.NewConsumer(service)

	accountConn, err := grpc.NewClient(cfg.AccountGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatalf("dial account grpc: %v", err)
	}
	defer func() {
		if err := accountConn.Close(); err != nil {
			logger.Log.Infof("close account grpc: %v", err)
		}
	}()
	accountClient := transportgrpc.NewAccountClientAdapter(accountpb.NewUsersServiceClient(accountConn))

	registerReader, err := pkgkafka.NewReader(cfg.KafkaBroker, cfg.RegisterTopic, cfg.GroupID+"-register")
	if err != nil {
		logger.Log.Fatalf("init register reader: %v", err)
	}
	defer registerReader.Close()

	dailyReader, err := pkgkafka.NewReader(cfg.KafkaBroker, cfg.DailySummaryTopic, cfg.GroupID+"-daily")
	if err != nil {
		logger.Log.Fatalf("init daily reader: %v", err)
	}
	defer dailyReader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)
	go consumer.ConsumeRegister(ctx, &readerAdapter{reader: registerReader}, errCh)
	go consumer.ConsumeDaily(ctx, &readerAdapter{reader: dailyReader}, accountClient, errCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		cancel()
	case err := <-errCh:
		if err != nil {
			logger.Log.Infof("consumer error: %v", err)
		}
		cancel()
	}
}

type readerAdapter struct {
	reader *kafka.Reader
}

func (r *readerAdapter) FetchMessage(ctx context.Context) (kafka2.Message, error) {
	msg, err := r.reader.FetchMessage(ctx)
	if err != nil {
		return kafka2.Message{}, err
	}
	return kafka2.Message{Value: msg.Value}, nil
}

func (r *readerAdapter) CommitMessages(ctx context.Context, msg kafka2.Message) error {
	return r.reader.CommitMessages(ctx, kafka.Message{Value: msg.Value})
}

type redisAdapter struct {
	client *pkgcache.Client
}

func (r redisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}
