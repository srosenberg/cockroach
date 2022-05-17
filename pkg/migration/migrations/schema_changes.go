package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/kr/pretty"
)

type operation struct {
	name redact.RedactableString

	schemaList []string

	query string

	schemaExistsFn func(catalog.TableDescriptor, catalog.TableDescriptor, string) (bool, error)
}

func migrateTable(
	ctx context.Context,
	_ clusterversion.ClusterVersion,
	d migration.TenantDeps,
	op operation,
	storedTableID descpb.ID,
	expectedTable catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(128581)
	for {
		__antithesis_instrumentation__.Notify(128582)

		log.Infof(ctx, "performing table migration operation %v", op.name)

		storedTable, err := readTableDescriptor(ctx, d, storedTableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(128588)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128589)
		}
		__antithesis_instrumentation__.Notify(128583)

		if mutations := storedTable.GetMutationJobs(); len(mutations) > 0 {
			__antithesis_instrumentation__.Notify(128590)
			for _, mutation := range mutations {
				__antithesis_instrumentation__.Notify(128592)
				log.Infof(ctx, "waiting for the mutation job %v to complete", mutation.JobID)
				if _, err := d.InternalExecutor.Exec(ctx, "migration-mutations-wait",
					nil, "SHOW JOB WHEN COMPLETE $1", mutation.JobID); err != nil {
					__antithesis_instrumentation__.Notify(128593)
					return err
				} else {
					__antithesis_instrumentation__.Notify(128594)
				}
			}
			__antithesis_instrumentation__.Notify(128591)
			continue
		} else {
			__antithesis_instrumentation__.Notify(128595)
		}
		__antithesis_instrumentation__.Notify(128584)

		var exists bool
		for i, schemaName := range op.schemaList {
			__antithesis_instrumentation__.Notify(128596)
			hasSchema, err := op.schemaExistsFn(storedTable, expectedTable, schemaName)
			if err != nil {
				__antithesis_instrumentation__.Notify(128599)
				return errors.Wrapf(err, "error while validating descriptors during"+
					" operation %s", op.name)
			} else {
				__antithesis_instrumentation__.Notify(128600)
			}
			__antithesis_instrumentation__.Notify(128597)
			if i > 0 && func() bool {
				__antithesis_instrumentation__.Notify(128601)
				return exists != hasSchema == true
			}() == true {
				__antithesis_instrumentation__.Notify(128602)
				return errors.Errorf("error while validating descriptors. observed"+
					" partial schema exists while performing %v", op.name)
			} else {
				__antithesis_instrumentation__.Notify(128603)
			}
			__antithesis_instrumentation__.Notify(128598)
			exists = hasSchema
		}
		__antithesis_instrumentation__.Notify(128585)
		if exists {
			__antithesis_instrumentation__.Notify(128604)
			log.Infof(ctx, "skipping %s operation as the schema change already exists.", op.name)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(128605)
		}
		__antithesis_instrumentation__.Notify(128586)

		log.Infof(ctx, "performing operation: %s", op.name)
		if _, err := d.InternalExecutor.ExecEx(
			ctx,
			fmt.Sprintf("migration-alter-table-%d", storedTableID),
			nil,
			sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
			op.query); err != nil {
			__antithesis_instrumentation__.Notify(128606)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128607)
		}
		__antithesis_instrumentation__.Notify(128587)
		return nil
	}
}

func readTableDescriptor(
	ctx context.Context, d migration.TenantDeps, tableID descpb.ID,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(128608)
	var t catalog.TableDescriptor

	if err := d.CollectionFactory.Txn(ctx, d.InternalExecutor, d.DB, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(128610)
		t, err = descriptors.GetImmutableTableByID(ctx, txn, tableID, tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				AvoidLeased: true,
				Required:    true,
			},
		})
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(128611)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(128612)
	}
	__antithesis_instrumentation__.Notify(128609)
	return t, nil
}

func ensureProtoMessagesAreEqual(expected, found protoutil.Message) error {
	__antithesis_instrumentation__.Notify(128613)
	expectedBytes, err := protoutil.Marshal(expected)
	if err != nil {
		__antithesis_instrumentation__.Notify(128617)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128618)
	}
	__antithesis_instrumentation__.Notify(128614)
	foundBytes, err := protoutil.Marshal(found)
	if err != nil {
		__antithesis_instrumentation__.Notify(128619)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128620)
	}
	__antithesis_instrumentation__.Notify(128615)
	if bytes.Equal(expectedBytes, foundBytes) {
		__antithesis_instrumentation__.Notify(128621)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128622)
	}
	__antithesis_instrumentation__.Notify(128616)
	return errors.Errorf("expected descriptor doesn't match "+
		"with found descriptor: %s", strings.Join(pretty.Diff(expected, found), "\n"))
}

func hasColumn(storedTable, expectedTable catalog.TableDescriptor, colName string) (bool, error) {
	__antithesis_instrumentation__.Notify(128623)
	storedCol, err := storedTable.FindColumnWithName(tree.Name(colName))
	if err != nil {
		__antithesis_instrumentation__.Notify(128627)
		if strings.Contains(err.Error(), "does not exist") {
			__antithesis_instrumentation__.Notify(128629)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(128630)
		}
		__antithesis_instrumentation__.Notify(128628)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(128631)
	}
	__antithesis_instrumentation__.Notify(128624)

	expectedCol, err := expectedTable.FindColumnWithName(tree.Name(colName))
	if err != nil {
		__antithesis_instrumentation__.Notify(128632)
		return false, errors.Wrapf(err, "columns name %s is invalid.", colName)
	} else {
		__antithesis_instrumentation__.Notify(128633)
	}
	__antithesis_instrumentation__.Notify(128625)

	expectedCopy := expectedCol.ColumnDescDeepCopy()
	storedCopy := storedCol.ColumnDescDeepCopy()

	storedCopy.ID = 0
	expectedCopy.ID = 0

	if err = ensureProtoMessagesAreEqual(&expectedCopy, &storedCopy); err != nil {
		__antithesis_instrumentation__.Notify(128634)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(128635)
	}
	__antithesis_instrumentation__.Notify(128626)
	return true, nil
}

func hasIndex(storedTable, expectedTable catalog.TableDescriptor, indexName string) (bool, error) {
	__antithesis_instrumentation__.Notify(128636)
	storedIdx, err := storedTable.FindIndexWithName(indexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(128640)
		if strings.Contains(err.Error(), "does not exist") {
			__antithesis_instrumentation__.Notify(128642)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(128643)
		}
		__antithesis_instrumentation__.Notify(128641)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(128644)
	}
	__antithesis_instrumentation__.Notify(128637)
	expectedIdx, err := expectedTable.FindIndexWithName(indexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(128645)
		return false, errors.Wrapf(err, "index name %s is invalid", indexName)
	} else {
		__antithesis_instrumentation__.Notify(128646)
	}
	__antithesis_instrumentation__.Notify(128638)
	storedCopy := storedIdx.IndexDescDeepCopy()
	expectedCopy := expectedIdx.IndexDescDeepCopy()

	storedCopy.ID = 0
	expectedCopy.ID = 0
	storedCopy.Version = 0
	expectedCopy.Version = 0

	storedCopy.CreatedExplicitly = false
	expectedCopy.CreatedExplicitly = false
	storedCopy.StoreColumnNames = []string{}
	expectedCopy.StoreColumnNames = []string{}
	storedCopy.StoreColumnIDs = []descpb.ColumnID{0, 0, 0}
	expectedCopy.StoreColumnIDs = []descpb.ColumnID{0, 0, 0}

	if err = ensureProtoMessagesAreEqual(&expectedCopy, &storedCopy); err != nil {
		__antithesis_instrumentation__.Notify(128647)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(128648)
	}
	__antithesis_instrumentation__.Notify(128639)
	return true, nil
}

func hasColumnFamily(
	storedTable, expectedTable catalog.TableDescriptor, colFamily string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(128649)
	var storedFamily, expectedFamily *descpb.ColumnFamilyDescriptor
	for _, fam := range storedTable.GetFamilies() {
		__antithesis_instrumentation__.Notify(128656)
		if fam.Name == colFamily {
			__antithesis_instrumentation__.Notify(128657)
			storedFamily = &fam
			break
		} else {
			__antithesis_instrumentation__.Notify(128658)
		}
	}
	__antithesis_instrumentation__.Notify(128650)
	if storedFamily == nil {
		__antithesis_instrumentation__.Notify(128659)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(128660)
	}
	__antithesis_instrumentation__.Notify(128651)

	for _, fam := range expectedTable.GetFamilies() {
		__antithesis_instrumentation__.Notify(128661)
		if fam.Name == colFamily {
			__antithesis_instrumentation__.Notify(128662)
			expectedFamily = &fam
			break
		} else {
			__antithesis_instrumentation__.Notify(128663)
		}
	}
	__antithesis_instrumentation__.Notify(128652)
	if expectedFamily == nil {
		__antithesis_instrumentation__.Notify(128664)
		return false, errors.Errorf("column family %s does not exist", colFamily)
	} else {
		__antithesis_instrumentation__.Notify(128665)
	}
	__antithesis_instrumentation__.Notify(128653)

	storedFamilyCols := storedFamily.ColumnNames
	expectedFamilyCols := expectedFamily.ColumnNames
	if len(storedFamilyCols) != len(expectedFamilyCols) {
		__antithesis_instrumentation__.Notify(128666)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(128667)
	}
	__antithesis_instrumentation__.Notify(128654)
	for i, storedCol := range storedFamilyCols {
		__antithesis_instrumentation__.Notify(128668)
		if storedCol != expectedFamilyCols[i] {
			__antithesis_instrumentation__.Notify(128669)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(128670)
		}
	}
	__antithesis_instrumentation__.Notify(128655)
	return true, nil
}
