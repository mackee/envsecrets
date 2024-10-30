package awsssm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mackee/envsecrets"
)

const (
	targetType = "aws_ssm"
)

type resolver struct {
	envs         []*envsecrets.Env
	clientLoader func(ctx context.Context) (*ssm.Client, error)
}

type ResolverOption func(*resolver)

func WithClient(cl func(ctx context.Context) (*ssm.Client, error)) ResolverOption {
	return func(r *resolver) {
		r.clientLoader = cl
	}
}

func NewResolver(opts ...ResolverOption) envsecrets.Resolver {
	r := &resolver{}
	for _, opt := range opts {
		opt(r)
	}
	if r.clientLoader == nil {
		r.clientLoader = func(ctx context.Context) (*ssm.Client, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}
			client := ssm.NewFromConfig(cfg)
			return client, nil
		}
	}

	return r
}

func (r *resolver) Push(env *envsecrets.Env) error {
	if env.Type != targetType {
		return envsecrets.ErrResolverNotInterested
	}
	r.envs = append(r.envs, env)
	return nil
}

func (r *resolver) Resolve(ctx context.Context) error {
	if len(r.envs) == 0 {
		return nil
	}
	slog.DebugContext(ctx, "resolve by aws systems manager parameter store")
	client, err := r.clientLoader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load client: %w", err)
	}
	nameMap := make(map[string]struct{})
	m := make(map[string]*envsecrets.Env)
	for _, env := range r.envs {
		idKey := strings.SplitN(env.Args, ".", 2)
		if len(idKey) == 2 {
			nameMap[idKey[0]] = struct{}{}
		} else {
			nameMap[env.Args] = struct{}{}
		}
		m[env.Args] = env
	}
	names := slices.Collect(maps.Keys(nameMap))

	resp, err := client.GetParameters(
		ctx,
		&ssm.GetParametersInput{
			Names: names,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get parameters: %w", err)
	}
	svs := make(map[string]string, len(resp.Parameters))
	for _, parameter := range resp.Parameters {
		key := *parameter.Name
		slog.DebugContext(ctx, "parameter", slog.String("key", key))
		var decoded map[string]string
		if err := json.Unmarshal([]byte(*parameter.Value), &decoded); err == nil {
			for k, v := range decoded {
				jk := strings.Join([]string{key, k}, ".")
				svs[jk] = v
				slog.DebugContext(ctx, "nested parameter", slog.String("key", jk))
			}
		}
		svs[key] = *parameter.Value
	}
	for k, v := range svs {
		env, ok := m[k]
		if !ok {
			continue
		}
		env.Value = v
		env.Resolved = true
	}

	return nil
}
