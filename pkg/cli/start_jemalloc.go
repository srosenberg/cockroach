//go:build !stdmalloc
// +build !stdmalloc

package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

// #cgo CPPFLAGS: -DJEMALLOC_NO_DEMANGLE
// #cgo LDFLAGS: -ljemalloc
// #cgo dragonfly freebsd LDFLAGS: -lm
// #cgo linux LDFLAGS: -lrt -lm -lpthread
//
// #include <jemalloc/jemalloc.h>
// #include <stddef.h>
//
// // Checks whether jemalloc profiling is enabled and active.
// // Returns true if profiling is enabled and active.
// // Returns false on any mallctl errors.
// bool is_profiling_enabled() {
//   bool enabled = false;
//   size_t enabledSize = sizeof(enabled);
//
//   // Check profiling flag.
//   if (je_mallctl("opt.prof", &enabled, &enabledSize, NULL, 0) != 0) {
//     return false;
//   }
//   if (!enabled) {
//     return false;
//   }
//
//   // Check prof_active flag.
//   if (je_mallctl("opt.prof_active", &enabled, &enabledSize, NULL, 0) != 0) {
//     return false;
//   }
//   return enabled;
// }
//
// // Write a heap profile to "filename". Returns true on success, false on error.
// int dump_heap_profile(const char *filename) {
//   return je_mallctl("prof.dump", NULL, NULL, &filename, sizeof(const char *));
// }
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/server/heapprofiler"
)

func init() {
	if C.is_profiling_enabled() {
		heapprofiler.SetJemallocHeapDumpFn(writeJemallocProfile)
	}
}

func writeJemallocProfile(filename string) error {
	__antithesis_instrumentation__.Notify(34340)
	cpath := C.CString(filename)
	defer C.free(unsafe.Pointer(cpath))

	if errCode := C.dump_heap_profile(cpath); errCode != 0 {
		__antithesis_instrumentation__.Notify(34342)
		return fmt.Errorf("error code %d", errCode)
	} else {
		__antithesis_instrumentation__.Notify(34343)
	}
	__antithesis_instrumentation__.Notify(34341)
	return nil
}
