//go:build integration

package awsenvsecrets_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	awsenvsecrets "github.com/mackee/envsecrets/dist/aws"
)

type testCase struct {
	name     string
	envs     map[string]string
	expected map[string]string
}

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestLoadEnv(t *testing.T) {
	testCases := []testCase{
		{
			name: "no envs",
			envs: map[string]string{
				"ENV_A": "value_a",
				"ENV_B": "value_b",
			},
			expected: map[string]string{
				"ENV_A": "value_a",
				"ENV_B": "value_b",
			},
		},
		{
			name: "aws_secretsmanager",
			envs: map[string]string{
				"ENV_AWS_SECRETSMANAGER_SINGLE":      "secretfrom:aws_secretsmanager:envsecrets_single",
				"ENV_AWS_SECRETSMANAGER_NESTED_KEY1": "secretfrom:aws_secretsmanager:envsecrets_nestedvalue.key1",
				"ENV_AWS_SECRETSMANAGER_NESTED_KEY2": "secretfrom:aws_secretsmanager:envsecrets_nestedvalue.key2",
				"ENV_NO_EFFECT":                      "no_effect",
			},
			expected: map[string]string{
				"ENV_AWS_SECRETSMANAGER_SINGLE":      "secretsmanager_singlevalue",
				"ENV_AWS_SECRETSMANAGER_NESTED_KEY1": "secretsmanager_nested_value1",
				"ENV_AWS_SECRETSMANAGER_NESTED_KEY2": "secretsmanager_nested_value2",
				"ENV_NO_EFFECT":                      "no_effect",
			},
		},
		{
			name: "aws_ssm",
			envs: map[string]string{
				"ENV_AWS_SSM_SINGLE":      "secretfrom:aws_ssm:/envsecrets/single",
				"ENV_AWS_SSM_NESTED_KEY1": "secretfrom:aws_ssm:/envsecrets/nestedvalue.key1",
				"ENV_AWS_SSM_NESTED_KEY2": "secretfrom:aws_ssm:/envsecrets/nestedvalue.key2",
				"ENV_NO_EFFECT":           "no_effect",
			},
			expected: map[string]string{
				"ENV_AWS_SSM_SINGLE":      "ssm_singlevalue",
				"ENV_AWS_SSM_NESTED_KEY1": "ssm_nested_value1",
				"ENV_AWS_SSM_NESTED_KEY2": "ssm_nested_value2",
				"ENV_NO_EFFECT":           "no_effect",
			},
		},
		{
			name: "aws_s3",
			envs: map[string]string{
				"ENV_AWS_S3_SINGLE":      "secretfrom:aws_s3:s3://envsecrets-testbucket/envsecrets/single",
				"ENV_AWS_S3_NESTED_KEY1": "secretfrom:aws_s3:s3://envsecrets-testbucket/envsecrets/nestedvalue.key1",
				"ENV_AWS_S3_NESTED_KEY2": "secretfrom:aws_s3:s3://envsecrets-testbucket/envsecrets/nestedvalue.key2",
				"ENV_NO_EFFECT":          "no_effect",
			},
			expected: map[string]string{
				"ENV_AWS_S3_SINGLE":      "s3_singlevalue",
				"ENV_AWS_S3_NESTED_KEY1": "s3_nested_value1",
				"ENV_AWS_S3_NESTED_KEY2": "s3_nested_value2",
				"ENV_NO_EFFECT":          "no_effect",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				if err := os.Setenv(k, v); err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				defer os.Unsetenv(k)
			}

			ctx := context.Background()
			if err := awsenvsecrets.Load(ctx); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for k, v := range tc.expected {
				env := os.Getenv(k)
				if v != env {
					t.Errorf("expected: %v, got: %v", v, env)
				}
			}
		})
	}
}
