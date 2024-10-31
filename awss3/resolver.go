package awss3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"slices"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mackee/envsecrets"
)

const (
	targetType = "aws_s3"
)

type resolver struct {
	envs         []*envsecrets.Env
	clientLoader func(ctx context.Context) (*s3.Client, error)
}

type ResolverOption func(*resolver)

func WithClient(cl func(ctx context.Context) (*s3.Client, error)) ResolverOption {
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
		r.clientLoader = func(ctx context.Context) (*s3.Client, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}
			client := s3.NewFromConfig(cfg)
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
	slog.DebugContext(ctx, "resolve by aws s3")
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
		u, err := url.Parse(name)
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}
		if u.Scheme != "s3" {
			slog.ErrorContext(ctx, "unsupported scheme", slog.String("scheme", u.Scheme), slog.String("args", name))
			continue
		}
		resp, err := client.GetObject(
			ctx,
			&s3.GetObjectInput{
				Bucket: &u.Host,
				Key:    &u.Path,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to get object %s: %w", name, err)
		}
		slog.DebugContext(ctx, "object", slog.String("name", name))
		bs := &bytes.Buffer{}
		if _, err := bs.ReadFrom(resp.Body); err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}
		svs[name] = bs.String()
		var decoded map[string]string
		if err := json.NewDecoder(bs).Decode(&decoded); err == nil {
			for k, v := range decoded {
				jk := strings.Join([]string{name, k}, ".")
				svs[jk] = v
				slog.DebugContext(ctx, "nested in object", slog.String("key", jk))
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
