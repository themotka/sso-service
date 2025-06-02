package suite

import (
	"context"
	ssov1 "github.com/themotka/proto/gen/go/sso"
	"github.com/themotka/sso-service/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strconv"
	"testing"
)

type Suite struct {
	*testing.T
	Config     *config.Config
	AuthClient ssov1.OAuthClient
}

func NewSuite(t *testing.T) (ctx context.Context, suite *Suite) {
	const op = "tests.suite.NewSuite"
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadWPath("../config/config.yaml")

	ctx, timeout := context.WithTimeout(context.Background(), cfg.Grpc.Timeout)

	t.Cleanup(func() {
		t.Helper()
		timeout()
	})

	connect, err := grpc.NewClient(
		address(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("%s: %v", op, err)
	}
	return ctx, &Suite{
		T:          t,
		Config:     cfg,
		AuthClient: ssov1.NewOAuthClient(connect),
	}
}

func address(cfg *config.Config) string {
	return net.JoinHostPort("localhost", strconv.Itoa(cfg.Grpc.Port))

}
