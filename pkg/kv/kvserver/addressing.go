package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

type metaAction func(*kv.Batch, roachpb.Key, *roachpb.RangeDescriptor)

func putMeta(b *kv.Batch, key roachpb.Key, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(94095)
	b.Put(key, desc)
}

func delMeta(b *kv.Batch, key roachpb.Key, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(94096)
	b.Del(key)
}

func splitRangeAddressing(b *kv.Batch, left, right *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(94097)
	if err := rangeAddressing(b, left, putMeta); err != nil {
		__antithesis_instrumentation__.Notify(94099)
		return err
	} else {
		__antithesis_instrumentation__.Notify(94100)
	}
	__antithesis_instrumentation__.Notify(94098)
	return rangeAddressing(b, right, putMeta)
}

func mergeRangeAddressing(b *kv.Batch, left, merged *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(94101)
	if err := rangeAddressing(b, left, delMeta); err != nil {
		__antithesis_instrumentation__.Notify(94103)
		return err
	} else {
		__antithesis_instrumentation__.Notify(94104)
	}
	__antithesis_instrumentation__.Notify(94102)
	return rangeAddressing(b, merged, putMeta)
}

func updateRangeAddressing(b *kv.Batch, desc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(94105)
	return rangeAddressing(b, desc, putMeta)
}

func rangeAddressing(b *kv.Batch, desc *roachpb.RangeDescriptor, action metaAction) error {
	__antithesis_instrumentation__.Notify(94106)

	if bytes.HasPrefix(desc.EndKey, keys.Meta1Prefix) || func() bool {
		__antithesis_instrumentation__.Notify(94109)
		return bytes.HasPrefix(desc.StartKey, keys.Meta1Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(94110)
		return errors.Errorf("meta1 addressing records cannot be split: %+v", desc)
	} else {
		__antithesis_instrumentation__.Notify(94111)
	}
	__antithesis_instrumentation__.Notify(94107)

	action(b, keys.RangeMetaKey(desc.EndKey).AsRawKey(), desc)

	if bytes.Compare(desc.StartKey, keys.MetaMax) < 0 && func() bool {
		__antithesis_instrumentation__.Notify(94112)
		return bytes.Compare(desc.EndKey, keys.MetaMax) >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94113)

		action(b, keys.Meta1KeyMax, desc)
	} else {
		__antithesis_instrumentation__.Notify(94114)
	}
	__antithesis_instrumentation__.Notify(94108)
	return nil
}
