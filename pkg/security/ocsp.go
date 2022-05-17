package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/ocsp"
	"golang.org/x/sync/errgroup"
)

func makeOCSPVerifier(settings TLSSettings) func([][]byte, [][]*x509.Certificate) error {
	__antithesis_instrumentation__.Notify(186666)
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		__antithesis_instrumentation__.Notify(186667)
		if !settings.ocspEnabled() {
			__antithesis_instrumentation__.Notify(186669)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(186670)
		}
		__antithesis_instrumentation__.Notify(186668)

		return contextutil.RunWithTimeout(context.Background(), "OCSP verification", settings.ocspTimeout(),
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(186671)

				telemetry.Inc(ocspChecksCounter)

				errG, gCtx := errgroup.WithContext(ctx)
				for _, chain := range verifiedChains {
					__antithesis_instrumentation__.Notify(186673)

					for i := 0; i < len(chain)-1; i++ {
						__antithesis_instrumentation__.Notify(186674)
						cert := chain[i]
						if len(cert.OCSPServer) > 0 {
							__antithesis_instrumentation__.Notify(186675)
							issuer := chain[i+1]
							errG.Go(func() error {
								__antithesis_instrumentation__.Notify(186676)
								return verifyOCSP(gCtx, settings, cert, issuer)
							})
						} else {
							__antithesis_instrumentation__.Notify(186677)
						}
					}
				}
				__antithesis_instrumentation__.Notify(186672)

				return errG.Wait()
			})
	}
}

var ocspChecksCounter = telemetry.GetCounterOnce("server.ocsp.conn-verifications")

var ocspCheckWithOCSPServerInCertCounter = telemetry.GetCounterOnce("server.ocsp.cert-verifications")

func verifyOCSP(ctx context.Context, settings TLSSettings, cert, issuer *x509.Certificate) error {
	__antithesis_instrumentation__.Notify(186678)
	if len(cert.OCSPServer) == 0 {
		__antithesis_instrumentation__.Notify(186682)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(186683)
	}
	__antithesis_instrumentation__.Notify(186679)

	telemetry.Inc(ocspCheckWithOCSPServerInCertCounter)

	var errs []error
	for _, url := range cert.OCSPServer {
		__antithesis_instrumentation__.Notify(186684)
		ok, err := queryOCSP(ctx, url, cert, issuer)
		if err != nil {
			__antithesis_instrumentation__.Notify(186687)
			errs = append(errs, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(186688)
		}
		__antithesis_instrumentation__.Notify(186685)
		if !ok {
			__antithesis_instrumentation__.Notify(186689)
			return errors.Newf("OCSP server says cert is revoked: %v", cert)
		} else {
			__antithesis_instrumentation__.Notify(186690)
		}
		__antithesis_instrumentation__.Notify(186686)
		return nil
	}
	__antithesis_instrumentation__.Notify(186680)

	if settings.ocspStrict() {
		__antithesis_instrumentation__.Notify(186691)
		switch len(errs) {
		case 0:
			__antithesis_instrumentation__.Notify(186692)
			panic("can't happen: OCSP failed but errs is empty")
		case 1:
			__antithesis_instrumentation__.Notify(186693)
			return errors.Wrap(errs[0], "OCSP check failed in strict mode")
		default:
			__antithesis_instrumentation__.Notify(186694)

			return errors.Wrap(errors.CombineErrors(errs[0], errs[1]), "OCSP check failed in strict mode")
		}
	} else {
		__antithesis_instrumentation__.Notify(186695)
	}
	__antithesis_instrumentation__.Notify(186681)

	log.Warningf(ctx, "OCSP check failed in non-strict mode: %v", errs)
	return nil
}

func queryOCSP(ctx context.Context, url string, cert, issuer *x509.Certificate) (bool, error) {
	__antithesis_instrumentation__.Notify(186696)
	ocspReq, err := ocsp.CreateRequest(cert, issuer, &ocsp.RequestOptions{

		Hash: crypto.SHA256,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(186705)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186706)
	}
	__antithesis_instrumentation__.Notify(186697)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(ocspReq))
	if err != nil {
		__antithesis_instrumentation__.Notify(186707)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186708)
	}
	__antithesis_instrumentation__.Notify(186698)
	httpReq.Header.Add("Content-Type", "application/ocsp-request")
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		__antithesis_instrumentation__.Notify(186709)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186710)
	}
	__antithesis_instrumentation__.Notify(186699)
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		__antithesis_instrumentation__.Notify(186711)
		return false, errors.Newf("OCSP server returned status code %v", errors.Safe(httpResp.StatusCode))
	} else {
		__antithesis_instrumentation__.Notify(186712)
	}
	__antithesis_instrumentation__.Notify(186700)
	if ct := httpResp.Header.Get("Content-Type"); ct != "application/ocsp-response" {
		__antithesis_instrumentation__.Notify(186713)
		return false, errors.Newf("OCSP server returned unexpected content-type %q", errors.Safe(ct))
	} else {
		__antithesis_instrumentation__.Notify(186714)
	}
	__antithesis_instrumentation__.Notify(186701)

	httpBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		__antithesis_instrumentation__.Notify(186715)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186716)
	}
	__antithesis_instrumentation__.Notify(186702)

	ocspResp, err := ocsp.ParseResponseForCert(httpBody, cert, issuer)
	if err != nil {
		__antithesis_instrumentation__.Notify(186717)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186718)
	}
	__antithesis_instrumentation__.Notify(186703)
	if ocspResp == nil {
		__antithesis_instrumentation__.Notify(186719)
		return false, errors.Newf("OCSP response for cert %v not found", cert)
	} else {
		__antithesis_instrumentation__.Notify(186720)
	}
	__antithesis_instrumentation__.Notify(186704)
	switch ocspResp.Status {
	case ocsp.Good:
		__antithesis_instrumentation__.Notify(186721)
		return true, nil
	case ocsp.Revoked:
		__antithesis_instrumentation__.Notify(186722)
		return false, nil
	default:
		__antithesis_instrumentation__.Notify(186723)
		return false, errors.Newf("OCSP returned status %v", errors.Safe(ocspResp.Status))
	}
}
