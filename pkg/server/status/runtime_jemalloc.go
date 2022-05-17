//go:build !stdmalloc
// +build !stdmalloc

package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

// #cgo CPPFLAGS: -DJEMALLOC_NO_DEMANGLE
// #cgo LDFLAGS: -ljemalloc
// #cgo dragonfly freebsd LDFLAGS: -lm
// #cgo linux LDFLAGS: -lrt -lm -lpthread
//
// #include <jemalloc/jemalloc.h>
//
// // https://github.com/jemalloc/jemalloc/wiki/Use-Case:-Introspection-Via-mallctl*()
// // https://github.com/jemalloc/jemalloc/blob/4.5.0/src/stats.c#L960:L969
//
// typedef struct {
//   size_t Allocated;
//   size_t Active;
//   size_t Metadata;
//   size_t Resident;
//   size_t Mapped;
//   size_t Retained;
// } JemallocStats;
//
// int jemalloc_get_stats(JemallocStats *stats) {
//   // Update the statistics cached by je_mallctl.
//   uint64_t epoch = 1;
//   size_t sz = sizeof(epoch);
//   je_mallctl("epoch", &epoch, &sz, &epoch, sz);
//
//   int err;
//
//   sz = sizeof(&stats->Allocated);
//   err = je_mallctl("stats.allocated", &stats->Allocated, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   sz = sizeof(&stats->Active);
//   err = je_mallctl("stats.active", &stats->Active, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   sz = sizeof(&stats->Metadata);
//   err = je_mallctl("stats.metadata", &stats->Metadata, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   sz = sizeof(&stats->Resident);
//   err = je_mallctl("stats.resident", &stats->Resident, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   sz = sizeof(&stats->Mapped);
//   err = je_mallctl("stats.mapped", &stats->Mapped, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   sz = sizeof(&stats->Retained);
//   err = je_mallctl("stats.retained", &stats->Retained, &sz, NULL, 0);
//   if (err != 0) {
//     return err;
//   }
//   return err;
// }
import "C"

import (
	"context"
	"math"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/dustin/go-humanize"
)

func init() {
	if getCgoMemStats != nil {
		panic("getCgoMemStats is already set")
	}
	getCgoMemStats = getJemallocStats
}

func getJemallocStats(ctx context.Context) (uint, uint, error) {
	__antithesis_instrumentation__.Notify(235686)
	var js C.JemallocStats

	if _, err := C.jemalloc_get_stats(&js); err != nil {
		__antithesis_instrumentation__.Notify(235690)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(235691)
	}
	__antithesis_instrumentation__.Notify(235687)

	if log.V(2) {
		__antithesis_instrumentation__.Notify(235692)

		v := reflect.ValueOf(js)
		t := v.Type()
		stats := make([]string, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			__antithesis_instrumentation__.Notify(235694)
			stats[i] = t.Field(i).Name + ": " + humanize.IBytes(uint64(v.Field(i).Interface().(C.size_t)))
		}
		__antithesis_instrumentation__.Notify(235693)
		log.Infof(ctx, "jemalloc stats: %s", strings.Join(stats, " "))
	} else {
		__antithesis_instrumentation__.Notify(235695)
	}
	__antithesis_instrumentation__.Notify(235688)

	if log.V(3) && func() bool {
		__antithesis_instrumentation__.Notify(235696)
		return !log.V(math.MaxInt32) == true
	}() == true {
		__antithesis_instrumentation__.Notify(235697)

		C.je_malloc_stats_print(nil, nil, nil)
	} else {
		__antithesis_instrumentation__.Notify(235698)
	}
	__antithesis_instrumentation__.Notify(235689)

	return uint(js.Allocated), uint(js.Resident), nil
}

func allocateMemory() {
	__antithesis_instrumentation__.Notify(235699)

	C.malloc(256 << 10)
}
