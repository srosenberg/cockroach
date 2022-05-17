package certmgr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var _ Cert = (*SelfSignedCert)(nil)

type SelfSignedCert struct {
	syncutil.Mutex
	years, months, days int
	secs                time.Duration
	err                 error
	cert                *tls.Certificate
}

func NewSelfSignedCert(years, months, days int, secs time.Duration) *SelfSignedCert {
	__antithesis_instrumentation__.Notify(186386)
	return &SelfSignedCert{years: years, months: months, days: days, secs: secs}
}

func (ssc *SelfSignedCert) Reload(ctx context.Context) {
	__antithesis_instrumentation__.Notify(186387)
	ssc.Lock()
	defer ssc.Unlock()

	if ssc.err != nil {
		__antithesis_instrumentation__.Notify(186391)
		return
	} else {
		__antithesis_instrumentation__.Notify(186392)
	}
	__antithesis_instrumentation__.Notify(186388)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(186393)
		ssc.err = errors.Wrapf(err, "could not generate key")
		return
	} else {
		__antithesis_instrumentation__.Notify(186394)
	}
	__antithesis_instrumentation__.Notify(186389)

	from := timeutil.Now()
	until := from.AddDate(ssc.years, ssc.months, ssc.days).Add(ssc.secs)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    from,
		NotAfter:     until,
		DNSNames:     []string{"localhost"},
		IsCA:         true,
	}
	cer, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		__antithesis_instrumentation__.Notify(186395)
		ssc.err = errors.Wrapf(err, "could not create certificate")
	} else {
		__antithesis_instrumentation__.Notify(186396)
	}
	__antithesis_instrumentation__.Notify(186390)

	ssc.cert = &tls.Certificate{
		Certificate: [][]byte{cer},
		PrivateKey:  priv,
	}
}

func (ssc *SelfSignedCert) Err() error {
	__antithesis_instrumentation__.Notify(186397)
	ssc.Lock()
	defer ssc.Unlock()
	return ssc.err
}

func (ssc *SelfSignedCert) ClearErr() {
	__antithesis_instrumentation__.Notify(186398)
	ssc.Lock()
	defer ssc.Unlock()
	ssc.err = nil
}

func (ssc *SelfSignedCert) TLSCert() *tls.Certificate {
	__antithesis_instrumentation__.Notify(186399)
	return ssc.cert
}
