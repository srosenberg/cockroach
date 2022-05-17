package option

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachprod"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

type StartOpts struct {
	RoachprodOpts install.StartOpts
	RoachtestOpts struct {
		Worker      bool
		DontEncrypt bool
	}
}

func DefaultStartOpts() StartOpts {
	__antithesis_instrumentation__.Notify(44330)
	return StartOpts{RoachprodOpts: roachprod.DefaultStartOpts()}
}

type StopOpts struct {
	RoachprodOpts roachprod.StopOpts
	RoachtestOpts struct {
		Worker bool
	}
}

func DefaultStopOpts() StopOpts {
	__antithesis_instrumentation__.Notify(44331)
	return StopOpts{RoachprodOpts: roachprod.DefaultStopOpts()}
}
