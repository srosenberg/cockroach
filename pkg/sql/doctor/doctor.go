// Package doctor provides utilities for checking the consistency of cockroach
// internal persisted metadata.
package doctor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type DescriptorTableRow struct {
	ID        int64
	DescBytes []byte
	ModTime   hlc.Timestamp
}

type DescriptorTable []DescriptorTableRow

type NamespaceTableRow struct {
	descpb.NameInfo
	ID int64
}

var _ catalog.NameEntry = (*NamespaceTableRow)(nil)

func (nsr *NamespaceTableRow) GetID() descpb.ID {
	__antithesis_instrumentation__.Notify(468577)
	return descpb.ID(nsr.ID)
}

type NamespaceTable []NamespaceTableRow

type JobsTable []jobs.JobMetadata

func (jt JobsTable) GetJobMetadata(jobID jobspb.JobID) (*jobs.JobMetadata, error) {
	__antithesis_instrumentation__.Notify(468578)
	for i := range jt {
		__antithesis_instrumentation__.Notify(468580)
		md := &jt[i]
		if md.ID == jobID {
			__antithesis_instrumentation__.Notify(468581)
			return md, nil
		} else {
			__antithesis_instrumentation__.Notify(468582)
		}
	}
	__antithesis_instrumentation__.Notify(468579)
	return nil, errors.Newf("job %d not found", jobID)
}

func processDescriptorTable(
	stdout io.Writer, descRows []DescriptorTableRow,
) (func(descpb.ID) catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(468583)
	m := make(map[int64]catalog.Descriptor, len(descRows))

	for _, r := range descRows {
		__antithesis_instrumentation__.Notify(468586)
		var d descpb.Descriptor
		if err := protoutil.Unmarshal(r.DescBytes, &d); err != nil {
			__antithesis_instrumentation__.Notify(468588)
			return nil, errors.Wrapf(err, "failed to unmarshal descriptor %d", r.ID)
		} else {
			__antithesis_instrumentation__.Notify(468589)
		}
		__antithesis_instrumentation__.Notify(468587)
		b := descbuilder.NewBuilderWithMVCCTimestamp(&d, r.ModTime)
		if b != nil {
			__antithesis_instrumentation__.Notify(468590)
			if err := b.RunPostDeserializationChanges(); err != nil {
				__antithesis_instrumentation__.Notify(468592)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err, "error during RunPostDeserializationChanges")
			} else {
				__antithesis_instrumentation__.Notify(468593)
			}
			__antithesis_instrumentation__.Notify(468591)
			m[r.ID] = b.BuildImmutable()
		} else {
			__antithesis_instrumentation__.Notify(468594)
		}
	}
	__antithesis_instrumentation__.Notify(468584)

	for _, r := range descRows {
		__antithesis_instrumentation__.Notify(468595)
		desc := m[r.ID]
		if desc == nil {
			__antithesis_instrumentation__.Notify(468599)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468600)
		}
		__antithesis_instrumentation__.Notify(468596)
		b := desc.NewBuilder()
		if err := b.RunPostDeserializationChanges(); err != nil {
			__antithesis_instrumentation__.Notify(468601)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err, "error during RunPostDeserializationChanges")
		} else {
			__antithesis_instrumentation__.Notify(468602)
		}
		__antithesis_instrumentation__.Notify(468597)
		if err := b.RunRestoreChanges(func(id descpb.ID) catalog.Descriptor {
			__antithesis_instrumentation__.Notify(468603)
			return m[int64(id)]
		}); err != nil {
			__antithesis_instrumentation__.Notify(468604)
			descReport(stdout, desc, "failed to upgrade descriptor: %v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468605)
		}
		__antithesis_instrumentation__.Notify(468598)
		m[r.ID] = b.BuildImmutable()
	}
	__antithesis_instrumentation__.Notify(468585)
	return func(id descpb.ID) catalog.Descriptor {
		__antithesis_instrumentation__.Notify(468606)
		return m[int64(id)]
	}, nil
}

func Examine(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	descTable DescriptorTable,
	namespaceTable NamespaceTable,
	jobsTable JobsTable,
	verbose bool,
	stdout io.Writer,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(468607)
	descOk, err := ExamineDescriptors(
		ctx,
		version,
		descTable,
		namespaceTable,
		jobsTable,
		verbose,
		stdout)
	if err != nil {
		__antithesis_instrumentation__.Notify(468610)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468611)
	}
	__antithesis_instrumentation__.Notify(468608)
	jobsOk, err := ExamineJobs(ctx, descTable, jobsTable, verbose, stdout)
	if err != nil {
		__antithesis_instrumentation__.Notify(468612)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468613)
	}
	__antithesis_instrumentation__.Notify(468609)
	return descOk && func() bool {
		__antithesis_instrumentation__.Notify(468614)
		return jobsOk == true
	}() == true, nil
}

func ExamineDescriptors(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	descTable DescriptorTable,
	namespaceTable NamespaceTable,
	jobsTable JobsTable,
	verbose bool,
	stdout io.Writer,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(468615)
	fmt.Fprintf(
		stdout, "Examining %d descriptors and %d namespace entries...\n",
		len(descTable), len(namespaceTable))
	descLookupFn, err := processDescriptorTable(stdout, descTable)
	if err != nil {
		__antithesis_instrumentation__.Notify(468621)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468622)
	}
	__antithesis_instrumentation__.Notify(468616)
	var problemsFound bool
	var cb nstree.MutableCatalog
	for _, row := range namespaceTable {
		__antithesis_instrumentation__.Notify(468623)
		cb.UpsertNamespaceEntry(row.NameInfo, descpb.ID(row.ID))
	}
	__antithesis_instrumentation__.Notify(468617)
	for _, row := range descTable {
		__antithesis_instrumentation__.Notify(468624)
		id := descpb.ID(row.ID)
		desc := descLookupFn(id)
		if desc == nil {
			__antithesis_instrumentation__.Notify(468627)

			log.Fatalf(ctx, "Descriptor ID %d not found", row.ID)
		} else {
			__antithesis_instrumentation__.Notify(468628)
		}
		__antithesis_instrumentation__.Notify(468625)
		if desc.GetID() != id {
			__antithesis_instrumentation__.Notify(468629)
			problemsFound = true
			descReport(stdout, desc, "different id in descriptor table: %d", row.ID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468630)
		}
		__antithesis_instrumentation__.Notify(468626)
		cb.UpsertDescriptorEntry(desc)
	}
	__antithesis_instrumentation__.Notify(468618)
	for _, row := range descTable {
		__antithesis_instrumentation__.Notify(468631)
		id := descpb.ID(row.ID)
		desc := descLookupFn(id)
		ve := cb.ValidateWithRecover(ctx, version, desc)
		for _, err := range ve {
			__antithesis_instrumentation__.Notify(468634)
			problemsFound = true
			descReport(stdout, desc, "%s", err)
		}
		__antithesis_instrumentation__.Notify(468632)

		jobs.ValidateJobReferencesInDescriptor(desc, jobsTable, func(err error) {
			__antithesis_instrumentation__.Notify(468635)
			problemsFound = true
			descReport(stdout, desc, "%s", err)
		})
		__antithesis_instrumentation__.Notify(468633)

		if verbose {
			__antithesis_instrumentation__.Notify(468636)
			descReport(stdout, desc, "processed")
		} else {
			__antithesis_instrumentation__.Notify(468637)
		}
	}
	__antithesis_instrumentation__.Notify(468619)
	for _, row := range namespaceTable {
		__antithesis_instrumentation__.Notify(468638)
		err := cb.ValidateNamespaceEntry(row)
		if err != nil {
			__antithesis_instrumentation__.Notify(468639)
			problemsFound = true
			nsReport(stdout, row, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(468640)
			if verbose {
				__antithesis_instrumentation__.Notify(468641)
				nsReport(stdout, row, "processed")
			} else {
				__antithesis_instrumentation__.Notify(468642)
			}
		}
	}
	__antithesis_instrumentation__.Notify(468620)

	return !problemsFound, err
}

func ExamineJobs(
	ctx context.Context,
	descTable DescriptorTable,
	jobsTable JobsTable,
	verbose bool,
	stdout io.Writer,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(468643)
	fmt.Fprintf(stdout, "Examining %d jobs...\n", len(jobsTable))
	descLookupFn, err := processDescriptorTable(stdout, descTable)
	if err != nil {
		__antithesis_instrumentation__.Notify(468646)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468647)
	}
	__antithesis_instrumentation__.Notify(468644)
	problemsFound := false
	for _, j := range jobsTable {
		__antithesis_instrumentation__.Notify(468648)
		if verbose {
			__antithesis_instrumentation__.Notify(468650)
			fmt.Fprintf(stdout, "Processing job %d\n", j.ID)
		} else {
			__antithesis_instrumentation__.Notify(468651)
		}
		__antithesis_instrumentation__.Notify(468649)
		jobs.ValidateDescriptorReferencesInJob(j, descLookupFn, func(err error) {
			__antithesis_instrumentation__.Notify(468652)
			problemsFound = true
			fmt.Fprintf(stdout, "job %d: %s.\n", j.ID, err)
		})
	}
	__antithesis_instrumentation__.Notify(468645)
	return !problemsFound, nil
}

func nsReport(stdout io.Writer, row NamespaceTableRow, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(468653)
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(stdout, "  ParentID %3d, ParentSchemaID %2d: namespace entry %q (%d): %s\n",
		row.ParentID, row.ParentSchemaID, row.Name, row.ID, msg)
}

func descReport(stdout io.Writer, desc catalog.Descriptor, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(468654)
	msg := fmt.Sprintf(format, args...)

	msgPrefix := fmt.Sprintf("%s %q (%d): ", desc.DescriptorType(), desc.GetName(), desc.GetID())
	if strings.HasPrefix(msg, msgPrefix) {
		__antithesis_instrumentation__.Notify(468656)
		msgPrefix = ""
	} else {
		__antithesis_instrumentation__.Notify(468657)
	}
	__antithesis_instrumentation__.Notify(468655)
	_, _ = fmt.Fprintf(stdout, "  ParentID %3d, ParentSchemaID %2d: %s%s\n",
		desc.GetParentID(), desc.GetParentSchemaID(), msgPrefix, msg)
}

func DumpSQL(out io.Writer, descTable DescriptorTable, namespaceTable NamespaceTable) error {
	__antithesis_instrumentation__.Notify(468658)

	ms := bootstrap.MakeMetadataSchema(keys.SystemSQLCodec, zonepb.DefaultZoneConfigRef(), zonepb.DefaultSystemZoneConfigRef())
	minUserDescID := ms.FirstNonSystemDescriptorID()
	minUserCreatedDescID := minUserDescID + descpb.ID(len(catalogkeys.DefaultUserDBs))*2

	fmt.Fprintln(out, `BEGIN;`)

	fmt.Fprintf(out,
		"SELECT 1/(1-sign(count(*))) FROM system.descriptor WHERE id >= %d;\n",
		minUserCreatedDescID)

	fmt.Fprintf(out,
		"SELECT crdb_internal.unsafe_delete_descriptor(id, true) FROM system.descriptor WHERE id >= %d;\n",
		minUserDescID)

	fmt.Fprintf(out,
		"SELECT crdb_internal.unsafe_delete_namespace_entry(\"parentID\", \"parentSchemaID\", name, id) "+
			"FROM system.namespace WHERE id >= %d;\n",
		minUserDescID)
	fmt.Fprintln(out, `COMMIT;`)

	fmt.Fprintln(out, `BEGIN;`)
	reverseNamespace := make(map[int64][]NamespaceTableRow, len(descTable))
	for _, row := range namespaceTable {
		__antithesis_instrumentation__.Notify(468662)
		reverseNamespace[row.ID] = append(reverseNamespace[row.ID], row)
	}
	__antithesis_instrumentation__.Notify(468659)
	for _, descRow := range descTable {
		__antithesis_instrumentation__.Notify(468663)

		updatedDescBytes, err := descriptorModifiedForInsert(descRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(468666)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468667)
		}
		__antithesis_instrumentation__.Notify(468664)
		if updatedDescBytes == nil {
			__antithesis_instrumentation__.Notify(468668)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468669)
		}
		__antithesis_instrumentation__.Notify(468665)
		fmt.Fprintf(out,
			"SELECT crdb_internal.unsafe_upsert_descriptor(%d, decode('%s', 'hex'), true);\n",
			descRow.ID, hex.EncodeToString(updatedDescBytes))
		for _, namespaceRow := range reverseNamespace[descRow.ID] {
			__antithesis_instrumentation__.Notify(468670)
			fmt.Fprintf(out,
				"SELECT crdb_internal.unsafe_upsert_namespace_entry(%d, %d, '%s', %d, true);\n",
				namespaceRow.ParentID, namespaceRow.ParentSchemaID, namespaceRow.Name, namespaceRow.ID)
		}
	}
	__antithesis_instrumentation__.Notify(468660)

	for _, namespaceRow := range namespaceTable {
		__antithesis_instrumentation__.Notify(468671)
		if namespaceRow.ID == keys.SystemDatabaseID || func() bool {
			__antithesis_instrumentation__.Notify(468674)
			return namespaceRow.ParentID == keys.SystemDatabaseID == true
		}() == true {
			__antithesis_instrumentation__.Notify(468675)

			continue
		} else {
			__antithesis_instrumentation__.Notify(468676)
		}
		__antithesis_instrumentation__.Notify(468672)
		if _, found := reverseNamespace[namespaceRow.ID]; found {
			__antithesis_instrumentation__.Notify(468677)

			continue
		} else {
			__antithesis_instrumentation__.Notify(468678)
		}
		__antithesis_instrumentation__.Notify(468673)
		fmt.Fprintf(out,
			"SELECT crdb_internal.unsafe_upsert_namespace_entry(%d, %d, '%s', %d, true);\n",
			namespaceRow.ParentID, namespaceRow.ParentSchemaID, namespaceRow.Name, namespaceRow.ID)
	}
	__antithesis_instrumentation__.Notify(468661)
	fmt.Fprintln(out, `COMMIT;`)
	return nil
}

func descriptorModifiedForInsert(r DescriptorTableRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(468679)
	var descProto descpb.Descriptor
	if err := protoutil.Unmarshal(r.DescBytes, &descProto); err != nil {
		__antithesis_instrumentation__.Notify(468684)
		return nil, errors.Wrapf(err, "failed to unmarshal descriptor %d", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(468685)
	}
	__antithesis_instrumentation__.Notify(468680)
	b := descbuilder.NewBuilderWithMVCCTimestamp(&descProto, r.ModTime)
	if b == nil {
		__antithesis_instrumentation__.Notify(468686)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(468687)
	}
	__antithesis_instrumentation__.Notify(468681)
	mut := b.BuildCreatedMutable()
	if catalog.IsSystemDescriptor(mut) {
		__antithesis_instrumentation__.Notify(468688)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(468689)
	}
	__antithesis_instrumentation__.Notify(468682)
	switch d := mut.(type) {
	case catalog.DatabaseDescriptor:
		__antithesis_instrumentation__.Notify(468690)
		d.DatabaseDesc().ModificationTime = hlc.Timestamp{}
		d.DatabaseDesc().Version = 1
	case catalog.SchemaDescriptor:
		__antithesis_instrumentation__.Notify(468691)
		d.SchemaDesc().ModificationTime = hlc.Timestamp{}
		d.SchemaDesc().Version = 1
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(468692)
		d.TypeDesc().ModificationTime = hlc.Timestamp{}
		d.TypeDesc().Version = 1
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(468693)
		d.TableDesc().ModificationTime = hlc.Timestamp{}
		d.TableDesc().CreateAsOfTime = hlc.Timestamp{}
		d.TableDesc().Version = 1
	default:
		__antithesis_instrumentation__.Notify(468694)
		return nil, nil
	}
	__antithesis_instrumentation__.Notify(468683)
	return protoutil.Marshal(mut.DescriptorProto())
}
