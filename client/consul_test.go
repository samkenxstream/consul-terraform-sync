package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelf_SKU(t *testing.T) {
	t.Parallel()

	t.Run("oss", func(t *testing.T) {
		s, ok := parseSKU(Self{
			"Config": {"Version": "v1.9.5"},
		})
		require.True(t, ok)
		require.Equal(t, "oss", s)
	})

	t.Run("oss dev", func(t *testing.T) {
		s, ok := parseSKU(Self{
			"Config": {"Version": "v1.9.5-dev"},
		})
		require.True(t, ok)
		require.Equal(t, "oss", s)
	})

	t.Run("ent", func(t *testing.T) {
		s, ok := parseSKU(Self{
			"Config": {"Version": "v1.9.5+ent"},
		})
		require.True(t, ok)
		require.Equal(t, "ent", s)
	})

	t.Run("ent dev", func(t *testing.T) {
		s, ok := parseSKU(Self{
			"Config": {"Version": "v1.9.5+ent-dev"},
		})
		require.True(t, ok)
		require.Equal(t, "ent", s)
	})

	t.Run("missing", func(t *testing.T) {
		_, ok := parseSKU(Self{
			"Config": {},
		})
		require.False(t, ok)
	})

	t.Run("malformed", func(t *testing.T) {
		_, ok := parseSKU(Self{
			"Config": {"Version": "***"},
		})
		require.False(t, ok)
	})
}
