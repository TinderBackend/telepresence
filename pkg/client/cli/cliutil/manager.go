package cliutil

import (
	"context"

<<<<<<< HEAD
	"google.golang.org/grpc"

	"github.com/TinderBackend/telepresence/rpc/v2/connector"
	"github.com/TinderBackend/telepresence/rpc/v2/manager"
=======
	"github.com/TinderBackend/telepresence/rpc/v2/connector"
	"github.com/TinderBackend/telepresence/rpc/v2/manager"
>>>>>>> upstream/release/v2
)

func WithManager(ctx context.Context, fn func(context.Context, manager.ManagerClient) error) error {
	return WithConnector(ctx, func(ctx context.Context, _ connector.ConnectorClient) error {
		conn := getConnectorConn(ctx)
		managerClient := manager.NewManagerClient(conn)
		return fn(ctx, managerClient)
	})
}
