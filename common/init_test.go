package common

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteDSNWithDefaults(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "plain path",
			dsn:  "data/one-api.db",
			want: "data/one-api.db?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_txlock=immediate",
		},
		{
			name: "unrelated existing query",
			dsn:  "file:one-api.db?cache=shared&mode=rwc",
			want: "file:one-api.db?cache=shared&mode=rwc&_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_txlock=immediate",
		},
		{
			name: "all explicit overrides",
			dsn:  "one-api.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(DELETE)&_txlock=deferred",
			want: "one-api.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(DELETE)&_txlock=deferred",
		},
		{
			name: "partial override",
			dsn:  "one-api.db?cache=shared&_pragma=journal_mode(TRUNCATE)",
			want: "one-api.db?cache=shared&_pragma=journal_mode(TRUNCATE)&_pragma=busy_timeout(30000)&_txlock=immediate",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, sqliteDSNWithDefaults(test.dsn))
		})
	}
}

func TestEnableInsecureSkipVerify(t *testing.T) {
	t.Run("clones existing TLS config", func(t *testing.T) {
		baseTLSConfig := &tls.Config{ServerName: "upstream.example"}
		transport := &http.Transport{TLSClientConfig: baseTLSConfig}

		enableInsecureSkipVerify(transport)

		require.NotSame(t, baseTLSConfig, transport.TLSClientConfig)
		assert.False(t, baseTLSConfig.InsecureSkipVerify)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
		assert.Equal(t, "upstream.example", transport.TLSClientConfig.ServerName)
	})

	t.Run("creates config when absent", func(t *testing.T) {
		transport := &http.Transport{}

		enableInsecureSkipVerify(transport)

		require.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}
