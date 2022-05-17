// Package spanconfigsqltranslator provides logic to translate sql descriptors
// and their corresponding zone configurations to constituent spans and span
// configurations.
package spanconfigsqltranslator

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

var _ spanconfig.SQLTranslator = &SQLTranslator{}

type SQLTranslator struct {
	ptsProvider protectedts.Provider
	codec       keys.SQLCodec
	knobs       *spanconfig.TestingKnobs

	txn      *kv.Txn
	descsCol *descs.Collection
}

type Factory struct {
	ptsProvider protectedts.Provider
	codec       keys.SQLCodec
	knobs       *spanconfig.TestingKnobs
}

func NewFactory(
	ptsProvider protectedts.Provider, codec keys.SQLCodec, knobs *spanconfig.TestingKnobs,
) *Factory {
	__antithesis_instrumentation__.Notify(241008)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(241010)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(241011)
	}
	__antithesis_instrumentation__.Notify(241009)
	return &Factory{
		ptsProvider: ptsProvider,
		codec:       codec,
		knobs:       knobs,
	}
}

func (f *Factory) NewSQLTranslator(txn *kv.Txn, descsCol *descs.Collection) *SQLTranslator {
	__antithesis_instrumentation__.Notify(241012)
	return &SQLTranslator{
		ptsProvider: f.ptsProvider,
		codec:       f.codec,
		knobs:       f.knobs,
		txn:         txn,
		descsCol:    descsCol,
	}
}

func (s *SQLTranslator) Translate(
	ctx context.Context, ids descpb.IDs, generateSystemSpanConfigurations bool,
) (records []spanconfig.Record, _ hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(241013)

	ptsState, err := s.ptsProvider.GetState(ctx, s.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(241021)
		return nil, hlc.Timestamp{}, errors.Wrap(err, "failed to get protected timestamp state")
	} else {
		__antithesis_instrumentation__.Notify(241022)
	}
	__antithesis_instrumentation__.Notify(241014)
	ptsStateReader := spanconfig.NewProtectedTimestampStateReader(ctx, ptsState)

	if generateSystemSpanConfigurations {
		__antithesis_instrumentation__.Notify(241023)
		records, err = s.generateSystemSpanConfigRecords(ptsStateReader)
		if err != nil {
			__antithesis_instrumentation__.Notify(241024)
			return nil, hlc.Timestamp{}, errors.Wrap(err, "failed to generate SystemTarget records")
		} else {
			__antithesis_instrumentation__.Notify(241025)
		}
	} else {
		__antithesis_instrumentation__.Notify(241026)
	}
	__antithesis_instrumentation__.Notify(241015)

	seen := make(map[descpb.ID]struct{})
	var leafIDs descpb.IDs
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(241027)
		descendantLeafIDs, err := s.findDescendantLeafIDs(ctx, id, s.txn, s.descsCol)
		if err != nil {
			__antithesis_instrumentation__.Notify(241029)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(241030)
		}
		__antithesis_instrumentation__.Notify(241028)
		for _, descendantLeafID := range descendantLeafIDs {
			__antithesis_instrumentation__.Notify(241031)
			if _, found := seen[descendantLeafID]; !found {
				__antithesis_instrumentation__.Notify(241032)
				seen[descendantLeafID] = struct{}{}
				leafIDs = append(leafIDs, descendantLeafID)
			} else {
				__antithesis_instrumentation__.Notify(241033)
			}
		}
	}
	__antithesis_instrumentation__.Notify(241016)

	pseudoTableRecords, err := s.maybeGeneratePseudoTableRecords(ctx, s.txn, ids)
	if err != nil {
		__antithesis_instrumentation__.Notify(241034)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(241035)
	}
	__antithesis_instrumentation__.Notify(241017)
	records = append(records, pseudoTableRecords...)

	scratchRangeRecord, err := s.maybeGenerateScratchRangeRecord(ctx, s.txn, ids)
	if err != nil {
		__antithesis_instrumentation__.Notify(241036)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(241037)
	}
	__antithesis_instrumentation__.Notify(241018)
	if !scratchRangeRecord.IsEmpty() {
		__antithesis_instrumentation__.Notify(241038)
		records = append(records, scratchRangeRecord)
	} else {
		__antithesis_instrumentation__.Notify(241039)
	}
	__antithesis_instrumentation__.Notify(241019)

	for _, leafID := range leafIDs {
		__antithesis_instrumentation__.Notify(241040)
		translatedRecords, err := s.generateSpanConfigurations(ctx, leafID, s.txn, s.descsCol, ptsStateReader)
		if err != nil {
			__antithesis_instrumentation__.Notify(241042)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(241043)
		}
		__antithesis_instrumentation__.Notify(241041)
		records = append(records, translatedRecords...)
	}
	__antithesis_instrumentation__.Notify(241020)

	return records, s.txn.CommitTimestamp(), nil
}

var descLookupFlags = tree.CommonLookupFlags{

	Required: true,

	IncludeDropped: true,
	IncludeOffline: true,

	AvoidLeased: true,
}

func (s *SQLTranslator) generateSystemSpanConfigRecords(
	ptsStateReader *spanconfig.ProtectedTimestampStateReader,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241044)
	tenantPrefix := s.codec.TenantPrefix()
	_, sourceTenantID, err := keys.DecodeTenantPrefix(tenantPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(241048)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241049)
	}
	__antithesis_instrumentation__.Notify(241045)
	records := make([]spanconfig.Record, 0)

	clusterProtections := ptsStateReader.GetProtectionPoliciesForCluster()
	if len(clusterProtections) != 0 {
		__antithesis_instrumentation__.Notify(241050)
		var systemTarget spanconfig.SystemTarget
		var err error
		if sourceTenantID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(241053)
			systemTarget = spanconfig.MakeEntireKeyspaceTarget()
		} else {
			__antithesis_instrumentation__.Notify(241054)
			systemTarget, err = spanconfig.MakeTenantKeyspaceTarget(sourceTenantID, sourceTenantID)
			if err != nil {
				__antithesis_instrumentation__.Notify(241055)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(241056)
			}
		}
		__antithesis_instrumentation__.Notify(241051)
		clusterSystemRecord, err := spanconfig.MakeRecord(
			spanconfig.MakeTargetFromSystemTarget(systemTarget),
			roachpb.SpanConfig{GCPolicy: roachpb.GCPolicy{ProtectionPolicies: clusterProtections}})
		if err != nil {
			__antithesis_instrumentation__.Notify(241057)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241058)
		}
		__antithesis_instrumentation__.Notify(241052)
		records = append(records, clusterSystemRecord)
	} else {
		__antithesis_instrumentation__.Notify(241059)
	}
	__antithesis_instrumentation__.Notify(241046)

	tenantProtections := ptsStateReader.GetProtectionPoliciesForTenants()
	for _, protection := range tenantProtections {
		__antithesis_instrumentation__.Notify(241060)
		tenantProtection := protection
		systemTarget, err := spanconfig.MakeTenantKeyspaceTarget(sourceTenantID, tenantProtection.GetTenantID())
		if err != nil {
			__antithesis_instrumentation__.Notify(241063)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241064)
		}
		__antithesis_instrumentation__.Notify(241061)
		tenantSystemRecord, err := spanconfig.MakeRecord(
			spanconfig.MakeTargetFromSystemTarget(systemTarget),
			roachpb.SpanConfig{GCPolicy: roachpb.GCPolicy{
				ProtectionPolicies: tenantProtection.GetTenantProtections()}})
		if err != nil {
			__antithesis_instrumentation__.Notify(241065)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241066)
		}
		__antithesis_instrumentation__.Notify(241062)
		records = append(records, tenantSystemRecord)
	}
	__antithesis_instrumentation__.Notify(241047)
	return records, nil
}

func (s *SQLTranslator) generateSpanConfigurations(
	ctx context.Context,
	id descpb.ID,
	txn *kv.Txn,
	descsCol *descs.Collection,
	ptsStateReader *spanconfig.ProtectedTimestampStateReader,
) (_ []spanconfig.Record, err error) {
	__antithesis_instrumentation__.Notify(241067)
	if zonepb.IsNamedZoneID(uint32(id)) {
		__antithesis_instrumentation__.Notify(241073)
		return s.generateSpanConfigurationsForNamedZone(ctx, txn, id)
	} else {
		__antithesis_instrumentation__.Notify(241074)
	}
	__antithesis_instrumentation__.Notify(241068)

	desc, err := descsCol.GetImmutableDescriptorByID(ctx, txn, id, descLookupFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(241075)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(241077)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(241078)
		}
		__antithesis_instrumentation__.Notify(241076)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241079)
	}
	__antithesis_instrumentation__.Notify(241069)
	if s.knobs.ExcludeDroppedDescriptorsFromLookup && func() bool {
		__antithesis_instrumentation__.Notify(241080)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241081)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(241082)
	}
	__antithesis_instrumentation__.Notify(241070)

	if desc.DescriptorType() != catalog.Table {
		__antithesis_instrumentation__.Notify(241083)
		return nil, errors.AssertionFailedf(
			"can only generate span configurations for tables, but got %s", desc.DescriptorType(),
		)
	} else {
		__antithesis_instrumentation__.Notify(241084)
	}
	__antithesis_instrumentation__.Notify(241071)

	table, ok := desc.(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(241085)
		return nil, errors.AssertionFailedf(
			"can only generate span configurations for tables, but got %s", desc.DescriptorType(),
		)
	} else {
		__antithesis_instrumentation__.Notify(241086)
	}
	__antithesis_instrumentation__.Notify(241072)
	return s.generateSpanConfigurationsForTable(ctx, txn, table, ptsStateReader)
}

func (s *SQLTranslator) generateSpanConfigurationsForNamedZone(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241087)
	name, ok := zonepb.NamedZonesByID[uint32(id)]
	if !ok {
		__antithesis_instrumentation__.Notify(241093)
		return nil, errors.AssertionFailedf("id %d does not belong to a named zone", id)
	} else {
		__antithesis_instrumentation__.Notify(241094)
	}
	__antithesis_instrumentation__.Notify(241088)

	if !s.codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(241095)
		return name != zonepb.DefaultZoneName == true
	}() == true {
		__antithesis_instrumentation__.Notify(241096)
		return nil,
			errors.AssertionFailedf("secondary tenants do not have the notion of %s named zone", name)
	} else {
		__antithesis_instrumentation__.Notify(241097)
	}
	__antithesis_instrumentation__.Notify(241089)

	var spans []roachpb.Span
	switch name {
	case zonepb.DefaultZoneName:
		__antithesis_instrumentation__.Notify(241098)
	case zonepb.MetaZoneName:
		__antithesis_instrumentation__.Notify(241099)
		spans = append(spans, roachpb.Span{Key: keys.Meta1Span.Key, EndKey: keys.NodeLivenessSpan.Key})
	case zonepb.LivenessZoneName:
		__antithesis_instrumentation__.Notify(241100)
		spans = append(spans, keys.NodeLivenessSpan)
	case zonepb.TimeseriesZoneName:
		__antithesis_instrumentation__.Notify(241101)
		spans = append(spans, keys.TimeseriesSpan)
	case zonepb.SystemZoneName:
		__antithesis_instrumentation__.Notify(241102)

		spans = append(spans, roachpb.Span{
			Key:    keys.NodeLivenessSpan.EndKey,
			EndKey: keys.TimeseriesSpan.Key,
		})
		spans = append(spans, roachpb.Span{
			Key:    keys.TimeseriesSpan.EndKey,
			EndKey: keys.SystemSpanConfigSpan.Key,
		})
	case zonepb.TenantsZoneName:
		__antithesis_instrumentation__.Notify(241103)
	default:
		__antithesis_instrumentation__.Notify(241104)
		return nil, errors.AssertionFailedf("unknown named zone config %s", name)
	}
	__antithesis_instrumentation__.Notify(241090)

	zoneConfig, err := sql.GetHydratedZoneConfigForNamedZone(ctx, txn, s.codec, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(241105)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241106)
	}
	__antithesis_instrumentation__.Notify(241091)
	spanConfig := zoneConfig.AsSpanConfig()
	var records []spanconfig.Record
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(241107)
		record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(span), spanConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(241109)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241110)
		}
		__antithesis_instrumentation__.Notify(241108)
		records = append(records, record)
	}
	__antithesis_instrumentation__.Notify(241092)
	return records, nil
}

func (s *SQLTranslator) generateSpanConfigurationsForTable(
	ctx context.Context,
	txn *kv.Txn,
	table catalog.TableDescriptor,
	ptsStateReader *spanconfig.ProtectedTimestampStateReader,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241111)

	if !table.IsPhysicalTable() {
		__antithesis_instrumentation__.Notify(241118)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(241119)
	}
	__antithesis_instrumentation__.Notify(241112)

	zone, err := sql.GetHydratedZoneConfigForTable(ctx, txn, s.codec, table.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(241120)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241121)
	}
	__antithesis_instrumentation__.Notify(241113)

	isSystemDesc := catalog.IsSystemDescriptor(table)
	tableStartKey := s.codec.TablePrefix(uint32(table.GetID()))
	tableEndKey := tableStartKey.PrefixEnd()
	tableSpanConfig := zone.AsSpanConfig()
	if isSystemDesc {
		__antithesis_instrumentation__.Notify(241122)

		tableSpanConfig.RangefeedEnabled = true

		tableSpanConfig.GCPolicy.IgnoreStrictEnforcement = true
	} else {
		__antithesis_instrumentation__.Notify(241123)
	}
	__antithesis_instrumentation__.Notify(241114)

	tableSpanConfig.GCPolicy.ProtectionPolicies = append(
		ptsStateReader.GetProtectionPoliciesForSchemaObject(table.GetID()),
		ptsStateReader.GetProtectionPoliciesForSchemaObject(table.GetParentID())...)

	tableSpanConfig.ExcludeDataFromBackup = table.GetExcludeDataFromBackup()

	records := make([]spanconfig.Record, 0)
	if table.GetID() == keys.DescriptorTableID {
		__antithesis_instrumentation__.Notify(241124)

		if !s.codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(241126)

			record, err := spanconfig.MakeRecord(
				spanconfig.MakeTargetFromSpan(roachpb.Span{
					Key:    s.codec.TenantPrefix(),
					EndKey: tableEndKey,
				}), tableSpanConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(241128)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(241129)
			}
			__antithesis_instrumentation__.Notify(241127)
			records = append(records, record)
		} else {
			__antithesis_instrumentation__.Notify(241130)

			record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(roachpb.Span{
				Key:    keys.SystemConfigSpan.Key,
				EndKey: tableEndKey,
			}), tableSpanConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(241132)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(241133)
			}
			__antithesis_instrumentation__.Notify(241131)
			records = append(records, record)
		}
		__antithesis_instrumentation__.Notify(241125)

		return records, nil

	} else {
		__antithesis_instrumentation__.Notify(241134)
	}
	__antithesis_instrumentation__.Notify(241115)

	prevEndKey := tableStartKey
	for i := range zone.SubzoneSpans {
		__antithesis_instrumentation__.Notify(241135)

		span := roachpb.Span{
			Key:    append(s.codec.TablePrefix(uint32(table.GetID())), zone.SubzoneSpans[i].Key...),
			EndKey: append(s.codec.TablePrefix(uint32(table.GetID())), zone.SubzoneSpans[i].EndKey...),
		}

		{
			__antithesis_instrumentation__.Notify(241140)

			if zone.SubzoneSpans[i].EndKey == nil {
				__antithesis_instrumentation__.Notify(241141)
				span.EndKey = span.Key.PrefixEnd()
			} else {
				__antithesis_instrumentation__.Notify(241142)
			}
		}
		__antithesis_instrumentation__.Notify(241136)

		if !prevEndKey.Equal(span.Key) {
			__antithesis_instrumentation__.Notify(241143)
			record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(
				roachpb.Span{Key: prevEndKey, EndKey: span.Key}), tableSpanConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(241145)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(241146)
			}
			__antithesis_instrumentation__.Notify(241144)
			records = append(records, record)
		} else {
			__antithesis_instrumentation__.Notify(241147)
		}
		__antithesis_instrumentation__.Notify(241137)

		subzoneSpanConfig := zone.Subzones[zone.SubzoneSpans[i].SubzoneIndex].Config.AsSpanConfig()

		subzoneSpanConfig.GCPolicy.ProtectionPolicies = tableSpanConfig.GCPolicy.ProtectionPolicies[:]
		subzoneSpanConfig.ExcludeDataFromBackup = tableSpanConfig.ExcludeDataFromBackup
		if isSystemDesc {
			__antithesis_instrumentation__.Notify(241148)
			subzoneSpanConfig.RangefeedEnabled = true
			subzoneSpanConfig.GCPolicy.IgnoreStrictEnforcement = true
		} else {
			__antithesis_instrumentation__.Notify(241149)
		}
		__antithesis_instrumentation__.Notify(241138)
		record, err := spanconfig.MakeRecord(
			spanconfig.MakeTargetFromSpan(roachpb.Span{Key: span.Key, EndKey: span.EndKey}), subzoneSpanConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(241150)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241151)
		}
		__antithesis_instrumentation__.Notify(241139)
		records = append(records, record)

		prevEndKey = span.EndKey
	}
	__antithesis_instrumentation__.Notify(241116)

	if !prevEndKey.Equal(tableEndKey) {
		__antithesis_instrumentation__.Notify(241152)
		record, err := spanconfig.MakeRecord(
			spanconfig.MakeTargetFromSpan(roachpb.Span{Key: prevEndKey, EndKey: tableEndKey}), tableSpanConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(241154)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241155)
		}
		__antithesis_instrumentation__.Notify(241153)
		records = append(records, record)
	} else {
		__antithesis_instrumentation__.Notify(241156)
	}
	__antithesis_instrumentation__.Notify(241117)
	return records, nil
}

func (s *SQLTranslator) findDescendantLeafIDs(
	ctx context.Context, id descpb.ID, txn *kv.Txn, descsCol *descs.Collection,
) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(241157)
	if zonepb.IsNamedZoneID(uint32(id)) {
		__antithesis_instrumentation__.Notify(241159)
		return s.findDescendantLeafIDsForNamedZone(ctx, id, txn, descsCol)
	} else {
		__antithesis_instrumentation__.Notify(241160)
	}
	__antithesis_instrumentation__.Notify(241158)

	return s.findDescendantLeafIDsForDescriptor(ctx, id, txn, descsCol)
}

func (s *SQLTranslator) findDescendantLeafIDsForDescriptor(
	ctx context.Context, id descpb.ID, txn *kv.Txn, descsCol *descs.Collection,
) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(241161)
	desc, err := descsCol.GetImmutableDescriptorByID(ctx, txn, id, descLookupFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(241168)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(241170)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(241171)
		}
		__antithesis_instrumentation__.Notify(241169)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241172)
	}
	__antithesis_instrumentation__.Notify(241162)
	if s.knobs.ExcludeDroppedDescriptorsFromLookup && func() bool {
		__antithesis_instrumentation__.Notify(241173)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241174)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(241175)
	}
	__antithesis_instrumentation__.Notify(241163)

	switch desc.DescriptorType() {
	case catalog.Type, catalog.Schema:
		__antithesis_instrumentation__.Notify(241176)

		return nil, nil
	case catalog.Table:
		__antithesis_instrumentation__.Notify(241177)

		return descpb.IDs{id}, nil
	case catalog.Database:
		__antithesis_instrumentation__.Notify(241178)

	default:
		__antithesis_instrumentation__.Notify(241179)
		return nil, errors.AssertionFailedf("unknown descriptor type: %s", desc.DescriptorType())
	}
	__antithesis_instrumentation__.Notify(241164)

	if desc.Offline() || func() bool {
		__antithesis_instrumentation__.Notify(241180)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241181)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(241182)
	}
	__antithesis_instrumentation__.Notify(241165)

	tables, err := descsCol.GetAllTableDescriptorsInDatabase(ctx, txn, desc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(241183)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241184)
	}
	__antithesis_instrumentation__.Notify(241166)
	ret := make(descpb.IDs, 0, len(tables))
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(241185)
		ret = append(ret, table.GetID())
	}
	__antithesis_instrumentation__.Notify(241167)
	return ret, nil
}

func (s *SQLTranslator) findDescendantLeafIDsForNamedZone(
	ctx context.Context, id descpb.ID, txn *kv.Txn, descsCol *descs.Collection,
) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(241186)
	name, ok := zonepb.NamedZonesByID[uint32(id)]
	if !ok {
		__antithesis_instrumentation__.Notify(241192)
		return nil, errors.AssertionFailedf("id %d does not belong to a named zone", id)
	} else {
		__antithesis_instrumentation__.Notify(241193)
	}
	__antithesis_instrumentation__.Notify(241187)

	if name != zonepb.DefaultZoneName {
		__antithesis_instrumentation__.Notify(241194)

		return descpb.IDs{id}, nil
	} else {
		__antithesis_instrumentation__.Notify(241195)
	}
	__antithesis_instrumentation__.Notify(241188)

	databases, err := descsCol.GetAllDatabaseDescriptors(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(241196)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241197)
	}
	__antithesis_instrumentation__.Notify(241189)
	var descendantIDs descpb.IDs
	for _, dbDesc := range databases {
		__antithesis_instrumentation__.Notify(241198)
		tableIDs, err := s.findDescendantLeafIDsForDescriptor(
			ctx, dbDesc.GetID(), txn, descsCol,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(241200)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241201)
		}
		__antithesis_instrumentation__.Notify(241199)
		descendantIDs = append(descendantIDs, tableIDs...)
	}
	__antithesis_instrumentation__.Notify(241190)

	if s.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(241202)
		for _, namedZone := range zonepb.NamedZonesList {
			__antithesis_instrumentation__.Notify(241203)

			if namedZone == zonepb.DefaultZoneName {
				__antithesis_instrumentation__.Notify(241205)
				continue
			} else {
				__antithesis_instrumentation__.Notify(241206)
			}
			__antithesis_instrumentation__.Notify(241204)
			descendantIDs = append(descendantIDs, descpb.ID(zonepb.NamedZones[namedZone]))
		}
	} else {
		__antithesis_instrumentation__.Notify(241207)
	}
	__antithesis_instrumentation__.Notify(241191)
	return descendantIDs, nil
}

func (s *SQLTranslator) maybeGeneratePseudoTableRecords(
	ctx context.Context, txn *kv.Txn, ids descpb.IDs,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241208)
	if !s.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(241211)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(241212)
	}
	__antithesis_instrumentation__.Notify(241209)

	for _, id := range ids {
		__antithesis_instrumentation__.Notify(241213)
		if id != keys.SystemDatabaseID && func() bool {
			__antithesis_instrumentation__.Notify(241217)
			return id != keys.RootNamespaceID == true
		}() == true {
			__antithesis_instrumentation__.Notify(241218)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241219)
		}
		__antithesis_instrumentation__.Notify(241214)

		zone, err := sql.GetHydratedZoneConfigForDatabase(ctx, txn, s.codec, keys.SystemDatabaseID)
		if err != nil {
			__antithesis_instrumentation__.Notify(241220)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(241221)
		}
		__antithesis_instrumentation__.Notify(241215)
		tableSpanConfig := zone.AsSpanConfig()
		var records []spanconfig.Record
		for _, pseudoTableID := range keys.PseudoTableIDs {
			__antithesis_instrumentation__.Notify(241222)
			tableStartKey := s.codec.TablePrefix(pseudoTableID)
			tableEndKey := tableStartKey.PrefixEnd()
			record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(roachpb.Span{
				Key:    tableStartKey,
				EndKey: tableEndKey,
			}), tableSpanConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(241224)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(241225)
			}
			__antithesis_instrumentation__.Notify(241223)
			records = append(records, record)
		}
		__antithesis_instrumentation__.Notify(241216)

		return records, nil
	}
	__antithesis_instrumentation__.Notify(241210)

	return nil, nil
}

func (s *SQLTranslator) maybeGenerateScratchRangeRecord(
	ctx context.Context, txn *kv.Txn, ids descpb.IDs,
) (spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241226)
	if !s.knobs.ConfigureScratchRange || func() bool {
		__antithesis_instrumentation__.Notify(241229)
		return !s.codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(241230)
		return spanconfig.Record{}, nil
	} else {
		__antithesis_instrumentation__.Notify(241231)
	}
	__antithesis_instrumentation__.Notify(241227)

	for _, id := range ids {
		__antithesis_instrumentation__.Notify(241232)
		if id != keys.RootNamespaceID {
			__antithesis_instrumentation__.Notify(241236)
			continue
		} else {
			__antithesis_instrumentation__.Notify(241237)
		}
		__antithesis_instrumentation__.Notify(241233)

		zone, err := sql.GetHydratedZoneConfigForDatabase(ctx, txn, s.codec, keys.RootNamespaceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(241238)
			return spanconfig.Record{}, err
		} else {
			__antithesis_instrumentation__.Notify(241239)
		}
		__antithesis_instrumentation__.Notify(241234)

		record, err := spanconfig.MakeRecord(
			spanconfig.MakeTargetFromSpan(roachpb.Span{
				Key:    keys.ScratchRangeMin,
				EndKey: keys.ScratchRangeMax,
			}), zone.AsSpanConfig())
		if err != nil {
			__antithesis_instrumentation__.Notify(241240)
			return spanconfig.Record{}, err
		} else {
			__antithesis_instrumentation__.Notify(241241)
		}
		__antithesis_instrumentation__.Notify(241235)
		return record, nil
	}
	__antithesis_instrumentation__.Notify(241228)

	return spanconfig.Record{}, nil
}
