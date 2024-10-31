package googlesecretmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/mackee/envsecrets"
)

const (
	targetType = "aws_s3"
)

type resolver struct {
	envs         []*envsecrets.Env
	clientLoader func(ctx context.Context) (*secretmanager.Client, error)
}

type ResolverOption func(*resolver)

func WithClient(cl func(ctx context.Context) (*secretmanager.Client, error)) ResolverOption {
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
		r.clientLoader = func(ctx context.Context) (*secretmanager.Client, error) {
			client, err := secretmanager.NewClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create client: %w", err)
			}
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
	slog.DebugContext(ctx, "resolve by google secret manager")
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
	svs := make(map[string]string, len(names))

	for _, name := range names {
		resp, err := client.AccessSecretVersion(
			ctx,
			&secretmanagerpb.AccessSecretVersionRequest{
				Name: name,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to get object %s: %w", name, err)
		}
		slog.DebugContext(ctx, "secret key", slog.String("key", name))
		svs[name] = resp.Payload.String()
		var decoded map[string]string
		if err := json.Unmarshal(resp.Payload.Data, &decoded); err == nil {
			for k, v := range decoded {
				jk := strings.Join([]string{name, k}, ".")
				svs[jk] = v
				slog.DebugContext(ctx, "nested in secret", slog.String("key", jk))
			}
		}
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
