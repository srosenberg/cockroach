package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	istatus "google.golang.org/grpc/internal/status"
)

type Status = istatus.Status

func New(_ int, _ string) *Status {
	__antithesis_instrumentation__.Notify(644786)
	return &Status{}
}
