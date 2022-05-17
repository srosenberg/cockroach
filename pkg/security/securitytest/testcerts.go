package securitytest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/security"
)

func CreateTestCerts(certsDir string) (cleanup func() error) {
	__antithesis_instrumentation__.Notify(187239)

	security.ResetAssetLoader()

	assets := []string{
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedCACert),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedCAKey),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedNodeCert),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedNodeKey),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedRootCert),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedRootKey),
		filepath.Join(security.EmbeddedCertsDir, security.EmbeddedTenantCACert),
	}

	for _, a := range assets {
		__antithesis_instrumentation__.Notify(187241)
		_, err := RestrictedCopy(a, certsDir, filepath.Base(a))
		if err != nil {
			__antithesis_instrumentation__.Notify(187242)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(187243)
		}
	}
	__antithesis_instrumentation__.Notify(187240)

	return func() error {
		__antithesis_instrumentation__.Notify(187244)
		security.SetAssetLoader(EmbeddedAssets)
		return os.RemoveAll(certsDir)
	}
}
