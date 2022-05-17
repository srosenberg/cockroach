package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/joberror"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/ingesting"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/rewrite"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/gcjob"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type importResumer struct {
	job      *jobs.Job
	settings *cluster.Settings
	res      roachpb.RowCount

	testingKnobs struct {
		afterImport            func(summary roachpb.RowCount) error
		alwaysFlushJobProgress bool
	}
}

func (r *importResumer) TestingSetAfterImportKnob(fn func(summary roachpb.RowCount) error) {
	__antithesis_instrumentation__.Notify(493448)
	r.testingKnobs.afterImport = fn
}

var _ jobs.TraceableJob = &importResumer{}

func (r *importResumer) ForceRealSpan() bool {
	__antithesis_instrumentation__.Notify(493449)
	return true
}

var _ jobs.Resumer = &importResumer{}

var processorsPerNode = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.import.processors_per_node",
	"number of input processors to run on each sql instance", 1,
	settings.PositiveInt,
)

type preparedSchemaMetadata struct {
	schemaPreparedDetails jobspb.ImportDetails
	schemaRewrites        jobspb.DescRewriteMap
	newSchemaIDToName     map[descpb.ID]string
	oldSchemaIDToName     map[descpb.ID]string
	queuedSchemaJobs      []jobspb.JobID
}

func (r *importResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(493450)
	p := execCtx.(sql.JobExecContext)
	if err := r.parseBundleSchemaIfNeeded(ctx, p); err != nil {
		__antithesis_instrumentation__.Notify(493467)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493468)
	}
	__antithesis_instrumentation__.Notify(493451)

	details := r.job.Details().(jobspb.ImportDetails)
	files := details.URIs
	format := details.Format

	tables := make(map[string]*execinfrapb.ReadImportDataSpec_ImportTable, len(details.Tables))
	if details.Tables != nil {
		__antithesis_instrumentation__.Notify(493469)

		if !details.PrepareComplete {
			__antithesis_instrumentation__.Notify(493472)
			var schemaMetadata *preparedSchemaMetadata
			if err := sql.DescsTxn(ctx, p.ExecCfg(), func(
				ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
			) error {
				__antithesis_instrumentation__.Notify(493475)
				var preparedDetails jobspb.ImportDetails
				schemaMetadata = &preparedSchemaMetadata{
					newSchemaIDToName: make(map[descpb.ID]string),
					oldSchemaIDToName: make(map[descpb.ID]string),
				}
				var err error
				curDetails := details
				if len(details.Schemas) != 0 {
					__antithesis_instrumentation__.Notify(493480)
					schemaMetadata, err = r.prepareSchemasForIngestion(ctx, p, curDetails, txn, descsCol)
					if err != nil {
						__antithesis_instrumentation__.Notify(493482)
						return err
					} else {
						__antithesis_instrumentation__.Notify(493483)
					}
					__antithesis_instrumentation__.Notify(493481)
					curDetails = schemaMetadata.schemaPreparedDetails
				} else {
					__antithesis_instrumentation__.Notify(493484)
				}
				__antithesis_instrumentation__.Notify(493476)

				if r.settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
					__antithesis_instrumentation__.Notify(493485)

					_, dbDesc, err := descsCol.GetImmutableDatabaseByID(ctx, txn, details.ParentID, tree.DatabaseLookupFlags{Required: true})
					if err != nil {
						__antithesis_instrumentation__.Notify(493487)
						return err
					} else {
						__antithesis_instrumentation__.Notify(493488)
					}
					__antithesis_instrumentation__.Notify(493486)
					schemaMetadata.oldSchemaIDToName[dbDesc.GetSchemaID(tree.PublicSchema)] = tree.PublicSchema
					schemaMetadata.newSchemaIDToName[dbDesc.GetSchemaID(tree.PublicSchema)] = tree.PublicSchema
				} else {
					__antithesis_instrumentation__.Notify(493489)
				}
				__antithesis_instrumentation__.Notify(493477)

				preparedDetails, err = r.prepareTablesForIngestion(ctx, p, curDetails, txn, descsCol,
					schemaMetadata)
				if err != nil {
					__antithesis_instrumentation__.Notify(493490)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493491)
				}
				__antithesis_instrumentation__.Notify(493478)

				for _, table := range preparedDetails.Tables {
					__antithesis_instrumentation__.Notify(493492)
					_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
						ctx, txn, table.Desc.GetParentID(), tree.DatabaseLookupFlags{Required: true})
					if err != nil {
						__antithesis_instrumentation__.Notify(493494)
						return err
					} else {
						__antithesis_instrumentation__.Notify(493495)
					}
					__antithesis_instrumentation__.Notify(493493)
					if dbDesc.IsMultiRegion() {
						__antithesis_instrumentation__.Notify(493496)
						telemetry.Inc(sqltelemetry.ImportIntoMultiRegionDatabaseCounter)
					} else {
						__antithesis_instrumentation__.Notify(493497)
					}
				}
				__antithesis_instrumentation__.Notify(493479)

				return r.job.Update(ctx, txn, func(
					txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater,
				) error {
					__antithesis_instrumentation__.Notify(493498)
					pl := md.Payload
					*pl.GetImport() = preparedDetails

					prev := md.Payload.DescriptorIDs
					if prev == nil {
						__antithesis_instrumentation__.Notify(493500)
						var descriptorIDs []descpb.ID
						for _, schema := range preparedDetails.Schemas {
							__antithesis_instrumentation__.Notify(493503)
							descriptorIDs = append(descriptorIDs, schema.Desc.GetID())
						}
						__antithesis_instrumentation__.Notify(493501)
						for _, table := range preparedDetails.Tables {
							__antithesis_instrumentation__.Notify(493504)
							descriptorIDs = append(descriptorIDs, table.Desc.GetID())
						}
						__antithesis_instrumentation__.Notify(493502)
						pl.DescriptorIDs = descriptorIDs
					} else {
						__antithesis_instrumentation__.Notify(493505)
					}
					__antithesis_instrumentation__.Notify(493499)
					ju.UpdatePayload(pl)
					return nil
				})
			}); err != nil {
				__antithesis_instrumentation__.Notify(493506)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493507)
			}
			__antithesis_instrumentation__.Notify(493473)

			if err := p.ExecCfg().JobRegistry.Run(ctx, p.ExecCfg().InternalExecutor,
				schemaMetadata.queuedSchemaJobs); err != nil {
				__antithesis_instrumentation__.Notify(493508)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493509)
			}
			__antithesis_instrumentation__.Notify(493474)

			details = r.job.Details().(jobspb.ImportDetails)
			emitImportJobEvent(ctx, p, jobs.StatusRunning, r.job)
		} else {
			__antithesis_instrumentation__.Notify(493510)
		}
		__antithesis_instrumentation__.Notify(493470)

		schemaIDToName := make(map[descpb.ID]string)
		for _, i := range details.Schemas {
			__antithesis_instrumentation__.Notify(493511)
			schemaIDToName[i.Desc.GetID()] = i.Desc.GetName()
		}
		__antithesis_instrumentation__.Notify(493471)

		for _, i := range details.Tables {
			__antithesis_instrumentation__.Notify(493512)
			var tableName string
			if i.Name != "" {
				__antithesis_instrumentation__.Notify(493515)
				tableName = i.Name
			} else {
				__antithesis_instrumentation__.Notify(493516)
				if i.Desc != nil {
					__antithesis_instrumentation__.Notify(493517)
					tableName = i.Desc.Name
				} else {
					__antithesis_instrumentation__.Notify(493518)
					return errors.New("invalid table specification")
				}
			}
			__antithesis_instrumentation__.Notify(493513)

			if details.Format.Format == roachpb.IOFileFormat_PgDump {
				__antithesis_instrumentation__.Notify(493519)
				schemaName := tree.PublicSchema
				if schema, ok := schemaIDToName[i.Desc.GetUnexposedParentSchemaID()]; ok {
					__antithesis_instrumentation__.Notify(493521)
					schemaName = schema
				} else {
					__antithesis_instrumentation__.Notify(493522)
				}
				__antithesis_instrumentation__.Notify(493520)
				tableName = fmt.Sprintf("%s.%s", schemaName, tableName)
			} else {
				__antithesis_instrumentation__.Notify(493523)
			}
			__antithesis_instrumentation__.Notify(493514)
			tables[tableName] = &execinfrapb.ReadImportDataSpec_ImportTable{
				Desc:       i.Desc,
				TargetCols: i.TargetCols,
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(493524)
	}
	__antithesis_instrumentation__.Notify(493452)

	typeDescs := make([]*descpb.TypeDescriptor, len(details.Types))
	for i, t := range details.Types {
		__antithesis_instrumentation__.Notify(493525)
		typeDescs[i] = t.Desc
	}
	__antithesis_instrumentation__.Notify(493453)

	if details.Walltime == 0 {
		__antithesis_instrumentation__.Notify(493526)

		details.Walltime = p.ExecCfg().Clock.Now().WallTime

		for i := range details.Tables {
			__antithesis_instrumentation__.Notify(493528)
			if !details.Tables[i].IsNew {
				__antithesis_instrumentation__.Notify(493529)
				tblDesc := tabledesc.NewBuilder(details.Tables[i].Desc).BuildImmutableTable()
				tblSpan := tblDesc.TableSpan(p.ExecCfg().Codec)
				res, err := p.ExecCfg().DB.Scan(ctx, tblSpan.Key, tblSpan.EndKey, 1)
				if err != nil {
					__antithesis_instrumentation__.Notify(493531)
					return errors.Wrap(err, "checking if existing table is empty")
				} else {
					__antithesis_instrumentation__.Notify(493532)
				}
				__antithesis_instrumentation__.Notify(493530)
				details.Tables[i].WasEmpty = len(res) == 0
			} else {
				__antithesis_instrumentation__.Notify(493533)
			}
		}
		__antithesis_instrumentation__.Notify(493527)

		if err := r.job.SetDetails(ctx, nil, details); err != nil {
			__antithesis_instrumentation__.Notify(493534)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493535)
		}
	} else {
		__antithesis_instrumentation__.Notify(493536)
	}
	__antithesis_instrumentation__.Notify(493454)

	procsPerNode := int(processorsPerNode.Get(&p.ExecCfg().Settings.SV))

	res, err := ingestWithRetry(ctx, p, r.job, tables, typeDescs, files, format, details.Walltime,
		r.testingKnobs.alwaysFlushJobProgress, procsPerNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(493537)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493538)
	}
	__antithesis_instrumentation__.Notify(493455)

	pkIDs := make(map[uint64]struct{}, len(details.Tables))
	for _, t := range details.Tables {
		__antithesis_instrumentation__.Notify(493539)
		pkIDs[roachpb.BulkOpSummaryID(uint64(t.Desc.ID), uint64(t.Desc.PrimaryIndex.ID))] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(493456)
	r.res.DataSize = res.DataSize
	for id, count := range res.EntryCounts {
		__antithesis_instrumentation__.Notify(493540)
		if _, ok := pkIDs[id]; ok {
			__antithesis_instrumentation__.Notify(493541)
			r.res.Rows += count
		} else {
			__antithesis_instrumentation__.Notify(493542)
			r.res.IndexEntries += count
		}
	}
	__antithesis_instrumentation__.Notify(493457)

	if r.testingKnobs.afterImport != nil {
		__antithesis_instrumentation__.Notify(493543)
		if err := r.testingKnobs.afterImport(r.res); err != nil {
			__antithesis_instrumentation__.Notify(493544)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493545)
		}
	} else {
		__antithesis_instrumentation__.Notify(493546)
	}
	__antithesis_instrumentation__.Notify(493458)
	if err := p.ExecCfg().JobRegistry.CheckPausepoint("import.after_ingest"); err != nil {
		__antithesis_instrumentation__.Notify(493547)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493548)
	}
	__antithesis_instrumentation__.Notify(493459)

	if err := r.checkVirtualConstraints(ctx, p.ExecCfg(), r.job); err != nil {
		__antithesis_instrumentation__.Notify(493549)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493550)
	}
	__antithesis_instrumentation__.Notify(493460)

	if err := r.checkForUDTModification(ctx, p.ExecCfg()); err != nil {
		__antithesis_instrumentation__.Notify(493551)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493552)
	}
	__antithesis_instrumentation__.Notify(493461)

	if err := r.publishSchemas(ctx, p.ExecCfg()); err != nil {
		__antithesis_instrumentation__.Notify(493553)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493554)
	}
	__antithesis_instrumentation__.Notify(493462)

	if err := r.publishTables(ctx, p.ExecCfg(), res); err != nil {
		__antithesis_instrumentation__.Notify(493555)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493556)
	}
	__antithesis_instrumentation__.Notify(493463)

	if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(493557)
		return r.releaseProtectedTimestamp(ctx, txn, p.ExecCfg().ProtectedTimestampProvider)
	}); err != nil {
		__antithesis_instrumentation__.Notify(493558)
		log.Errorf(ctx, "failed to release protected timestamp: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(493559)
	}
	__antithesis_instrumentation__.Notify(493464)

	emitImportJobEvent(ctx, p, jobs.StatusSucceeded, r.job)

	addToFileFormatTelemetry(details.Format.Format.String(), "succeeded")
	telemetry.CountBucketed("import.rows", r.res.Rows)
	const mb = 1 << 20
	sizeMb := r.res.DataSize / mb
	telemetry.CountBucketed("import.size-mb", sizeMb)

	sec := int64(timeutil.Since(timeutil.FromUnixMicros(r.job.Payload().StartedMicros)).Seconds())
	var mbps int64
	if sec > 0 {
		__antithesis_instrumentation__.Notify(493560)
		mbps = mb / sec
	} else {
		__antithesis_instrumentation__.Notify(493561)
	}
	__antithesis_instrumentation__.Notify(493465)
	telemetry.CountBucketed("import.duration-sec.succeeded", sec)
	telemetry.CountBucketed("import.speed-mbps", mbps)

	if sizeMb > 10 {
		__antithesis_instrumentation__.Notify(493562)
		telemetry.CountBucketed("import.speed-mbps.over10mb", mbps)
	} else {
		__antithesis_instrumentation__.Notify(493563)
	}
	__antithesis_instrumentation__.Notify(493466)

	return nil
}

func (r *importResumer) prepareTablesForIngestion(
	ctx context.Context,
	p sql.JobExecContext,
	details jobspb.ImportDetails,
	txn *kv.Txn,
	descsCol *descs.Collection,
	schemaMetadata *preparedSchemaMetadata,
) (jobspb.ImportDetails, error) {
	__antithesis_instrumentation__.Notify(493564)
	importDetails := details
	importDetails.Tables = make([]jobspb.ImportDetails_Table, len(details.Tables))

	newSchemaAndTableNameToIdx := make(map[string]int, len(importDetails.Tables))
	var hasExistingTables bool
	var err error
	var newTableDescs []jobspb.ImportDetails_Table
	var desc *descpb.TableDescriptor
	for i, table := range details.Tables {
		__antithesis_instrumentation__.Notify(493568)
		if !table.IsNew {
			__antithesis_instrumentation__.Notify(493569)
			desc, err = prepareExistingTablesForIngestion(ctx, txn, descsCol, table.Desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(493571)
				return importDetails, err
			} else {
				__antithesis_instrumentation__.Notify(493572)
			}
			__antithesis_instrumentation__.Notify(493570)
			importDetails.Tables[i] = jobspb.ImportDetails_Table{
				Desc: desc, Name: table.Name,
				SeqVal:     table.SeqVal,
				IsNew:      table.IsNew,
				TargetCols: table.TargetCols,
			}

			hasExistingTables = true
		} else {
			__antithesis_instrumentation__.Notify(493573)

			key, err := constructSchemaAndTableKey(ctx, table.Desc, schemaMetadata.oldSchemaIDToName, p.ExecCfg().Settings.Version)
			if err != nil {
				__antithesis_instrumentation__.Notify(493575)
				return importDetails, err
			} else {
				__antithesis_instrumentation__.Notify(493576)
			}
			__antithesis_instrumentation__.Notify(493574)
			newSchemaAndTableNameToIdx[key.String()] = i

			newTableDescs = append(newTableDescs,
				*protoutil.Clone(&table).(*jobspb.ImportDetails_Table))
		}
	}
	__antithesis_instrumentation__.Notify(493565)

	if len(newTableDescs) != 0 {
		__antithesis_instrumentation__.Notify(493577)
		res, err := prepareNewTablesForIngestion(
			ctx, txn, descsCol, p, newTableDescs, importDetails.ParentID, schemaMetadata.schemaRewrites)
		if err != nil {
			__antithesis_instrumentation__.Notify(493579)
			return importDetails, err
		} else {
			__antithesis_instrumentation__.Notify(493580)
		}
		__antithesis_instrumentation__.Notify(493578)

		for _, desc := range res {
			__antithesis_instrumentation__.Notify(493581)
			key, err := constructSchemaAndTableKey(ctx, desc, schemaMetadata.newSchemaIDToName, p.ExecCfg().Settings.Version)
			if err != nil {
				__antithesis_instrumentation__.Notify(493583)
				return importDetails, err
			} else {
				__antithesis_instrumentation__.Notify(493584)
			}
			__antithesis_instrumentation__.Notify(493582)
			i := newSchemaAndTableNameToIdx[key.String()]
			table := details.Tables[i]
			importDetails.Tables[i] = jobspb.ImportDetails_Table{
				Desc:       desc,
				Name:       table.Name,
				SeqVal:     table.SeqVal,
				IsNew:      table.IsNew,
				TargetCols: table.TargetCols,
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(493585)
	}
	__antithesis_instrumentation__.Notify(493566)

	importDetails.PrepareComplete = true

	if !hasExistingTables {
		__antithesis_instrumentation__.Notify(493586)
		importDetails.Walltime = p.ExecCfg().Clock.Now().WallTime
	} else {
		__antithesis_instrumentation__.Notify(493587)
		importDetails.Walltime = 0
	}
	__antithesis_instrumentation__.Notify(493567)

	return importDetails, nil
}

func prepareExistingTablesForIngestion(
	ctx context.Context, txn *kv.Txn, descsCol *descs.Collection, desc *descpb.TableDescriptor,
) (*descpb.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(493588)
	if len(desc.Mutations) > 0 {
		__antithesis_instrumentation__.Notify(493593)
		return nil, errors.Errorf("cannot IMPORT INTO a table with schema changes in progress -- try again later (pending mutation %s)", desc.Mutations[0].String())
	} else {
		__antithesis_instrumentation__.Notify(493594)
	}
	__antithesis_instrumentation__.Notify(493589)

	importing, err := descsCol.GetMutableTableVersionByID(ctx, desc.ID, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(493595)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493596)
	}
	__antithesis_instrumentation__.Notify(493590)

	if got, exp := importing.Version, desc.Version; got != exp {
		__antithesis_instrumentation__.Notify(493597)
		return nil, errors.Errorf("another operation is currently operating on the table")
	} else {
		__antithesis_instrumentation__.Notify(493598)
	}
	__antithesis_instrumentation__.Notify(493591)

	importing.SetOffline("importing")

	if err := descsCol.WriteDesc(
		ctx, false, importing, txn,
	); err != nil {
		__antithesis_instrumentation__.Notify(493599)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493600)
	}
	__antithesis_instrumentation__.Notify(493592)

	return importing.TableDesc(), nil
}

func prepareNewTablesForIngestion(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	p sql.JobExecContext,
	importTables []jobspb.ImportDetails_Table,
	parentID descpb.ID,
	schemaRewrites jobspb.DescRewriteMap,
) ([]*descpb.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(493601)
	newMutableTableDescriptors := make([]*tabledesc.Mutable, len(importTables))
	for i := range importTables {
		__antithesis_instrumentation__.Notify(493611)
		newMutableTableDescriptors[i] = tabledesc.NewBuilder(importTables[i].Desc).BuildCreatedMutableTable()
	}
	__antithesis_instrumentation__.Notify(493602)

	tableRewrites := schemaRewrites
	if tableRewrites == nil {
		__antithesis_instrumentation__.Notify(493612)
		tableRewrites = make(jobspb.DescRewriteMap)
	} else {
		__antithesis_instrumentation__.Notify(493613)
	}
	__antithesis_instrumentation__.Notify(493603)
	seqVals := make(map[descpb.ID]int64, len(importTables))
	for _, tableDesc := range importTables {
		__antithesis_instrumentation__.Notify(493614)
		id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(493617)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493618)
		}
		__antithesis_instrumentation__.Notify(493615)
		oldParentSchemaID := tableDesc.Desc.GetUnexposedParentSchemaID()
		parentSchemaID := oldParentSchemaID
		if rw, ok := schemaRewrites[oldParentSchemaID]; ok {
			__antithesis_instrumentation__.Notify(493619)
			parentSchemaID = rw.ID
		} else {
			__antithesis_instrumentation__.Notify(493620)
		}
		__antithesis_instrumentation__.Notify(493616)
		tableRewrites[tableDesc.Desc.ID] = &jobspb.DescriptorRewrite{
			ID:             id,
			ParentSchemaID: parentSchemaID,
			ParentID:       parentID,
		}
		seqVals[id] = tableDesc.SeqVal
	}
	__antithesis_instrumentation__.Notify(493604)
	if err := rewrite.TableDescs(
		newMutableTableDescriptors, tableRewrites, "",
	); err != nil {
		__antithesis_instrumentation__.Notify(493621)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493622)
	}
	__antithesis_instrumentation__.Notify(493605)

	for i := range newMutableTableDescriptors {
		__antithesis_instrumentation__.Notify(493623)
		tbl := newMutableTableDescriptors[i]
		err := descsCol.Direct().CheckObjectCollision(
			ctx,
			txn,
			tbl.GetParentID(),
			tbl.GetParentSchemaID(),
			tree.NewUnqualifiedTableName(tree.Name(tbl.GetName())),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(493624)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493625)
		}
	}
	__antithesis_instrumentation__.Notify(493606)

	tableDescs := make([]catalog.TableDescriptor, len(newMutableTableDescriptors))
	for i := range tableDescs {
		__antithesis_instrumentation__.Notify(493626)
		newMutableTableDescriptors[i].SetOffline("importing")
		tableDescs[i] = newMutableTableDescriptors[i]
	}
	__antithesis_instrumentation__.Notify(493607)

	var seqValKVs []roachpb.KeyValue
	for _, desc := range newMutableTableDescriptors {
		__antithesis_instrumentation__.Notify(493627)
		if v, ok := seqVals[desc.GetID()]; ok && func() bool {
			__antithesis_instrumentation__.Notify(493628)
			return v != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(493629)
			key, val, err := sql.MakeSequenceKeyVal(p.ExecCfg().Codec, desc, v, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(493631)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493632)
			}
			__antithesis_instrumentation__.Notify(493630)
			kv := roachpb.KeyValue{Key: key}
			kv.Value.SetInt(val)
			seqValKVs = append(seqValKVs, kv)
		} else {
			__antithesis_instrumentation__.Notify(493633)
		}
	}
	__antithesis_instrumentation__.Notify(493608)

	if err := ingesting.WriteDescriptors(ctx, p.ExecCfg().Codec, txn, p.User(), descsCol,
		nil, nil,
		tableDescs, nil, tree.RequestedDescriptors, seqValKVs, ""); err != nil {
		__antithesis_instrumentation__.Notify(493634)
		return nil, errors.Wrapf(err, "creating importTables")
	} else {
		__antithesis_instrumentation__.Notify(493635)
	}
	__antithesis_instrumentation__.Notify(493609)

	newPreparedTableDescs := make([]*descpb.TableDescriptor, len(newMutableTableDescriptors))
	for i := range newMutableTableDescriptors {
		__antithesis_instrumentation__.Notify(493636)
		newPreparedTableDescs[i] = newMutableTableDescriptors[i].TableDesc()
	}
	__antithesis_instrumentation__.Notify(493610)

	return newPreparedTableDescs, nil
}

func (r *importResumer) prepareSchemasForIngestion(
	ctx context.Context,
	p sql.JobExecContext,
	details jobspb.ImportDetails,
	txn *kv.Txn,
	descsCol *descs.Collection,
) (*preparedSchemaMetadata, error) {
	__antithesis_instrumentation__.Notify(493637)
	schemaMetadata := &preparedSchemaMetadata{
		schemaPreparedDetails: details,
		newSchemaIDToName:     make(map[descpb.ID]string),
		oldSchemaIDToName:     make(map[descpb.ID]string),
	}

	schemaMetadata.schemaPreparedDetails.Schemas = make([]jobspb.ImportDetails_Schema,
		len(details.Schemas))

	desc, err := descsCol.GetMutableDescriptorByID(ctx, txn, details.ParentID)
	if err != nil {
		__antithesis_instrumentation__.Notify(493643)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493644)
	}
	__antithesis_instrumentation__.Notify(493638)

	dbDesc, ok := desc.(*dbdesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(493645)
		return nil, errors.Newf("expected ID %d to refer to the database being imported into",
			details.ParentID)
	} else {
		__antithesis_instrumentation__.Notify(493646)
	}
	__antithesis_instrumentation__.Notify(493639)

	schemaMetadata.schemaRewrites = make(jobspb.DescRewriteMap)
	mutableSchemaDescs := make([]*schemadesc.Mutable, 0)
	for _, desc := range details.Schemas {
		__antithesis_instrumentation__.Notify(493647)
		schemaMetadata.oldSchemaIDToName[desc.Desc.GetID()] = desc.Desc.GetName()
		newMutableSchemaDescriptor := schemadesc.NewBuilder(desc.Desc).BuildCreatedMutable().(*schemadesc.Mutable)

		id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(493649)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493650)
		}
		__antithesis_instrumentation__.Notify(493648)
		newMutableSchemaDescriptor.Version = 1
		newMutableSchemaDescriptor.ID = id
		mutableSchemaDescs = append(mutableSchemaDescs, newMutableSchemaDescriptor)

		schemaMetadata.newSchemaIDToName[id] = newMutableSchemaDescriptor.GetName()

		dbDesc.AddSchemaToDatabase(newMutableSchemaDescriptor.Name,
			descpb.DatabaseDescriptor_SchemaInfo{ID: newMutableSchemaDescriptor.ID})

		schemaMetadata.schemaRewrites[desc.Desc.ID] = &jobspb.DescriptorRewrite{
			ID: id,
		}
	}
	__antithesis_instrumentation__.Notify(493640)

	schemaMetadata.queuedSchemaJobs, err = writeNonDropDatabaseChange(ctx, dbDesc, txn, descsCol, p,
		fmt.Sprintf("updating parent database %s when importing new schemas", dbDesc.GetName()))
	if err != nil {
		__antithesis_instrumentation__.Notify(493651)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493652)
	}
	__antithesis_instrumentation__.Notify(493641)

	for i, mutDesc := range mutableSchemaDescs {
		__antithesis_instrumentation__.Notify(493653)
		nameKey := catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, dbDesc.ID, mutDesc.GetName())
		err = createSchemaDescriptorWithID(ctx, nameKey, mutDesc.ID, mutDesc, p, descsCol, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(493655)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493656)
		}
		__antithesis_instrumentation__.Notify(493654)
		schemaMetadata.schemaPreparedDetails.Schemas[i] = jobspb.ImportDetails_Schema{
			Desc: mutDesc.SchemaDesc(),
		}
	}
	__antithesis_instrumentation__.Notify(493642)

	return schemaMetadata, err
}

func createSchemaDescriptorWithID(
	ctx context.Context,
	idKey roachpb.Key,
	id descpb.ID,
	descriptor catalog.Descriptor,
	p sql.JobExecContext,
	descsCol *descs.Collection,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(493657)
	if descriptor.GetID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(493664)
		return errors.AssertionFailedf("cannot create descriptor with an empty ID: %v", descriptor)
	} else {
		__antithesis_instrumentation__.Notify(493665)
	}
	__antithesis_instrumentation__.Notify(493658)
	if descriptor.GetID() != id {
		__antithesis_instrumentation__.Notify(493666)
		return errors.AssertionFailedf("cannot create descriptor with an ID %v; expected ID %v; descriptor %v",
			id, descriptor.GetID(), descriptor)
	} else {
		__antithesis_instrumentation__.Notify(493667)
	}
	__antithesis_instrumentation__.Notify(493659)
	b := &kv.Batch{}
	descID := descriptor.GetID()
	if p.ExtendedEvalContext().Tracing.KVTracingEnabled() {
		__antithesis_instrumentation__.Notify(493668)
		log.VEventf(ctx, 2, "CPut %s -> %d", idKey, descID)
	} else {
		__antithesis_instrumentation__.Notify(493669)
	}
	__antithesis_instrumentation__.Notify(493660)
	b.CPut(idKey, descID, nil)
	if err := descsCol.Direct().WriteNewDescToBatch(
		ctx,
		p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		b,
		descriptor,
	); err != nil {
		__antithesis_instrumentation__.Notify(493670)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493671)
	}
	__antithesis_instrumentation__.Notify(493661)

	mutDesc, ok := descriptor.(catalog.MutableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(493672)
		return errors.Newf("unexpected type %T when creating descriptor", descriptor)
	} else {
		__antithesis_instrumentation__.Notify(493673)
	}
	__antithesis_instrumentation__.Notify(493662)
	switch mutDesc.(type) {
	case *schemadesc.Mutable:
		__antithesis_instrumentation__.Notify(493674)
		if err := descsCol.AddUncommittedDescriptor(mutDesc); err != nil {
			__antithesis_instrumentation__.Notify(493676)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493677)
		}
	default:
		__antithesis_instrumentation__.Notify(493675)
		return errors.Newf("unexpected type %T when creating descriptor", mutDesc)
	}
	__antithesis_instrumentation__.Notify(493663)

	return txn.Run(ctx, b)
}

func (r *importResumer) parseBundleSchemaIfNeeded(ctx context.Context, phs interface{}) error {
	__antithesis_instrumentation__.Notify(493678)
	p := phs.(sql.JobExecContext)
	seqVals := make(map[descpb.ID]int64)
	details := r.job.Details().(jobspb.ImportDetails)
	skipFKs := details.SkipFKs
	parentID := details.ParentID
	files := details.URIs
	format := details.Format

	owner := r.job.Payload().UsernameProto.Decode()

	p.SessionDataMutatorIterator().SetSessionDefaultIntSize(details.DefaultIntSize)

	if details.ParseBundleSchema {
		__antithesis_instrumentation__.Notify(493680)
		var span *tracing.Span
		ctx, span = tracing.ChildSpan(ctx, "import-parsing-bundle-schema")
		defer span.Finish()

		if err := r.job.RunningStatus(ctx, nil, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(493687)
			return runningStatusImportBundleParseSchema, nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(493688)
			return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(r.job.ID()))
		} else {
			__antithesis_instrumentation__.Notify(493689)
		}
		__antithesis_instrumentation__.Notify(493681)

		var dbDesc catalog.DatabaseDescriptor
		{
			__antithesis_instrumentation__.Notify(493690)
			if err := sql.DescsTxn(ctx, p.ExecCfg(), func(
				ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
			) (err error) {
				__antithesis_instrumentation__.Notify(493691)
				_, dbDesc, err = descriptors.GetImmutableDatabaseByID(ctx, txn, parentID, tree.DatabaseLookupFlags{
					Required:    true,
					AvoidLeased: true,
				})
				if err != nil {
					__antithesis_instrumentation__.Notify(493693)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493694)
				}
				__antithesis_instrumentation__.Notify(493692)
				return err
			}); err != nil {
				__antithesis_instrumentation__.Notify(493695)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493696)
			}
		}
		__antithesis_instrumentation__.Notify(493682)

		var schemaDescs []*schemadesc.Mutable
		var tableDescs []*tabledesc.Mutable
		var err error
		walltime := p.ExecCfg().Clock.Now().WallTime

		if tableDescs, schemaDescs, err = parseAndCreateBundleTableDescs(
			ctx, p, details, seqVals, skipFKs, dbDesc, files, format, walltime, owner,
			r.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(493697)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493698)
		}
		__antithesis_instrumentation__.Notify(493683)

		schemaDetails := make([]jobspb.ImportDetails_Schema, len(schemaDescs))
		for i, schemaDesc := range schemaDescs {
			__antithesis_instrumentation__.Notify(493699)
			schemaDetails[i] = jobspb.ImportDetails_Schema{Desc: schemaDesc.SchemaDesc()}
		}
		__antithesis_instrumentation__.Notify(493684)
		details.Schemas = schemaDetails

		tableDetails := make([]jobspb.ImportDetails_Table, len(tableDescs))
		for i, tableDesc := range tableDescs {
			__antithesis_instrumentation__.Notify(493700)
			tableDetails[i] = jobspb.ImportDetails_Table{
				Name:   tableDesc.GetName(),
				Desc:   tableDesc.TableDesc(),
				SeqVal: seqVals[tableDescs[i].ID],
				IsNew:  true,
			}
		}
		__antithesis_instrumentation__.Notify(493685)
		details.Tables = tableDetails

		for _, tbl := range tableDescs {
			__antithesis_instrumentation__.Notify(493701)

			for _, col := range tbl.Columns {
				__antithesis_instrumentation__.Notify(493702)
				if col.Type.UserDefined() {
					__antithesis_instrumentation__.Notify(493703)
					return errors.Newf("IMPORT cannot be used with user defined types; use IMPORT INTO instead")
				} else {
					__antithesis_instrumentation__.Notify(493704)
				}
			}
		}
		__antithesis_instrumentation__.Notify(493686)

		details.ParseBundleSchema = false
		if err := r.job.SetDetails(ctx, nil, details); err != nil {
			__antithesis_instrumentation__.Notify(493705)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493706)
		}
	} else {
		__antithesis_instrumentation__.Notify(493707)
	}
	__antithesis_instrumentation__.Notify(493679)
	return nil
}

func getPublicSchemaDescForDatabase(
	ctx context.Context, execCfg *sql.ExecutorConfig, db catalog.DatabaseDescriptor,
) (scDesc catalog.SchemaDescriptor, err error) {
	__antithesis_instrumentation__.Notify(493708)
	if !execCfg.Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(493711)
		return schemadesc.GetPublicSchema(), err
	} else {
		__antithesis_instrumentation__.Notify(493712)
	}
	__antithesis_instrumentation__.Notify(493709)
	if err := sql.DescsTxn(ctx, execCfg, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(493713)
		publicSchemaID := db.GetSchemaID(tree.PublicSchema)
		scDesc, err = descriptors.GetImmutableSchemaByID(ctx, txn, publicSchemaID, tree.SchemaLookupFlags{Required: true})
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(493714)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493715)
	}
	__antithesis_instrumentation__.Notify(493710)

	return scDesc, nil
}

func parseAndCreateBundleTableDescs(
	ctx context.Context,
	p sql.JobExecContext,
	details jobspb.ImportDetails,
	seqVals map[descpb.ID]int64,
	skipFKs bool,
	parentDB catalog.DatabaseDescriptor,
	files []string,
	format roachpb.IOFileFormat,
	walltime int64,
	owner security.SQLUsername,
	jobID jobspb.JobID,
) ([]*tabledesc.Mutable, []*schemadesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(493716)

	var schemaDescs []*schemadesc.Mutable
	var tableDescs []*tabledesc.Mutable
	var tableName string

	if len(details.Tables) > 0 {
		__antithesis_instrumentation__.Notify(493724)
		tableName = details.Tables[0].Name
	} else {
		__antithesis_instrumentation__.Notify(493725)
	}
	__antithesis_instrumentation__.Notify(493717)

	store, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, files[0], p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(493726)
		return tableDescs, schemaDescs, err
	} else {
		__antithesis_instrumentation__.Notify(493727)
	}
	__antithesis_instrumentation__.Notify(493718)
	defer store.Close()

	raw, err := store.ReadFile(ctx, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(493728)
		return tableDescs, schemaDescs, err
	} else {
		__antithesis_instrumentation__.Notify(493729)
	}
	__antithesis_instrumentation__.Notify(493719)
	defer raw.Close(ctx)
	reader, err := decompressingReader(ioctx.ReaderCtxAdapter(ctx, raw), files[0], format.Compression)
	if err != nil {
		__antithesis_instrumentation__.Notify(493730)
		return tableDescs, schemaDescs, err
	} else {
		__antithesis_instrumentation__.Notify(493731)
	}
	__antithesis_instrumentation__.Notify(493720)
	defer reader.Close()

	fks := fkHandler{skip: skipFKs, allowed: true, resolver: fkResolver{
		tableNameToDesc: make(map[string]*tabledesc.Mutable),
	}}
	switch format.Format {
	case roachpb.IOFileFormat_Mysqldump:
		__antithesis_instrumentation__.Notify(493732)
		id, err := descidgen.PeekNextUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(493736)
			return tableDescs, schemaDescs, err
		} else {
			__antithesis_instrumentation__.Notify(493737)
		}
		__antithesis_instrumentation__.Notify(493733)
		fks.resolver.format.Format = roachpb.IOFileFormat_Mysqldump
		evalCtx := &p.ExtendedEvalContext().EvalContext
		tableDescs, err = readMysqlCreateTable(
			ctx, reader, evalCtx, p, id, parentDB, tableName, fks,
			seqVals, owner, walltime,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(493738)
			return tableDescs, schemaDescs, err
		} else {
			__antithesis_instrumentation__.Notify(493739)
		}
	case roachpb.IOFileFormat_PgDump:
		__antithesis_instrumentation__.Notify(493734)
		fks.resolver.format.Format = roachpb.IOFileFormat_PgDump
		evalCtx := &p.ExtendedEvalContext().EvalContext

		unsupportedStmtLogger := makeUnsupportedStmtLogger(ctx, p.User(), int64(jobID),
			format.PgDump.IgnoreUnsupported, format.PgDump.IgnoreUnsupportedLog, schemaParsing,
			p.ExecCfg().DistSQLSrv.ExternalStorage)

		tableDescs, schemaDescs, err = readPostgresCreateTable(ctx, reader, evalCtx, p, tableName,
			parentDB, walltime, fks, int(format.PgDump.MaxRowSize), owner, unsupportedStmtLogger)

		logErr := unsupportedStmtLogger.flush()
		if logErr != nil {
			__antithesis_instrumentation__.Notify(493740)
			return nil, nil, logErr
		} else {
			__antithesis_instrumentation__.Notify(493741)
		}

	default:
		__antithesis_instrumentation__.Notify(493735)
		return tableDescs, schemaDescs, errors.Errorf(
			"non-bundle format %q does not support reading schemas", format.Format.String())
	}
	__antithesis_instrumentation__.Notify(493721)

	if err != nil {
		__antithesis_instrumentation__.Notify(493742)
		return tableDescs, schemaDescs, err
	} else {
		__antithesis_instrumentation__.Notify(493743)
	}
	__antithesis_instrumentation__.Notify(493722)

	if tableDescs == nil && func() bool {
		__antithesis_instrumentation__.Notify(493744)
		return len(details.Tables) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(493745)
		return tableDescs, schemaDescs, errors.Errorf("table definition not found for %q", tableName)
	} else {
		__antithesis_instrumentation__.Notify(493746)
	}
	__antithesis_instrumentation__.Notify(493723)

	return tableDescs, schemaDescs, err
}

func (r *importResumer) publishTables(
	ctx context.Context, execCfg *sql.ExecutorConfig, res roachpb.BulkOpSummary,
) error {
	__antithesis_instrumentation__.Notify(493747)
	details := r.job.Details().(jobspb.ImportDetails)

	if details.TablesPublished {
		__antithesis_instrumentation__.Notify(493752)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(493753)
	}
	__antithesis_instrumentation__.Notify(493748)

	r.writeStubStatisticsForImportedTables(ctx, execCfg, res)

	log.Event(ctx, "making tables live")

	err := sql.DescsTxn(ctx, execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(493754)
		b := txn.NewBatch()
		for _, tbl := range details.Tables {
			__antithesis_instrumentation__.Notify(493758)
			newTableDesc, err := descsCol.GetMutableTableVersionByID(ctx, tbl.Desc.ID, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(493761)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493762)
			}
			__antithesis_instrumentation__.Notify(493759)
			newTableDesc.SetPublic()

			if !tbl.IsNew {
				__antithesis_instrumentation__.Notify(493763)

				newTableDesc.OutboundFKs = make([]descpb.ForeignKeyConstraint, len(newTableDesc.OutboundFKs))
				copy(newTableDesc.OutboundFKs, tbl.Desc.OutboundFKs)
				for i := range newTableDesc.OutboundFKs {
					__antithesis_instrumentation__.Notify(493765)
					newTableDesc.OutboundFKs[i].Validity = descpb.ConstraintValidity_Unvalidated
				}
				__antithesis_instrumentation__.Notify(493764)

				for _, c := range newTableDesc.AllActiveAndInactiveChecks() {
					__antithesis_instrumentation__.Notify(493766)
					c.Validity = descpb.ConstraintValidity_Unvalidated
				}
			} else {
				__antithesis_instrumentation__.Notify(493767)
			}
			__antithesis_instrumentation__.Notify(493760)

			if err := descsCol.WriteDescToBatch(
				ctx, false, newTableDesc, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(493768)
				return errors.Wrapf(err, "publishing table %d", newTableDesc.ID)
			} else {
				__antithesis_instrumentation__.Notify(493769)
			}
		}
		__antithesis_instrumentation__.Notify(493755)
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(493770)
			return errors.Wrap(err, "publishing tables")
		} else {
			__antithesis_instrumentation__.Notify(493771)
		}
		__antithesis_instrumentation__.Notify(493756)

		details.TablesPublished = true
		err := r.job.SetDetails(ctx, txn, details)
		if err != nil {
			__antithesis_instrumentation__.Notify(493772)
			return errors.Wrap(err, "updating job details after publishing tables")
		} else {
			__antithesis_instrumentation__.Notify(493773)
		}
		__antithesis_instrumentation__.Notify(493757)
		return nil
	})
	__antithesis_instrumentation__.Notify(493749)
	if err != nil {
		__antithesis_instrumentation__.Notify(493774)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493775)
	}
	__antithesis_instrumentation__.Notify(493750)

	for i := range details.Tables {
		__antithesis_instrumentation__.Notify(493776)
		desc := tabledesc.NewBuilder(details.Tables[i].Desc).BuildImmutableTable()
		execCfg.StatsRefresher.NotifyMutation(desc, math.MaxInt32)
	}
	__antithesis_instrumentation__.Notify(493751)

	return nil
}

func (r *importResumer) writeStubStatisticsForImportedTables(
	ctx context.Context, execCfg *sql.ExecutorConfig, res roachpb.BulkOpSummary,
) {
	__antithesis_instrumentation__.Notify(493777)
	details := r.job.Details().(jobspb.ImportDetails)
	for _, tbl := range details.Tables {
		__antithesis_instrumentation__.Notify(493778)
		if tbl.IsNew {
			__antithesis_instrumentation__.Notify(493779)
			desc := tabledesc.NewBuilder(tbl.Desc).BuildImmutableTable()
			id := roachpb.BulkOpSummaryID(uint64(desc.GetID()), uint64(desc.GetPrimaryIndexID()))
			rowCount := uint64(res.EntryCounts[id])

			distinctCount := uint64(float64(rowCount) * memo.UnknownDistinctCountRatio)
			nullCount := uint64(float64(rowCount) * memo.UnknownNullCountRatio)
			avgRowSize := uint64(memo.UnknownAvgRowSize)

			multiColEnabled := false
			statistics, err := sql.StubTableStats(desc, jobspb.ImportStatsName, multiColEnabled)
			if err == nil {
				__antithesis_instrumentation__.Notify(493781)
				for _, statistic := range statistics {
					__antithesis_instrumentation__.Notify(493783)
					statistic.RowCount = rowCount
					statistic.DistinctCount = distinctCount
					statistic.NullCount = nullCount
					statistic.AvgSize = avgRowSize
				}
				__antithesis_instrumentation__.Notify(493782)

				err = stats.InsertNewStats(ctx, execCfg.Settings, execCfg.InternalExecutor, nil, statistics)
			} else {
				__antithesis_instrumentation__.Notify(493784)
			}
			__antithesis_instrumentation__.Notify(493780)
			if err != nil {
				__antithesis_instrumentation__.Notify(493785)

				log.Warningf(
					ctx, "error while creating statistics during import of %q: %v",
					desc.GetName(), err,
				)
			} else {
				__antithesis_instrumentation__.Notify(493786)
			}
		} else {
			__antithesis_instrumentation__.Notify(493787)
		}
	}
}

func (r *importResumer) publishSchemas(ctx context.Context, execCfg *sql.ExecutorConfig) error {
	__antithesis_instrumentation__.Notify(493788)
	details := r.job.Details().(jobspb.ImportDetails)

	if details.SchemasPublished {
		__antithesis_instrumentation__.Notify(493790)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(493791)
	}
	__antithesis_instrumentation__.Notify(493789)
	log.Event(ctx, "making schemas live")

	return sql.DescsTxn(ctx, execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(493792)
		b := txn.NewBatch()
		for _, schema := range details.Schemas {
			__antithesis_instrumentation__.Notify(493796)
			newDesc, err := descsCol.GetMutableDescriptorByID(ctx, txn, schema.Desc.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(493799)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493800)
			}
			__antithesis_instrumentation__.Notify(493797)
			newSchemaDesc, ok := newDesc.(*schemadesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(493801)
				return errors.Newf("expected schema descriptor with ID %v, got %v",
					schema.Desc.GetID(), newDesc)
			} else {
				__antithesis_instrumentation__.Notify(493802)
			}
			__antithesis_instrumentation__.Notify(493798)
			newSchemaDesc.SetPublic()
			if err := descsCol.WriteDescToBatch(
				ctx, false, newSchemaDesc, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(493803)
				return errors.Wrapf(err, "publishing schema %d", newSchemaDesc.ID)
			} else {
				__antithesis_instrumentation__.Notify(493804)
			}
		}
		__antithesis_instrumentation__.Notify(493793)
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(493805)
			return errors.Wrap(err, "publishing schemas")
		} else {
			__antithesis_instrumentation__.Notify(493806)
		}
		__antithesis_instrumentation__.Notify(493794)

		details.SchemasPublished = true
		err := r.job.SetDetails(ctx, txn, details)
		if err != nil {
			__antithesis_instrumentation__.Notify(493807)
			return errors.Wrap(err, "updating job details after publishing schemas")
		} else {
			__antithesis_instrumentation__.Notify(493808)
		}
		__antithesis_instrumentation__.Notify(493795)
		return nil
	})
}

func (*importResumer) checkVirtualConstraints(
	ctx context.Context, execCfg *sql.ExecutorConfig, job *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(493809)
	for _, tbl := range job.Details().(jobspb.ImportDetails).Tables {
		__antithesis_instrumentation__.Notify(493811)
		desc := tabledesc.NewBuilder(tbl.Desc).BuildExistingMutableTable()
		desc.SetPublic()

		if sql.HasVirtualUniqueConstraints(desc) {
			__antithesis_instrumentation__.Notify(493813)
			if err := job.RunningStatus(ctx, nil, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
				__antithesis_instrumentation__.Notify(493814)
				return jobs.RunningStatus(fmt.Sprintf("re-validating %s", desc.GetName())), nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(493815)
				return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(job.ID()))
			} else {
				__antithesis_instrumentation__.Notify(493816)
			}
		} else {
			__antithesis_instrumentation__.Notify(493817)
		}
		__antithesis_instrumentation__.Notify(493812)

		if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(493818)
			ie := execCfg.InternalExecutorFactory(ctx, sql.NewFakeSessionData(execCfg.SV()))
			return ie.WithSyntheticDescriptors([]catalog.Descriptor{desc}, func() error {
				__antithesis_instrumentation__.Notify(493819)
				return sql.RevalidateUniqueConstraintsInTable(ctx, txn, ie, desc)
			})
		}); err != nil {
			__antithesis_instrumentation__.Notify(493820)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493821)
		}
	}
	__antithesis_instrumentation__.Notify(493810)
	return nil
}

func (r *importResumer) checkForUDTModification(
	ctx context.Context, execCfg *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(493822)
	details := r.job.Details().(jobspb.ImportDetails)
	if details.Types == nil {
		__antithesis_instrumentation__.Notify(493827)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(493828)
	}
	__antithesis_instrumentation__.Notify(493823)

	typeDescsAreEquivalent := func(a, b *descpb.TypeDescriptor) (bool, error) {
		__antithesis_instrumentation__.Notify(493829)
		clearIgnoredFields := func(d *descpb.TypeDescriptor) *descpb.TypeDescriptor {
			__antithesis_instrumentation__.Notify(493833)
			d = protoutil.Clone(d).(*descpb.TypeDescriptor)
			d.ModificationTime = hlc.Timestamp{}
			d.Privileges = nil
			d.Version = 0
			d.ReferencingDescriptorIDs = nil
			return d
		}
		__antithesis_instrumentation__.Notify(493830)
		aData, err := protoutil.Marshal(clearIgnoredFields(a))
		if err != nil {
			__antithesis_instrumentation__.Notify(493834)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(493835)
		}
		__antithesis_instrumentation__.Notify(493831)
		bData, err := protoutil.Marshal(clearIgnoredFields(b))
		if err != nil {
			__antithesis_instrumentation__.Notify(493836)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(493837)
		}
		__antithesis_instrumentation__.Notify(493832)
		return bytes.Equal(aData, bData), nil
	}
	__antithesis_instrumentation__.Notify(493824)

	checkTypeIsEquivalent := func(
		ctx context.Context, txn *kv.Txn, col *descs.Collection,
		savedTypeDesc *descpb.TypeDescriptor,
	) error {
		__antithesis_instrumentation__.Notify(493838)
		typeDesc, err := col.Direct().MustGetTypeDescByID(ctx, txn, savedTypeDesc.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(493843)
			return errors.Wrap(err, "resolving type descriptor when checking version mismatch")
		} else {
			__antithesis_instrumentation__.Notify(493844)
		}
		__antithesis_instrumentation__.Notify(493839)
		if typeDesc.GetModificationTime() == savedTypeDesc.GetModificationTime() {
			__antithesis_instrumentation__.Notify(493845)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(493846)
		}
		__antithesis_instrumentation__.Notify(493840)
		equivalent, err := typeDescsAreEquivalent(typeDesc.TypeDesc(), savedTypeDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(493847)
			return errors.NewAssertionErrorWithWrappedErrf(
				err, "failed to check for type descriptor equivalence for type %q (%d)",
				typeDesc.GetName(), typeDesc.GetID())
		} else {
			__antithesis_instrumentation__.Notify(493848)
		}
		__antithesis_instrumentation__.Notify(493841)
		if equivalent {
			__antithesis_instrumentation__.Notify(493849)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(493850)
		}
		__antithesis_instrumentation__.Notify(493842)
		return errors.WithHint(
			errors.Newf(
				"type descriptor %q (%d) has been modified, potentially incompatibly,"+
					" since import planning; aborting to avoid possible corruption",
				typeDesc.GetName(), typeDesc.GetID(),
			),
			"retrying the IMPORT operation may succeed if the operation concurrently"+
				" modifying the descriptor does not reoccur during the retry attempt",
		)
	}
	__antithesis_instrumentation__.Notify(493825)
	checkTypesAreEquivalent := func(
		ctx context.Context, txn *kv.Txn, col *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(493851)
		for _, savedTypeDesc := range details.Types {
			__antithesis_instrumentation__.Notify(493853)
			if err := checkTypeIsEquivalent(
				ctx, txn, col, savedTypeDesc.Desc,
			); err != nil {
				__antithesis_instrumentation__.Notify(493854)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493855)
			}
		}
		__antithesis_instrumentation__.Notify(493852)
		return nil
	}
	__antithesis_instrumentation__.Notify(493826)
	return sql.DescsTxn(ctx, execCfg, checkTypesAreEquivalent)
}

func ingestWithRetry(
	ctx context.Context,
	execCtx sql.JobExecContext,
	job *jobs.Job,
	tables map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
	typeDescs []*descpb.TypeDescriptor,
	from []string,
	format roachpb.IOFileFormat,
	walltime int64,
	alwaysFlushProgress bool,
	procsPerNode int,
) (roachpb.BulkOpSummary, error) {
	__antithesis_instrumentation__.Notify(493856)
	resumerSpan := tracing.SpanFromContext(ctx)

	retryOpts := retry.Options{
		MaxBackoff: 1 * time.Second,
		MaxRetries: 5,
	}

	var res roachpb.BulkOpSummary
	var err error
	var retryCount int32
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(493859)
		retryCount++
		resumerSpan.RecordStructured(&roachpb.RetryTracingEvent{
			Operation:     "importResumer.ingestWithRetry",
			AttemptNumber: retryCount,
			RetryError:    tracing.RedactAndTruncateError(err),
		})
		res, err = distImport(ctx, execCtx, job, tables, typeDescs, from, format, walltime,
			alwaysFlushProgress, procsPerNode)
		if err == nil {
			__antithesis_instrumentation__.Notify(493864)
			break
		} else {
			__antithesis_instrumentation__.Notify(493865)
		}
		__antithesis_instrumentation__.Notify(493860)

		if errors.HasType(err, &roachpb.InsufficientSpaceError{}) {
			__antithesis_instrumentation__.Notify(493866)
			return res, jobs.MarkPauseRequestError(errors.UnwrapAll(err))
		} else {
			__antithesis_instrumentation__.Notify(493867)
		}
		__antithesis_instrumentation__.Notify(493861)

		if joberror.IsPermanentBulkJobError(err) {
			__antithesis_instrumentation__.Notify(493868)
			return res, err
		} else {
			__antithesis_instrumentation__.Notify(493869)
		}
		__antithesis_instrumentation__.Notify(493862)

		reloadedJob, reloadErr := execCtx.ExecCfg().JobRegistry.LoadClaimedJob(ctx, job.ID())
		if reloadErr != nil {
			__antithesis_instrumentation__.Notify(493870)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(493872)
				return res, ctx.Err()
			} else {
				__antithesis_instrumentation__.Notify(493873)
			}
			__antithesis_instrumentation__.Notify(493871)
			log.Warningf(ctx, `IMPORT job %d could not reload job progress when retrying: %+v`,
				int64(job.ID()), reloadErr)
		} else {
			__antithesis_instrumentation__.Notify(493874)
			job = reloadedJob
		}
		__antithesis_instrumentation__.Notify(493863)
		log.Warningf(ctx, `encountered retryable error: %+v`, err)
	}
	__antithesis_instrumentation__.Notify(493857)

	if err != nil {
		__antithesis_instrumentation__.Notify(493875)
		return roachpb.BulkOpSummary{}, errors.Wrap(err, "exhausted retries")
	} else {
		__antithesis_instrumentation__.Notify(493876)
	}
	__antithesis_instrumentation__.Notify(493858)
	return res, nil
}

func emitImportJobEvent(
	ctx context.Context, p sql.JobExecContext, status jobs.Status, job *jobs.Job,
) {
	__antithesis_instrumentation__.Notify(493877)
	var importEvent eventpb.Import
	if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(493878)
		return sql.LogEventForJobs(ctx, p.ExecCfg(), txn, &importEvent, int64(job.ID()),
			job.Payload(), p.User(), status)
	}); err != nil {
		__antithesis_instrumentation__.Notify(493879)
		log.Warningf(ctx, "failed to log event: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(493880)
	}
}

func constructSchemaAndTableKey(
	ctx context.Context,
	tableDesc *descpb.TableDescriptor,
	schemaIDToName map[descpb.ID]string,
	version clusterversion.Handle,
) (schemaAndTableName, error) {
	__antithesis_instrumentation__.Notify(493881)
	if !version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(493884)
		if tableDesc.UnexposedParentSchemaID == keys.PublicSchemaIDForBackup {
			__antithesis_instrumentation__.Notify(493885)
			return schemaAndTableName{schema: "", table: tableDesc.GetName()}, nil
		} else {
			__antithesis_instrumentation__.Notify(493886)
		}
	} else {
		__antithesis_instrumentation__.Notify(493887)
	}
	__antithesis_instrumentation__.Notify(493882)
	schemaName, ok := schemaIDToName[tableDesc.GetUnexposedParentSchemaID()]
	if !ok && func() bool {
		__antithesis_instrumentation__.Notify(493888)
		return schemaName != tree.PublicSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(493889)
		return schemaAndTableName{}, errors.Newf("invalid parent schema %s with ID %d for table %s",
			schemaName, tableDesc.UnexposedParentSchemaID, tableDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(493890)
	}
	__antithesis_instrumentation__.Notify(493883)

	return schemaAndTableName{schema: schemaName, table: tableDesc.GetName()}, nil
}

func writeNonDropDatabaseChange(
	ctx context.Context,
	desc *dbdesc.Mutable,
	txn *kv.Txn,
	descsCol *descs.Collection,
	p sql.JobExecContext,
	jobDesc string,
) ([]jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(493891)
	var job *jobs.Job
	var err error
	if job, err = createNonDropDatabaseChangeJob(p.User(), desc.ID, jobDesc, p, txn); err != nil {
		__antithesis_instrumentation__.Notify(493894)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493895)
	}
	__antithesis_instrumentation__.Notify(493892)

	queuedJob := []jobspb.JobID{job.ID()}
	b := txn.NewBatch()
	err = descsCol.WriteDescToBatch(
		ctx,
		p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		desc,
		b,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(493896)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493897)
	}
	__antithesis_instrumentation__.Notify(493893)
	return queuedJob, txn.Run(ctx, b)
}

func createNonDropDatabaseChangeJob(
	user security.SQLUsername,
	databaseID descpb.ID,
	jobDesc string,
	p sql.JobExecContext,
	txn *kv.Txn,
) (*jobs.Job, error) {
	__antithesis_instrumentation__.Notify(493898)
	jobRecord := jobs.Record{
		Description: jobDesc,
		Username:    user,
		Details: jobspb.SchemaChangeDetails{
			DescID:        databaseID,
			FormatVersion: jobspb.DatabaseJobFormatVersion,
		},
		Progress:      jobspb.SchemaChangeProgress{},
		NonCancelable: true,
	}

	jobID := p.ExecCfg().JobRegistry.MakeJobID()
	return p.ExecCfg().JobRegistry.CreateJobWithTxn(
		p.ExtendedEvalContext().Context,
		jobRecord,
		jobID,
		txn,
	)
}

func (r *importResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(493899)
	p := execCtx.(sql.JobExecContext)

	emitImportJobEvent(ctx, p, jobs.StatusReverting, r.job)

	details := r.job.Details().(jobspb.ImportDetails)
	addToFileFormatTelemetry(details.Format.Format.String(), "failed")
	cfg := execCtx.(sql.JobExecContext).ExecCfg()
	var jobsToRunAfterTxnCommit []jobspb.JobID
	if err := sql.DescsTxn(ctx, cfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(493902)
		if err := r.dropTables(ctx, txn, descsCol, cfg); err != nil {
			__antithesis_instrumentation__.Notify(493905)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493906)
		}
		__antithesis_instrumentation__.Notify(493903)

		var err error
		jobsToRunAfterTxnCommit, err = r.dropSchemas(ctx, txn, descsCol, cfg, p)
		if err != nil {
			__antithesis_instrumentation__.Notify(493907)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493908)
		}
		__antithesis_instrumentation__.Notify(493904)

		return r.releaseProtectedTimestamp(ctx, txn, cfg.ProtectedTimestampProvider)
	}); err != nil {
		__antithesis_instrumentation__.Notify(493909)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493910)
	}
	__antithesis_instrumentation__.Notify(493900)

	if len(jobsToRunAfterTxnCommit) != 0 {
		__antithesis_instrumentation__.Notify(493911)
		if err := p.ExecCfg().JobRegistry.Run(ctx, p.ExecCfg().InternalExecutor,
			jobsToRunAfterTxnCommit); err != nil {
			__antithesis_instrumentation__.Notify(493912)
			return errors.Wrap(err, "failed to run jobs that drop the imported schemas")
		} else {
			__antithesis_instrumentation__.Notify(493913)
		}
	} else {
		__antithesis_instrumentation__.Notify(493914)
	}
	__antithesis_instrumentation__.Notify(493901)

	emitImportJobEvent(ctx, p, jobs.StatusFailed, r.job)

	return nil
}

func (r *importResumer) dropTables(
	ctx context.Context, txn *kv.Txn, descsCol *descs.Collection, execCfg *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(493915)
	details := r.job.Details().(jobspb.ImportDetails)
	dropTime := int64(1)

	if !details.PrepareComplete {
		__antithesis_instrumentation__.Notify(493924)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(493925)
	}
	__antithesis_instrumentation__.Notify(493916)

	var revert []catalog.TableDescriptor
	var empty []catalog.TableDescriptor
	for _, tbl := range details.Tables {
		__antithesis_instrumentation__.Notify(493926)
		if !tbl.IsNew {
			__antithesis_instrumentation__.Notify(493927)
			desc, err := descsCol.GetMutableTableVersionByID(ctx, tbl.Desc.ID, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(493929)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493930)
			}
			__antithesis_instrumentation__.Notify(493928)
			imm := desc.ImmutableCopy().(catalog.TableDescriptor)
			if tbl.WasEmpty {
				__antithesis_instrumentation__.Notify(493931)
				empty = append(empty, imm)
			} else {
				__antithesis_instrumentation__.Notify(493932)
				revert = append(revert, imm)
			}
		} else {
			__antithesis_instrumentation__.Notify(493933)
		}
	}
	__antithesis_instrumentation__.Notify(493917)

	if details.Walltime != 0 && func() bool {
		__antithesis_instrumentation__.Notify(493934)
		return len(revert) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(493935)

		ts := hlc.Timestamp{WallTime: details.Walltime}.Prev()

		const ignoreGC = true
		if err := sql.RevertTables(ctx, txn.DB(), execCfg, revert, ts, ignoreGC, sql.RevertTableDefaultBatchSize); err != nil {
			__antithesis_instrumentation__.Notify(493936)
			return errors.Wrap(err, "rolling back partially completed IMPORT")
		} else {
			__antithesis_instrumentation__.Notify(493937)
		}
	} else {
		__antithesis_instrumentation__.Notify(493938)
	}
	__antithesis_instrumentation__.Notify(493918)

	for i := range empty {
		__antithesis_instrumentation__.Notify(493939)

		empty[i].TableDesc().DropTime = dropTime
		if err := gcjob.ClearTableData(
			ctx, execCfg.DB, execCfg.DistSender, execCfg.Codec, &execCfg.Settings.SV, empty[i],
		); err != nil {
			__antithesis_instrumentation__.Notify(493940)
			return errors.Wrapf(err, "clearing data for table %d", empty[i].GetID())
		} else {
			__antithesis_instrumentation__.Notify(493941)
		}
	}
	__antithesis_instrumentation__.Notify(493919)

	b := txn.NewBatch()
	tablesToGC := make([]descpb.ID, 0, len(details.Tables))
	toWrite := make([]*tabledesc.Mutable, 0, len(details.Tables))
	for _, tbl := range details.Tables {
		__antithesis_instrumentation__.Notify(493942)
		newTableDesc, err := descsCol.GetMutableTableVersionByID(ctx, tbl.Desc.ID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(493945)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493946)
		}
		__antithesis_instrumentation__.Notify(493943)
		if tbl.IsNew {
			__antithesis_instrumentation__.Notify(493947)
			newTableDesc.SetDropped()

			newTableDesc.DropTime = dropTime
			b.Del(catalogkeys.EncodeNameKey(execCfg.Codec, newTableDesc))
			tablesToGC = append(tablesToGC, newTableDesc.ID)
			descsCol.AddDeletedDescriptor(newTableDesc.GetID())
		} else {
			__antithesis_instrumentation__.Notify(493948)

			newTableDesc.SetPublic()
		}
		__antithesis_instrumentation__.Notify(493944)

		toWrite = append(toWrite, newTableDesc)
	}
	__antithesis_instrumentation__.Notify(493920)
	for _, d := range toWrite {
		__antithesis_instrumentation__.Notify(493949)
		const kvTrace = false
		if err := descsCol.WriteDescToBatch(ctx, kvTrace, d, b); err != nil {
			__antithesis_instrumentation__.Notify(493950)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493951)
		}
	}
	__antithesis_instrumentation__.Notify(493921)

	gcDetails := jobspb.SchemaChangeGCDetails{}
	for _, tableID := range tablesToGC {
		__antithesis_instrumentation__.Notify(493952)
		gcDetails.Tables = append(gcDetails.Tables, jobspb.SchemaChangeGCDetails_DroppedID{
			ID:       tableID,
			DropTime: dropTime,
		})
	}
	__antithesis_instrumentation__.Notify(493922)
	gcJobRecord := jobs.Record{
		Description:   fmt.Sprintf("GC for %s", r.job.Payload().Description),
		Username:      r.job.Payload().UsernameProto.Decode(),
		DescriptorIDs: tablesToGC,
		Details:       gcDetails,
		Progress:      jobspb.SchemaChangeGCProgress{},
		NonCancelable: true,
	}
	if _, err := execCfg.JobRegistry.CreateJobWithTxn(
		ctx, gcJobRecord, execCfg.JobRegistry.MakeJobID(), txn); err != nil {
		__antithesis_instrumentation__.Notify(493953)
		return err
	} else {
		__antithesis_instrumentation__.Notify(493954)
	}
	__antithesis_instrumentation__.Notify(493923)

	return errors.Wrap(txn.Run(ctx, b), "rolling back tables")
}

func (r *importResumer) dropSchemas(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	execCfg *sql.ExecutorConfig,
	p sql.JobExecContext,
) ([]jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(493955)
	details := r.job.Details().(jobspb.ImportDetails)

	if !details.PrepareComplete || func() bool {
		__antithesis_instrumentation__.Notify(493962)
		return len(details.Schemas) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(493963)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(493964)
	}
	__antithesis_instrumentation__.Notify(493956)

	desc, err := descsCol.GetMutableDescriptorByID(ctx, txn, details.ParentID)
	if err != nil {
		__antithesis_instrumentation__.Notify(493965)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493966)
	}
	__antithesis_instrumentation__.Notify(493957)

	dbDesc, ok := desc.(*dbdesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(493967)
		return nil, errors.Newf("expected ID %d to refer to the database being imported into",
			details.ParentID)
	} else {
		__antithesis_instrumentation__.Notify(493968)
	}
	__antithesis_instrumentation__.Notify(493958)

	droppedSchemaIDs := make([]descpb.ID, 0)
	for _, schema := range details.Schemas {
		__antithesis_instrumentation__.Notify(493969)
		desc, err := descsCol.GetMutableDescriptorByID(ctx, txn, schema.Desc.ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(493974)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493975)
		}
		__antithesis_instrumentation__.Notify(493970)
		var schemaDesc *schemadesc.Mutable
		var ok bool
		if schemaDesc, ok = desc.(*schemadesc.Mutable); !ok {
			__antithesis_instrumentation__.Notify(493976)
			return nil, errors.Newf("unable to resolve schema desc with ID %d", schema.Desc.ID)
		} else {
			__antithesis_instrumentation__.Notify(493977)
		}
		__antithesis_instrumentation__.Notify(493971)

		schemaDesc.SetDropped()
		droppedSchemaIDs = append(droppedSchemaIDs, schemaDesc.GetID())

		b := txn.NewBatch()

		if execCfg.Settings.Version.IsActive(ctx, clusterversion.AvoidDrainingNames) {
			__antithesis_instrumentation__.Notify(493978)
			if dbDesc.Schemas != nil {
				__antithesis_instrumentation__.Notify(493980)
				delete(dbDesc.Schemas, schemaDesc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(493981)
			}
			__antithesis_instrumentation__.Notify(493979)
			b.Del(catalogkeys.EncodeNameKey(p.ExecCfg().Codec, schemaDesc))
		} else {
			__antithesis_instrumentation__.Notify(493982)

			schemaDesc.AddDrainingName(descpb.NameInfo{
				ParentID:       details.ParentID,
				ParentSchemaID: keys.RootNamespaceID,
				Name:           schemaDesc.Name,
			})

			dbDesc.AddSchemaToDatabase(schema.Desc.Name, descpb.DatabaseDescriptor_SchemaInfo{ID: dbDesc.ID, Dropped: true})
		}
		__antithesis_instrumentation__.Notify(493972)

		if err := descsCol.WriteDescToBatch(ctx, p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
			schemaDesc, b); err != nil {
			__antithesis_instrumentation__.Notify(493983)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493984)
		}
		__antithesis_instrumentation__.Notify(493973)
		err = txn.Run(ctx, b)
		if err != nil {
			__antithesis_instrumentation__.Notify(493985)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493986)
		}
	}
	__antithesis_instrumentation__.Notify(493959)

	queuedJob, err := writeNonDropDatabaseChange(ctx, dbDesc, txn, descsCol, p, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(493987)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493988)
	}
	__antithesis_instrumentation__.Notify(493960)

	dropSchemaJobRecord := jobs.Record{
		Description:   "dropping schemas as part of an import job rollback",
		Username:      p.User(),
		DescriptorIDs: droppedSchemaIDs,
		Details: jobspb.SchemaChangeDetails{
			DroppedSchemas:    droppedSchemaIDs,
			DroppedDatabaseID: descpb.InvalidID,
			FormatVersion:     jobspb.DatabaseJobFormatVersion,
		},
		Progress:      jobspb.SchemaChangeProgress{},
		NonCancelable: true,
	}
	jobID := p.ExecCfg().JobRegistry.MakeJobID()
	job, err := execCfg.JobRegistry.CreateJobWithTxn(ctx, dropSchemaJobRecord, jobID, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(493989)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493990)
	}
	__antithesis_instrumentation__.Notify(493961)
	queuedJob = append(queuedJob, job.ID())

	return queuedJob, nil
}

func (r *importResumer) releaseProtectedTimestamp(
	ctx context.Context, txn *kv.Txn, pts protectedts.Storage,
) error {
	__antithesis_instrumentation__.Notify(493991)
	details := r.job.Details().(jobspb.ImportDetails)
	ptsID := details.ProtectedTimestampRecord

	if ptsID == nil {
		__antithesis_instrumentation__.Notify(493994)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(493995)
	}
	__antithesis_instrumentation__.Notify(493992)
	err := pts.Release(ctx, txn, *ptsID)
	if errors.Is(err, protectedts.ErrNotExists) {
		__antithesis_instrumentation__.Notify(493996)

		log.Warningf(ctx, "failed to release protected which seems not to exist: %v", err)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(493997)
	}
	__antithesis_instrumentation__.Notify(493993)
	return err
}

func (r *importResumer) ReportResults(ctx context.Context, resultsCh chan<- tree.Datums) error {
	__antithesis_instrumentation__.Notify(493998)
	select {
	case resultsCh <- tree.Datums{
		tree.NewDInt(tree.DInt(r.job.ID())),
		tree.NewDString(string(jobs.StatusSucceeded)),
		tree.NewDFloat(tree.DFloat(1.0)),
		tree.NewDInt(tree.DInt(r.res.Rows)),
		tree.NewDInt(tree.DInt(r.res.IndexEntries)),
		tree.NewDInt(tree.DInt(r.res.DataSize)),
	}:
		__antithesis_instrumentation__.Notify(493999)
		return nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(494000)
		return ctx.Err()
	}
}

func init() {
	jobs.RegisterConstructor(
		jobspb.TypeImport,
		func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
			return &importResumer{
				job:      job,
				settings: settings,
			}
		},
	)
}
