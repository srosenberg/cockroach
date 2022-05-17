package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
)

type intervalSpan roachpb.Span

var _ interval.Interface = intervalSpan{}

func (ie intervalSpan) ID() uintptr { __antithesis_instrumentation__.Notify(12526); return 0 }

func (ie intervalSpan) Range() interval.Range {
	__antithesis_instrumentation__.Notify(12527)
	return interval.Range{Start: []byte(ie.Key), End: []byte(ie.EndKey)}
}

func makeSimpleImportSpans(
	requiredSpans []roachpb.Span,
	backups []BackupManifest,
	backupLocalityMap map[int]storeByLocalityKV,
	lowWaterMark roachpb.Key,
) []execinfrapb.RestoreSpanEntry {
	__antithesis_instrumentation__.Notify(12528)
	if len(backups) < 1 {
		__antithesis_instrumentation__.Notify(12532)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12533)
	}
	__antithesis_instrumentation__.Notify(12529)

	for i := range backups {
		__antithesis_instrumentation__.Notify(12534)
		sort.Sort(BackupFileDescriptors(backups[i].Files))
	}
	__antithesis_instrumentation__.Notify(12530)

	var cover []execinfrapb.RestoreSpanEntry
	for _, span := range requiredSpans {
		__antithesis_instrumentation__.Notify(12535)
		if span.EndKey.Compare(lowWaterMark) < 0 {
			__antithesis_instrumentation__.Notify(12538)
			continue
		} else {
			__antithesis_instrumentation__.Notify(12539)
		}
		__antithesis_instrumentation__.Notify(12536)
		if span.Key.Compare(lowWaterMark) < 0 {
			__antithesis_instrumentation__.Notify(12540)
			span.Key = lowWaterMark
		} else {
			__antithesis_instrumentation__.Notify(12541)
		}
		__antithesis_instrumentation__.Notify(12537)

		spanCoverStart := len(cover)

		for layer := range backups {
			__antithesis_instrumentation__.Notify(12542)
			covPos := spanCoverStart

			for _, f := range backups[layer].Files {
				__antithesis_instrumentation__.Notify(12543)
				if sp := span.Intersect(f.Span); sp.Valid() {
					__antithesis_instrumentation__.Notify(12544)
					fileSpec := execinfrapb.RestoreFileSpec{Path: f.Path, Dir: backups[layer].Dir}
					if dir, ok := backupLocalityMap[layer][f.LocalityKV]; ok {
						__antithesis_instrumentation__.Notify(12546)
						fileSpec = execinfrapb.RestoreFileSpec{Path: f.Path, Dir: dir}
					} else {
						__antithesis_instrumentation__.Notify(12547)
					}
					__antithesis_instrumentation__.Notify(12545)
					if len(cover) == spanCoverStart {
						__antithesis_instrumentation__.Notify(12548)
						cover = append(cover, makeEntry(span.Key, sp.EndKey, fileSpec))
					} else {
						__antithesis_instrumentation__.Notify(12549)

						for i := covPos; i < len(cover) && func() bool {
							__antithesis_instrumentation__.Notify(12551)
							return cover[i].Span.Key.Compare(sp.EndKey) < 0 == true
						}() == true; i++ {
							__antithesis_instrumentation__.Notify(12552)
							if cover[i].Span.Overlaps(sp) {
								__antithesis_instrumentation__.Notify(12554)
								cover[i].Files = append(cover[i].Files, fileSpec)
							} else {
								__antithesis_instrumentation__.Notify(12555)
							}
							__antithesis_instrumentation__.Notify(12553)

							if cover[i].Span.EndKey.Compare(sp.Key) <= 0 {
								__antithesis_instrumentation__.Notify(12556)
								covPos = i + 1
							} else {
								__antithesis_instrumentation__.Notify(12557)
							}
						}
						__antithesis_instrumentation__.Notify(12550)

						if covEnd := cover[len(cover)-1].Span.EndKey; sp.EndKey.Compare(covEnd) > 0 {
							__antithesis_instrumentation__.Notify(12558)
							cover = append(cover, makeEntry(covEnd, sp.EndKey, fileSpec))
						} else {
							__antithesis_instrumentation__.Notify(12559)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(12560)
					if span.EndKey.Compare(f.Span.Key) <= 0 {
						__antithesis_instrumentation__.Notify(12561)

						break
					} else {
						__antithesis_instrumentation__.Notify(12562)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(12531)

	return cover
}

func makeEntry(start, end roachpb.Key, f execinfrapb.RestoreFileSpec) execinfrapb.RestoreSpanEntry {
	__antithesis_instrumentation__.Notify(12563)
	return execinfrapb.RestoreSpanEntry{
		Span: roachpb.Span{Key: start, EndKey: end}, Files: []execinfrapb.RestoreFileSpec{f},
	}
}
