package awssecretsmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/mackee/envsecrets"
)

const (
	targetType = "aws_secretsmanager"
)

type resolver struct {
	envs         []*envsecrets.Env
	clientLoader func(ctx context.Context) (*secretsmanager.Client, error)
}

type ResolverOption func(*resolver)

func WithClient(cl func(ctx context.Context) (*secretsmanager.Client, error)) ResolverOption {
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
		r.clientLoader = func(ctx context.Context) (*secretsmanager.Client, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}
			client := secretsmanager.NewFromConfig(cfg)
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
	slog.DebugContext(ctx, "resolve by aws secrets manager")
	client, err := r.clientLoader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load client: %w", err)
	}
	idm := make(map[string]struct{})
	m := make(map[string]*envsecrets.Env)
	for _, env := range r.envs {
		idKey := strings.SplitN(env.Args, ".", 2)
		if len(idKey) == 2 {
			idm[idKey[0]] = struct{}{}
		} else {
			idm[env.Args] = struct{}{}
		}
		m[env.Args] = env
	}
	ids := slices.Collect(maps.Keys(idm))

	resp, err := client.BatchGetSecretValue(
		ctx,
		&secretsmanager.BatchGetSecretValueInput{
			SecretIdList: ids,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to batch get secret value: %w", err)
	}
	svs := make(map[string]string, len(resp.SecretValues))
	for _, secret := range resp.SecretValues {
		key := *secret.Name
		slog.DebugContext(ctx, "secret", slog.String("key", key))
		var decoded map[string]string
		if err := json.Unmarshal([]byte(*secret.SecretString), &decoded); err == nil {
			for k, v := range decoded {
				jk := strings.Join([]string{key, k}, ".")
				svs[jk] = v
				slog.DebugContext(ctx, "nested secret", slog.String("key", jk))
			}
		}
		svs[key] = *secret.SecretString
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
