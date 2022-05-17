package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"crypto/x509"
	"math"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
)

func GetAddJoinDialOptions(certPool *x509.CertPool) []grpc.DialOption {
	__antithesis_instrumentation__.Notify(183980)

	var dialOpts []grpc.DialOption

	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
		grpc.MaxCallSendMsgSize(math.MaxInt32),
	))
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(grpc.UseCompressor((snappyCompressor{}).Name())))
	dialOpts = append(dialOpts, grpc.WithNoProxy())
	backoffConfig := backoff.DefaultConfig
	backoffConfig.MaxDelay = maxBackoff
	dialOpts = append(dialOpts, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoffConfig,
		MinConnectTimeout: minConnectionTimeout}))
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(clientKeepalive))
	dialOpts = append(dialOpts,
		grpc.WithInitialWindowSize(initialWindowSize),
		grpc.WithInitialConnWindowSize(initialConnWindowSize))

	var tlsConf tls.Config
	if certPool != nil {
		__antithesis_instrumentation__.Notify(183982)
		tlsConf = tls.Config{
			RootCAs: certPool,
		}
	} else {
		__antithesis_instrumentation__.Notify(183983)

		tlsConf = tls.Config{InsecureSkipVerify: true}
	}
	__antithesis_instrumentation__.Notify(183981)

	creds := credentials.NewTLS(&tlsConf)
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))

	return dialOpts
}
