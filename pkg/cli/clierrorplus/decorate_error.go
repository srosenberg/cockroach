package clierrorplus

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var reGRPCConnRefused = regexp.MustCompile(`Error while dialing dial tcp .*: connection.* refused`)

var reGRPCNoTLS = regexp.MustCompile(`authentication handshake failed: tls: first record does not look like a TLS handshake`)

var reGRPCAuthFailure = regexp.MustCompile(`authentication handshake failed: x509`)

var reGRPCConnFailed = regexp.MustCompile(`desc = (transport is closing|all SubConns are in TransientFailure)`)

func MaybeDecorateError(
	wrapped func(*cobra.Command, []string) error,
) func(*cobra.Command, []string) error {
	__antithesis_instrumentation__.Notify(28352)
	return func(cmd *cobra.Command, args []string) (err error) {
		__antithesis_instrumentation__.Notify(28353)
		err = wrapped(cmd, args)

		if err == nil {
			__antithesis_instrumentation__.Notify(28372)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(28373)
		}
		__antithesis_instrumentation__.Notify(28354)

		defer func() {
			__antithesis_instrumentation__.Notify(28374)
			err = clierror.NewFormattedError(err, true, false)
		}()
		__antithesis_instrumentation__.Notify(28355)

		connFailed := func() error {
			__antithesis_instrumentation__.Notify(28375)

			const format = "cannot dial server.\n" +
				"Is the server running?\n" +
				"If the server is running, check --host client-side and --advertise server-side.\n\n%v"
			return errors.Errorf(format, err)
		}
		__antithesis_instrumentation__.Notify(28356)

		connSecurityHint := func() error {
			__antithesis_instrumentation__.Notify(28376)

			const format = "SSL authentication error while connecting.\n%v"
			return errors.Errorf(format, err)
		}
		__antithesis_instrumentation__.Notify(28357)

		connInsecureHint := func() error {
			__antithesis_instrumentation__.Notify(28377)

			const format = "cannot establish secure connection to insecure server.\n" +
				"Maybe use --insecure?\n\n%v"
			return errors.Errorf(format, err)
		}
		__antithesis_instrumentation__.Notify(28358)

		connRefused := func() error {
			__antithesis_instrumentation__.Notify(28378)

			const format = "server closed the connection.\n" +
				"Is this a CockroachDB node?\n%v"
			return errors.Errorf(format, err)
		}
		__antithesis_instrumentation__.Notify(28359)

		if errors.Is(err, pq.ErrSSLNotSupported) {
			__antithesis_instrumentation__.Notify(28379)

			return connInsecureHint()
		} else {
			__antithesis_instrumentation__.Notify(28380)
		}
		__antithesis_instrumentation__.Notify(28360)

		if wErr := (*security.Error)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28381)

			const format = "cannot load certificates.\n" +
				"Check your certificate settings, set --certs-dir, or use --insecure for insecure clusters.\n\n%v"
			return errors.Errorf(format, err)
		} else {
			__antithesis_instrumentation__.Notify(28382)
		}
		__antithesis_instrumentation__.Notify(28361)

		if wErr := (*x509.UnknownAuthorityError)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28383)

			return connSecurityHint()
		} else {
			__antithesis_instrumentation__.Notify(28384)
		}
		__antithesis_instrumentation__.Notify(28362)

		if wErr := (*clisqlclient.InitialSQLConnectionError)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28385)

			return connRefused()
		} else {
			__antithesis_instrumentation__.Notify(28386)
		}
		__antithesis_instrumentation__.Notify(28363)

		if wErr := (*pq.Error)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28387)

			if pgcode.MakeCode(string(wErr.Code)) == pgcode.ProtocolViolation {
				__antithesis_instrumentation__.Notify(28390)
				return connSecurityHint()
			} else {
				__antithesis_instrumentation__.Notify(28391)
			}
			__antithesis_instrumentation__.Notify(28388)

			if strings.Contains(wErr.Message, "column \"membership\" does not exist") {
				__antithesis_instrumentation__.Notify(28392)

				return fmt.Errorf("cannot use a v20.2 cli against servers running v20.1")
			} else {
				__antithesis_instrumentation__.Notify(28393)
			}
			__antithesis_instrumentation__.Notify(28389)

			return err
		} else {
			__antithesis_instrumentation__.Notify(28394)
		}
		__antithesis_instrumentation__.Notify(28364)

		if wErr := (*net.OpError)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28395)

			if msg := wErr.Err.Error(); strings.HasPrefix(msg, "tls: ") {
				__antithesis_instrumentation__.Notify(28397)

				return connSecurityHint()
			} else {
				__antithesis_instrumentation__.Notify(28398)
			}
			__antithesis_instrumentation__.Notify(28396)
			return connFailed()
		} else {
			__antithesis_instrumentation__.Notify(28399)
		}
		__antithesis_instrumentation__.Notify(28365)

		if wErr := (*netutil.InitialHeartbeatFailedError)(nil); errors.As(err, &wErr) {
			__antithesis_instrumentation__.Notify(28400)

			msg := wErr.Error()
			if reGRPCConnRefused.MatchString(msg) {
				__antithesis_instrumentation__.Notify(28404)
				return connFailed()
			} else {
				__antithesis_instrumentation__.Notify(28405)
			}
			__antithesis_instrumentation__.Notify(28401)
			if reGRPCNoTLS.MatchString(msg) {
				__antithesis_instrumentation__.Notify(28406)
				return connInsecureHint()
			} else {
				__antithesis_instrumentation__.Notify(28407)
			}
			__antithesis_instrumentation__.Notify(28402)
			if reGRPCAuthFailure.MatchString(msg) {
				__antithesis_instrumentation__.Notify(28408)
				return connSecurityHint()
			} else {
				__antithesis_instrumentation__.Notify(28409)
			}
			__antithesis_instrumentation__.Notify(28403)
			if reGRPCConnFailed.MatchString(msg) || func() bool {
				__antithesis_instrumentation__.Notify(28410)
				return status.Code(errors.Cause(err)) == codes.Unavailable == true
			}() == true {
				__antithesis_instrumentation__.Notify(28411)
				return connRefused()
			} else {
				__antithesis_instrumentation__.Notify(28412)
			}

		} else {
			__antithesis_instrumentation__.Notify(28413)
		}
		__antithesis_instrumentation__.Notify(28366)

		opTimeout := func() error {
			__antithesis_instrumentation__.Notify(28414)

			const format = "operation timed out.\n\n%v"
			return errors.Errorf(format, err)
		}
		__antithesis_instrumentation__.Notify(28367)

		if errors.IsAny(err,
			context.DeadlineExceeded,
			context.Canceled) {
			__antithesis_instrumentation__.Notify(28415)
			return opTimeout()
		} else {
			__antithesis_instrumentation__.Notify(28416)
		}
		__antithesis_instrumentation__.Notify(28368)

		if code := status.Code(errors.Cause(err)); code == codes.DeadlineExceeded {
			__antithesis_instrumentation__.Notify(28417)
			return opTimeout()
		} else {
			__antithesis_instrumentation__.Notify(28418)
			if code == codes.Unimplemented && func() bool {
				__antithesis_instrumentation__.Notify(28419)
				return strings.Contains(err.Error(), "unknown method Decommission") == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(28420)
				return strings.Contains(err.Error(), "unknown service cockroach.server.serverpb.Init") == true
			}() == true {
				__antithesis_instrumentation__.Notify(28421)
				return fmt.Errorf(
					"incompatible client and server versions (likely server version: v1.0, required: >=v1.1)")
			} else {
				__antithesis_instrumentation__.Notify(28422)
				if grpcutil.IsClosedConnection(err) {
					__antithesis_instrumentation__.Notify(28423)

					const format = "connection lost.\n\n%v"
					return errors.Errorf(format, err)
				} else {
					__antithesis_instrumentation__.Notify(28424)
				}
			}
		}
		__antithesis_instrumentation__.Notify(28369)

		if strings.Contains(err.Error(), "pq: unknown authentication response: 7") {
			__antithesis_instrumentation__.Notify(28425)
			return fmt.Errorf(
				"server requires GSSAPI authentication for this user.\n" +
					"The CockroachDB CLI does not support GSSAPI authentication; use 'psql' instead")
		} else {
			__antithesis_instrumentation__.Notify(28426)
		}
		__antithesis_instrumentation__.Notify(28370)

		if strings.Contains(err.Error(), server.ErrClusterInitialized.Error()) {
			__antithesis_instrumentation__.Notify(28427)

			return server.ErrClusterInitialized
		} else {
			__antithesis_instrumentation__.Notify(28428)
		}
		__antithesis_instrumentation__.Notify(28371)

		return err
	}
}
