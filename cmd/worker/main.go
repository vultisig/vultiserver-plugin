package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultiserver-plugin/config"
	"github.com/vultisig/vultiserver-plugin/internal/syncer"
	"github.com/vultisig/vultiserver-plugin/internal/tasks"
	"github.com/vultisig/vultiserver-plugin/service"
	"github.com/vultisig/vultiserver-plugin/storage"
)

func main() {
	cfg, err := config.GetConfigure()
	if err != nil {
		panic(err)
	}

	sdClient, err := statsd.New(cfg.Datadog.Host + ":" + cfg.Datadog.Port)
	if err != nil {
		panic(err)
	}
	blockStorage, err := storage.NewBlockStorage(*cfg)
	if err != nil {
		panic(err)
	}
	redisOptions := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	logger := logrus.StandardLogger()
	verifierConfig, err := config.ReadConfig("config-verifier")
	if err != nil {
		panic(err)
	}
	syncerService := syncer.NewPolicySyncer(logger.WithField("service", "syncer").Logger, verifierConfig.Server.Host, verifierConfig.Server.Port)
	authService := service.NewAuthService(cfg.Server.JWTSecret)

	client := asynq.NewClient(redisOptions)
	inspector := asynq.NewInspector(redisOptions)

	workerService, err := service.NewWorker(*cfg, verifierConfig.Server.Port, client, sdClient, syncerService, authService, blockStorage, inspector)
	if err != nil {
		panic(err)
	}

	srv := asynq.NewServer(
		redisOptions,
		asynq.Config{
			Logger:      logger,
			Concurrency: 10,
			Queues: map[string]int{
				tasks.QUEUE_NAME:         10,
				tasks.EMAIL_QUEUE_NAME:   100,
				"scheduled_plugin_queue": 10, // new queue
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeKeyGeneration, workerService.HandleKeyGeneration)
	mux.HandleFunc(tasks.TypeKeySign, workerService.HandleKeySign)
	mux.HandleFunc(tasks.TypeEmailVaultBackup, workerService.HandleEmailVaultBackup)
	mux.HandleFunc(tasks.TypeReshare, workerService.HandleReshare)
	mux.HandleFunc(tasks.TypePluginTransaction, workerService.HandlePluginTransaction)
	mux.HandleFunc(tasks.TypeKeyGenerationDKLS, workerService.HandleKeyGenerationDKLS)
	mux.HandleFunc(tasks.TypeKeySignDKLS, workerService.HandleKeySignDKLS)
	mux.HandleFunc(tasks.TypeReshareDKLS, workerService.HandleReshareDKLS)
	mux.HandleFunc(tasks.TypeMigrate, workerService.HandleMigrateDKLS)
	if err := srv.Run(mux); err != nil {
		panic(fmt.Errorf("could not run server: %w", err))
	}
}
