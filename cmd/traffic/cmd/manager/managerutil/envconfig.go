package managerutil

import (
	"context"
	"strings"

	"github.com/sethvargo/go-envconfig"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/TinderBackend/telepresence/v2/pkg/k8sapi"
	"github.com/TinderBackend/telepresence/v2/pkg/version"
)

type Env struct {
	User        string `env:"USER,default="`
	ServerHost  string `env:"SERVER_HOST,default="`
	ServerPort  string `env:"SERVER_PORT,default=8081"`
	SystemAHost string `env:"SYSTEMA_HOST,default=app.getambassador.io"`
	SystemAPort string `env:"SYSTEMA_PORT,default=443"`

	ManagerNamespace    string                     `env:"MANAGER_NAMESPACE,default="`
	AgentRegistry       string                     `env:"TELEPRESENCE_REGISTRY,default=docker.io/datawire"`
	AgentImage          string                     `env:"TELEPRESENCE_AGENT_IMAGE,default="`
	AgentPort           int32                      `env:"TELEPRESENCE_AGENT_PORT,default=9900"`
	APIPort             int32                      `env:"TELEPRESENCE_API_PORT,default="`
	MaxReceiveSize      resource.Quantity          `env:"TELEPRESENCE_MAX_RECEIVE_SIZE,default=4Mi"`
	AppProtocolStrategy k8sapi.AppProtocolStrategy `env:"TELEPRESENCE_APP_PROTO_STRATEGY,default="`

	PodCIDRStrategy string `env:"POD_CIDR_STRATEGY,default=auto"`
	PodCIDRs        string `env:"POD_CIDRS,default="`
	PodIP           string `env:"TELEPRESENCE_MANAGER_POD_IP,default="`
}

type envKey struct{}

func LoadEnv(ctx context.Context) (context.Context, error) {
	var env Env
	if err := envconfig.Process(ctx, &env); err != nil {
		return ctx, err
	}
	if env.AgentImage == "" {
		env.AgentImage = "tel2:" + strings.TrimPrefix(version.Version, "v")
	}
	return WithEnv(ctx, &env), nil
}

func WithEnv(ctx context.Context, env *Env) context.Context {
	return context.WithValue(ctx, envKey{}, env)
}

func GetEnv(ctx context.Context) *Env {
	env, ok := ctx.Value(envKey{}).(*Env)
	if !ok {
		return nil
	}
	return env
}
