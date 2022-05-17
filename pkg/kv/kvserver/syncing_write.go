package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/time/rate"
)

const bulkIOWriteBurst = 512 << 10

const bulkIOWriteLimiterLongWait = 500 * time.Millisecond

func limitBulkIOWrite(ctx context.Context, limiter *rate.Limiter, cost int) error {
	__antithesis_instrumentation__.Notify(126606)

	if cost > bulkIOWriteBurst {
		__antithesis_instrumentation__.Notify(126610)
		cost = bulkIOWriteBurst
	} else {
		__antithesis_instrumentation__.Notify(126611)
	}
	__antithesis_instrumentation__.Notify(126607)

	begin := timeutil.Now()
	if err := limiter.WaitN(ctx, cost); err != nil {
		__antithesis_instrumentation__.Notify(126612)
		return errors.Wrapf(err, "error rate limiting bulk io write")
	} else {
		__antithesis_instrumentation__.Notify(126613)
	}
	__antithesis_instrumentation__.Notify(126608)

	if d := timeutil.Since(begin); d > bulkIOWriteLimiterLongWait {
		__antithesis_instrumentation__.Notify(126614)
		log.Warningf(ctx, "bulk io write limiter took %s (>%s):\n%s",
			d, bulkIOWriteLimiterLongWait, debug.Stack())
	} else {
		__antithesis_instrumentation__.Notify(126615)
	}
	__antithesis_instrumentation__.Notify(126609)
	return nil
}

var sstWriteSyncRate = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.bulk_sst.sync_size",
	"threshold after which non-Rocks SST writes must fsync (0 disables)",
	bulkIOWriteBurst,
)

func writeFileSyncing(
	ctx context.Context,
	filename string,
	data []byte,
	eng storage.Engine,
	perm os.FileMode,
	settings *cluster.Settings,
	limiter *rate.Limiter,
) error {
	__antithesis_instrumentation__.Notify(126616)
	chunkSize := sstWriteSyncRate.Get(&settings.SV)
	sync := true
	if chunkSize == 0 {
		__antithesis_instrumentation__.Notify(126621)
		chunkSize = bulkIOWriteBurst
		sync = false
	} else {
		__antithesis_instrumentation__.Notify(126622)
	}
	__antithesis_instrumentation__.Notify(126617)

	f, err := eng.Create(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(126623)
		if strings.Contains(err.Error(), "No such file or directory") {
			__antithesis_instrumentation__.Notify(126625)
			return os.ErrNotExist
		} else {
			__antithesis_instrumentation__.Notify(126626)
		}
		__antithesis_instrumentation__.Notify(126624)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126627)
	}
	__antithesis_instrumentation__.Notify(126618)

	for i := int64(0); i < int64(len(data)); i += chunkSize {
		__antithesis_instrumentation__.Notify(126628)
		end := i + chunkSize
		if l := int64(len(data)); end > l {
			__antithesis_instrumentation__.Notify(126632)
			end = l
		} else {
			__antithesis_instrumentation__.Notify(126633)
		}
		__antithesis_instrumentation__.Notify(126629)
		chunk := data[i:end]

		if err = limitBulkIOWrite(ctx, limiter, len(chunk)); err != nil {
			__antithesis_instrumentation__.Notify(126634)
			break
		} else {
			__antithesis_instrumentation__.Notify(126635)
		}
		__antithesis_instrumentation__.Notify(126630)
		if _, err = f.Write(chunk); err != nil {
			__antithesis_instrumentation__.Notify(126636)
			break
		} else {
			__antithesis_instrumentation__.Notify(126637)
		}
		__antithesis_instrumentation__.Notify(126631)
		if sync {
			__antithesis_instrumentation__.Notify(126638)
			if err = f.Sync(); err != nil {
				__antithesis_instrumentation__.Notify(126639)
				break
			} else {
				__antithesis_instrumentation__.Notify(126640)
			}
		} else {
			__antithesis_instrumentation__.Notify(126641)
		}
	}
	__antithesis_instrumentation__.Notify(126619)

	closeErr := f.Close()
	if err == nil {
		__antithesis_instrumentation__.Notify(126642)
		err = closeErr
	} else {
		__antithesis_instrumentation__.Notify(126643)
	}
	__antithesis_instrumentation__.Notify(126620)
	return err
}
