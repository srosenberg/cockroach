package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func init() {

	config.ZoneConfigHook = zoneConfigHook
}

var errNoZoneConfigApplies = errors.New("no zone config applies")

func getZoneConfig(
	codec keys.SQLCodec,
	id descpb.ID,
	getKey func(roachpb.Key) (*roachpb.Value, error),
	getInheritedDefault bool,
	mayBeTable bool,
) (descpb.ID, *zonepb.ZoneConfig, descpb.ID, *zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(632869)
	var placeholder *zonepb.ZoneConfig
	var placeholderID descpb.ID
	if !getInheritedDefault {
		__antithesis_instrumentation__.Notify(632873)

		if zoneVal, err := getKey(config.MakeZoneKey(codec, id)); err != nil {
			__antithesis_instrumentation__.Notify(632874)
			return 0, nil, 0, nil, err
		} else {
			__antithesis_instrumentation__.Notify(632875)
			if zoneVal != nil {
				__antithesis_instrumentation__.Notify(632876)

				var zone zonepb.ZoneConfig
				if err := zoneVal.GetProto(&zone); err != nil {
					__antithesis_instrumentation__.Notify(632879)
					return 0, nil, 0, nil, err
				} else {
					__antithesis_instrumentation__.Notify(632880)
				}
				__antithesis_instrumentation__.Notify(632877)

				if !zone.IsSubzonePlaceholder() {
					__antithesis_instrumentation__.Notify(632881)
					return id, &zone, 0, nil, nil
				} else {
					__antithesis_instrumentation__.Notify(632882)
				}
				__antithesis_instrumentation__.Notify(632878)

				placeholder = &zone
				placeholderID = id
			} else {
				__antithesis_instrumentation__.Notify(632883)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(632884)
	}
	__antithesis_instrumentation__.Notify(632870)

	if mayBeTable {
		__antithesis_instrumentation__.Notify(632885)
		if descVal, err := getKey(catalogkeys.MakeDescMetadataKey(codec, id)); err != nil {
			__antithesis_instrumentation__.Notify(632886)
			return 0, nil, 0, nil, err
		} else {
			__antithesis_instrumentation__.Notify(632887)
			if descVal != nil {
				__antithesis_instrumentation__.Notify(632888)
				var desc descpb.Descriptor
				if err := descVal.GetProto(&desc); err != nil {
					__antithesis_instrumentation__.Notify(632890)
					return 0, nil, 0, nil, err
				} else {
					__antithesis_instrumentation__.Notify(632891)
				}
				__antithesis_instrumentation__.Notify(632889)
				tableDesc, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(&desc, descVal.Timestamp)
				if tableDesc != nil {
					__antithesis_instrumentation__.Notify(632892)

					dbID, zone, _, _, err := getZoneConfig(
						codec,
						tableDesc.ParentID,
						getKey,
						false,
						false)
					if err != nil {
						__antithesis_instrumentation__.Notify(632894)
						return 0, nil, 0, nil, err
					} else {
						__antithesis_instrumentation__.Notify(632895)
					}
					__antithesis_instrumentation__.Notify(632893)
					return dbID, zone, placeholderID, placeholder, nil
				} else {
					__antithesis_instrumentation__.Notify(632896)
				}
			} else {
				__antithesis_instrumentation__.Notify(632897)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(632898)
	}
	__antithesis_instrumentation__.Notify(632871)

	if id != keys.RootNamespaceID {
		__antithesis_instrumentation__.Notify(632899)
		rootID, zone, _, _, err := getZoneConfig(
			codec,
			keys.RootNamespaceID,
			getKey,
			false,
			false)
		if err != nil {
			__antithesis_instrumentation__.Notify(632901)
			return 0, nil, 0, nil, err
		} else {
			__antithesis_instrumentation__.Notify(632902)
		}
		__antithesis_instrumentation__.Notify(632900)
		return rootID, zone, placeholderID, placeholder, nil
	} else {
		__antithesis_instrumentation__.Notify(632903)
	}
	__antithesis_instrumentation__.Notify(632872)

	return 0, nil, 0, nil, errNoZoneConfigApplies
}

func completeZoneConfig(
	cfg *zonepb.ZoneConfig,
	codec keys.SQLCodec,
	id descpb.ID,
	getKey func(roachpb.Key) (*roachpb.Value, error),
) error {
	__antithesis_instrumentation__.Notify(632904)
	if cfg.IsComplete() {
		__antithesis_instrumentation__.Notify(632909)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(632910)
	}
	__antithesis_instrumentation__.Notify(632905)

	if descVal, err := getKey(catalogkeys.MakeDescMetadataKey(codec, id)); err != nil {
		__antithesis_instrumentation__.Notify(632911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632912)
		if descVal != nil {
			__antithesis_instrumentation__.Notify(632913)
			var desc descpb.Descriptor
			if err := descVal.GetProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(632915)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632916)
			}
			__antithesis_instrumentation__.Notify(632914)
			tableDesc, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(&desc, descVal.Timestamp)
			if tableDesc != nil {
				__antithesis_instrumentation__.Notify(632917)
				_, dbzone, _, _, err := getZoneConfig(
					codec, tableDesc.ParentID, getKey, false, false)
				if err != nil {
					__antithesis_instrumentation__.Notify(632919)
					return err
				} else {
					__antithesis_instrumentation__.Notify(632920)
				}
				__antithesis_instrumentation__.Notify(632918)
				cfg.InheritFromParent(dbzone)
			} else {
				__antithesis_instrumentation__.Notify(632921)
			}
		} else {
			__antithesis_instrumentation__.Notify(632922)
		}
	}
	__antithesis_instrumentation__.Notify(632906)

	if cfg.IsComplete() {
		__antithesis_instrumentation__.Notify(632923)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(632924)
	}
	__antithesis_instrumentation__.Notify(632907)
	_, defaultZone, _, _, err := getZoneConfig(codec, keys.RootNamespaceID, getKey, false, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(632925)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632926)
	}
	__antithesis_instrumentation__.Notify(632908)
	cfg.InheritFromParent(defaultZone)
	return nil
}

func zoneConfigHook(
	cfg *config.SystemConfig, codec keys.SQLCodec, id config.ObjectID,
) (*zonepb.ZoneConfig, *zonepb.ZoneConfig, bool, error) {
	__antithesis_instrumentation__.Notify(632927)
	getKey := func(key roachpb.Key) (*roachpb.Value, error) {
		__antithesis_instrumentation__.Notify(632931)
		return cfg.GetValue(key), nil
	}
	__antithesis_instrumentation__.Notify(632928)
	const mayBeTable = true
	zoneID, zone, _, placeholder, err := getZoneConfig(
		codec, descpb.ID(id), getKey, false, mayBeTable)
	if errors.Is(err, errNoZoneConfigApplies) {
		__antithesis_instrumentation__.Notify(632932)
		return nil, nil, true, nil
	} else {
		__antithesis_instrumentation__.Notify(632933)
		if err != nil {
			__antithesis_instrumentation__.Notify(632934)
			return nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(632935)
		}
	}
	__antithesis_instrumentation__.Notify(632929)
	if err = completeZoneConfig(zone, codec, zoneID, getKey); err != nil {
		__antithesis_instrumentation__.Notify(632936)
		return nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(632937)
	}
	__antithesis_instrumentation__.Notify(632930)
	return zone, placeholder, true, nil
}

func GetZoneConfigInTxn(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	id descpb.ID,
	index catalog.Index,
	partition string,
	getInheritedDefault bool,
) (descpb.ID, *zonepb.ZoneConfig, *zonepb.Subzone, error) {
	__antithesis_instrumentation__.Notify(632938)
	getKey := func(key roachpb.Key) (*roachpb.Value, error) {
		__antithesis_instrumentation__.Notify(632943)
		kv, err := txn.Get(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(632945)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(632946)
		}
		__antithesis_instrumentation__.Notify(632944)
		return kv.Value, nil
	}
	__antithesis_instrumentation__.Notify(632939)
	zoneID, zone, placeholderID, placeholder, err := getZoneConfig(
		codec, id, getKey, getInheritedDefault, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(632947)
		return 0, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(632948)
	}
	__antithesis_instrumentation__.Notify(632940)
	if err = completeZoneConfig(zone, codec, zoneID, getKey); err != nil {
		__antithesis_instrumentation__.Notify(632949)
		return 0, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(632950)
	}
	__antithesis_instrumentation__.Notify(632941)
	var subzone *zonepb.Subzone
	if index != nil {
		__antithesis_instrumentation__.Notify(632951)
		indexID := uint32(index.GetID())
		if placeholder != nil {
			__antithesis_instrumentation__.Notify(632952)
			if subzone = placeholder.GetSubzone(indexID, partition); subzone != nil {
				__antithesis_instrumentation__.Notify(632953)
				if indexSubzone := placeholder.GetSubzone(indexID, ""); indexSubzone != nil {
					__antithesis_instrumentation__.Notify(632955)
					subzone.Config.InheritFromParent(&indexSubzone.Config)
				} else {
					__antithesis_instrumentation__.Notify(632956)
				}
				__antithesis_instrumentation__.Notify(632954)
				subzone.Config.InheritFromParent(zone)
				return placeholderID, placeholder, subzone, nil
			} else {
				__antithesis_instrumentation__.Notify(632957)
			}
		} else {
			__antithesis_instrumentation__.Notify(632958)
			if subzone = zone.GetSubzone(indexID, partition); subzone != nil {
				__antithesis_instrumentation__.Notify(632959)
				if indexSubzone := zone.GetSubzone(indexID, ""); indexSubzone != nil {
					__antithesis_instrumentation__.Notify(632961)
					subzone.Config.InheritFromParent(&indexSubzone.Config)
				} else {
					__antithesis_instrumentation__.Notify(632962)
				}
				__antithesis_instrumentation__.Notify(632960)
				subzone.Config.InheritFromParent(zone)
			} else {
				__antithesis_instrumentation__.Notify(632963)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(632964)
	}
	__antithesis_instrumentation__.Notify(632942)
	return zoneID, zone, subzone, nil
}

func GetHydratedZoneConfigForNamedZone(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, zoneName zonepb.NamedZone,
) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(632965)
	getKey := func(key roachpb.Key) (*roachpb.Value, error) {
		__antithesis_instrumentation__.Notify(632970)
		kv, err := txn.Get(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(632972)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(632973)
		}
		__antithesis_instrumentation__.Notify(632971)
		return kv.Value, nil
	}
	__antithesis_instrumentation__.Notify(632966)
	id, found := zonepb.NamedZones[zoneName]
	if !found {
		__antithesis_instrumentation__.Notify(632974)
		return nil, errors.AssertionFailedf("id %d does not belong to a named zone", id)
	} else {
		__antithesis_instrumentation__.Notify(632975)
	}
	__antithesis_instrumentation__.Notify(632967)
	zoneID, zone, _, _, err := getZoneConfig(
		codec, descpb.ID(id), getKey, false, false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(632976)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(632977)
	}
	__antithesis_instrumentation__.Notify(632968)
	if err := completeZoneConfig(zone, codec, zoneID, getKey); err != nil {
		__antithesis_instrumentation__.Notify(632978)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(632979)
	}
	__antithesis_instrumentation__.Notify(632969)
	return zone, nil
}

func GetHydratedZoneConfigForTable(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, id descpb.ID,
) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(632980)
	getKey := func(key roachpb.Key) (*roachpb.Value, error) {
		__antithesis_instrumentation__.Notify(632986)
		kv, err := txn.Get(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(632988)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(632989)
		}
		__antithesis_instrumentation__.Notify(632987)
		return kv.Value, nil
	}
	__antithesis_instrumentation__.Notify(632981)

	zoneID, zone, _, placeholder, err := getZoneConfig(
		codec, id, getKey, false, true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(632990)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(632991)
	}
	__antithesis_instrumentation__.Notify(632982)
	if err := completeZoneConfig(zone, codec, zoneID, getKey); err != nil {
		__antithesis_instrumentation__.Notify(632992)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(632993)
	}
	__antithesis_instrumentation__.Notify(632983)

	if placeholder != nil {
		__antithesis_instrumentation__.Notify(632994)

		if len(zone.Subzones) != 0 {
			__antithesis_instrumentation__.Notify(632996)
			return nil, errors.AssertionFailedf("placeholder %v exists in conjunction with subzones on zone config %v", *zone, *placeholder)
		} else {
			__antithesis_instrumentation__.Notify(632997)
		}
		__antithesis_instrumentation__.Notify(632995)
		zone.Subzones = placeholder.Subzones
		zone.SubzoneSpans = placeholder.SubzoneSpans
	} else {
		__antithesis_instrumentation__.Notify(632998)
	}
	__antithesis_instrumentation__.Notify(632984)

	for i, subzone := range zone.Subzones {
		__antithesis_instrumentation__.Notify(632999)

		indexSubzone := zone.GetSubzone(subzone.IndexID, "")

		if indexSubzone != nil {
			__antithesis_instrumentation__.Notify(633001)
			zone.Subzones[i].Config.InheritFromParent(&indexSubzone.Config)
		} else {
			__antithesis_instrumentation__.Notify(633002)
		}
		__antithesis_instrumentation__.Notify(633000)

		zone.Subzones[i].Config.InheritFromParent(zone)
	}
	__antithesis_instrumentation__.Notify(632985)

	return zone, nil
}

func GetHydratedZoneConfigForDatabase(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, id descpb.ID,
) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(633003)
	getKey := func(key roachpb.Key) (*roachpb.Value, error) {
		__antithesis_instrumentation__.Notify(633007)
		kv, err := txn.Get(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(633009)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(633010)
		}
		__antithesis_instrumentation__.Notify(633008)
		return kv.Value, nil
	}
	__antithesis_instrumentation__.Notify(633004)
	zoneID, zone, _, _, err := getZoneConfig(
		codec, id, getKey, false, false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(633011)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633012)
	}
	__antithesis_instrumentation__.Notify(633005)
	if err := completeZoneConfig(zone, codec, zoneID, getKey); err != nil {
		__antithesis_instrumentation__.Notify(633013)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633014)
	}
	__antithesis_instrumentation__.Notify(633006)

	return zone, nil
}

func zoneSpecifierNotFoundError(zs tree.ZoneSpecifier) error {
	__antithesis_instrumentation__.Notify(633015)
	if zs.NamedZone != "" {
		__antithesis_instrumentation__.Notify(633016)
		return pgerror.Newf(
			pgcode.InvalidCatalogName, "zone %q does not exist", zs.NamedZone)
	} else {
		__antithesis_instrumentation__.Notify(633017)
		if zs.Database != "" {
			__antithesis_instrumentation__.Notify(633018)
			return sqlerrors.NewUndefinedDatabaseError(string(zs.Database))
		} else {
			__antithesis_instrumentation__.Notify(633019)
			return sqlerrors.NewUndefinedRelationError(&zs.TableOrIndex)
		}
	}
}

func (p *planner) resolveTableForZone(
	ctx context.Context, zs *tree.ZoneSpecifier,
) (res catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(633020)
	if zs.TargetsIndex() {
		__antithesis_instrumentation__.Notify(633022)
		var mutRes *tabledesc.Mutable
		_, mutRes, err = expandMutableIndexName(ctx, p, &zs.TableOrIndex, true)
		if mutRes != nil {
			__antithesis_instrumentation__.Notify(633023)
			res = mutRes
		} else {
			__antithesis_instrumentation__.Notify(633024)
		}
	} else {
		__antithesis_instrumentation__.Notify(633025)
		if zs.TargetsTable() {
			__antithesis_instrumentation__.Notify(633026)
			var immutRes catalog.TableDescriptor
			p.runWithOptions(resolveFlags{skipCache: true}, func() {
				__antithesis_instrumentation__.Notify(633028)
				flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveAnyTableKind)
				flags.IncludeOffline = true
				_, immutRes, err = resolver.ResolveExistingTableObject(ctx, p, &zs.TableOrIndex.Table, flags)
			})
			__antithesis_instrumentation__.Notify(633027)
			if err != nil {
				__antithesis_instrumentation__.Notify(633029)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(633030)
				if immutRes != nil {
					__antithesis_instrumentation__.Notify(633031)
					res = immutRes
				} else {
					__antithesis_instrumentation__.Notify(633032)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(633033)
		}
	}
	__antithesis_instrumentation__.Notify(633021)
	return res, err
}

func resolveZone(
	ctx context.Context,
	txn *kv.Txn,
	col *descs.Collection,
	zs *tree.ZoneSpecifier,
	version clusterversion.Handle,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(633034)
	errMissingKey := errors.New("missing key")
	id, err := zonepb.ResolveZoneSpecifier(ctx, zs,
		func(parentID uint32, schemaID uint32, name string) (uint32, error) {
			__antithesis_instrumentation__.Notify(633037)
			id, err := col.Direct().LookupObjectID(ctx, txn, descpb.ID(parentID), descpb.ID(schemaID), name)
			if err != nil {
				__antithesis_instrumentation__.Notify(633040)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(633041)
			}
			__antithesis_instrumentation__.Notify(633038)
			if id == descpb.InvalidID {
				__antithesis_instrumentation__.Notify(633042)
				return 0, errMissingKey
			} else {
				__antithesis_instrumentation__.Notify(633043)
			}
			__antithesis_instrumentation__.Notify(633039)
			return uint32(id), nil
		},
		version,
	)
	__antithesis_instrumentation__.Notify(633035)
	if err != nil {
		__antithesis_instrumentation__.Notify(633044)
		if errors.Is(err, errMissingKey) {
			__antithesis_instrumentation__.Notify(633046)
			return 0, zoneSpecifierNotFoundError(*zs)
		} else {
			__antithesis_instrumentation__.Notify(633047)
		}
		__antithesis_instrumentation__.Notify(633045)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(633048)
	}
	__antithesis_instrumentation__.Notify(633036)
	return descpb.ID(id), nil
}

func resolveSubzone(
	zs *tree.ZoneSpecifier, table catalog.TableDescriptor,
) (catalog.Index, string, error) {
	__antithesis_instrumentation__.Notify(633049)
	if !zs.TargetsTable() || func() bool {
		__antithesis_instrumentation__.Notify(633053)
		return (zs.TableOrIndex.Index == "" && func() bool {
			__antithesis_instrumentation__.Notify(633054)
			return zs.Partition == "" == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(633055)
		return nil, "", nil
	} else {
		__antithesis_instrumentation__.Notify(633056)
	}
	__antithesis_instrumentation__.Notify(633050)

	indexName := string(zs.TableOrIndex.Index)
	var index catalog.Index
	if indexName == "" {
		__antithesis_instrumentation__.Notify(633057)
		index = table.GetPrimaryIndex()
		indexName = index.GetName()
	} else {
		__antithesis_instrumentation__.Notify(633058)
		var err error
		index, err = table.FindIndexWithName(indexName)
		if err != nil {
			__antithesis_instrumentation__.Notify(633059)
			return nil, "", err
		} else {
			__antithesis_instrumentation__.Notify(633060)
		}
	}
	__antithesis_instrumentation__.Notify(633051)

	partitionName := string(zs.Partition)
	if partitionName != "" {
		__antithesis_instrumentation__.Notify(633061)
		if index.GetPartitioning().FindPartitionByName(partitionName) == nil {
			__antithesis_instrumentation__.Notify(633062)
			return nil, "", fmt.Errorf("partition %q does not exist on index %q", partitionName, indexName)
		} else {
			__antithesis_instrumentation__.Notify(633063)
		}
	} else {
		__antithesis_instrumentation__.Notify(633064)
	}
	__antithesis_instrumentation__.Notify(633052)

	return index, partitionName, nil
}

func prepareRemovedPartitionZoneConfigs(
	ctx context.Context,
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	indexID descpb.IndexID,
	oldPart catalog.Partitioning,
	newPart catalog.Partitioning,
	execCfg *ExecutorConfig,
) (*zoneConfigUpdate, error) {
	__antithesis_instrumentation__.Notify(633065)
	newNames := map[string]struct{}{}
	_ = newPart.ForEachPartitionName(func(newName string) error {
		__antithesis_instrumentation__.Notify(633071)
		newNames[newName] = struct{}{}
		return nil
	})
	__antithesis_instrumentation__.Notify(633066)
	removedNames := make([]string, 0, len(newNames))
	_ = oldPart.ForEachPartitionName(func(oldName string) error {
		__antithesis_instrumentation__.Notify(633072)
		if _, exists := newNames[oldName]; !exists {
			__antithesis_instrumentation__.Notify(633074)
			removedNames = append(removedNames, oldName)
		} else {
			__antithesis_instrumentation__.Notify(633075)
		}
		__antithesis_instrumentation__.Notify(633073)
		return nil
	})
	__antithesis_instrumentation__.Notify(633067)
	if len(removedNames) == 0 {
		__antithesis_instrumentation__.Notify(633076)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(633077)
	}
	__antithesis_instrumentation__.Notify(633068)
	zone, err := getZoneConfigRaw(ctx, txn, execCfg.Codec, execCfg.Settings, tableDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(633078)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633079)
		if zone == nil {
			__antithesis_instrumentation__.Notify(633080)
			zone = zonepb.NewZoneConfig()
		} else {
			__antithesis_instrumentation__.Notify(633081)
		}
	}
	__antithesis_instrumentation__.Notify(633069)
	for _, n := range removedNames {
		__antithesis_instrumentation__.Notify(633082)
		zone.DeleteSubzone(uint32(indexID), n)
	}
	__antithesis_instrumentation__.Notify(633070)
	return prepareZoneConfigWrites(
		ctx, execCfg, tableDesc.GetID(), tableDesc, zone, false,
	)
}

func deleteRemovedPartitionZoneConfigs(
	ctx context.Context,
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	indexID descpb.IndexID,
	oldPart catalog.Partitioning,
	newPart catalog.Partitioning,
	execCfg *ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(633083)
	update, err := prepareRemovedPartitionZoneConfigs(
		ctx, txn, tableDesc, indexID, oldPart, newPart, execCfg,
	)
	if update == nil || func() bool {
		__antithesis_instrumentation__.Notify(633085)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(633086)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633087)
	}
	__antithesis_instrumentation__.Notify(633084)
	_, err = writeZoneConfigUpdate(ctx, txn, execCfg, update)
	return err
}
