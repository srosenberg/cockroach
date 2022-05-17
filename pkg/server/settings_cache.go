package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/settingswatcher"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type settingsCacheWriter struct {
	stopper *stop.Stopper
	eng     storage.Engine

	mu struct {
		syncutil.Mutex
		currentlyWriting bool

		queuedToWrite []roachpb.KeyValue
	}
}

func newSettingsCacheWriter(eng storage.Engine, stopper *stop.Stopper) *settingsCacheWriter {
	__antithesis_instrumentation__.Notify(235058)
	return &settingsCacheWriter{
		eng:     eng,
		stopper: stopper,
	}
}

func (s *settingsCacheWriter) SnapshotKVs(ctx context.Context, kvs []roachpb.KeyValue) {
	__antithesis_instrumentation__.Notify(235059)
	if !s.queueSnapshot(kvs) {
		__antithesis_instrumentation__.Notify(235061)
		return
	} else {
		__antithesis_instrumentation__.Notify(235062)
	}
	__antithesis_instrumentation__.Notify(235060)
	if err := s.stopper.RunAsyncTask(ctx, "snapshot-settings-cache", func(
		ctx context.Context,
	) {
		__antithesis_instrumentation__.Notify(235063)
		defer s.doneWriting()
		for toWrite, ok := s.getToWrite(); ok; toWrite, ok = s.getToWrite() {
			__antithesis_instrumentation__.Notify(235064)
			if err := storeCachedSettingsKVs(ctx, s.eng, toWrite); err != nil {
				__antithesis_instrumentation__.Notify(235065)
				log.Warningf(ctx, "failed to write settings snapshot: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(235066)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(235067)
		s.doneWriting()
	} else {
		__antithesis_instrumentation__.Notify(235068)
	}
}

func (s *settingsCacheWriter) queueSnapshot(kvs []roachpb.KeyValue) (shouldRun bool) {
	__antithesis_instrumentation__.Notify(235069)
	s.mu.Lock()
	if s.mu.currentlyWriting {
		__antithesis_instrumentation__.Notify(235071)
		s.mu.queuedToWrite = kvs
		s.mu.Unlock()
		return false
	} else {
		__antithesis_instrumentation__.Notify(235072)
	}
	__antithesis_instrumentation__.Notify(235070)
	s.mu.currentlyWriting = true
	s.mu.queuedToWrite = kvs
	s.mu.Unlock()
	return true
}

func (s *settingsCacheWriter) doneWriting() {
	__antithesis_instrumentation__.Notify(235073)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.currentlyWriting = false
}

func (s *settingsCacheWriter) getToWrite() (toWrite []roachpb.KeyValue, ok bool) {
	__antithesis_instrumentation__.Notify(235074)
	s.mu.Lock()
	defer s.mu.Unlock()
	toWrite, s.mu.queuedToWrite = s.mu.queuedToWrite, nil
	return toWrite, toWrite != nil
}

var _ settingswatcher.Storage = (*settingsCacheWriter)(nil)

func storeCachedSettingsKVs(ctx context.Context, eng storage.Engine, kvs []roachpb.KeyValue) error {
	__antithesis_instrumentation__.Notify(235075)
	batch := eng.NewBatch()
	defer batch.Close()
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(235077)
		kv.Value.Timestamp = hlc.Timestamp{}
		if err := storage.MVCCPut(
			ctx, batch, nil, keys.StoreCachedSettingsKey(kv.Key), hlc.Timestamp{}, kv.Value, nil,
		); err != nil {
			__antithesis_instrumentation__.Notify(235078)
			return err
		} else {
			__antithesis_instrumentation__.Notify(235079)
		}
	}
	__antithesis_instrumentation__.Notify(235076)
	return batch.Commit(false)
}

func loadCachedSettingsKVs(_ context.Context, eng storage.Engine) ([]roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(235080)
	var settingsKVs []roachpb.KeyValue
	if err := eng.MVCCIterate(
		keys.LocalStoreCachedSettingsKeyMin,
		keys.LocalStoreCachedSettingsKeyMax,
		storage.MVCCKeyAndIntentsIterKind,
		func(kv storage.MVCCKeyValue) error {
			__antithesis_instrumentation__.Notify(235082)
			settingKey, err := keys.DecodeStoreCachedSettingsKey(kv.Key.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(235085)
				return err
			} else {
				__antithesis_instrumentation__.Notify(235086)
			}
			__antithesis_instrumentation__.Notify(235083)
			meta := enginepb.MVCCMetadata{}
			if err := protoutil.Unmarshal(kv.Value, &meta); err != nil {
				__antithesis_instrumentation__.Notify(235087)
				return err
			} else {
				__antithesis_instrumentation__.Notify(235088)
			}
			__antithesis_instrumentation__.Notify(235084)
			settingsKVs = append(settingsKVs, roachpb.KeyValue{
				Key:   settingKey,
				Value: roachpb.Value{RawBytes: meta.RawBytes},
			})
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(235089)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235090)
	}
	__antithesis_instrumentation__.Notify(235081)
	return settingsKVs, nil
}

func initializeCachedSettings(
	ctx context.Context, codec keys.SQLCodec, updater settings.Updater, kvs []roachpb.KeyValue,
) error {
	__antithesis_instrumentation__.Notify(235091)
	dec := settingswatcher.MakeRowDecoder(codec)
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(235093)
		settings, val, _, err := dec.DecodeRow(kv)
		if err != nil {
			__antithesis_instrumentation__.Notify(235095)
			return errors.Wrap(err, `while decoding settings data
-this likely indicates the settings table structure or encoding has been altered;
-skipping settings updates`)
		} else {
			__antithesis_instrumentation__.Notify(235096)
		}
		__antithesis_instrumentation__.Notify(235094)
		if err := updater.Set(ctx, settings, val); err != nil {
			__antithesis_instrumentation__.Notify(235097)
			log.Warningf(ctx, "setting %q to %v failed: %+v", settings, val, err)
		} else {
			__antithesis_instrumentation__.Notify(235098)
		}
	}
	__antithesis_instrumentation__.Notify(235092)
	updater.ResetRemaining(ctx)
	return nil
}
