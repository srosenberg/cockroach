package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	yaml "gopkg.in/yaml.v2"
)

var showZoneConfigColumns = colinfo.ResultColumns{
	{Name: "zone_id", Typ: types.Int, Hidden: true},
	{Name: "subzone_id", Typ: types.Int, Hidden: true},
	{Name: "target", Typ: types.String},
	{Name: "range_name", Typ: types.String, Hidden: true},
	{Name: "database_name", Typ: types.String, Hidden: true},
	{Name: "schema_name", Typ: types.String, Hidden: true},
	{Name: "table_name", Typ: types.String, Hidden: true},
	{Name: "index_name", Typ: types.String, Hidden: true},
	{Name: "partition_name", Typ: types.String, Hidden: true},
	{Name: "raw_config_yaml", Typ: types.String, Hidden: true},
	{Name: "raw_config_sql", Typ: types.String},
	{Name: "raw_config_protobuf", Typ: types.Bytes, Hidden: true},
	{Name: "full_config_yaml", Typ: types.String, Hidden: true},
	{Name: "full_config_sql", Typ: types.String, Hidden: true},
}

const (
	zoneIDCol int = iota
	subZoneIDCol
	targetCol
	rangeNameCol
	databaseNameCol
	schemaNameCol
	tableNameCol
	indexNameCol
	partitionNameCol
	rawConfigYAMLCol
	rawConfigSQLCol
	rawConfigProtobufCol
	fullConfigYamlCol
	fullConfigSQLCol
)

func (p *planner) ShowZoneConfig(ctx context.Context, n *tree.ShowZoneConfig) (planNode, error) {
	__antithesis_instrumentation__.Notify(623359)
	return &delayedNode{
		name:    n.String(),
		columns: showZoneConfigColumns,
		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(623360)
			v := p.newContainerValuesNode(showZoneConfigColumns, 0)

			if n.ZoneSpecifier == (tree.ZoneSpecifier{}) {
				__antithesis_instrumentation__.Notify(623364)
				return nil, errors.AssertionFailedf("zone must be specified")
			} else {
				__antithesis_instrumentation__.Notify(623365)
			}
			__antithesis_instrumentation__.Notify(623361)

			row, err := getShowZoneConfigRow(ctx, p, n.ZoneSpecifier)
			if err != nil {
				__antithesis_instrumentation__.Notify(623366)
				v.Close(ctx)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623367)
			}
			__antithesis_instrumentation__.Notify(623362)
			if _, err := v.rows.AddRow(ctx, row); err != nil {
				__antithesis_instrumentation__.Notify(623368)
				v.Close(ctx)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623369)
			}
			__antithesis_instrumentation__.Notify(623363)
			return v, nil
		},
	}, nil
}

func getShowZoneConfigRow(
	ctx context.Context, p *planner, zoneSpecifier tree.ZoneSpecifier,
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(623370)
	tblDesc, err := p.resolveTableForZone(ctx, &zoneSpecifier)
	if err != nil {
		__antithesis_instrumentation__.Notify(623377)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623378)
	}
	__antithesis_instrumentation__.Notify(623371)

	if zoneSpecifier.TableOrIndex.Table.ObjectName != "" {
		__antithesis_instrumentation__.Notify(623379)
		if err = p.CheckAnyPrivilege(ctx, tblDesc); err != nil {
			__antithesis_instrumentation__.Notify(623380)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623381)
		}
	} else {
		__antithesis_instrumentation__.Notify(623382)
		if zoneSpecifier.Database != "" {
			__antithesis_instrumentation__.Notify(623383)
			database, err := p.Descriptors().GetImmutableDatabaseByName(
				ctx,
				p.txn,
				string(zoneSpecifier.Database),
				tree.DatabaseLookupFlags{Required: true},
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(623385)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623386)
			}
			__antithesis_instrumentation__.Notify(623384)
			if err = p.CheckAnyPrivilege(ctx, database); err != nil {
				__antithesis_instrumentation__.Notify(623387)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623388)
			}
		} else {
			__antithesis_instrumentation__.Notify(623389)
		}
	}
	__antithesis_instrumentation__.Notify(623372)

	targetID, err := resolveZone(ctx, p.txn, p.Descriptors(), &zoneSpecifier, p.ExecCfg().Settings.Version)
	if err != nil {
		__antithesis_instrumentation__.Notify(623390)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623391)
	}
	__antithesis_instrumentation__.Notify(623373)

	index, partition, err := resolveSubzone(&zoneSpecifier, tblDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(623392)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623393)
	}
	__antithesis_instrumentation__.Notify(623374)

	subZoneIdx := uint32(0)
	zoneID, zone, subzone, err := GetZoneConfigInTxn(
		ctx, p.txn, p.ExecCfg().Codec, targetID, index, partition, false,
	)
	if errors.Is(err, errNoZoneConfigApplies) {
		__antithesis_instrumentation__.Notify(623394)

		zone = p.execCfg.DefaultZoneConfig
		zoneID = keys.RootNamespaceID
	} else {
		__antithesis_instrumentation__.Notify(623395)
		if err != nil {
			__antithesis_instrumentation__.Notify(623396)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623397)
			if subzone != nil {
				__antithesis_instrumentation__.Notify(623398)
				for i := range zone.Subzones {
					__antithesis_instrumentation__.Notify(623400)
					subZoneIdx++
					if subzone == &zone.Subzones[i] {
						__antithesis_instrumentation__.Notify(623401)
						break
					} else {
						__antithesis_instrumentation__.Notify(623402)
					}
				}
				__antithesis_instrumentation__.Notify(623399)
				zone = &subzone.Config
			} else {
				__antithesis_instrumentation__.Notify(623403)
			}
		}
	}
	__antithesis_instrumentation__.Notify(623375)

	zs := ascendZoneSpecifier(zoneSpecifier, targetID, zoneID, subzone)

	zone.Subzones = nil
	zone.SubzoneSpans = nil

	vals := make(tree.Datums, len(showZoneConfigColumns))
	if err := generateZoneConfigIntrospectionValues(
		vals, tree.NewDInt(tree.DInt(zoneID)), tree.NewDInt(tree.DInt(subZoneIdx)), &zs, zone, nil,
	); err != nil {
		__antithesis_instrumentation__.Notify(623404)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623405)
	}
	__antithesis_instrumentation__.Notify(623376)
	return vals, nil
}

func zoneConfigToSQL(zs *tree.ZoneSpecifier, zone *zonepb.ZoneConfig) (string, error) {
	__antithesis_instrumentation__.Notify(623406)

	yaml.FutureLineWrap()
	constraints, err := yamlMarshalFlow(zonepb.ConstraintsList{
		Constraints: zone.Constraints,
		Inherited:   zone.InheritedConstraints})
	if err != nil {
		__antithesis_instrumentation__.Notify(623420)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(623421)
	}
	__antithesis_instrumentation__.Notify(623407)
	constraints = strings.TrimSpace(constraints)
	voterConstraints, err := yamlMarshalFlow(zonepb.ConstraintsList{
		Constraints: zone.VoterConstraints,
		Inherited:   zone.InheritedVoterConstraints(),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(623422)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(623423)
	}
	__antithesis_instrumentation__.Notify(623408)
	voterConstraints = strings.TrimSpace(voterConstraints)
	prefs, err := yamlMarshalFlow(zone.LeasePreferences)
	if err != nil {
		__antithesis_instrumentation__.Notify(623424)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(623425)
	}
	__antithesis_instrumentation__.Notify(623409)
	prefs = strings.TrimSpace(prefs)

	useComma := false
	maybeWriteComma := func(f *tree.FmtCtx) {
		__antithesis_instrumentation__.Notify(623426)
		if useComma {
			__antithesis_instrumentation__.Notify(623428)
			f.Printf(",\n")
		} else {
			__antithesis_instrumentation__.Notify(623429)
		}
		__antithesis_instrumentation__.Notify(623427)
		useComma = true
	}
	__antithesis_instrumentation__.Notify(623410)

	f := tree.NewFmtCtx(tree.FmtParsable)
	f.WriteString("ALTER ")
	f.FormatNode(zs)
	f.WriteString(" CONFIGURE ZONE USING\n")
	if zone.RangeMinBytes != nil {
		__antithesis_instrumentation__.Notify(623430)
		maybeWriteComma(f)
		f.Printf("\trange_min_bytes = %d", *zone.RangeMinBytes)
	} else {
		__antithesis_instrumentation__.Notify(623431)
	}
	__antithesis_instrumentation__.Notify(623411)
	if zone.RangeMaxBytes != nil {
		__antithesis_instrumentation__.Notify(623432)
		maybeWriteComma(f)
		f.Printf("\trange_max_bytes = %d", *zone.RangeMaxBytes)
	} else {
		__antithesis_instrumentation__.Notify(623433)
	}
	__antithesis_instrumentation__.Notify(623412)
	if zone.GC != nil {
		__antithesis_instrumentation__.Notify(623434)
		maybeWriteComma(f)
		f.Printf("\tgc.ttlseconds = %d", zone.GC.TTLSeconds)
	} else {
		__antithesis_instrumentation__.Notify(623435)
	}
	__antithesis_instrumentation__.Notify(623413)
	if zone.GlobalReads != nil {
		__antithesis_instrumentation__.Notify(623436)
		maybeWriteComma(f)
		f.Printf("\tglobal_reads = %t", *zone.GlobalReads)
	} else {
		__antithesis_instrumentation__.Notify(623437)
	}
	__antithesis_instrumentation__.Notify(623414)
	if zone.NumReplicas != nil {
		__antithesis_instrumentation__.Notify(623438)
		maybeWriteComma(f)
		f.Printf("\tnum_replicas = %d", *zone.NumReplicas)
	} else {
		__antithesis_instrumentation__.Notify(623439)
	}
	__antithesis_instrumentation__.Notify(623415)
	if zone.NumVoters != nil {
		__antithesis_instrumentation__.Notify(623440)
		maybeWriteComma(f)
		f.Printf("\tnum_voters = %d", *zone.NumVoters)
	} else {
		__antithesis_instrumentation__.Notify(623441)
	}
	__antithesis_instrumentation__.Notify(623416)
	if !zone.InheritedConstraints {
		__antithesis_instrumentation__.Notify(623442)
		maybeWriteComma(f)
		f.Printf("\tconstraints = %s", lexbase.EscapeSQLString(constraints))
	} else {
		__antithesis_instrumentation__.Notify(623443)
	}
	__antithesis_instrumentation__.Notify(623417)
	if !zone.InheritedVoterConstraints() && func() bool {
		__antithesis_instrumentation__.Notify(623444)
		return zone.NumVoters != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(623445)
		return *zone.NumVoters > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(623446)
		maybeWriteComma(f)
		f.Printf("\tvoter_constraints = %s", lexbase.EscapeSQLString(voterConstraints))
	} else {
		__antithesis_instrumentation__.Notify(623447)
	}
	__antithesis_instrumentation__.Notify(623418)
	if !zone.InheritedLeasePreferences {
		__antithesis_instrumentation__.Notify(623448)
		maybeWriteComma(f)
		f.Printf("\tlease_preferences = %s", lexbase.EscapeSQLString(prefs))
	} else {
		__antithesis_instrumentation__.Notify(623449)
	}
	__antithesis_instrumentation__.Notify(623419)
	return f.String(), nil
}

func generateZoneConfigIntrospectionValues(
	values tree.Datums,
	zoneID tree.Datum,
	subZoneID tree.Datum,
	zs *tree.ZoneSpecifier,
	zone *zonepb.ZoneConfig,
	fullZoneConfig *zonepb.ZoneConfig,
) error {
	__antithesis_instrumentation__.Notify(623450)

	values[zoneIDCol] = zoneID
	values[subZoneIDCol] = subZoneID

	values[targetCol] = tree.DNull
	values[rangeNameCol] = tree.DNull
	values[databaseNameCol] = tree.DNull
	values[schemaNameCol] = tree.DNull
	values[tableNameCol] = tree.DNull
	values[indexNameCol] = tree.DNull
	values[partitionNameCol] = tree.DNull
	if zs != nil {
		__antithesis_instrumentation__.Notify(623458)
		values[targetCol] = tree.NewDString(zs.String())
		if zs.NamedZone != "" {
			__antithesis_instrumentation__.Notify(623463)
			values[rangeNameCol] = tree.NewDString(string(zs.NamedZone))
		} else {
			__antithesis_instrumentation__.Notify(623464)
		}
		__antithesis_instrumentation__.Notify(623459)
		if zs.Database != "" {
			__antithesis_instrumentation__.Notify(623465)
			values[databaseNameCol] = tree.NewDString(string(zs.Database))
		} else {
			__antithesis_instrumentation__.Notify(623466)
		}
		__antithesis_instrumentation__.Notify(623460)
		if zs.TableOrIndex.Table.ObjectName != "" {
			__antithesis_instrumentation__.Notify(623467)
			values[databaseNameCol] = tree.NewDString(string(zs.TableOrIndex.Table.CatalogName))
			values[schemaNameCol] = tree.NewDString(string(zs.TableOrIndex.Table.SchemaName))
			values[tableNameCol] = tree.NewDString(string(zs.TableOrIndex.Table.ObjectName))
		} else {
			__antithesis_instrumentation__.Notify(623468)
		}
		__antithesis_instrumentation__.Notify(623461)
		if zs.TableOrIndex.Index != "" {
			__antithesis_instrumentation__.Notify(623469)
			values[indexNameCol] = tree.NewDString(string(zs.TableOrIndex.Index))
		} else {
			__antithesis_instrumentation__.Notify(623470)
		}
		__antithesis_instrumentation__.Notify(623462)
		if zs.Partition != "" {
			__antithesis_instrumentation__.Notify(623471)
			values[partitionNameCol] = tree.NewDString(string(zs.Partition))
		} else {
			__antithesis_instrumentation__.Notify(623472)
		}
	} else {
		__antithesis_instrumentation__.Notify(623473)
	}
	__antithesis_instrumentation__.Notify(623451)

	yamlConfig, err := yaml.Marshal(zone)
	if err != nil {
		__antithesis_instrumentation__.Notify(623474)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623475)
	}
	__antithesis_instrumentation__.Notify(623452)
	values[rawConfigYAMLCol] = tree.NewDString(string(yamlConfig))

	if zs == nil {
		__antithesis_instrumentation__.Notify(623476)
		values[rawConfigSQLCol] = tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(623477)
		sqlStr, err := zoneConfigToSQL(zs, zone)
		if err != nil {
			__antithesis_instrumentation__.Notify(623479)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623480)
		}
		__antithesis_instrumentation__.Notify(623478)
		values[rawConfigSQLCol] = tree.NewDString(sqlStr)
	}
	__antithesis_instrumentation__.Notify(623453)

	protoConfig, err := protoutil.Marshal(zone)
	if err != nil {
		__antithesis_instrumentation__.Notify(623481)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623482)
	}
	__antithesis_instrumentation__.Notify(623454)
	values[rawConfigProtobufCol] = tree.NewDBytes(tree.DBytes(protoConfig))

	inheritedConfig := fullZoneConfig
	if inheritedConfig == nil {
		__antithesis_instrumentation__.Notify(623483)
		inheritedConfig = zone
	} else {
		__antithesis_instrumentation__.Notify(623484)
	}
	__antithesis_instrumentation__.Notify(623455)

	yamlConfig, err = yaml.Marshal(inheritedConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(623485)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623486)
	}
	__antithesis_instrumentation__.Notify(623456)
	values[fullConfigYamlCol] = tree.NewDString(string(yamlConfig))

	if zs == nil {
		__antithesis_instrumentation__.Notify(623487)
		values[fullConfigSQLCol] = tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(623488)
		sqlStr, err := zoneConfigToSQL(zs, inheritedConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(623490)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623491)
		}
		__antithesis_instrumentation__.Notify(623489)
		values[fullConfigSQLCol] = tree.NewDString(sqlStr)
	}
	__antithesis_instrumentation__.Notify(623457)
	return nil
}

func yamlMarshalFlow(v interface{}) (string, error) {
	__antithesis_instrumentation__.Notify(623492)
	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	e.UseStyle(yaml.FlowStyle)
	if err := e.Encode(v); err != nil {
		__antithesis_instrumentation__.Notify(623495)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(623496)
	}
	__antithesis_instrumentation__.Notify(623493)
	if err := e.Close(); err != nil {
		__antithesis_instrumentation__.Notify(623497)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(623498)
	}
	__antithesis_instrumentation__.Notify(623494)
	return buf.String(), nil
}

func ascendZoneSpecifier(
	zs tree.ZoneSpecifier, resolvedID, actualID descpb.ID, actualSubzone *zonepb.Subzone,
) tree.ZoneSpecifier {
	__antithesis_instrumentation__.Notify(623499)
	if actualID == keys.RootNamespaceID {
		__antithesis_instrumentation__.Notify(623501)

		zs.NamedZone = tree.UnrestrictedName(zonepb.DefaultZoneName)
		zs.Database = ""
		zs.TableOrIndex = tree.TableIndexName{}

		zs.Partition = ""
	} else {
		__antithesis_instrumentation__.Notify(623502)
		if resolvedID != actualID {
			__antithesis_instrumentation__.Notify(623503)

			zs.Database = zs.TableOrIndex.Table.CatalogName
			zs.TableOrIndex = tree.TableIndexName{}

			zs.Partition = ""
		} else {
			__antithesis_instrumentation__.Notify(623504)
			if actualSubzone == nil {
				__antithesis_instrumentation__.Notify(623505)

				zs.TableOrIndex.Index = ""
				zs.Partition = ""
			} else {
				__antithesis_instrumentation__.Notify(623506)
				if actualSubzone.PartitionName == "" {
					__antithesis_instrumentation__.Notify(623507)

					zs.Partition = ""
				} else {
					__antithesis_instrumentation__.Notify(623508)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(623500)
	return zs
}
