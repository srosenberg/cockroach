package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type ObjectID uint32

type zoneConfigHook func(
	sysCfg *SystemConfig, codec keys.SQLCodec, objectID ObjectID,
) (zone *zonepb.ZoneConfig, placeholder *zonepb.ZoneConfig, cache bool, err error)

var (
	ZoneConfigHook zoneConfigHook

	testingLargestIDHook func(maxID ObjectID) ObjectID
)

type zoneEntry struct {
	zone        *zonepb.ZoneConfig
	placeholder *zonepb.ZoneConfig

	combined *zonepb.ZoneConfig
}

type SystemConfig struct {
	SystemConfigEntries
	DefaultZoneConfig *zonepb.ZoneConfig
	mu                struct {
		syncutil.RWMutex
		zoneCache        map[ObjectID]zoneEntry
		shouldSplitCache map[ObjectID]bool
	}
}

func NewSystemConfig(defaultZoneConfig *zonepb.ZoneConfig) *SystemConfig {
	__antithesis_instrumentation__.Notify(55795)
	sc := &SystemConfig{}
	sc.DefaultZoneConfig = defaultZoneConfig
	sc.mu.zoneCache = map[ObjectID]zoneEntry{}
	sc.mu.shouldSplitCache = map[ObjectID]bool{}
	return sc
}

func (s *SystemConfig) Equal(other *SystemConfigEntries) bool {
	__antithesis_instrumentation__.Notify(55796)
	if len(s.Values) != len(other.Values) {
		__antithesis_instrumentation__.Notify(55799)
		return false
	} else {
		__antithesis_instrumentation__.Notify(55800)
	}
	__antithesis_instrumentation__.Notify(55797)
	for i := range s.Values {
		__antithesis_instrumentation__.Notify(55801)
		leftKV, rightKV := s.Values[i], other.Values[i]
		if !leftKV.Key.Equal(rightKV.Key) {
			__antithesis_instrumentation__.Notify(55804)
			return false
		} else {
			__antithesis_instrumentation__.Notify(55805)
		}
		__antithesis_instrumentation__.Notify(55802)
		leftVal, rightVal := leftKV.Value, rightKV.Value
		if !leftVal.EqualTagAndData(rightVal) {
			__antithesis_instrumentation__.Notify(55806)
			return false
		} else {
			__antithesis_instrumentation__.Notify(55807)
		}
		__antithesis_instrumentation__.Notify(55803)
		if leftVal.Timestamp != rightVal.Timestamp {
			__antithesis_instrumentation__.Notify(55808)
			return false
		} else {
			__antithesis_instrumentation__.Notify(55809)
		}
	}
	__antithesis_instrumentation__.Notify(55798)
	return true
}

func (s *SystemConfig) getSystemTenantDesc(key roachpb.Key) *roachpb.Value {
	__antithesis_instrumentation__.Notify(55810)
	if getVal := s.GetValue(key); getVal != nil {
		__antithesis_instrumentation__.Notify(55814)
		return getVal
	} else {
		__antithesis_instrumentation__.Notify(55815)
	}
	__antithesis_instrumentation__.Notify(55811)

	id, err := keys.SystemSQLCodec.DecodeDescMetadataID(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(55816)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(55817)
	}
	__antithesis_instrumentation__.Notify(55812)

	testingLock.Lock()
	_, ok := testingZoneConfig[ObjectID(id)]
	testingLock.Unlock()

	if ok {
		__antithesis_instrumentation__.Notify(55818)

		desc := tabledesc.NewBuilder(&descpb.TableDescriptor{}).BuildImmutable().DescriptorProto()
		var val roachpb.Value
		if err := val.SetProto(desc); err != nil {
			__antithesis_instrumentation__.Notify(55820)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(55821)
		}
		__antithesis_instrumentation__.Notify(55819)
		return &val
	} else {
		__antithesis_instrumentation__.Notify(55822)
	}
	__antithesis_instrumentation__.Notify(55813)
	return nil
}

func (s *SystemConfig) GetValue(key roachpb.Key) *roachpb.Value {
	__antithesis_instrumentation__.Notify(55823)
	if kv := s.get(key); kv != nil {
		__antithesis_instrumentation__.Notify(55825)
		return &kv.Value
	} else {
		__antithesis_instrumentation__.Notify(55826)
	}
	__antithesis_instrumentation__.Notify(55824)
	return nil
}

func (s *SystemConfig) get(key roachpb.Key) *roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(55827)
	if i, ok := s.GetIndex(key); ok {
		__antithesis_instrumentation__.Notify(55829)
		return &s.Values[i]
	} else {
		__antithesis_instrumentation__.Notify(55830)
	}
	__antithesis_instrumentation__.Notify(55828)
	return nil
}

func (s *SystemConfig) GetIndex(key roachpb.Key) (int, bool) {
	__antithesis_instrumentation__.Notify(55831)
	i := s.getIndexBound(key)
	if i == len(s.Values) || func() bool {
		__antithesis_instrumentation__.Notify(55833)
		return !key.Equal(s.Values[i].Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(55834)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(55835)
	}
	__antithesis_instrumentation__.Notify(55832)
	return i, true
}

func (s *SystemConfig) getIndexBound(key roachpb.Key) int {
	__antithesis_instrumentation__.Notify(55836)
	return sort.Search(len(s.Values), func(i int) bool {
		__antithesis_instrumentation__.Notify(55837)
		return key.Compare(s.Values[i].Key) <= 0
	})
}

func (s *SystemConfig) GetLargestObjectID(
	maxReservedDescID ObjectID, pseudoIDs []uint32,
) (ObjectID, error) {
	__antithesis_instrumentation__.Notify(55838)
	testingLock.Lock()
	hook := testingLargestIDHook
	testingLock.Unlock()
	if hook != nil {
		__antithesis_instrumentation__.Notify(55849)
		return hook(maxReservedDescID), nil
	} else {
		__antithesis_instrumentation__.Notify(55850)
	}
	__antithesis_instrumentation__.Notify(55839)

	lowBound := keys.SystemSQLCodec.TablePrefix(keys.DescriptorTableID)
	lowIndex := s.getIndexBound(lowBound)
	highBound := keys.SystemSQLCodec.TablePrefix(keys.DescriptorTableID + 1)
	highIndex := s.getIndexBound(highBound)
	if lowIndex == highIndex {
		__antithesis_instrumentation__.Notify(55851)
		return 0, fmt.Errorf("descriptor table not found in system config of %d values", len(s.Values))
	} else {
		__antithesis_instrumentation__.Notify(55852)
	}
	__antithesis_instrumentation__.Notify(55840)

	maxPseudoID := ObjectID(0)
	for _, id := range pseudoIDs {
		__antithesis_instrumentation__.Notify(55853)
		objID := ObjectID(id)
		if objID > maxPseudoID && func() bool {
			__antithesis_instrumentation__.Notify(55854)
			return (maxReservedDescID == 0 || func() bool {
				__antithesis_instrumentation__.Notify(55855)
				return objID <= maxReservedDescID == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(55856)
			maxPseudoID = objID
		} else {
			__antithesis_instrumentation__.Notify(55857)
		}
	}
	__antithesis_instrumentation__.Notify(55841)

	if maxReservedDescID == 0 {
		__antithesis_instrumentation__.Notify(55858)
		id, err := keys.SystemSQLCodec.DecodeDescMetadataID(s.Values[highIndex-1].Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(55861)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(55862)
		}
		__antithesis_instrumentation__.Notify(55859)
		objID := ObjectID(id)
		if objID < maxPseudoID {
			__antithesis_instrumentation__.Notify(55863)
			objID = maxPseudoID
		} else {
			__antithesis_instrumentation__.Notify(55864)
		}
		__antithesis_instrumentation__.Notify(55860)
		return objID, nil
	} else {
		__antithesis_instrumentation__.Notify(55865)
	}
	__antithesis_instrumentation__.Notify(55842)

	searchSlice := s.Values[lowIndex:highIndex]
	var err error
	maxIdx := sort.Search(len(searchSlice), func(i int) bool {
		__antithesis_instrumentation__.Notify(55866)
		if err != nil {
			__antithesis_instrumentation__.Notify(55868)
			return false
		} else {
			__antithesis_instrumentation__.Notify(55869)
		}
		__antithesis_instrumentation__.Notify(55867)
		var id uint32
		id, err = keys.SystemSQLCodec.DecodeDescMetadataID(searchSlice[i].Key)
		return uint32(maxReservedDescID) < id
	})
	__antithesis_instrumentation__.Notify(55843)
	if err != nil {
		__antithesis_instrumentation__.Notify(55870)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(55871)
	}
	__antithesis_instrumentation__.Notify(55844)

	if maxIdx < len(searchSlice) {
		__antithesis_instrumentation__.Notify(55872)
		id, err := keys.SystemSQLCodec.DecodeDescMetadataID(searchSlice[maxIdx].Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(55874)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(55875)
		}
		__antithesis_instrumentation__.Notify(55873)
		if id <= uint32(maxReservedDescID) {
			__antithesis_instrumentation__.Notify(55876)
			return ObjectID(id), nil
		} else {
			__antithesis_instrumentation__.Notify(55877)
		}
	} else {
		__antithesis_instrumentation__.Notify(55878)
	}
	__antithesis_instrumentation__.Notify(55845)

	if maxIdx == 0 {
		__antithesis_instrumentation__.Notify(55879)
		return 0, fmt.Errorf("no system descriptors present")
	} else {
		__antithesis_instrumentation__.Notify(55880)
	}
	__antithesis_instrumentation__.Notify(55846)

	id, err := keys.SystemSQLCodec.DecodeDescMetadataID(searchSlice[maxIdx-1].Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(55881)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(55882)
	}
	__antithesis_instrumentation__.Notify(55847)
	objID := ObjectID(id)
	if objID < maxPseudoID {
		__antithesis_instrumentation__.Notify(55883)
		objID = maxPseudoID
	} else {
		__antithesis_instrumentation__.Notify(55884)
	}
	__antithesis_instrumentation__.Notify(55848)
	return objID, nil
}

func TestingGetSystemTenantZoneConfigForKey(
	s *SystemConfig, key roachpb.RKey,
) (ObjectID, *zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(55885)
	return s.getZoneConfigForKey(keys.SystemSQLCodec, key)
}

func (s *SystemConfig) getZoneConfigForKey(
	codec keys.SQLCodec, key roachpb.RKey,
) (ObjectID, *zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(55886)
	id, suffix := DecodeKeyIntoZoneIDAndSuffix(codec, key)
	entry, err := s.getZoneEntry(codec, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(55889)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(55890)
	}
	__antithesis_instrumentation__.Notify(55887)
	if entry.zone != nil {
		__antithesis_instrumentation__.Notify(55891)
		if entry.placeholder != nil {
			__antithesis_instrumentation__.Notify(55893)
			if subzone, _ := entry.placeholder.GetSubzoneForKeySuffix(suffix); subzone != nil {
				__antithesis_instrumentation__.Notify(55894)
				if indexSubzone := entry.placeholder.GetSubzone(subzone.IndexID, ""); indexSubzone != nil {
					__antithesis_instrumentation__.Notify(55896)
					subzone.Config.InheritFromParent(&indexSubzone.Config)
				} else {
					__antithesis_instrumentation__.Notify(55897)
				}
				__antithesis_instrumentation__.Notify(55895)
				subzone.Config.InheritFromParent(entry.zone)
				return id, &subzone.Config, nil
			} else {
				__antithesis_instrumentation__.Notify(55898)
			}
		} else {
			__antithesis_instrumentation__.Notify(55899)
			if subzone, _ := entry.zone.GetSubzoneForKeySuffix(suffix); subzone != nil {
				__antithesis_instrumentation__.Notify(55900)
				if indexSubzone := entry.zone.GetSubzone(subzone.IndexID, ""); indexSubzone != nil {
					__antithesis_instrumentation__.Notify(55902)
					subzone.Config.InheritFromParent(&indexSubzone.Config)
				} else {
					__antithesis_instrumentation__.Notify(55903)
				}
				__antithesis_instrumentation__.Notify(55901)
				subzone.Config.InheritFromParent(entry.zone)
				return id, &subzone.Config, nil
			} else {
				__antithesis_instrumentation__.Notify(55904)
			}
		}
		__antithesis_instrumentation__.Notify(55892)
		return id, entry.zone, nil
	} else {
		__antithesis_instrumentation__.Notify(55905)
	}
	__antithesis_instrumentation__.Notify(55888)
	return id, s.DefaultZoneConfig, nil
}

func (s *SystemConfig) GetSpanConfigForKey(
	ctx context.Context, key roachpb.RKey,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(55906)
	id, zone, err := s.getZoneConfigForKey(keys.SystemSQLCodec, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(55909)
		return roachpb.SpanConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(55910)
	}
	__antithesis_instrumentation__.Notify(55907)
	spanConfig := zone.AsSpanConfig()
	if id <= keys.MaxReservedDescID {
		__antithesis_instrumentation__.Notify(55911)

		spanConfig.RangefeedEnabled = true

		spanConfig.GCPolicy.IgnoreStrictEnforcement = true
	} else {
		__antithesis_instrumentation__.Notify(55912)
	}
	__antithesis_instrumentation__.Notify(55908)
	return spanConfig, nil
}

func DecodeKeyIntoZoneIDAndSuffix(
	codec keys.SQLCodec, key roachpb.RKey,
) (id ObjectID, keySuffix []byte) {
	__antithesis_instrumentation__.Notify(55913)
	objectID, keySuffix, ok := DecodeObjectID(codec, key)
	if !ok {
		__antithesis_instrumentation__.Notify(55916)

		objectID = keys.RootNamespaceID
	} else {
		__antithesis_instrumentation__.Notify(55917)
		if objectID <= keys.MaxSystemConfigDescID || func() bool {
			__antithesis_instrumentation__.Notify(55918)
			return isPseudoTableID(uint32(objectID)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(55919)

			objectID = keys.SystemDatabaseID
		} else {
			__antithesis_instrumentation__.Notify(55920)
		}
	}
	__antithesis_instrumentation__.Notify(55914)

	if key.Equal(roachpb.RKeyMin) || func() bool {
		__antithesis_instrumentation__.Notify(55921)
		return bytes.HasPrefix(key, keys.Meta1Prefix) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(55922)
		return bytes.HasPrefix(key, keys.Meta2Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(55923)
		objectID = keys.MetaRangesID
	} else {
		__antithesis_instrumentation__.Notify(55924)
		if bytes.HasPrefix(key, keys.SystemPrefix) {
			__antithesis_instrumentation__.Notify(55925)
			if bytes.HasPrefix(key, keys.NodeLivenessPrefix) {
				__antithesis_instrumentation__.Notify(55926)
				objectID = keys.LivenessRangesID
			} else {
				__antithesis_instrumentation__.Notify(55927)
				if bytes.HasPrefix(key, keys.TimeseriesPrefix) {
					__antithesis_instrumentation__.Notify(55928)
					objectID = keys.TimeseriesRangesID
				} else {
					__antithesis_instrumentation__.Notify(55929)
					objectID = keys.SystemRangesID
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(55930)
			if bytes.HasPrefix(key, keys.TenantPrefix) {
				__antithesis_instrumentation__.Notify(55931)
				objectID = keys.TenantsRangesID
			} else {
				__antithesis_instrumentation__.Notify(55932)
			}
		}
	}
	__antithesis_instrumentation__.Notify(55915)
	return objectID, keySuffix
}

func isPseudoTableID(id uint32) bool {
	__antithesis_instrumentation__.Notify(55933)
	for _, pseudoTableID := range keys.PseudoTableIDs {
		__antithesis_instrumentation__.Notify(55935)
		if id == pseudoTableID {
			__antithesis_instrumentation__.Notify(55936)
			return true
		} else {
			__antithesis_instrumentation__.Notify(55937)
		}
	}
	__antithesis_instrumentation__.Notify(55934)
	return false
}

func (s *SystemConfig) GetZoneConfigForObject(
	codec keys.SQLCodec, version clusterversion.ClusterVersion, id ObjectID,
) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(55938)
	var entry zoneEntry
	var err error

	if !codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(55941)
		return (id == 0 || func() bool {
			__antithesis_instrumentation__.Notify(55942)
			return !version.IsActive(clusterversion.EnableSpanConfigStore) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(55943)
		codec, id = keys.SystemSQLCodec, keys.TenantsRangesID
	} else {
		__antithesis_instrumentation__.Notify(55944)
	}
	__antithesis_instrumentation__.Notify(55939)
	entry, err = s.getZoneEntry(codec, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(55945)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55946)
	}
	__antithesis_instrumentation__.Notify(55940)
	return entry.combined, nil
}

func (s *SystemConfig) getZoneEntry(codec keys.SQLCodec, id ObjectID) (zoneEntry, error) {
	__antithesis_instrumentation__.Notify(55947)
	s.mu.RLock()
	entry, ok := s.mu.zoneCache[id]
	s.mu.RUnlock()
	if ok {
		__antithesis_instrumentation__.Notify(55951)
		return entry, nil
	} else {
		__antithesis_instrumentation__.Notify(55952)
	}
	__antithesis_instrumentation__.Notify(55948)
	testingLock.Lock()
	hook := ZoneConfigHook
	testingLock.Unlock()
	zone, placeholder, cache, err := hook(s, codec, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(55953)
		return zoneEntry{}, err
	} else {
		__antithesis_instrumentation__.Notify(55954)
	}
	__antithesis_instrumentation__.Notify(55949)
	if zone != nil {
		__antithesis_instrumentation__.Notify(55955)
		entry := zoneEntry{zone: zone, placeholder: placeholder, combined: zone}
		if placeholder != nil {
			__antithesis_instrumentation__.Notify(55958)

			combined := *zone
			combined.Subzones = placeholder.Subzones
			combined.SubzoneSpans = placeholder.SubzoneSpans
			entry.combined = &combined
		} else {
			__antithesis_instrumentation__.Notify(55959)
		}
		__antithesis_instrumentation__.Notify(55956)

		if cache {
			__antithesis_instrumentation__.Notify(55960)
			s.mu.Lock()
			s.mu.zoneCache[id] = entry
			s.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(55961)
		}
		__antithesis_instrumentation__.Notify(55957)
		return entry, nil
	} else {
		__antithesis_instrumentation__.Notify(55962)
	}
	__antithesis_instrumentation__.Notify(55950)
	return zoneEntry{}, nil
}

var staticSplits = []roachpb.RKey{
	roachpb.RKey(keys.NodeLivenessPrefix),
	roachpb.RKey(keys.NodeLivenessKeyMax),
	roachpb.RKey(keys.TimeseriesPrefix),
	roachpb.RKey(keys.TimeseriesPrefix.PrefixEnd()),
	roachpb.RKey(keys.TableDataMin),
}

func StaticSplits() []roachpb.RKey {
	__antithesis_instrumentation__.Notify(55963)
	return staticSplits
}

func (s *SystemConfig) ComputeSplitKey(
	ctx context.Context, startKey, endKey roachpb.RKey,
) (rr roachpb.RKey) {
	__antithesis_instrumentation__.Notify(55964)

	for _, split := range staticSplits {
		__antithesis_instrumentation__.Notify(55967)
		if startKey.Less(split) {
			__antithesis_instrumentation__.Notify(55968)
			if split.Less(endKey) {
				__antithesis_instrumentation__.Notify(55970)

				return split
			} else {
				__antithesis_instrumentation__.Notify(55971)
			}
			__antithesis_instrumentation__.Notify(55969)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(55972)
		}

	}
	__antithesis_instrumentation__.Notify(55965)

	if split := s.systemTenantTableBoundarySplitKey(ctx, startKey, endKey); split != nil {
		__antithesis_instrumentation__.Notify(55973)
		return split
	} else {
		__antithesis_instrumentation__.Notify(55974)
	}
	__antithesis_instrumentation__.Notify(55966)

	return s.tenantBoundarySplitKey(ctx, startKey, endKey)
}

func (s *SystemConfig) systemTenantTableBoundarySplitKey(
	ctx context.Context, startKey, endKey roachpb.RKey,
) roachpb.RKey {
	__antithesis_instrumentation__.Notify(55975)
	if bytes.HasPrefix(startKey, keys.TenantPrefix) {
		__antithesis_instrumentation__.Notify(55981)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(55982)
	}
	__antithesis_instrumentation__.Notify(55976)

	startID, _, ok := DecodeObjectID(keys.SystemSQLCodec, startKey)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(55983)
		return startID <= keys.MaxSystemConfigDescID == true
	}() == true {
		__antithesis_instrumentation__.Notify(55984)

		startID = keys.MaxSystemConfigDescID + 1
	} else {
		__antithesis_instrumentation__.Notify(55985)
	}
	__antithesis_instrumentation__.Notify(55977)

	findSplitKey := func(startID, endID ObjectID) roachpb.RKey {
		__antithesis_instrumentation__.Notify(55986)

		for id := startID; id <= endID; id++ {
			__antithesis_instrumentation__.Notify(55988)
			tableKey := roachpb.RKey(keys.SystemSQLCodec.TablePrefix(uint32(id)))

			if startKey.Less(tableKey) && func() bool {
				__antithesis_instrumentation__.Notify(55992)
				return s.shouldSplitOnSystemTenantObject(id) == true
			}() == true {
				__antithesis_instrumentation__.Notify(55993)
				if tableKey.Less(endKey) {
					__antithesis_instrumentation__.Notify(55995)
					return tableKey
				} else {
					__antithesis_instrumentation__.Notify(55996)
				}
				__antithesis_instrumentation__.Notify(55994)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(55997)
			}
			__antithesis_instrumentation__.Notify(55989)

			zoneVal := s.GetValue(MakeZoneKey(keys.SystemSQLCodec, descpb.ID(id)))
			if zoneVal == nil {
				__antithesis_instrumentation__.Notify(55998)
				continue
			} else {
				__antithesis_instrumentation__.Notify(55999)
			}
			__antithesis_instrumentation__.Notify(55990)
			var zone zonepb.ZoneConfig
			if err := zoneVal.GetProto(&zone); err != nil {
				__antithesis_instrumentation__.Notify(56000)

				continue
			} else {
				__antithesis_instrumentation__.Notify(56001)
			}
			__antithesis_instrumentation__.Notify(55991)

			for _, s := range zone.SubzoneSplits() {
				__antithesis_instrumentation__.Notify(56002)
				subzoneKey := append(tableKey, s...)
				if startKey.Less(subzoneKey) {
					__antithesis_instrumentation__.Notify(56003)
					if subzoneKey.Less(endKey) {
						__antithesis_instrumentation__.Notify(56005)
						return subzoneKey
					} else {
						__antithesis_instrumentation__.Notify(56006)
					}
					__antithesis_instrumentation__.Notify(56004)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(56007)
				}
			}
		}
		__antithesis_instrumentation__.Notify(55987)
		return nil
	}
	__antithesis_instrumentation__.Notify(55978)

	if uint32(startID) <= keys.MaxReservedDescID {
		__antithesis_instrumentation__.Notify(56008)
		endID, err := s.GetLargestObjectID(keys.MaxReservedDescID, keys.PseudoTableIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(56011)
			log.Errorf(ctx, "unable to determine largest reserved object ID from system config: %s", err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(56012)
		}
		__antithesis_instrumentation__.Notify(56009)
		if splitKey := findSplitKey(startID, endID); splitKey != nil {
			__antithesis_instrumentation__.Notify(56013)
			return splitKey
		} else {
			__antithesis_instrumentation__.Notify(56014)
		}
		__antithesis_instrumentation__.Notify(56010)
		startID = ObjectID(keys.MaxReservedDescID + 1)
	} else {
		__antithesis_instrumentation__.Notify(56015)
	}
	__antithesis_instrumentation__.Notify(55979)

	endID, err := s.GetLargestObjectID(0, keys.PseudoTableIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(56016)
		log.Errorf(ctx, "unable to determine largest object ID from system config: %s", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(56017)
	}
	__antithesis_instrumentation__.Notify(55980)
	return findSplitKey(startID, endID)
}

func (s *SystemConfig) tenantBoundarySplitKey(
	ctx context.Context, startKey, endKey roachpb.RKey,
) roachpb.RKey {
	__antithesis_instrumentation__.Notify(56018)

	searchSpan := roachpb.Span{Key: startKey.AsRawKey(), EndKey: endKey.AsRawKey()}
	tenantSpan := roachpb.Span{Key: keys.TenantTableDataMin, EndKey: keys.TenantTableDataMax}
	if !searchSpan.Overlaps(tenantSpan) {
		__antithesis_instrumentation__.Notify(56026)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(56027)
	}
	__antithesis_instrumentation__.Notify(56019)

	var lowTenID, highTenID roachpb.TenantID
	if searchSpan.Key.Compare(tenantSpan.Key) < 0 {
		__antithesis_instrumentation__.Notify(56028)

		lowTenID = roachpb.MinTenantID
	} else {
		__antithesis_instrumentation__.Notify(56029)

		_, lowTenIDExcl, err := keys.DecodeTenantPrefix(searchSpan.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(56032)
			log.Errorf(ctx, "unable to decode tenant ID from start key: %s", err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(56033)
		}
		__antithesis_instrumentation__.Notify(56030)
		if lowTenIDExcl == roachpb.MaxTenantID {
			__antithesis_instrumentation__.Notify(56034)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(56035)
		}
		__antithesis_instrumentation__.Notify(56031)
		lowTenID = roachpb.MakeTenantID(lowTenIDExcl.ToUint64() + 1)
	}
	__antithesis_instrumentation__.Notify(56020)
	if searchSpan.EndKey.Compare(tenantSpan.EndKey) >= 0 {
		__antithesis_instrumentation__.Notify(56036)

		highTenID = roachpb.MaxTenantID
	} else {
		__antithesis_instrumentation__.Notify(56037)
		rem, highTenIDExcl, err := keys.DecodeTenantPrefix(searchSpan.EndKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(56039)
			log.Errorf(ctx, "unable to decode tenant ID from end key: %s", err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(56040)
		}
		__antithesis_instrumentation__.Notify(56038)
		if len(rem) == 0 {
			__antithesis_instrumentation__.Notify(56041)

			highTenID = roachpb.MakeTenantID(highTenIDExcl.ToUint64() - 1)
		} else {
			__antithesis_instrumentation__.Notify(56042)
			highTenID = highTenIDExcl
		}
	}
	__antithesis_instrumentation__.Notify(56021)

	if lowTenID.ToUint64() > highTenID.ToUint64() {
		__antithesis_instrumentation__.Notify(56043)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(56044)
	}
	__antithesis_instrumentation__.Notify(56022)

	lowBound := keys.SystemSQLCodec.TenantMetadataKey(lowTenID)
	lowIndex := s.getIndexBound(lowBound)
	if lowIndex == len(s.Values) {
		__antithesis_instrumentation__.Notify(56045)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(56046)
	}
	__antithesis_instrumentation__.Notify(56023)

	splitKey := s.Values[lowIndex].Key
	splitTenID, err := keys.SystemSQLCodec.DecodeTenantMetadataID(splitKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(56047)
		log.Errorf(ctx, "unable to decode tenant ID from system config: %s", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(56048)
	}
	__antithesis_instrumentation__.Notify(56024)
	if splitTenID.ToUint64() > highTenID.ToUint64() {
		__antithesis_instrumentation__.Notify(56049)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(56050)
	}
	__antithesis_instrumentation__.Notify(56025)
	return roachpb.RKey(keys.MakeTenantPrefix(splitTenID))
}

func (s *SystemConfig) NeedsSplit(ctx context.Context, startKey, endKey roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(56051)
	return len(s.ComputeSplitKey(ctx, startKey, endKey)) > 0
}

func (s *SystemConfig) shouldSplitOnSystemTenantObject(id ObjectID) bool {
	__antithesis_instrumentation__.Notify(56052)

	{
		__antithesis_instrumentation__.Notify(56055)
		s.mu.RLock()
		shouldSplit, ok := s.mu.shouldSplitCache[id]
		s.mu.RUnlock()
		if ok {
			__antithesis_instrumentation__.Notify(56056)
			return shouldSplit
		} else {
			__antithesis_instrumentation__.Notify(56057)
		}
	}
	__antithesis_instrumentation__.Notify(56053)

	var shouldSplit bool
	if uint32(id) <= keys.MaxReservedDescID {
		__antithesis_instrumentation__.Notify(56058)

		shouldSplit = true
	} else {
		__antithesis_instrumentation__.Notify(56059)
		desc := s.getSystemTenantDesc(keys.SystemSQLCodec.DescMetadataKey(uint32(id)))
		shouldSplit = desc != nil && func() bool {
			__antithesis_instrumentation__.Notify(56060)
			return systemschema.ShouldSplitAtDesc(desc) == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(56054)

	s.mu.Lock()
	s.mu.shouldSplitCache[id] = shouldSplit
	s.mu.Unlock()
	return shouldSplit
}
