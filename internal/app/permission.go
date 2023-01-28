package app

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/grpc/permission"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"permission-service-go/internal/config"
	"permission-service-go/internal/notifier"
	"permission-service-go/internal/repository"
	"permission-service-go/internal/service"
)

func Run(ctx context.Context, cfg *config.Config, logger *zap.SugaredLogger) {
	repo, err := repository.NewMongoRepository(ctx, cfg.MongoDB)
	if err != nil {
		logger.Fatalw("failed to create repository", "error", err)
	}

	notif, err := notifier.NewRabbitMqNotifier(cfg.RabbitMQ)
	if err != nil {
		logger.Fatalw("failed to create notifier", "error", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatalw("failed to listen", "error", err)
	}

	s := grpc.NewServer()
	permission.RegisterPermissionServiceServer(s, service.NewPermissionService(repo, notif))
	logger.Infow("listening on port", "port", cfg.Port)

	err = s.Serve(lis)
	if err != nil {
		logger.Fatalw("failed to serve", "error", err)
		return
	}
}
