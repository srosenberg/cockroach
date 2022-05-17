package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	gojson "encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const (
	confluentSubjectSuffixKey   = `-key`
	confluentSubjectSuffixValue = `-value`
)

type encodeRow struct {
	datums rowenc.EncDatumRow

	updated hlc.Timestamp

	mvccTimestamp hlc.Timestamp

	deleted bool

	tableDesc catalog.TableDescriptor

	familyID descpb.FamilyID

	prevDatums rowenc.EncDatumRow

	prevDeleted bool

	prevTableDesc catalog.TableDescriptor

	prevFamilyID descpb.FamilyID

	topic string
}

type Encoder interface {
	EncodeKey(context.Context, encodeRow) ([]byte, error)

	EncodeValue(context.Context, encodeRow) ([]byte, error)

	EncodeResolvedTimestamp(context.Context, string, hlc.Timestamp) ([]byte, error)
}

func getEncoder(
	opts map[string]string, targets []jobspb.ChangefeedTargetSpecification,
) (Encoder, error) {
	__antithesis_instrumentation__.Notify(16713)
	switch changefeedbase.FormatType(opts[changefeedbase.OptFormat]) {
	case ``, changefeedbase.OptFormatJSON:
		__antithesis_instrumentation__.Notify(16714)
		return makeJSONEncoder(opts, targets)
	case changefeedbase.OptFormatAvro, changefeedbase.DeprecatedOptFormatAvro:
		__antithesis_instrumentation__.Notify(16715)
		return newConfluentAvroEncoder(opts, targets)
	case changefeedbase.OptFormatNative:
		__antithesis_instrumentation__.Notify(16716)
		return &nativeEncoder{}, nil
	default:
		__antithesis_instrumentation__.Notify(16717)
		return nil, errors.Errorf(`unknown %s: %s`, changefeedbase.OptFormat, opts[changefeedbase.OptFormat])
	}
}

type jsonEncoder struct {
	updatedField, mvccTimestampField, beforeField, wrapped, keyOnly, keyInValue, topicInValue bool

	targets                 []jobspb.ChangefeedTargetSpecification
	alloc                   tree.DatumAlloc
	buf                     bytes.Buffer
	virtualColumnVisibility string
}

var _ Encoder = &jsonEncoder{}

func makeJSONEncoder(
	opts map[string]string, targets []jobspb.ChangefeedTargetSpecification,
) (*jsonEncoder, error) {
	__antithesis_instrumentation__.Notify(16718)
	e := &jsonEncoder{
		targets:                 targets,
		keyOnly:                 changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) == changefeedbase.OptEnvelopeKeyOnly,
		wrapped:                 changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) == changefeedbase.OptEnvelopeWrapped,
		virtualColumnVisibility: opts[changefeedbase.OptVirtualColumns],
	}
	_, e.updatedField = opts[changefeedbase.OptUpdatedTimestamps]
	_, e.mvccTimestampField = opts[changefeedbase.OptMVCCTimestamps]
	_, e.beforeField = opts[changefeedbase.OptDiff]
	if e.beforeField && func() bool {
		__antithesis_instrumentation__.Notify(16722)
		return !e.wrapped == true
	}() == true {
		__antithesis_instrumentation__.Notify(16723)
		return nil, errors.Errorf(`%s is only usable with %s=%s`,
			changefeedbase.OptDiff, changefeedbase.OptEnvelope, changefeedbase.OptEnvelopeWrapped)
	} else {
		__antithesis_instrumentation__.Notify(16724)
	}
	__antithesis_instrumentation__.Notify(16719)
	_, e.keyInValue = opts[changefeedbase.OptKeyInValue]
	if e.keyInValue && func() bool {
		__antithesis_instrumentation__.Notify(16725)
		return !e.wrapped == true
	}() == true {
		__antithesis_instrumentation__.Notify(16726)
		return nil, errors.Errorf(`%s is only usable with %s=%s`,
			changefeedbase.OptKeyInValue, changefeedbase.OptEnvelope, changefeedbase.OptEnvelopeWrapped)
	} else {
		__antithesis_instrumentation__.Notify(16727)
	}
	__antithesis_instrumentation__.Notify(16720)
	_, e.topicInValue = opts[changefeedbase.OptTopicInValue]
	if e.topicInValue && func() bool {
		__antithesis_instrumentation__.Notify(16728)
		return !e.wrapped == true
	}() == true {
		__antithesis_instrumentation__.Notify(16729)
		return nil, errors.Errorf(`%s is only usable with %s=%s`,
			changefeedbase.OptTopicInValue, changefeedbase.OptEnvelope, changefeedbase.OptEnvelopeWrapped)
	} else {
		__antithesis_instrumentation__.Notify(16730)
	}
	__antithesis_instrumentation__.Notify(16721)
	return e, nil
}

func (e *jsonEncoder) EncodeKey(_ context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16731)
	jsonEntries, err := e.encodeKeyRaw(row)
	if err != nil {
		__antithesis_instrumentation__.Notify(16734)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16735)
	}
	__antithesis_instrumentation__.Notify(16732)
	j, err := json.MakeJSON(jsonEntries)
	if err != nil {
		__antithesis_instrumentation__.Notify(16736)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16737)
	}
	__antithesis_instrumentation__.Notify(16733)
	e.buf.Reset()
	j.Format(&e.buf)
	return e.buf.Bytes(), nil
}

func (e *jsonEncoder) encodeKeyRaw(row encodeRow) ([]interface{}, error) {
	__antithesis_instrumentation__.Notify(16738)
	colIdxByID := catalog.ColumnIDToOrdinalMap(row.tableDesc.PublicColumns())
	primaryIndex := row.tableDesc.GetPrimaryIndex()
	jsonEntries := make([]interface{}, primaryIndex.NumKeyColumns())
	for i := 0; i < primaryIndex.NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(16740)
		colID := primaryIndex.GetKeyColumnID(i)
		idx, ok := colIdxByID.Get(colID)
		if !ok {
			__antithesis_instrumentation__.Notify(16743)
			return nil, errors.Errorf(`unknown column id: %d`, colID)
		} else {
			__antithesis_instrumentation__.Notify(16744)
		}
		__antithesis_instrumentation__.Notify(16741)
		datum, col := row.datums[idx], row.tableDesc.PublicColumns()[idx]
		if err := datum.EnsureDecoded(col.GetType(), &e.alloc); err != nil {
			__antithesis_instrumentation__.Notify(16745)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16746)
		}
		__antithesis_instrumentation__.Notify(16742)
		var err error
		jsonEntries[i], err = tree.AsJSON(
			datum.Datum,
			sessiondatapb.DataConversionConfig{},
			time.UTC,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(16747)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16748)
		}
	}
	__antithesis_instrumentation__.Notify(16739)
	return jsonEntries, nil
}

func (e *jsonEncoder) EncodeValue(_ context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16749)
	if e.keyOnly || func() bool {
		__antithesis_instrumentation__.Notify(16756)
		return (!e.wrapped && func() bool {
			__antithesis_instrumentation__.Notify(16757)
			return row.deleted == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(16758)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(16759)
	}
	__antithesis_instrumentation__.Notify(16750)

	var after map[string]interface{}
	if !row.deleted {
		__antithesis_instrumentation__.Notify(16760)
		family, err := row.tableDesc.FindFamilyByID(row.familyID)
		if err != nil {
			__antithesis_instrumentation__.Notify(16763)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16764)
		}
		__antithesis_instrumentation__.Notify(16761)
		include := make(map[descpb.ColumnID]struct{}, len(family.ColumnIDs))
		var yes struct{}
		for _, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(16765)
			include[colID] = yes
		}
		__antithesis_instrumentation__.Notify(16762)
		after = make(map[string]interface{})
		for i, col := range row.tableDesc.PublicColumns() {
			__antithesis_instrumentation__.Notify(16766)
			_, inFamily := include[col.GetID()]
			virtual := col.IsVirtual() && func() bool {
				__antithesis_instrumentation__.Notify(16767)
				return e.virtualColumnVisibility == string(changefeedbase.OptVirtualColumnsNull) == true
			}() == true
			if inFamily || func() bool {
				__antithesis_instrumentation__.Notify(16768)
				return virtual == true
			}() == true {
				__antithesis_instrumentation__.Notify(16769)
				datum := row.datums[i]
				if err := datum.EnsureDecoded(col.GetType(), &e.alloc); err != nil {
					__antithesis_instrumentation__.Notify(16771)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(16772)
				}
				__antithesis_instrumentation__.Notify(16770)
				after[col.GetName()], err = tree.AsJSON(
					datum.Datum,
					sessiondatapb.DataConversionConfig{},
					time.UTC,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(16773)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(16774)
				}
			} else {
				__antithesis_instrumentation__.Notify(16775)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(16776)
	}
	__antithesis_instrumentation__.Notify(16751)

	var before map[string]interface{}
	if row.prevDatums != nil && func() bool {
		__antithesis_instrumentation__.Notify(16777)
		return !row.prevDeleted == true
	}() == true {
		__antithesis_instrumentation__.Notify(16778)
		family, err := row.prevTableDesc.FindFamilyByID(row.prevFamilyID)
		if err != nil {
			__antithesis_instrumentation__.Notify(16781)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16782)
		}
		__antithesis_instrumentation__.Notify(16779)
		include := make(map[descpb.ColumnID]struct{})
		var yes struct{}
		for _, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(16783)
			include[colID] = yes
		}
		__antithesis_instrumentation__.Notify(16780)
		before = make(map[string]interface{})
		for i, col := range row.prevTableDesc.PublicColumns() {
			__antithesis_instrumentation__.Notify(16784)
			_, inFamily := include[col.GetID()]
			virtual := col.IsVirtual() && func() bool {
				__antithesis_instrumentation__.Notify(16785)
				return e.virtualColumnVisibility == string(changefeedbase.OptVirtualColumnsNull) == true
			}() == true
			if inFamily || func() bool {
				__antithesis_instrumentation__.Notify(16786)
				return virtual == true
			}() == true {
				__antithesis_instrumentation__.Notify(16787)
				datum := row.prevDatums[i]
				if err := datum.EnsureDecoded(col.GetType(), &e.alloc); err != nil {
					__antithesis_instrumentation__.Notify(16789)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(16790)
				}
				__antithesis_instrumentation__.Notify(16788)
				before[col.GetName()], err = tree.AsJSON(
					datum.Datum,
					sessiondatapb.DataConversionConfig{},
					time.UTC,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(16791)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(16792)
				}
			} else {
				__antithesis_instrumentation__.Notify(16793)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(16794)
	}
	__antithesis_instrumentation__.Notify(16752)

	var jsonEntries map[string]interface{}
	if e.wrapped {
		__antithesis_instrumentation__.Notify(16795)
		if after != nil {
			__antithesis_instrumentation__.Notify(16799)
			jsonEntries = map[string]interface{}{`after`: after}
		} else {
			__antithesis_instrumentation__.Notify(16800)
			jsonEntries = map[string]interface{}{`after`: nil}
		}
		__antithesis_instrumentation__.Notify(16796)
		if e.beforeField {
			__antithesis_instrumentation__.Notify(16801)
			if before != nil {
				__antithesis_instrumentation__.Notify(16802)
				jsonEntries[`before`] = before
			} else {
				__antithesis_instrumentation__.Notify(16803)
				jsonEntries[`before`] = nil
			}
		} else {
			__antithesis_instrumentation__.Notify(16804)
		}
		__antithesis_instrumentation__.Notify(16797)
		if e.keyInValue {
			__antithesis_instrumentation__.Notify(16805)
			keyEntries, err := e.encodeKeyRaw(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(16807)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(16808)
			}
			__antithesis_instrumentation__.Notify(16806)
			jsonEntries[`key`] = keyEntries
		} else {
			__antithesis_instrumentation__.Notify(16809)
		}
		__antithesis_instrumentation__.Notify(16798)
		if e.topicInValue {
			__antithesis_instrumentation__.Notify(16810)
			jsonEntries[`topic`] = row.topic
		} else {
			__antithesis_instrumentation__.Notify(16811)
		}
	} else {
		__antithesis_instrumentation__.Notify(16812)
		jsonEntries = after
	}
	__antithesis_instrumentation__.Notify(16753)

	if e.updatedField || func() bool {
		__antithesis_instrumentation__.Notify(16813)
		return e.mvccTimestampField == true
	}() == true {
		__antithesis_instrumentation__.Notify(16814)
		var meta map[string]interface{}
		if e.wrapped {
			__antithesis_instrumentation__.Notify(16817)
			meta = jsonEntries
		} else {
			__antithesis_instrumentation__.Notify(16818)
			meta = make(map[string]interface{}, 1)
			jsonEntries[jsonMetaSentinel] = meta
		}
		__antithesis_instrumentation__.Notify(16815)
		if e.updatedField {
			__antithesis_instrumentation__.Notify(16819)
			meta[`updated`] = row.updated.AsOfSystemTime()
		} else {
			__antithesis_instrumentation__.Notify(16820)
		}
		__antithesis_instrumentation__.Notify(16816)
		if e.mvccTimestampField {
			__antithesis_instrumentation__.Notify(16821)
			meta[`mvcc_timestamp`] = row.mvccTimestamp.AsOfSystemTime()
		} else {
			__antithesis_instrumentation__.Notify(16822)
		}
	} else {
		__antithesis_instrumentation__.Notify(16823)
	}
	__antithesis_instrumentation__.Notify(16754)

	j, err := json.MakeJSON(jsonEntries)
	if err != nil {
		__antithesis_instrumentation__.Notify(16824)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16825)
	}
	__antithesis_instrumentation__.Notify(16755)
	e.buf.Reset()
	j.Format(&e.buf)
	return e.buf.Bytes(), nil
}

func (e *jsonEncoder) EncodeResolvedTimestamp(
	_ context.Context, _ string, resolved hlc.Timestamp,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16826)
	meta := map[string]interface{}{
		`resolved`: tree.TimestampToDecimalDatum(resolved).Decimal.String(),
	}
	var jsonEntries interface{}
	if e.wrapped {
		__antithesis_instrumentation__.Notify(16828)
		jsonEntries = meta
	} else {
		__antithesis_instrumentation__.Notify(16829)
		jsonEntries = map[string]interface{}{
			jsonMetaSentinel: meta,
		}
	}
	__antithesis_instrumentation__.Notify(16827)
	return gojson.Marshal(jsonEntries)
}

type confluentAvroEncoder struct {
	schemaRegistry                     schemaRegistry
	schemaPrefix                       string
	updatedField, beforeField, keyOnly bool
	virtualColumnVisibility            string
	targets                            []jobspb.ChangefeedTargetSpecification

	keyCache   *cache.UnorderedCache
	valueCache *cache.UnorderedCache

	resolvedCache map[string]confluentRegisteredEnvelopeSchema
}

type tableIDAndVersion struct {
	tableID  descpb.ID
	version  descpb.DescriptorVersion
	familyID descpb.FamilyID
}
type tableIDAndVersionPair [2]tableIDAndVersion // [before, after]

type confluentRegisteredKeySchema struct {
	schema     *avroDataRecord
	registryID int32
}

type confluentRegisteredEnvelopeSchema struct {
	schema     *avroEnvelopeRecord
	registryID int32
}

var _ Encoder = &confluentAvroEncoder{}

var encoderCacheConfig = cache.Config{
	Policy: cache.CacheFIFO,

	ShouldEvict: func(size int, _ interface{}, _ interface{}) bool {
		__antithesis_instrumentation__.Notify(16830)
		return size > 1024
	},
}

func newConfluentAvroEncoder(
	opts map[string]string, targets []jobspb.ChangefeedTargetSpecification,
) (*confluentAvroEncoder, error) {
	__antithesis_instrumentation__.Notify(16831)
	e := &confluentAvroEncoder{
		schemaPrefix:            opts[changefeedbase.OptAvroSchemaPrefix],
		targets:                 targets,
		virtualColumnVisibility: opts[changefeedbase.OptVirtualColumns],
	}

	switch opts[changefeedbase.OptEnvelope] {
	case string(changefeedbase.OptEnvelopeKeyOnly):
		__antithesis_instrumentation__.Notify(16839)
		e.keyOnly = true
	case string(changefeedbase.OptEnvelopeWrapped):
		__antithesis_instrumentation__.Notify(16840)
	default:
		__antithesis_instrumentation__.Notify(16841)
		return nil, errors.Errorf(`%s=%s is not supported with %s=%s`,
			changefeedbase.OptEnvelope, opts[changefeedbase.OptEnvelope], changefeedbase.OptFormat, changefeedbase.OptFormatAvro)
	}
	__antithesis_instrumentation__.Notify(16832)
	_, e.updatedField = opts[changefeedbase.OptUpdatedTimestamps]
	if e.updatedField && func() bool {
		__antithesis_instrumentation__.Notify(16842)
		return e.keyOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(16843)
		return nil, errors.Errorf(`%s is only usable with %s=%s`,
			changefeedbase.OptUpdatedTimestamps, changefeedbase.OptEnvelope, changefeedbase.OptEnvelopeWrapped)
	} else {
		__antithesis_instrumentation__.Notify(16844)
	}
	__antithesis_instrumentation__.Notify(16833)
	_, e.beforeField = opts[changefeedbase.OptDiff]
	if e.beforeField && func() bool {
		__antithesis_instrumentation__.Notify(16845)
		return e.keyOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(16846)
		return nil, errors.Errorf(`%s is only usable with %s=%s`,
			changefeedbase.OptDiff, changefeedbase.OptEnvelope, changefeedbase.OptEnvelopeWrapped)
	} else {
		__antithesis_instrumentation__.Notify(16847)
	}
	__antithesis_instrumentation__.Notify(16834)

	if _, ok := opts[changefeedbase.OptKeyInValue]; ok {
		__antithesis_instrumentation__.Notify(16848)
		return nil, errors.Errorf(`%s is not supported with %s=%s`,
			changefeedbase.OptKeyInValue, changefeedbase.OptFormat, changefeedbase.OptFormatAvro)
	} else {
		__antithesis_instrumentation__.Notify(16849)
	}
	__antithesis_instrumentation__.Notify(16835)
	if _, ok := opts[changefeedbase.OptTopicInValue]; ok {
		__antithesis_instrumentation__.Notify(16850)
		return nil, errors.Errorf(`%s is not supported with %s=%s`,
			changefeedbase.OptTopicInValue, changefeedbase.OptFormat, changefeedbase.OptFormatAvro)
	} else {
		__antithesis_instrumentation__.Notify(16851)
	}
	__antithesis_instrumentation__.Notify(16836)
	if len(opts[changefeedbase.OptConfluentSchemaRegistry]) == 0 {
		__antithesis_instrumentation__.Notify(16852)
		return nil, errors.Errorf(`WITH option %s is required for %s=%s`,
			changefeedbase.OptConfluentSchemaRegistry, changefeedbase.OptFormat, changefeedbase.OptFormatAvro)
	} else {
		__antithesis_instrumentation__.Notify(16853)
	}
	__antithesis_instrumentation__.Notify(16837)

	reg, err := newConfluentSchemaRegistry(opts[changefeedbase.OptConfluentSchemaRegistry])
	if err != nil {
		__antithesis_instrumentation__.Notify(16854)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16855)
	}
	__antithesis_instrumentation__.Notify(16838)

	e.schemaRegistry = reg
	e.keyCache = cache.NewUnorderedCache(encoderCacheConfig)
	e.valueCache = cache.NewUnorderedCache(encoderCacheConfig)
	e.resolvedCache = make(map[string]confluentRegisteredEnvelopeSchema)
	return e, nil
}

func (e *confluentAvroEncoder) rawTableName(
	desc catalog.TableDescriptor, familyID descpb.FamilyID,
) (string, error) {
	__antithesis_instrumentation__.Notify(16856)
	for _, target := range e.targets {
		__antithesis_instrumentation__.Notify(16858)
		if target.TableID == desc.GetID() {
			__antithesis_instrumentation__.Notify(16859)
			switch target.Type {
			case jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY:
				__antithesis_instrumentation__.Notify(16861)
				return e.schemaPrefix + target.StatementTimeName, nil
			case jobspb.ChangefeedTargetSpecification_EACH_FAMILY:
				__antithesis_instrumentation__.Notify(16862)
				family, err := desc.FindFamilyByID(familyID)
				if err != nil {
					__antithesis_instrumentation__.Notify(16868)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(16869)
				}
				__antithesis_instrumentation__.Notify(16863)
				return fmt.Sprintf("%s%s.%s", e.schemaPrefix, target.StatementTimeName, family.Name), nil
			case jobspb.ChangefeedTargetSpecification_COLUMN_FAMILY:
				__antithesis_instrumentation__.Notify(16864)
				family, err := desc.FindFamilyByID(familyID)
				if err != nil {
					__antithesis_instrumentation__.Notify(16870)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(16871)
				}
				__antithesis_instrumentation__.Notify(16865)
				if family.Name != target.FamilyName {
					__antithesis_instrumentation__.Notify(16872)

					continue
				} else {
					__antithesis_instrumentation__.Notify(16873)
				}
				__antithesis_instrumentation__.Notify(16866)
				return fmt.Sprintf("%s%s.%s", e.schemaPrefix, target.StatementTimeName, target.FamilyName), nil
			default:
				__antithesis_instrumentation__.Notify(16867)

			}
			__antithesis_instrumentation__.Notify(16860)
			return "", errors.AssertionFailedf("Found a matching target with unimplemented type %s", target.Type)
		} else {
			__antithesis_instrumentation__.Notify(16874)
		}
	}
	__antithesis_instrumentation__.Notify(16857)
	return desc.GetName(), errors.Newf("Could not find TargetSpecification for descriptor %v", desc)
}

func (e *confluentAvroEncoder) EncodeKey(ctx context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16875)

	cacheKey := tableIDAndVersion{tableID: row.tableDesc.GetID(), version: row.tableDesc.GetVersion()}

	var registered confluentRegisteredKeySchema
	v, ok := e.keyCache.Get(cacheKey)
	if ok {
		__antithesis_instrumentation__.Notify(16877)
		registered = v.(confluentRegisteredKeySchema)
		registered.schema.refreshTypeMetadata(row.tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(16878)
		var err error
		tableName, err := e.rawTableName(row.tableDesc, row.familyID)
		if err != nil {
			__antithesis_instrumentation__.Notify(16882)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16883)
		}
		__antithesis_instrumentation__.Notify(16879)
		registered.schema, err = indexToAvroSchema(row.tableDesc, row.tableDesc.GetPrimaryIndex(), tableName, e.schemaPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(16884)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16885)
		}
		__antithesis_instrumentation__.Notify(16880)

		subject := SQLNameToKafkaName(tableName) + confluentSubjectSuffixKey
		registered.registryID, err = e.register(ctx, &registered.schema.avroRecord, subject)
		if err != nil {
			__antithesis_instrumentation__.Notify(16886)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16887)
		}
		__antithesis_instrumentation__.Notify(16881)
		e.keyCache.Add(cacheKey, registered)
	}
	__antithesis_instrumentation__.Notify(16876)

	header := []byte{
		changefeedbase.ConfluentAvroWireFormatMagic,
		0, 0, 0, 0,
	}
	binary.BigEndian.PutUint32(header[1:5], uint32(registered.registryID))
	return registered.schema.BinaryFromRow(header, row.datums)
}

func (e *confluentAvroEncoder) EncodeValue(ctx context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16888)
	if e.keyOnly {
		__antithesis_instrumentation__.Notify(16895)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(16896)
	}
	__antithesis_instrumentation__.Notify(16889)

	var cacheKey tableIDAndVersionPair
	if e.beforeField && func() bool {
		__antithesis_instrumentation__.Notify(16897)
		return row.prevTableDesc != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(16898)
		cacheKey[0] = tableIDAndVersion{
			tableID: row.prevTableDesc.GetID(), version: row.prevTableDesc.GetVersion(), familyID: row.prevFamilyID,
		}
	} else {
		__antithesis_instrumentation__.Notify(16899)
	}
	__antithesis_instrumentation__.Notify(16890)
	cacheKey[1] = tableIDAndVersion{
		tableID: row.tableDesc.GetID(), version: row.tableDesc.GetVersion(), familyID: row.familyID,
	}

	var registered confluentRegisteredEnvelopeSchema
	v, ok := e.valueCache.Get(cacheKey)
	if ok {
		__antithesis_instrumentation__.Notify(16900)
		registered = v.(confluentRegisteredEnvelopeSchema)
		registered.schema.after.refreshTypeMetadata(row.tableDesc)
		if row.prevTableDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(16901)
			return registered.schema.before != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(16902)
			registered.schema.before.refreshTypeMetadata(row.prevTableDesc)
		} else {
			__antithesis_instrumentation__.Notify(16903)
		}
	} else {
		__antithesis_instrumentation__.Notify(16904)
		var beforeDataSchema *avroDataRecord
		if e.beforeField && func() bool {
			__antithesis_instrumentation__.Notify(16910)
			return row.prevTableDesc != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(16911)
			var err error
			beforeDataSchema, err = tableToAvroSchema(row.prevTableDesc, row.prevFamilyID, `before`, e.schemaPrefix, e.virtualColumnVisibility)
			if err != nil {
				__antithesis_instrumentation__.Notify(16912)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(16913)
			}
		} else {
			__antithesis_instrumentation__.Notify(16914)
		}
		__antithesis_instrumentation__.Notify(16905)

		afterDataSchema, err := tableToAvroSchema(row.tableDesc, row.familyID, avroSchemaNoSuffix, e.schemaPrefix, e.virtualColumnVisibility)
		if err != nil {
			__antithesis_instrumentation__.Notify(16915)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16916)
		}
		__antithesis_instrumentation__.Notify(16906)

		opts := avroEnvelopeOpts{afterField: true, beforeField: e.beforeField, updatedField: e.updatedField}
		name, err := e.rawTableName(row.tableDesc, row.familyID)
		if err != nil {
			__antithesis_instrumentation__.Notify(16917)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16918)
		}
		__antithesis_instrumentation__.Notify(16907)
		registered.schema, err = envelopeToAvroSchema(name, opts, beforeDataSchema, afterDataSchema, e.schemaPrefix)

		if err != nil {
			__antithesis_instrumentation__.Notify(16919)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16920)
		}
		__antithesis_instrumentation__.Notify(16908)

		subject := SQLNameToKafkaName(name) + confluentSubjectSuffixValue
		registered.registryID, err = e.register(ctx, &registered.schema.avroRecord, subject)
		if err != nil {
			__antithesis_instrumentation__.Notify(16921)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16922)
		}
		__antithesis_instrumentation__.Notify(16909)
		e.valueCache.Add(cacheKey, registered)
	}
	__antithesis_instrumentation__.Notify(16891)

	var meta avroMetadata
	if registered.schema.opts.updatedField {
		__antithesis_instrumentation__.Notify(16923)
		meta = map[string]interface{}{
			`updated`: row.updated,
		}
	} else {
		__antithesis_instrumentation__.Notify(16924)
	}
	__antithesis_instrumentation__.Notify(16892)
	var beforeDatums, afterDatums rowenc.EncDatumRow
	if row.prevDatums != nil && func() bool {
		__antithesis_instrumentation__.Notify(16925)
		return !row.prevDeleted == true
	}() == true {
		__antithesis_instrumentation__.Notify(16926)
		beforeDatums = row.prevDatums
	} else {
		__antithesis_instrumentation__.Notify(16927)
	}
	__antithesis_instrumentation__.Notify(16893)
	if !row.deleted {
		__antithesis_instrumentation__.Notify(16928)
		afterDatums = row.datums
	} else {
		__antithesis_instrumentation__.Notify(16929)
	}
	__antithesis_instrumentation__.Notify(16894)

	header := []byte{
		changefeedbase.ConfluentAvroWireFormatMagic,
		0, 0, 0, 0,
	}
	binary.BigEndian.PutUint32(header[1:5], uint32(registered.registryID))
	return registered.schema.BinaryFromRow(header, meta, beforeDatums, afterDatums)
}

func (e *confluentAvroEncoder) EncodeResolvedTimestamp(
	ctx context.Context, topic string, resolved hlc.Timestamp,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16930)
	registered, ok := e.resolvedCache[topic]
	if !ok {
		__antithesis_instrumentation__.Notify(16933)
		opts := avroEnvelopeOpts{resolvedField: true}
		var err error
		registered.schema, err = envelopeToAvroSchema(topic, opts, nil, nil, e.schemaPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(16936)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16937)
		}
		__antithesis_instrumentation__.Notify(16934)

		subject := SQLNameToKafkaName(topic) + confluentSubjectSuffixValue
		registered.registryID, err = e.register(ctx, &registered.schema.avroRecord, subject)
		if err != nil {
			__antithesis_instrumentation__.Notify(16938)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16939)
		}
		__antithesis_instrumentation__.Notify(16935)

		e.resolvedCache[topic] = registered
	} else {
		__antithesis_instrumentation__.Notify(16940)
	}
	__antithesis_instrumentation__.Notify(16931)
	var meta avroMetadata
	if registered.schema.opts.resolvedField {
		__antithesis_instrumentation__.Notify(16941)
		meta = map[string]interface{}{
			`resolved`: resolved,
		}
	} else {
		__antithesis_instrumentation__.Notify(16942)
	}
	__antithesis_instrumentation__.Notify(16932)

	header := []byte{
		changefeedbase.ConfluentAvroWireFormatMagic,
		0, 0, 0, 0,
	}
	binary.BigEndian.PutUint32(header[1:5], uint32(registered.registryID))
	return registered.schema.BinaryFromRow(header, meta, nil, nil)
}

func (e *confluentAvroEncoder) register(
	ctx context.Context, schema *avroRecord, subject string,
) (int32, error) {
	__antithesis_instrumentation__.Notify(16943)
	return e.schemaRegistry.RegisterSchemaForSubject(ctx, subject, schema.codec.Schema())
}

type nativeEncoder struct{}

func (e *nativeEncoder) EncodeKey(ctx context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16944)
	return nil, errors.New("EncodeKey unexpectedly called on nativeEncoder")
}

func (e *nativeEncoder) EncodeValue(ctx context.Context, row encodeRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16945)
	return nil, errors.New("EncodeValue unexpectedly called on nativeEncoder")
}

func (e *nativeEncoder) EncodeResolvedTimestamp(
	ctx context.Context, s string, ts hlc.Timestamp,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(16946)
	return protoutil.Marshal(&ts)
}

var _ Encoder = &nativeEncoder{}
