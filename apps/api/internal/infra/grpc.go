package infra

import (
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	historyv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	recommendationv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/recommendation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitializeGRPC(cfg *config.Config) (
	historyv1.HistoryServiceClient,
	recommendationv1.RecommendationServiceClient,
	func(),
	error,
) {
	historyAddr := cfg.HistoryGRPCAddr
	historyConn, err := grpc.NewClient(
		historyAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to history ms: %w", err)
	}

	historyClient := historyv1.NewHistoryServiceClient(historyConn)

	recAddr := cfg.RecommendationGRPCAddr
	recConn, err := grpc.NewClient(
		recAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		historyConn.Close()
		return nil, nil, nil, fmt.Errorf("failed to connect to recommendation ms: %w", err)
	}
	recClient := recommendationv1.NewRecommendationServiceClient(recConn)

	cleanup := func() {
		historyConn.Close()
		recConn.Close()
	}

	return historyClient, recClient, cleanup, nil
}
