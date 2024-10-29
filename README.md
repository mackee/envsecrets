# envsecrets

"envsecrets" is a meta framework fo environment variable loader. It is designed to load environment variables from various sources.

## Installation

```bash
$ go get github.com/mackee/envsecrets
```

## Usage

### Environment variable format

If you want set from AWS Secrets Manager, you should set environment variable like below.

```bash
export SECRET_ENV=secretfrom:awssecretsmanager:<secret-name>
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

| Source | Type | Description | Status |
| --- | --- | --- | --- |
| AWS Secrets Manager | `awssecretsmanager` | Load secret from AWS Secrets Manager | ✅ |
| AWS Systems Manager ParameterStore | `awsssm` | Load secret from AWS Systems Manager ParameterStore | 🔍 |
| Amazon S3 | `awss3` | Load secret from Amazon S3 | 🔍 |
