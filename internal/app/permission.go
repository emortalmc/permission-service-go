package app

import (
	"context"
	"go.uber.org/zap"
	"permission-service/internal/config"
	"permission-service/internal/messaging/notifier"
	"permission-service/internal/repository"
	"permission-service/internal/service"
	"sync"
)

func Run(cfg *config.Config, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}

	delayedCtx, repoCancel := context.WithCancel(ctx)
	delayedWg := &sync.WaitGroup{}

	repo, err := repository.NewMongoRepository(delayedCtx, logger, delayedWg, cfg.MongoDB)
	if err != nil {
		logger.Fatalw("failed to create repository", "error", err)
	}

	notif := notifier.NewKafkaNotifier(delayedCtx, delayedWg, logger, cfg.Kafka)

	service.RunServices(ctx, logger, wg, cfg, repo, notif)

	wg.Wait()
	logger.Info("shutting down")

	logger.Info("shutting down delayed services")
	repoCancel()
	delayedWg.Wait()
}
