package envsecrets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	targetEnvPrefix = "secretfrom"
)

type Loader interface {
	Load(ctx context.Context) error
}

type loader struct {
	resolvers []Resolver
	envLoader func(ctx context.Context) ([]*Env, error)
}

type LoaderOption func(*loader)

func WithResolver(resolvers ...Resolver) LoaderOption {
	return func(l *loader) {
		l.resolvers = append(l.resolvers, resolvers...)
	}
}

func WithEnvLoader(_l func(ctx context.Context) ([]*Env, error)) LoaderOption {
	return func(l *loader) {
		l.envLoader = _l
	}
}

func NewLoader(opts ...LoaderOption) Loader {
	l := &loader{
		resolvers: make([]Resolver, 0),
		envLoader: func(ctx context.Context) ([]*Env, error) {
			_env := os.Environ()
			envs, err := ParseEnvs(_env)
			if err != nil {
				return nil, fmt.Errorf("failed to parse envs: %w", err)
			}
			return envs, nil
		},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *loader) Load(ctx context.Context) error {
	envs, err := l.envLoader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load envs: %w", err)
	}

	for _, env := range envs {
		for _, r := range l.resolvers {
			if err := r.Push(env); err != nil && !errors.Is(err, ErrResolverNotInterested) {
				return fmt.Errorf("failed to push env: %w", err)
			}
		}
	}

	for _, r := range l.resolvers {
		if err := r.Resolve(ctx); err != nil {
			return fmt.Errorf("failed to resolve envs: %w", err)
		}
	}

	for _, env := range envs {
		if !env.Resolved {
			slog.WarnContext(ctx, "env not resolved", slog.String("key", env.Key))
		}
		if err := os.Setenv(env.Key, env.Value); err != nil {
			return fmt.Errorf("failed to set env: %w", err)
		}
	}

	return nil
}

func ParseEnvs(src []string) ([]*Env, error) {
	envs := make([]*Env, 0, len(src))
	for _, e := range src {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("invalid env format")
		}
		key, value := kv[0], kv[1]
		if !strings.HasPrefix(value, targetEnvPrefix) {
			continue
		}
		typeAndArgs := strings.SplitN(value, ":", 3)
		if len(typeAndArgs) != 3 {
			return nil, errors.New("invalid env format")
		}
		t, args := typeAndArgs[1], typeAndArgs[2]
		envs = append(envs, &Env{
			Key:      key,
			Type:     t,
			Args:     args,
			Resolved: false,
		})
	}

	return envs, nil
}

type Env struct {
	Key      string
	Type     string
	Args     string
	Value    string
	Resolved bool
}
