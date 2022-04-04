package client

import (
	"context"
	"testing"

	"github.com/hashicorp/consul-terraform-sync/logging"
	"github.com/stretchr/testify/assert"
)

func Test_isConsulEnterprise(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		info             ConsulAgentConfig
		expectEnterprise bool
		expectError      bool
	}{
		{
			name: "oss",
			info: ConsulAgentConfig{
				"Config": {"Version": "v1.9.5"},
			},
			expectEnterprise: false,
			expectError:      false,
		},
		{
			name: "oss dev",
			info: ConsulAgentConfig{
				"Config": {"Version": "v1.9.5-dev"},
			},
			expectEnterprise: false,
			expectError:      false,
		},
		{
			name: "ent",
			info: ConsulAgentConfig{
				"Config": {"Version": "v1.9.5+ent"},
			},
			expectEnterprise: true,
			expectError:      false,
		},
		{
			name: "ent dev",
			info: ConsulAgentConfig{
				"Config": {"Version": "v1.9.5+ent-dev"},
			},
			expectEnterprise: true,
			expectError:      false,
		},
		{
			name: "missing",
			info: ConsulAgentConfig{
				"Config": {},
			},
			expectEnterprise: false,
			expectError:      true,
		},
		{
			name: "malformed",
			info: ConsulAgentConfig{
				"Config": {"Version": "***"},
			},
			expectEnterprise: false,
			expectError:      true,
		},
		{
			name: "bad key",
			info: ConsulAgentConfig{
				"NoConfig": {"Version": "***"},
			},
			expectEnterprise: false,
			expectError:      true,
		},
	}

	ctx := logging.WithContext(context.Background(), logging.NewNullLogger())

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isEnterprise, err := isConsulEnterprise(ctx, tc.info)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectEnterprise, isEnterprise)
			}
		})
	}
}
