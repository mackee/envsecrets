package awsenvsecrets

import (
	"context"

	"github.com/mackee/envsecrets"
	"github.com/mackee/envsecrets/awssecretsmanager"
)

func Load(ctx context.Context) error {
	loader := envsecrets.NewLoader(
		envsecrets.WithResolver(
			awssecretsmanager.NewResolver(),
		),
	)

	return loader.Load(ctx)
}
