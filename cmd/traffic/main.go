package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TinderBackend/telepresence/v2/cmd/traffic/cmd/agent"
	"github.com/TinderBackend/telepresence/v2/cmd/traffic/cmd/agentinit"
	"github.com/TinderBackend/telepresence/v2/cmd/traffic/cmd/manager"
	"github.com/TinderBackend/telepresence/v2/pkg/log"
	"github.com/datawire/dlib/dlog"
)

func doMain(fn func(ctx context.Context, args ...string) error, logLevel string, args ...string) {
	ctx := log.MakeBaseLogger(context.Background(), logLevel)

	if err := fn(ctx, args...); err != nil {
		dlog.Errorf(ctx, "quit: %v", err)
		os.Exit(1)
	}
}

func main() {
	level := os.Getenv("LOG_LEVEL")
	if len(os.Args) > 1 {
		switch name := os.Args[1]; name {
		case "agent":
			doMain(agent.Main, agent.GetLogLevel(), os.Args[2:]...)
		case "manager":
			doMain(manager.Main, level, os.Args[2:]...)
		case "agent-init":
			doMain(agentinit.Main, level, os.Args[2:]...)
		default:
			fmt.Println("traffic: unknown command:", name)
			os.Exit(127)
		}
		return
	}

	switch name := filepath.Base(os.Args[0]); name {
	case "traffic-agent":
		doMain(agent.Main, agent.GetLogLevel(), os.Args[1:]...)
	case "traffic-agent-init":
		doMain(agentinit.Main, level, os.Args[1:]...)
	case "traffic-manager":
		fallthrough
	default:
		doMain(manager.Main, level, os.Args[1:]...)
	}
}
