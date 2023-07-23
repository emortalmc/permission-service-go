package service

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/grpc/permission"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"permission-service/internal/config"
	"permission-service/internal/kafka/notifier"
	"permission-service/internal/repository"
	"sync"
)

func RunServices(ctx context.Context, logger *zap.SugaredLogger, wg *sync.WaitGroup, cfg *config.Config,
	repo repository.Repository, notif notifier.Notifier) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatalw("failed to listen", "error", err)
	}

	s := grpc.NewServer()

	if cfg.Development {
		reflection.Register(s)
	}

	permission.RegisterPermissionServiceServer(s, newPermissionService(logger, repo, notif))
	logger.Infow("listening for gRPC requests", "port", cfg.Port)

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatalw("failed to serve", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.GracefulStop()
	}()

}
