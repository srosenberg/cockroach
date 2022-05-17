package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type RangeCounter interface {
	NumRangesInSpanContainedBy(
		ctx context.Context, span roachpb.Span, containedBy []roachpb.Span,
	) (total, inContainedBy int, _ error)
}

func newBackfillTrackerConfig(
	ctx context.Context, codec keys.SQLCodec, db *kv.DB, rc RangeCounter, job *jobs.Job,
) backfillTrackerConfig {
	__antithesis_instrumentation__.Notify(580440)
	return backfillTrackerConfig{
		numRangesInSpanContainedBy: rc.NumRangesInSpanContainedBy,
		writeProgressFraction: func(_ context.Context, fractionProgressed float32) error {
			__antithesis_instrumentation__.Notify(580441)
			if err := job.FractionProgressed(
				ctx, nil, jobs.FractionUpdater(fractionProgressed),
			); err != nil {
				__antithesis_instrumentation__.Notify(580443)
				return jobs.SimplifyInvalidStatusError(err)
			} else {
				__antithesis_instrumentation__.Notify(580444)
			}
			__antithesis_instrumentation__.Notify(580442)
			return nil
		},
		writeCheckpoint: func(ctx context.Context, progresses []scexec.BackfillProgress) error {
			__antithesis_instrumentation__.Notify(580445)
			return job.Update(ctx, nil, func(
				txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater,
			) error {
				__antithesis_instrumentation__.Notify(580446)
				pl := md.Payload
				jobProgress, err := convertToJobBackfillProgress(codec, progresses)
				if err != nil {
					__antithesis_instrumentation__.Notify(580448)
					return err
				} else {
					__antithesis_instrumentation__.Notify(580449)
				}
				__antithesis_instrumentation__.Notify(580447)
				sc := pl.GetNewSchemaChange()
				sc.BackfillProgress = jobProgress
				ju.UpdatePayload(pl)
				return nil
			})
		},
	}
}

func convertToJobBackfillProgress(
	codec keys.SQLCodec, progresses []scexec.BackfillProgress,
) ([]jobspb.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(580450)
	ret := make([]jobspb.BackfillProgress, 0, len(progresses))
	for _, bp := range progresses {
		__antithesis_instrumentation__.Notify(580452)
		strippedSpans, err := removeTenantPrefixFromSpans(codec, bp.CompletedSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(580454)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(580455)
		}
		__antithesis_instrumentation__.Notify(580453)
		ret = append(ret, jobspb.BackfillProgress{
			TableID:        bp.TableID,
			SourceIndexID:  bp.SourceIndexID,
			DestIndexIDs:   bp.DestIndexIDs,
			WriteTimestamp: bp.MinimumWriteTimestamp,
			CompletedSpans: strippedSpans,
		})
	}
	__antithesis_instrumentation__.Notify(580451)
	return ret, nil
}

func convertFromJobBackfillProgress(
	codec keys.SQLCodec, progresses []jobspb.BackfillProgress,
) []scexec.BackfillProgress {
	__antithesis_instrumentation__.Notify(580456)
	ret := make([]scexec.BackfillProgress, 0, len(progresses))
	for _, bp := range progresses {
		__antithesis_instrumentation__.Notify(580458)
		ret = append(ret, scexec.BackfillProgress{
			Backfill: scexec.Backfill{
				TableID:       bp.TableID,
				SourceIndexID: bp.SourceIndexID,
				DestIndexIDs:  bp.DestIndexIDs,
			},
			MinimumWriteTimestamp: bp.WriteTimestamp,
			CompletedSpans:        addTenantPrefixToSpans(codec, bp.CompletedSpans),
		})
	}
	__antithesis_instrumentation__.Notify(580457)
	return ret
}

func addTenantPrefixToSpans(codec keys.SQLCodec, spans []roachpb.Span) []roachpb.Span {
	__antithesis_instrumentation__.Notify(580459)
	prefix := codec.TenantPrefix()
	prefix = prefix[:len(prefix):len(prefix)]
	ret := make([]roachpb.Span, 0, len(spans))
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(580461)
		ret = append(ret, roachpb.Span{
			Key:    append(prefix, sp.Key...),
			EndKey: append(prefix, sp.EndKey...),
		})
	}
	__antithesis_instrumentation__.Notify(580460)
	return ret
}

func removeTenantPrefixFromSpans(
	codec keys.SQLCodec, spans []roachpb.Span,
) (ret []roachpb.Span, err error) {
	__antithesis_instrumentation__.Notify(580462)
	ret = make([]roachpb.Span, 0, len(spans))
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(580464)
		var stripped roachpb.Span
		if stripped.Key, err = codec.StripTenantPrefix(sp.Key); err != nil {
			__antithesis_instrumentation__.Notify(580467)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(580468)
		}
		__antithesis_instrumentation__.Notify(580465)
		if stripped.EndKey, err = codec.StripTenantPrefix(sp.EndKey); err != nil {
			__antithesis_instrumentation__.Notify(580469)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(580470)
		}
		__antithesis_instrumentation__.Notify(580466)
		ret = append(ret, stripped)
	}
	__antithesis_instrumentation__.Notify(580463)
	return ret, nil
}

type backfillTracker struct {
	backfillTrackerConfig
	codec keys.SQLCodec

	mu struct {
		syncutil.Mutex

		progress map[tableIndexKey]*progress
	}
}

type backfillTrackerConfig struct {
	numRangesInSpanContainedBy func(
		context.Context, roachpb.Span, []roachpb.Span,
	) (total, contained int, _ error)

	writeProgressFraction func(_ context.Context, fractionProgressed float32) error

	writeCheckpoint func(context.Context, []scexec.BackfillProgress) error
}

var _ scexec.BackfillTracker = (*backfillTracker)(nil)

func newBackfillTracker(
	codec keys.SQLCodec, cfg backfillTrackerConfig, initialProgress []scexec.BackfillProgress,
) *backfillTracker {
	__antithesis_instrumentation__.Notify(580471)
	bt := &backfillTracker{
		codec:                 codec,
		backfillTrackerConfig: cfg,
	}
	bt.mu.progress = make(map[tableIndexKey]*progress)
	for _, p := range initialProgress {
		__antithesis_instrumentation__.Notify(580473)
		bp := newProgress(codec, p)

		bt.mu.progress[toKey(p.Backfill)] = bp
	}
	__antithesis_instrumentation__.Notify(580472)
	return bt
}

func (b *backfillTracker) GetBackfillProgress(
	ctx context.Context, bf scexec.Backfill,
) (scexec.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(580474)
	p, ok := b.getTableIndexBackfillProgress(bf)
	if !ok {
		__antithesis_instrumentation__.Notify(580477)
		return scexec.BackfillProgress{
			Backfill: bf,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(580478)
	}
	__antithesis_instrumentation__.Notify(580475)
	if err := p.matches(bf); err != nil {
		__antithesis_instrumentation__.Notify(580479)
		return scexec.BackfillProgress{}, err
	} else {
		__antithesis_instrumentation__.Notify(580480)
	}
	__antithesis_instrumentation__.Notify(580476)
	return p.BackfillProgress, nil
}

func (b *backfillTracker) SetBackfillProgress(
	ctx context.Context, progress scexec.BackfillProgress,
) error {
	__antithesis_instrumentation__.Notify(580481)
	b.mu.Lock()
	defer b.mu.Unlock()
	p, err := b.getBackfillProgressLocked(progress.Backfill)
	if err != nil {
		__antithesis_instrumentation__.Notify(580484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(580485)
	}
	__antithesis_instrumentation__.Notify(580482)
	if p == nil {
		__antithesis_instrumentation__.Notify(580486)
		p = newProgress(b.codec, progress)
		b.mu.progress[toKey(progress.Backfill)] = p
	} else {
		__antithesis_instrumentation__.Notify(580487)
		p.BackfillProgress = progress
	}
	__antithesis_instrumentation__.Notify(580483)
	p.needsCheckpointFlush = true
	p.needsFractionFlush = true
	return nil
}

func newProgress(codec keys.SQLCodec, bp scexec.BackfillProgress) *progress {
	__antithesis_instrumentation__.Notify(580488)
	indexPrefix := codec.IndexPrefix(uint32(bp.TableID), uint32(bp.SourceIndexID))
	indexSpan := roachpb.Span{
		Key:    indexPrefix,
		EndKey: indexPrefix.PrefixEnd(),
	}
	return &progress{
		BackfillProgress: bp,
		totalSpan:        indexSpan,
	}
}

func (b *backfillTracker) FlushFractionCompleted(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580489)
	updated, fractionRangesFinished, err := b.getFractionRangesFinished(ctx)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(580491)
		return !updated == true
	}() == true {
		__antithesis_instrumentation__.Notify(580492)
		return err
	} else {
		__antithesis_instrumentation__.Notify(580493)
	}
	__antithesis_instrumentation__.Notify(580490)
	return b.writeProgressFraction(ctx, fractionRangesFinished)
}

func (b *backfillTracker) FlushCheckpoint(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580494)
	needsFlush, progress := b.collectProgressForCheckpointFlush()
	if !needsFlush {
		__antithesis_instrumentation__.Notify(580497)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580498)
	}
	__antithesis_instrumentation__.Notify(580495)
	sort.Slice(progress, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(580499)
		if progress[i].TableID != progress[j].TableID {
			__antithesis_instrumentation__.Notify(580501)
			return progress[i].TableID < progress[j].TableID
		} else {
			__antithesis_instrumentation__.Notify(580502)
		}
		__antithesis_instrumentation__.Notify(580500)
		return progress[i].SourceIndexID < progress[j].SourceIndexID
	})
	__antithesis_instrumentation__.Notify(580496)
	return b.writeCheckpoint(ctx, progress)
}

func (b *backfillTracker) getTableIndexBackfillProgress(bf scexec.Backfill) (progress, bool) {
	__antithesis_instrumentation__.Notify(580503)
	b.mu.Lock()
	defer b.mu.Unlock()
	if p, ok := b.getTableIndexBackfillProgressLocked(bf); ok {
		__antithesis_instrumentation__.Notify(580505)
		return *p, true
	} else {
		__antithesis_instrumentation__.Notify(580506)
	}
	__antithesis_instrumentation__.Notify(580504)
	return progress{}, false
}

func (b *backfillTracker) getBackfillProgressLocked(bf scexec.Backfill) (*progress, error) {
	__antithesis_instrumentation__.Notify(580507)
	p, ok := b.getTableIndexBackfillProgressLocked(bf)
	if !ok {
		__antithesis_instrumentation__.Notify(580510)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580511)
	}
	__antithesis_instrumentation__.Notify(580508)
	if err := p.matches(bf); err != nil {
		__antithesis_instrumentation__.Notify(580512)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(580513)
	}
	__antithesis_instrumentation__.Notify(580509)
	return p, nil
}

func (b *backfillTracker) getTableIndexBackfillProgressLocked(
	bf scexec.Backfill,
) (*progress, bool) {
	__antithesis_instrumentation__.Notify(580514)
	if p, ok := b.mu.progress[toKey(bf)]; ok {
		__antithesis_instrumentation__.Notify(580516)
		return p, true
	} else {
		__antithesis_instrumentation__.Notify(580517)
	}
	__antithesis_instrumentation__.Notify(580515)
	return nil, false
}

func (b *backfillTracker) getFractionRangesFinished(
	ctx context.Context,
) (updated bool, _ float32, _ error) {
	__antithesis_instrumentation__.Notify(580518)
	needsFlush, progresses := b.collectFractionProgressSpansForFlush()
	if !needsFlush {
		__antithesis_instrumentation__.Notify(580522)
		return false, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(580523)
	}
	__antithesis_instrumentation__.Notify(580519)
	var totalRanges int
	var completedRanges int
	for _, p := range progresses {
		__antithesis_instrumentation__.Notify(580524)
		total, completed, err := b.numRangesInSpanContainedBy(ctx, p.total, p.completed)
		if err != nil {
			__antithesis_instrumentation__.Notify(580526)
			return false, 0, err
		} else {
			__antithesis_instrumentation__.Notify(580527)
		}
		__antithesis_instrumentation__.Notify(580525)
		totalRanges += total
		completedRanges += completed
	}
	__antithesis_instrumentation__.Notify(580520)
	if totalRanges == 0 {
		__antithesis_instrumentation__.Notify(580528)
		return true, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(580529)
	}
	__antithesis_instrumentation__.Notify(580521)
	return true, float32(completedRanges) / float32(totalRanges), nil
}

type fractionProgressSpans struct {
	total     roachpb.Span
	completed []roachpb.Span
}

func (b *backfillTracker) collectFractionProgressSpansForFlush() (
	needsFlush bool,
	progress []fractionProgressSpans,
) {
	__antithesis_instrumentation__.Notify(580530)
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, p := range b.mu.progress {
		__antithesis_instrumentation__.Notify(580534)
		needsFlush = needsFlush || func() bool {
			__antithesis_instrumentation__.Notify(580535)
			return p.needsFractionFlush == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(580531)
	if !needsFlush {
		__antithesis_instrumentation__.Notify(580536)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(580537)
	}
	__antithesis_instrumentation__.Notify(580532)
	progress = make([]fractionProgressSpans, 0, len(b.mu.progress))
	for _, p := range b.mu.progress {
		__antithesis_instrumentation__.Notify(580538)
		p.needsFractionFlush = false
		progress = append(progress, fractionProgressSpans{
			total:     p.totalSpan,
			completed: p.CompletedSpans,
		})
	}
	__antithesis_instrumentation__.Notify(580533)
	return true, progress
}

func (b *backfillTracker) collectProgressForCheckpointFlush() (
	needsFlush bool,
	progress []scexec.BackfillProgress,
) {
	__antithesis_instrumentation__.Notify(580539)
	for _, p := range b.mu.progress {
		__antithesis_instrumentation__.Notify(580543)
		if p.needsCheckpointFlush {
			__antithesis_instrumentation__.Notify(580544)
			needsFlush = true
			break
		} else {
			__antithesis_instrumentation__.Notify(580545)
		}
	}
	__antithesis_instrumentation__.Notify(580540)
	if !needsFlush {
		__antithesis_instrumentation__.Notify(580546)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(580547)
	}
	__antithesis_instrumentation__.Notify(580541)
	for _, p := range b.mu.progress {
		__antithesis_instrumentation__.Notify(580548)
		p.needsCheckpointFlush = false
		progress = append(progress, p.BackfillProgress)
	}
	__antithesis_instrumentation__.Notify(580542)
	return needsFlush, progress
}

type progress struct {
	scexec.BackfillProgress

	needsCheckpointFlush bool

	needsFractionFlush bool

	totalSpan roachpb.Span
}

func (p progress) matches(bf scexec.Backfill) error {
	__antithesis_instrumentation__.Notify(580549)
	if bf.TableID == p.TableID && func() bool {
		__antithesis_instrumentation__.Notify(580551)
		return bf.SourceIndexID == p.SourceIndexID == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(580552)
		return sameIndexIDSet(bf.DestIndexIDs, p.DestIndexIDs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(580553)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580554)
	}
	__antithesis_instrumentation__.Notify(580550)
	return errors.AssertionFailedf(
		"backfill %v does not match stored progress for %v",
		bf, p.Backfill,
	)
}

func sameIndexIDSet(ds []descpb.IndexID, ds2 []descpb.IndexID) bool {
	__antithesis_instrumentation__.Notify(580555)
	toSet := func(ids []descpb.IndexID) (s util.FastIntSet) {
		__antithesis_instrumentation__.Notify(580557)
		for _, id := range ids {
			__antithesis_instrumentation__.Notify(580559)
			s.Add(int(id))
		}
		__antithesis_instrumentation__.Notify(580558)
		return s
	}
	__antithesis_instrumentation__.Notify(580556)
	return toSet(ds).Equals(toSet(ds2))
}

type tableIndexKey struct {
	tableID       descpb.ID
	sourceIndexID descpb.IndexID
}

func toKey(bf scexec.Backfill) tableIndexKey {
	__antithesis_instrumentation__.Notify(580560)
	return tableIndexKey{
		tableID:       bf.TableID,
		sourceIndexID: bf.SourceIndexID,
	}
}
