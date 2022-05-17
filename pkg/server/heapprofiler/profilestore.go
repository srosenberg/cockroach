package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var (
	maxProfiles = settings.RegisterIntSetting(
		settings.TenantWritable,
		"server.mem_profile.max_profiles",
		"maximum number of profiles to be kept per ramp-up of memory usage. "+
			"A ramp-up is defined as a sequence of profiles with increasing usage.",
		5,
	)

	maxCombinedFileSize = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"server.mem_profile.total_dump_size_limit",
		"maximum combined disk size of preserved memory profiles",
		128<<20,
	)
)

func init() {
	s := settings.RegisterIntSetting(
		settings.TenantWritable,
		"server.heap_profile.max_profiles", "use server.mem_profile.max_profiles instead", 5)
	s.SetRetired()

	b := settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"server.heap_profile.total_dump_size_limit",
		"use server.mem_profile.total_dump_size_limit instead",
		128<<20,
	)
	b.SetRetired()
}

type profileStore struct {
	*dumpstore.DumpStore
	prefix string
	suffix string
	st     *cluster.Settings
}

func newProfileStore(
	store *dumpstore.DumpStore, prefix, suffix string, st *cluster.Settings,
) *profileStore {
	__antithesis_instrumentation__.Notify(193602)
	s := &profileStore{DumpStore: store, prefix: prefix, suffix: suffix, st: st}
	return s
}

func (s *profileStore) gcProfiles(ctx context.Context, now time.Time) {
	__antithesis_instrumentation__.Notify(193603)
	s.GC(ctx, now, s)
}

func (s *profileStore) makeNewFileName(timestamp time.Time, curHeap int64) string {
	__antithesis_instrumentation__.Notify(193604)

	fileName := fmt.Sprintf("%s.%s.%d%s",
		s.prefix, timestamp.Format(timestampFormat), curHeap, s.suffix)
	return s.GetFullPath(fileName)
}

func (s *profileStore) PreFilter(
	ctx context.Context, files []os.FileInfo, cleanupFn func(fileName string) error,
) (preserved map[int]bool, _ error) {
	__antithesis_instrumentation__.Notify(193605)
	maxP := maxProfiles.Get(&s.st.SV)
	preserved = s.cleanupLastRampup(ctx, files, maxP, cleanupFn)
	return
}

func (s *profileStore) CheckOwnsFile(ctx context.Context, fi os.FileInfo) bool {
	__antithesis_instrumentation__.Notify(193606)
	ok, _, _ := s.parseFileName(ctx, fi.Name())
	return ok
}

func (s *profileStore) cleanupLastRampup(
	ctx context.Context, files []os.FileInfo, maxP int64, fn func(string) error,
) (preserved map[int]bool) {
	__antithesis_instrumentation__.Notify(193607)
	preserved = make(map[int]bool)
	curMaxHeap := uint64(math.MaxUint64)
	numFiles := int64(0)
	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(193609)
		ok, _, curHeap := s.parseFileName(ctx, files[i].Name())
		if !ok {
			__antithesis_instrumentation__.Notify(193612)
			continue
		} else {
			__antithesis_instrumentation__.Notify(193613)
		}
		__antithesis_instrumentation__.Notify(193610)

		if curHeap > curMaxHeap {
			__antithesis_instrumentation__.Notify(193614)

			break
		} else {
			__antithesis_instrumentation__.Notify(193615)
		}
		__antithesis_instrumentation__.Notify(193611)

		curMaxHeap = curHeap

		numFiles++

		if numFiles > maxP {
			__antithesis_instrumentation__.Notify(193616)

			if err := fn(files[i].Name()); err != nil {
				__antithesis_instrumentation__.Notify(193617)
				log.Warningf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(193618)
			}
		} else {
			__antithesis_instrumentation__.Notify(193619)

			preserved[i] = true
		}
	}
	__antithesis_instrumentation__.Notify(193608)

	return preserved
}

func (s *profileStore) parseFileName(
	ctx context.Context, fileName string,
) (ok bool, timestamp time.Time, heapUsage uint64) {
	__antithesis_instrumentation__.Notify(193620)
	parts := strings.Split(fileName, ".")
	numParts := 4
	if len(parts) < numParts || func() bool {
		__antithesis_instrumentation__.Notify(193625)
		return parts[0] != s.prefix == true
	}() == true {
		__antithesis_instrumentation__.Notify(193626)

		return
	} else {
		__antithesis_instrumentation__.Notify(193627)
	}
	__antithesis_instrumentation__.Notify(193621)
	if len(parts[2]) < 3 {
		__antithesis_instrumentation__.Notify(193628)

		parts[2] += "000"[:3-len(parts[2])]
	} else {
		__antithesis_instrumentation__.Notify(193629)
	}
	__antithesis_instrumentation__.Notify(193622)
	maybeTimestamp := parts[1] + "." + parts[2]
	var err error
	timestamp, err = time.Parse(timestampFormat, maybeTimestamp)
	if err != nil {
		__antithesis_instrumentation__.Notify(193630)
		log.Warningf(ctx, "%v", errors.Wrapf(err, "%s", fileName))
		return
	} else {
		__antithesis_instrumentation__.Notify(193631)
	}
	__antithesis_instrumentation__.Notify(193623)
	heapUsage, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(193632)
		log.Warningf(ctx, "%v", errors.Wrapf(err, "%s", fileName))
		return
	} else {
		__antithesis_instrumentation__.Notify(193633)
	}
	__antithesis_instrumentation__.Notify(193624)
	ok = true
	return
}
