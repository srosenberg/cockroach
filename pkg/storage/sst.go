package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/sstable"
)

func CheckSSTConflicts(
	ctx context.Context,
	sst []byte,
	reader Reader,
	start, end MVCCKey,
	disallowShadowing bool,
	disallowShadowingBelow hlc.Timestamp,
	maxIntents int64,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(643689)
	var statsDiff enginepb.MVCCStats
	var intents []roachpb.Intent

	extIter := reader.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{UpperBound: end.Key})
	defer extIter.Close()
	extIter.SeekGE(start)

	sstIter, err := NewMemSSTIterator(sst, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(643695)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(643696)
	}
	__antithesis_instrumentation__.Notify(643690)
	defer sstIter.Close()
	sstIter.SeekGE(start)

	extOK, extErr := extIter.Valid()
	sstOK, sstErr := sstIter.Valid()
	for extErr == nil && func() bool {
		__antithesis_instrumentation__.Notify(643697)
		return sstErr == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643698)
		return extOK == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643699)
		return sstOK == true
	}() == true {
		__antithesis_instrumentation__.Notify(643700)
		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(643711)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(643712)
		}
		__antithesis_instrumentation__.Notify(643701)

		extKey, extValue := extIter.UnsafeKey(), extIter.UnsafeValue()
		sstKey, sstValue := sstIter.UnsafeKey(), sstIter.UnsafeValue()

		if cmp := bytes.Compare(extKey.Key, sstKey.Key); cmp < 0 {
			__antithesis_instrumentation__.Notify(643713)
			extIter.SeekGE(MVCCKey{Key: sstKey.Key})
			extOK, extErr = extIter.Valid()
			continue
		} else {
			__antithesis_instrumentation__.Notify(643714)
			if cmp > 0 {
				__antithesis_instrumentation__.Notify(643715)
				sstIter.SeekGE(MVCCKey{Key: extKey.Key})
				sstOK, sstErr = sstIter.Valid()
				continue
			} else {
				__antithesis_instrumentation__.Notify(643716)
			}
		}
		__antithesis_instrumentation__.Notify(643702)

		if !sstKey.IsValue() {
			__antithesis_instrumentation__.Notify(643717)
			return enginepb.MVCCStats{}, errors.New("SST keys must have timestamps")
		} else {
			__antithesis_instrumentation__.Notify(643718)
		}
		__antithesis_instrumentation__.Notify(643703)
		if len(sstValue) == 0 {
			__antithesis_instrumentation__.Notify(643719)
			return enginepb.MVCCStats{}, errors.New("SST values cannot be tombstones")
		} else {
			__antithesis_instrumentation__.Notify(643720)
		}
		__antithesis_instrumentation__.Notify(643704)
		if !extKey.IsValue() {
			__antithesis_instrumentation__.Notify(643721)
			var mvccMeta enginepb.MVCCMetadata
			if err = extIter.ValueProto(&mvccMeta); err != nil {
				__antithesis_instrumentation__.Notify(643723)
				return enginepb.MVCCStats{}, err
			} else {
				__antithesis_instrumentation__.Notify(643724)
			}
			__antithesis_instrumentation__.Notify(643722)
			if len(mvccMeta.RawBytes) > 0 {
				__antithesis_instrumentation__.Notify(643725)
				return enginepb.MVCCStats{}, errors.New("inline values are unsupported")
			} else {
				__antithesis_instrumentation__.Notify(643726)
				if mvccMeta.Txn == nil {
					__antithesis_instrumentation__.Notify(643727)
					return enginepb.MVCCStats{}, errors.New("found intent without transaction")
				} else {
					__antithesis_instrumentation__.Notify(643728)

					intents = append(intents, roachpb.MakeIntent(mvccMeta.Txn, extIter.Key().Key))
					if int64(len(intents)) >= maxIntents {
						__antithesis_instrumentation__.Notify(643731)
						return enginepb.MVCCStats{}, &roachpb.WriteIntentError{Intents: intents}
					} else {
						__antithesis_instrumentation__.Notify(643732)
					}
					__antithesis_instrumentation__.Notify(643729)
					sstIter.NextKey()
					sstOK, sstErr = sstIter.Valid()
					if sstOK {
						__antithesis_instrumentation__.Notify(643733)
						extIter.SeekGE(MVCCKey{Key: sstIter.UnsafeKey().Key})
					} else {
						__antithesis_instrumentation__.Notify(643734)
					}
					__antithesis_instrumentation__.Notify(643730)
					extOK, extErr = extIter.Valid()
					continue
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(643735)
		}
		__antithesis_instrumentation__.Notify(643705)

		allowIdempotent := (!disallowShadowingBelow.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(643736)
			return disallowShadowingBelow.LessEq(extKey.Timestamp) == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(643737)
			return (disallowShadowingBelow.IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(643738)
				return disallowShadowing == true
			}() == true) == true
		}() == true
		if allowIdempotent && func() bool {
			__antithesis_instrumentation__.Notify(643739)
			return sstKey.Timestamp.Equal(extKey.Timestamp) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(643740)
			return bytes.Equal(extValue, sstValue) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643741)

			metaKeySize := int64(len(sstKey.Key) + 1)
			metaValSize := int64(0)
			totalBytes := metaKeySize + metaValSize

			statsDiff.LiveBytes -= totalBytes
			statsDiff.LiveCount--
			statsDiff.KeyBytes -= metaKeySize
			statsDiff.ValBytes -= metaValSize
			statsDiff.KeyCount--

			totalBytes = int64(len(sstValue)) + MVCCVersionTimestampSize
			statsDiff.LiveBytes -= totalBytes
			statsDiff.KeyBytes -= MVCCVersionTimestampSize
			statsDiff.ValBytes -= int64(len(sstValue))
			statsDiff.ValCount--

			sstIter.NextKey()
			sstOK, sstErr = sstIter.Valid()
			if sstOK {
				__antithesis_instrumentation__.Notify(643743)
				extIter.SeekGE(MVCCKey{Key: sstIter.UnsafeKey().Key})
			} else {
				__antithesis_instrumentation__.Notify(643744)
			}
			__antithesis_instrumentation__.Notify(643742)
			extOK, extErr = extIter.Valid()
			continue
		} else {
			__antithesis_instrumentation__.Notify(643745)
		}
		__antithesis_instrumentation__.Notify(643706)

		if len(extValue) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(643746)
			return (!disallowShadowingBelow.IsEmpty() || func() bool {
				__antithesis_instrumentation__.Notify(643747)
				return disallowShadowing == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643748)
			allowShadow := !disallowShadowingBelow.IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(643749)
				return disallowShadowingBelow.LessEq(extKey.Timestamp) == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(643750)
				return bytes.Equal(extValue, sstValue) == true
			}() == true
			if !allowShadow {
				__antithesis_instrumentation__.Notify(643751)
				return enginepb.MVCCStats{}, errors.Errorf(
					"ingested key collides with an existing one: %s", sstKey.Key)
			} else {
				__antithesis_instrumentation__.Notify(643752)
			}
		} else {
			__antithesis_instrumentation__.Notify(643753)
		}
		__antithesis_instrumentation__.Notify(643707)

		if sstKey.Timestamp.LessEq(extKey.Timestamp) {
			__antithesis_instrumentation__.Notify(643754)
			return enginepb.MVCCStats{}, roachpb.NewWriteTooOldError(
				sstKey.Timestamp, extKey.Timestamp.Next(), sstKey.Key)
		} else {
			__antithesis_instrumentation__.Notify(643755)
		}
		__antithesis_instrumentation__.Notify(643708)

		statsDiff.KeyCount--
		statsDiff.KeyBytes -= int64(len(extKey.Key) + 1)
		if len(extValue) > 0 {
			__antithesis_instrumentation__.Notify(643756)
			statsDiff.LiveCount--
			statsDiff.LiveBytes -= int64(len(extKey.Key) + 1)
			statsDiff.LiveBytes -= int64(len(extValue)) + MVCCVersionTimestampSize
		} else {
			__antithesis_instrumentation__.Notify(643757)
		}
		__antithesis_instrumentation__.Notify(643709)

		sstIter.NextKey()
		sstOK, sstErr = sstIter.Valid()
		if sstOK {
			__antithesis_instrumentation__.Notify(643758)
			extIter.SeekGE(MVCCKey{Key: sstIter.UnsafeKey().Key})
		} else {
			__antithesis_instrumentation__.Notify(643759)
		}
		__antithesis_instrumentation__.Notify(643710)
		extOK, extErr = extIter.Valid()
	}
	__antithesis_instrumentation__.Notify(643691)

	if extErr != nil {
		__antithesis_instrumentation__.Notify(643760)
		return enginepb.MVCCStats{}, extErr
	} else {
		__antithesis_instrumentation__.Notify(643761)
	}
	__antithesis_instrumentation__.Notify(643692)
	if sstErr != nil {
		__antithesis_instrumentation__.Notify(643762)
		return enginepb.MVCCStats{}, sstErr
	} else {
		__antithesis_instrumentation__.Notify(643763)
	}
	__antithesis_instrumentation__.Notify(643693)
	if len(intents) > 0 {
		__antithesis_instrumentation__.Notify(643764)
		return enginepb.MVCCStats{}, &roachpb.WriteIntentError{Intents: intents}
	} else {
		__antithesis_instrumentation__.Notify(643765)
	}
	__antithesis_instrumentation__.Notify(643694)

	return statsDiff, nil
}

func UpdateSSTTimestamps(
	ctx context.Context, st *cluster.Settings, sst []byte, from, to hlc.Timestamp, concurrency int,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(643766)
	if from.IsEmpty() {
		__antithesis_instrumentation__.Notify(643773)
		return nil, errors.Errorf("from timestamp not given")
	} else {
		__antithesis_instrumentation__.Notify(643774)
	}
	__antithesis_instrumentation__.Notify(643767)
	if to.IsEmpty() {
		__antithesis_instrumentation__.Notify(643775)
		return nil, errors.Errorf("to timestamp not given")
	} else {
		__antithesis_instrumentation__.Notify(643776)
	}
	__antithesis_instrumentation__.Notify(643768)

	sstOut := &MemFile{}
	sstOut.Buffer.Grow(len(sst))

	if concurrency > 0 {
		__antithesis_instrumentation__.Notify(643777)
		defaults := DefaultPebbleOptions()
		opts := defaults.MakeReaderOptions()
		if fp := defaults.Levels[0].FilterPolicy; fp != nil && func() bool {
			__antithesis_instrumentation__.Notify(643780)
			return len(opts.Filters) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(643781)
			opts.Filters = map[string]sstable.FilterPolicy{fp.Name(): fp}
		} else {
			__antithesis_instrumentation__.Notify(643782)
		}
		__antithesis_instrumentation__.Notify(643778)
		if _, err := sstable.RewriteKeySuffixes(sst,
			opts,
			sstOut,
			MakeIngestionWriterOptions(ctx, st),
			EncodeMVCCTimestampSuffix(from),
			EncodeMVCCTimestampSuffix(to),
			concurrency,
		); err != nil {
			__antithesis_instrumentation__.Notify(643783)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(643784)
		}
		__antithesis_instrumentation__.Notify(643779)
		return sstOut.Bytes(), nil
	} else {
		__antithesis_instrumentation__.Notify(643785)
	}
	__antithesis_instrumentation__.Notify(643769)

	writer := MakeIngestionSSTWriter(ctx, st, sstOut)
	defer writer.Close()

	iter, err := NewMemSSTIterator(sst, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(643786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(643787)
	}
	__antithesis_instrumentation__.Notify(643770)
	defer iter.Close()

	for iter.SeekGE(MVCCKey{Key: keys.MinKey}); ; iter.Next() {
		__antithesis_instrumentation__.Notify(643788)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(643791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(643792)
			if !ok {
				__antithesis_instrumentation__.Notify(643793)
				break
			} else {
				__antithesis_instrumentation__.Notify(643794)
			}
		}
		__antithesis_instrumentation__.Notify(643789)
		key := iter.UnsafeKey()
		if key.Timestamp != from {
			__antithesis_instrumentation__.Notify(643795)
			return nil, errors.Errorf("unexpected timestamp %s (expected %s) for key %s",
				key.Timestamp, from, key.Key)
		} else {
			__antithesis_instrumentation__.Notify(643796)
		}
		__antithesis_instrumentation__.Notify(643790)
		err = writer.PutMVCC(MVCCKey{Key: iter.UnsafeKey().Key, Timestamp: to}, iter.UnsafeValue())
		if err != nil {
			__antithesis_instrumentation__.Notify(643797)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(643798)
		}
	}
	__antithesis_instrumentation__.Notify(643771)

	if err = writer.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(643799)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(643800)
	}
	__antithesis_instrumentation__.Notify(643772)

	return sstOut.Bytes(), nil
}
