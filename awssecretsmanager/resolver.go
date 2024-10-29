package awssecretsmanager

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/mackee/go-envsecrets"
)

const (
	targetType = "awssecretsmanager"
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
	client, err := r.clientLoader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load client: %w", err)
	}
	ids := make([]string, 0, len(r.envs))
	m := make(map[string]*envsecrets.Env)
	for _, env := range r.envs {
		ids = append(ids, env.Args)
		m[env.Args] = env
	}

	resp, err := client.BatchGetSecretValue(
		ctx,
		&secretsmanager.BatchGetSecretValueInput{
			SecretIdList: ids,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to batch get secret value: %w", err)
	}
	for _, secret := range resp.SecretValues {
		env, ok := m[*secret.Name]
		if !ok {
			return fmt.Errorf("secret not found: %s", *secret.Name)
		}
		env.Value = *secret.SecretString
		env.Resolved = true
	}

	return nil
}
