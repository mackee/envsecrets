package awsenvsecrets

import (
	"context"

	"github.com/mackee/envsecrets"
	"github.com/mackee/envsecrets/awssecretsmanager"
	"github.com/mackee/envsecrets/awsssm"
)

func Load(ctx context.Context) error {
	loader := envsecrets.NewLoader(
		envsecrets.WithResolver(
			awssecretsmanager.NewResolver(),
			awsssm.NewResolver(),
		),
	)

	return loader.Load(ctx)
}
