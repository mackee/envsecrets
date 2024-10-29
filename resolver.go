package envsecrets

import (
	"context"
	"errors"
)

type Resolver interface {
	Push(env *Env) error
	Resolve(ctx context.Context) error
}

var ErrResolverNotInterested = errors.New("resolver not interested")
