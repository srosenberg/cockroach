package debug

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
)

type CPUProfileOptions struct {
	Seconds int32

	WithLabels bool
}

func (opts CPUProfileOptions) Type() cluster.CPUProfileType {
	__antithesis_instrumentation__.Notify(190190)
	typ := cluster.CPUProfileDefault
	if opts.WithLabels {
		__antithesis_instrumentation__.Notify(190192)
		typ = cluster.CPUProfileWithLabels
	} else {
		__antithesis_instrumentation__.Notify(190193)
	}
	__antithesis_instrumentation__.Notify(190191)
	return typ
}

func CPUProfileOptionsFromRequest(r *http.Request) CPUProfileOptions {
	__antithesis_instrumentation__.Notify(190194)
	seconds, err := strconv.ParseInt(r.FormValue("seconds"), 10, 32)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(190196)
		return seconds <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(190197)
		seconds = 30
	} else {
		__antithesis_instrumentation__.Notify(190198)
	}
	__antithesis_instrumentation__.Notify(190195)

	withLabels := r.FormValue("labels") != "false"
	return CPUProfileOptions{
		Seconds:    int32(seconds),
		WithLabels: withLabels,
	}
}

func CPUProfileDo(st *cluster.Settings, typ cluster.CPUProfileType, do func() error) error {
	__antithesis_instrumentation__.Notify(190199)
	if err := st.SetCPUProfiling(typ); err != nil {
		__antithesis_instrumentation__.Notify(190202)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190203)
	}
	__antithesis_instrumentation__.Notify(190200)
	defer func() { __antithesis_instrumentation__.Notify(190204); _ = st.SetCPUProfiling(cluster.CPUProfileNone) }()
	__antithesis_instrumentation__.Notify(190201)
	return do()
}

func CPUProfileHandler(st *cluster.Settings, w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190205)
	opts := CPUProfileOptionsFromRequest(r)
	if err := CPUProfileDo(st, opts.Type(), func() error {
		__antithesis_instrumentation__.Notify(190206)
		pprof.Profile(w, r)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(190207)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(190208)
	}
}
