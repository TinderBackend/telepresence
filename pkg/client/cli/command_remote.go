package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	empty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/TinderBackend/telepresence/rpc/v2/connector"
	"github.com/TinderBackend/telepresence/rpc/v2/daemon"
	"github.com/TinderBackend/telepresence/v2/pkg/client/cli/cliutil"
	"github.com/TinderBackend/telepresence/v2/pkg/client/userd/commands"
)

func getRemoteCommands(ctx context.Context) (groups cliutil.CommandGroups, err error) {
	err = cliutil.WithStartedConnector(ctx, false, func(ctx context.Context, connectorClient connector.ConnectorClient) error {
		remote, err := connectorClient.ListCommands(ctx, &empty.Empty{})
		if err != nil {
			return fmt.Errorf("unable to call ListCommands: %w", err)
		}
		if groups, err = cliutil.RPCToCommands(remote, runRemote); err != nil {
			groups = commands.GetCommandsForLocal(err)
		}
		userDaemonRunning = true
		return nil
	})
	if err != nil && err != cliutil.ErrNoUserDaemon {
		return nil, err
	}
	if !userDaemonRunning {
		groups = commands.GetCommandsForLocal(err)
	}
	return groups, nil
}

func runRemote(cmd *cobra.Command, _ []string) error {
	return cliutil.WithNetwork(cmd.Context(), func(ctx context.Context, _ daemon.DaemonClient) error {
		return cliutil.WithConnector(ctx, func(ctx context.Context, connectorClient connector.ConnectorClient) error {
			result, err := connectorClient.RunCommand(ctx, &connector.RunCommandRequest{OsArgs: os.Args[1:]})
			if err != nil {
				return err
			}
			_, _ = cmd.OutOrStdout().Write(result.GetStdout())
			_, _ = cmd.ErrOrStderr().Write(result.GetStderr())
			return nil
		})
	})
}
