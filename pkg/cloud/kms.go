package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/errors"
)

type KMS interface {
	MasterKeyID() (string, error)

	Encrypt(ctx context.Context, data []byte) ([]byte, error)

	Decrypt(ctx context.Context, data []byte) ([]byte, error)

	Close() error
}

type KMSEnv interface {
	ClusterSettings() *cluster.Settings
	KMSConfig() *base.ExternalIODirConfig
}

type KMSFromURIFactory func(uri string, env KMSEnv) (KMS, error)

var kmsFactoryMap = make(map[string]KMSFromURIFactory)

func RegisterKMSFromURIFactory(factory KMSFromURIFactory, scheme string) {
	__antithesis_instrumentation__.Notify(36541)
	if _, ok := kmsFactoryMap[scheme]; ok {
		__antithesis_instrumentation__.Notify(36543)
		panic("factory method for " + scheme + " has already been registered")
	} else {
		__antithesis_instrumentation__.Notify(36544)
	}
	__antithesis_instrumentation__.Notify(36542)
	kmsFactoryMap[scheme] = factory
}

func KMSFromURI(uri string, env KMSEnv) (KMS, error) {
	__antithesis_instrumentation__.Notify(36545)
	var kmsURL *url.URL
	var err error
	if kmsURL, err = url.ParseRequestURI(uri); err != nil {
		__antithesis_instrumentation__.Notify(36548)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36549)
	}
	__antithesis_instrumentation__.Notify(36546)

	var factory KMSFromURIFactory
	var ok bool
	if factory, ok = kmsFactoryMap[kmsURL.Scheme]; !ok {
		__antithesis_instrumentation__.Notify(36550)
		return nil, errors.Newf("no factory method found for scheme %s", kmsURL.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(36551)
	}
	__antithesis_instrumentation__.Notify(36547)

	return factory(uri, env)
}
