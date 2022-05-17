package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/stretchr/testify/require"
)

type TestKMSEnv struct {
	Settings         *cluster.Settings
	ExternalIOConfig *base.ExternalIODirConfig
}

var _ KMSEnv = &TestKMSEnv{}

func (e *TestKMSEnv) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36552)
	return e.Settings
}

func (e *TestKMSEnv) KMSConfig() *base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36553)
	return e.ExternalIOConfig
}

func KMSEncryptDecrypt(t *testing.T, kmsURI string, env TestKMSEnv) {
	__antithesis_instrumentation__.Notify(36554)
	ctx := context.Background()
	kms, err := KMSFromURI(kmsURI, &env)
	require.NoError(t, err)

	t.Run("simple encrypt decrypt", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(36555)
		sampleBytes := "hello world"

		encryptedBytes, err := kms.Encrypt(ctx, []byte(sampleBytes))
		require.NoError(t, err)

		decryptedBytes, err := kms.Decrypt(ctx, encryptedBytes)
		require.NoError(t, err)

		require.True(t, bytes.Equal(decryptedBytes, []byte(sampleBytes)))

		require.NoError(t, kms.Close())
	})
}
