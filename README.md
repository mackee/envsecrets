# envsecrets

"envsecrets" is a meta framework for environment variable loader. It is designed to load environment variables from various sources.

## Installation

```bash
$ go get github.com/mackee/envsecrets
```

## Usage

### Environment variable format

If you want set from AWS Secrets Manager, you should set environment variable like below.

```bash
export <SECRET_ENVNAME>=secretfrom:aws_secretsmanager:<secret-name>
```

If set a json value to the secret, you can access the value by specifying the key.

```bash
export <SECRET_ENVNAME>=secretfrom:aws_secretsmanager:<secret-name>.<key>
```

### Load environment variables

```go
import (
    "github.com/mackee/envsecrets/dist/aws"
)

func main() {
    ctx := context.Background()
    if err := awsenvsecrets.Load(ctx); err != nil {
        log.Fatalf("failed to load environment variables: %v", err)
    }
}
```

## Supported sources

✅: Implemented

🔍: Not implemented yet

| Source | Type | Description | Format | Status |
| --- | --- | --- | --- | --- |
| AWS Secrets Manager | `aws_secretsmanager` | Load secret from AWS Secrets Manager | `secretfrom:aws_secretsmanager:<id>[.<key>]` | ✅ |
| AWS Systems Manager ParameterStore | `aws_ssm` | Load secret from AWS Systems Manager ParameterStore | `secretfrom:aws_ssm:<name>[.<key>]` | ✅ |
| Amazon S3 | `aws_s3` | Load secret from Amazon S3 | `secretfrom:aws_s3:s3://<bucket-name>/<object-key>[.<key>]` | ✅ |
| Google Cloud Secret Manager | `google_secretmanager` | Load secret from Google Cloud Secret Manager | `secretfrom:google_secretmanager:projects/<project>/secrets/<name>/versions/<version>[.<key>]` |  ✅ |
| 1Password | `onepassword` | Load secret from 1Password | | 🔍 |

## License

Copyright (c) 2024- [mackee](https://github.com/mackee)

Licensed under MIT License.
