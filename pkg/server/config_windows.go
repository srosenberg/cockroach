package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/storage"

func setOpenFileLimitInner(physicalStoreCount int) (uint64, error) {
	__antithesis_instrumentation__.Notify(190189)
	return storage.RecommendedMaxOpenFiles, nil
}
