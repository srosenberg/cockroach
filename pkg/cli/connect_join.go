package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const nodeJoinTimeout = 1 * time.Minute

var connectJoinCmd = &cobra.Command{
	Use:   "join <join-token>",
	Short: "request the TLS certs for a new node from an existing node",
	Args:  cobra.MinimumNArgs(1),
	RunE:  clierrorplus.MaybeDecorateError(runConnectJoin),
}

func requestPeerCA(
	ctx context.Context, stopper *stop.Stopper, peer string, jt security.JoinToken,
) (*x509.CertPool, error) {
	__antithesis_instrumentation__.Notify(30286)
	dialOpts := rpc.GetAddJoinDialOptions(nil)

	conn, err := grpc.DialContext(ctx, peer, dialOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(30293)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30294)
	}
	__antithesis_instrumentation__.Notify(30287)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(30295)
		_ = conn.Close()
	}))
	__antithesis_instrumentation__.Notify(30288)

	s := serverpb.NewAdminClient(conn)

	req := serverpb.CARequest{}
	resp, err := s.RequestCA(ctx, &req)
	if err != nil {
		__antithesis_instrumentation__.Notify(30296)
		return nil, errors.Wrap(
			err, "failed grpc call to request CA from peer")
	} else {
		__antithesis_instrumentation__.Notify(30297)
	}
	__antithesis_instrumentation__.Notify(30289)

	if !jt.VerifySignature(resp.CaCert) {
		__antithesis_instrumentation__.Notify(30298)
		return nil, errors.New("resp.CaCert failed cryptologic validation")
	} else {
		__antithesis_instrumentation__.Notify(30299)
	}
	__antithesis_instrumentation__.Notify(30290)

	pemBlock, _ := pem.Decode(resp.CaCert)
	if pemBlock == nil {
		__antithesis_instrumentation__.Notify(30300)
		return nil, errors.New("failed to parse valid PEM from resp.CaCert")
	} else {
		__antithesis_instrumentation__.Notify(30301)
	}
	__antithesis_instrumentation__.Notify(30291)
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(30302)
		return nil, errors.New("failed to parse valid x509 cert from resp.CaCert")
	} else {
		__antithesis_instrumentation__.Notify(30303)
	}
	__antithesis_instrumentation__.Notify(30292)
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	return certPool, nil
}

func requestCertBundle(
	ctx context.Context,
	stopper *stop.Stopper,
	peerAddr string,
	certPool *x509.CertPool,
	jt security.JoinToken,
) (*server.CertificateBundle, error) {
	__antithesis_instrumentation__.Notify(30304)
	dialOpts := rpc.GetAddJoinDialOptions(certPool)

	conn, err := grpc.DialContext(ctx, peerAddr, dialOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(30309)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30310)
	}
	__antithesis_instrumentation__.Notify(30305)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(30311)
		_ = conn.Close()
	}))
	__antithesis_instrumentation__.Notify(30306)

	s := serverpb.NewAdminClient(conn)
	req := serverpb.CertBundleRequest{
		TokenID:      jt.TokenID.String(),
		SharedSecret: jt.SharedSecret,
	}
	resp, err := s.RequestCertBundle(ctx, &req)
	if err != nil {
		__antithesis_instrumentation__.Notify(30312)
		return nil, errors.Wrapf(
			err,
			"failed to RequestCertBundle from %q",
			peerAddr,
		)
	} else {
		__antithesis_instrumentation__.Notify(30313)
	}
	__antithesis_instrumentation__.Notify(30307)

	var certBundle server.CertificateBundle
	err = json.Unmarshal(resp.Bundle, &certBundle)
	if err != nil {
		__antithesis_instrumentation__.Notify(30314)
		return nil, errors.Wrapf(
			err,
			"failed to unmarshal CertBundle from %q",
			peerAddr,
		)
	} else {
		__antithesis_instrumentation__.Notify(30315)
	}
	__antithesis_instrumentation__.Notify(30308)

	return &certBundle, nil
}

func runConnectJoin(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30316)
	return contextutil.RunWithTimeout(context.Background(), "init handshake", nodeJoinTimeout, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(30317)
		ctx = logtags.AddTag(ctx, "init-tls-handshake", nil)

		stopper := stop.NewStopper()
		defer stopper.Stop(ctx)

		if err := validateNodeJoinFlags(cmd); err != nil {
			__antithesis_instrumentation__.Notify(30321)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30322)
		}
		__antithesis_instrumentation__.Notify(30318)

		joinTokenArg := args[0]
		var jt security.JoinToken
		if err := jt.UnmarshalText([]byte(joinTokenArg)); err != nil {
			__antithesis_instrumentation__.Notify(30323)
			return errors.Wrap(err, "failed to parse join token")
		} else {
			__antithesis_instrumentation__.Notify(30324)
		}
		__antithesis_instrumentation__.Notify(30319)

		for _, peer := range serverCfg.JoinList {
			__antithesis_instrumentation__.Notify(30325)
			certPool, err := requestPeerCA(ctx, stopper, peer, jt)
			if err != nil {
				__antithesis_instrumentation__.Notify(30329)

				log.Errorf(ctx, "failed requesting peer CA from %s: %s", peer, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(30330)
			}
			__antithesis_instrumentation__.Notify(30326)

			certBundle, err := requestCertBundle(ctx, stopper, peer, certPool, jt)
			if err != nil {
				__antithesis_instrumentation__.Notify(30331)
				return errors.Wrapf(err,
					"failed requesting certBundle from peer %q, token may have been consumed", peer)
			} else {
				__antithesis_instrumentation__.Notify(30332)
			}
			__antithesis_instrumentation__.Notify(30327)

			err = certBundle.InitializeNodeFromBundle(ctx, *baseCfg)
			if err != nil {
				__antithesis_instrumentation__.Notify(30333)
				return errors.Wrap(
					err,
					"failed to initialize node after consuming join-token",
				)
			} else {
				__antithesis_instrumentation__.Notify(30334)
			}
			__antithesis_instrumentation__.Notify(30328)
			return nil
		}
		__antithesis_instrumentation__.Notify(30320)

		return errors.New("could not successfully authenticate with any listed nodes")
	})
}

func validateNodeJoinFlags(_ *cobra.Command) error {
	__antithesis_instrumentation__.Notify(30335)
	if len(serverCfg.JoinList) == 0 {
		__antithesis_instrumentation__.Notify(30337)
		return errors.Newf("flag --%s must specify address of at least one node to join",
			cliflags.Join.Name)
	} else {
		__antithesis_instrumentation__.Notify(30338)
	}
	__antithesis_instrumentation__.Notify(30336)
	return nil
}
