package certmgr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"io/ioutil"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var _ Cert = (*FileCert)(nil)

type FileCert struct {
	syncutil.Mutex
	certFile string
	keyFile  string
	err      error
	cert     *tls.Certificate
}

func NewFileCert(certFile, keyFile string) *FileCert {
	__antithesis_instrumentation__.Notify(186369)
	return &FileCert{
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (fc *FileCert) Reload(ctx context.Context) {
	__antithesis_instrumentation__.Notify(186370)
	fc.Lock()
	defer fc.Unlock()

	if fc.err != nil {
		__antithesis_instrumentation__.Notify(186375)
		return
	} else {
		__antithesis_instrumentation__.Notify(186376)
	}
	__antithesis_instrumentation__.Notify(186371)

	certBytes, err := ioutil.ReadFile(fc.certFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(186377)
		fc.err = errors.Wrapf(err, "could not reload cert file %s", fc.certFile)
		return
	} else {
		__antithesis_instrumentation__.Notify(186378)
	}
	__antithesis_instrumentation__.Notify(186372)

	keyBytes, err := ioutil.ReadFile(fc.keyFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(186379)
		fc.err = errors.Wrapf(err, "could not reload cert key file %s", fc.keyFile)
		return
	} else {
		__antithesis_instrumentation__.Notify(186380)
	}
	__antithesis_instrumentation__.Notify(186373)

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(186381)
		fc.err = errors.Wrapf(err, "could not construct cert from cert %s and key %s", fc.certFile, fc.keyFile)
		return
	} else {
		__antithesis_instrumentation__.Notify(186382)
	}
	__antithesis_instrumentation__.Notify(186374)

	fc.cert = &cert
}

func (fc *FileCert) Err() error {
	__antithesis_instrumentation__.Notify(186383)
	fc.Lock()
	defer fc.Unlock()
	return fc.err
}

func (fc *FileCert) ClearErr() {
	__antithesis_instrumentation__.Notify(186384)
	fc.Lock()
	defer fc.Unlock()
	fc.err = nil
}

func (fc *FileCert) TLSCert() *tls.Certificate {
	__antithesis_instrumentation__.Notify(186385)
	return fc.cert
}
