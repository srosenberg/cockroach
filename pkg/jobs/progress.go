package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var (
	progressTimeThreshold             = 15 * time.Second
	progressFractionThreshold float32 = 0.05
)

func TestingSetProgressThresholds() func() {
	__antithesis_instrumentation__.Notify(84219)
	oldFraction := progressFractionThreshold
	oldDuration := progressTimeThreshold

	progressFractionThreshold = 0.0001
	progressTimeThreshold = time.Microsecond

	return func() {
		__antithesis_instrumentation__.Notify(84220)
		progressFractionThreshold = oldFraction
		progressTimeThreshold = oldDuration
	}
}

type ChunkProgressLogger struct {
	expectedChunks       int
	completedChunks      int
	perChunkContribution float32

	batcher ProgressUpdateBatcher
}

var ProgressUpdateOnly func(context.Context, jobspb.ProgressDetails)

func NewChunkProgressLogger(
	j *Job,
	expectedChunks int,
	startFraction float32,
	progressedFn func(context.Context, jobspb.ProgressDetails),
) *ChunkProgressLogger {
	__antithesis_instrumentation__.Notify(84221)
	return &ChunkProgressLogger{
		expectedChunks:       expectedChunks,
		perChunkContribution: (1.0 - startFraction) * 1.0 / float32(expectedChunks),
		batcher: ProgressUpdateBatcher{
			completed: startFraction,
			reported:  startFraction,
			Report: func(ctx context.Context, pct float32) error {
				__antithesis_instrumentation__.Notify(84222)
				return j.FractionProgressed(ctx, nil, func(ctx context.Context, details jobspb.ProgressDetails) float32 {
					__antithesis_instrumentation__.Notify(84223)
					if progressedFn != nil {
						__antithesis_instrumentation__.Notify(84225)
						progressedFn(ctx, details)
					} else {
						__antithesis_instrumentation__.Notify(84226)
					}
					__antithesis_instrumentation__.Notify(84224)
					return pct
				})
			},
		},
	}
}

func (jpl *ChunkProgressLogger) chunkFinished(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(84227)
	jpl.completedChunks++
	return jpl.batcher.Add(ctx, jpl.perChunkContribution)
}

func (jpl *ChunkProgressLogger) Loop(ctx context.Context, chunkCh <-chan struct{}) error {
	__antithesis_instrumentation__.Notify(84228)
	for {
		__antithesis_instrumentation__.Notify(84229)
		select {
		case _, ok := <-chunkCh:
			__antithesis_instrumentation__.Notify(84230)
			if !ok {
				__antithesis_instrumentation__.Notify(84234)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(84235)
			}
			__antithesis_instrumentation__.Notify(84231)
			if err := jpl.chunkFinished(ctx); err != nil {
				__antithesis_instrumentation__.Notify(84236)
				return err
			} else {
				__antithesis_instrumentation__.Notify(84237)
			}
			__antithesis_instrumentation__.Notify(84232)
			if jpl.completedChunks == jpl.expectedChunks {
				__antithesis_instrumentation__.Notify(84238)
				if err := jpl.batcher.Done(ctx); err != nil {
					__antithesis_instrumentation__.Notify(84239)
					return err
				} else {
					__antithesis_instrumentation__.Notify(84240)
				}
			} else {
				__antithesis_instrumentation__.Notify(84241)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(84233)
			return ctx.Err()
		}
	}
}

type ProgressUpdateBatcher struct {
	Report func(context.Context, float32) error

	syncutil.Mutex

	completed float32

	reported float32

	lastReported time.Time
}

func (p *ProgressUpdateBatcher) Add(ctx context.Context, delta float32) error {
	__antithesis_instrumentation__.Notify(84242)
	p.Lock()
	p.completed += delta
	completed := p.completed
	shouldReport := p.completed-p.reported > progressFractionThreshold
	shouldReport = shouldReport && func() bool {
		__antithesis_instrumentation__.Notify(84245)
		return p.lastReported.Add(progressTimeThreshold).Before(timeutil.Now()) == true
	}() == true

	if shouldReport {
		__antithesis_instrumentation__.Notify(84246)
		p.reported = p.completed
		p.lastReported = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(84247)
	}
	__antithesis_instrumentation__.Notify(84243)
	p.Unlock()

	if shouldReport {
		__antithesis_instrumentation__.Notify(84248)
		return p.Report(ctx, completed)
	} else {
		__antithesis_instrumentation__.Notify(84249)
	}
	__antithesis_instrumentation__.Notify(84244)
	return nil
}

func (p *ProgressUpdateBatcher) Done(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(84250)
	p.Lock()
	completed := p.completed
	shouldReport := completed-p.reported > progressFractionThreshold
	p.Unlock()

	if shouldReport {
		__antithesis_instrumentation__.Notify(84252)
		return p.Report(ctx, completed)
	} else {
		__antithesis_instrumentation__.Notify(84253)
	}
	__antithesis_instrumentation__.Notify(84251)
	return nil
}
