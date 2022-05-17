package sctestdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func WaitForNoRunningSchemaChanges(t *testing.T, tdb *sqlutils.SQLRunner) {
	__antithesis_instrumentation__.Notify(580804)
	tdb.CheckQueryResultsRetry(t, `
SELECT count(*) 
FROM [SHOW JOBS] 
WHERE job_type = 'SCHEMA CHANGE' 
  AND status NOT IN ('succeeded', 'failed', 'aborted')`,
		[][]string{{"0"}})
}

func ReadDescriptorsFromDB(
	ctx context.Context, t *testing.T, tdb *sqlutils.SQLRunner,
) nstree.MutableCatalog {
	__antithesis_instrumentation__.Notify(580805)
	var cb nstree.MutableCatalog
	hexDescRows := tdb.QueryStr(t, `
SELECT encode(descriptor, 'hex'), crdb_internal_mvcc_timestamp 
FROM system.descriptor 
ORDER BY id`)
	for _, hexDescRow := range hexDescRows {
		__antithesis_instrumentation__.Notify(580807)
		descBytes, err := hex.DecodeString(hexDescRow[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(580814)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580815)
		}
		__antithesis_instrumentation__.Notify(580808)
		ts, err := hlc.ParseTimestamp(hexDescRow[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(580816)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580817)
		}
		__antithesis_instrumentation__.Notify(580809)
		descProto := &descpb.Descriptor{}
		err = protoutil.Unmarshal(descBytes, descProto)
		if err != nil {
			__antithesis_instrumentation__.Notify(580818)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580819)
		}
		__antithesis_instrumentation__.Notify(580810)
		b := descbuilder.NewBuilderWithMVCCTimestamp(descProto, ts)
		if err := b.RunPostDeserializationChanges(); err != nil {
			__antithesis_instrumentation__.Notify(580820)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580821)
		}
		__antithesis_instrumentation__.Notify(580811)
		desc := b.BuildCreatedMutable()
		if desc.GetID() == keys.SystemDatabaseID || func() bool {
			__antithesis_instrumentation__.Notify(580822)
			return desc.GetParentID() == keys.SystemDatabaseID == true
		}() == true {
			__antithesis_instrumentation__.Notify(580823)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580824)
		}
		__antithesis_instrumentation__.Notify(580812)

		switch t := desc.(type) {
		case *dbdesc.Mutable:
			__antithesis_instrumentation__.Notify(580825)
			t.ModificationTime = hlc.Timestamp{}
			t.DefaultPrivileges = nil
		case *schemadesc.Mutable:
			__antithesis_instrumentation__.Notify(580826)
			t.ModificationTime = hlc.Timestamp{}
		case *tabledesc.Mutable:
			__antithesis_instrumentation__.Notify(580827)
			t.TableDescriptor.ModificationTime = hlc.Timestamp{}
			if t.TableDescriptor.CreateAsOfTime != (hlc.Timestamp{}) {
				__antithesis_instrumentation__.Notify(580830)
				t.TableDescriptor.CreateAsOfTime = hlc.Timestamp{WallTime: 1}
			} else {
				__antithesis_instrumentation__.Notify(580831)
			}
			__antithesis_instrumentation__.Notify(580828)
			if t.TableDescriptor.DropTime != 0 {
				__antithesis_instrumentation__.Notify(580832)
				t.TableDescriptor.DropTime = 1
			} else {
				__antithesis_instrumentation__.Notify(580833)
			}
		case *typedesc.Mutable:
			__antithesis_instrumentation__.Notify(580829)
			t.TypeDescriptor.ModificationTime = hlc.Timestamp{}
		}
		__antithesis_instrumentation__.Notify(580813)

		cb.UpsertDescriptorEntry(desc)
	}
	__antithesis_instrumentation__.Notify(580806)
	return cb
}

func ReadNamespaceFromDB(t *testing.T, tdb *sqlutils.SQLRunner) nstree.MutableCatalog {
	__antithesis_instrumentation__.Notify(580834)

	var cb nstree.MutableCatalog
	nsRows := tdb.QueryStr(t, `
SELECT "parentID", "parentSchemaID", name, id 
FROM system.namespace
ORDER BY id`)
	for _, nsRow := range nsRows {
		__antithesis_instrumentation__.Notify(580836)
		parentID, err := strconv.Atoi(nsRow[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(580841)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580842)
		}
		__antithesis_instrumentation__.Notify(580837)
		parentSchemaID, err := strconv.Atoi(nsRow[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(580843)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580844)
		}
		__antithesis_instrumentation__.Notify(580838)
		name := nsRow[2]
		id, err := strconv.Atoi(nsRow[3])
		if err != nil {
			__antithesis_instrumentation__.Notify(580845)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(580846)
		}
		__antithesis_instrumentation__.Notify(580839)
		if id == keys.SystemDatabaseID || func() bool {
			__antithesis_instrumentation__.Notify(580847)
			return parentID == keys.SystemDatabaseID == true
		}() == true {
			__antithesis_instrumentation__.Notify(580848)

			continue
		} else {
			__antithesis_instrumentation__.Notify(580849)
		}
		__antithesis_instrumentation__.Notify(580840)
		key := descpb.NameInfo{
			ParentID:       descpb.ID(parentID),
			ParentSchemaID: descpb.ID(parentSchemaID),
			Name:           name,
		}
		cb.UpsertNamespaceEntry(key, descpb.ID(id))
	}
	__antithesis_instrumentation__.Notify(580835)
	return cb
}

func ReadCurrentDatabaseFromDB(t *testing.T, tdb *sqlutils.SQLRunner) (db string) {
	__antithesis_instrumentation__.Notify(580850)
	tdb.QueryRow(t, `SELECT current_database()`).Scan(&db)
	return db
}

func ReadSessionDataFromDB(
	t *testing.T, tdb *sqlutils.SQLRunner, override func(sd *sessiondata.SessionData),
) (sd sessiondata.SessionData) {
	__antithesis_instrumentation__.Notify(580851)
	hexSessionData := tdb.QueryStr(t, `SELECT encode(crdb_internal.serialize_session(), 'hex')`)
	if len(hexSessionData) == 0 {
		__antithesis_instrumentation__.Notify(580856)
		t.Fatal("Empty session data query results.")
	} else {
		__antithesis_instrumentation__.Notify(580857)
	}
	__antithesis_instrumentation__.Notify(580852)
	sessionDataBytes, err := hex.DecodeString(hexSessionData[0][0])
	if err != nil {
		__antithesis_instrumentation__.Notify(580858)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(580859)
	}
	__antithesis_instrumentation__.Notify(580853)
	sessionDataProto := sessiondatapb.SessionData{}
	err = protoutil.Unmarshal(sessionDataBytes, &sessionDataProto)
	if err != nil {
		__antithesis_instrumentation__.Notify(580860)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(580861)
	}
	__antithesis_instrumentation__.Notify(580854)
	sessionData, err := sessiondata.UnmarshalNonLocal(sessionDataProto)
	if err != nil {
		__antithesis_instrumentation__.Notify(580862)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(580863)
	}
	__antithesis_instrumentation__.Notify(580855)
	sd = *sessionData
	override(&sd)
	return sd
}
