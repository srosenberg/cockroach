package scexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type gcJobs struct {
	dbs     []gcJobForDB
	tables  []gcJobForTable
	indexes []gcJobForIndex
}

type gcJobForTable struct {
	parentID, id descpb.ID
	statement    scop.StatementForDropJob
}

type gcJobForIndex struct {
	tableID   descpb.ID
	indexID   descpb.IndexID
	statement scop.StatementForDropJob
}

type gcJobForDB struct {
	id        descpb.ID
	statement scop.StatementForDropJob
}

func (gj *gcJobs) AddNewGCJobForTable(
	stmt scop.StatementForDropJob, table catalog.TableDescriptor,
) {
	__antithesis_instrumentation__.Notify(581620)
	gj.tables = append(gj.tables, gcJobForTable{
		parentID:  table.GetParentID(),
		id:        table.GetID(),
		statement: stmt,
	})
}

func (gj *gcJobs) AddNewGCJobForDatabase(
	stmt scop.StatementForDropJob, db catalog.DatabaseDescriptor,
) {
	__antithesis_instrumentation__.Notify(581621)
	gj.dbs = append(gj.dbs, gcJobForDB{
		id:        db.GetID(),
		statement: stmt,
	})
}

func (gj *gcJobs) AddNewGCJobForIndex(
	stmt scop.StatementForDropJob, tbl catalog.TableDescriptor, index catalog.Index,
) {
	__antithesis_instrumentation__.Notify(581622)
	gj.indexes = append(gj.indexes, gcJobForIndex{
		tableID:   tbl.GetID(),
		indexID:   index.GetID(),
		statement: stmt,
	})
}

func (gj gcJobs) makeRecords(
	mkJobID func() jobspb.JobID,
) (dbZoneConfigsToRemove catalog.DescriptorIDSet, gcJobRecords []jobs.Record) {
	__antithesis_instrumentation__.Notify(581623)
	type stmts struct {
		s   []scop.StatementForDropJob
		set util.FastIntSet
	}
	addStmt := func(s *stmts, stmt scop.StatementForDropJob) {
		__antithesis_instrumentation__.Notify(581629)
		if id := int(stmt.StatementID); !s.set.Contains(id) {
			__antithesis_instrumentation__.Notify(581630)
			s.set.Add(id)
			s.s = append(s.s, stmt)
		} else {
			__antithesis_instrumentation__.Notify(581631)
		}
	}
	__antithesis_instrumentation__.Notify(581624)
	formatStatements := func(s *stmts) string {
		__antithesis_instrumentation__.Notify(581632)
		sort.Slice(s.s, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(581636)
			return s.s[i].StatementID < s.s[j].StatementID
		})
		__antithesis_instrumentation__.Notify(581633)
		var buf strings.Builder
		if len(s.s) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(581637)
			return s.s[0].Rollback == true
		}() == true {
			__antithesis_instrumentation__.Notify(581638)
			buf.WriteString("ROLLBACK of ")
		} else {
			__antithesis_instrumentation__.Notify(581639)
		}
		__antithesis_instrumentation__.Notify(581634)
		for i, s := range s.s {
			__antithesis_instrumentation__.Notify(581640)
			if i > 0 {
				__antithesis_instrumentation__.Notify(581642)
				buf.WriteString("; ")
			} else {
				__antithesis_instrumentation__.Notify(581643)
			}
			__antithesis_instrumentation__.Notify(581641)
			buf.WriteString(s.Statement)
		}
		__antithesis_instrumentation__.Notify(581635)
		return buf.String()
	}
	__antithesis_instrumentation__.Notify(581625)

	var tablesBeingDropped catalog.DescriptorIDSet
	now := timeutil.Now()
	addTableToJob := func(s *stmts, j *jobspb.SchemaChangeGCDetails, t gcJobForTable) {
		__antithesis_instrumentation__.Notify(581644)
		addStmt(s, t.statement)
		tablesBeingDropped.Add(t.id)
		j.Tables = append(j.Tables, jobspb.SchemaChangeGCDetails_DroppedID{
			ID:       t.id,
			DropTime: now.UnixNano(),
		})
	}
	__antithesis_instrumentation__.Notify(581626)

	gj.sort()
	for _, db := range gj.dbs {
		__antithesis_instrumentation__.Notify(581645)
		i := sort.Search(len(gj.tables), func(i int) bool {
			__antithesis_instrumentation__.Notify(581650)
			return gj.tables[i].parentID >= db.id
		})
		__antithesis_instrumentation__.Notify(581646)
		toDropIsID := func(i int) bool {
			__antithesis_instrumentation__.Notify(581651)
			return i < len(gj.tables) && func() bool {
				__antithesis_instrumentation__.Notify(581652)
				return gj.tables[i].parentID == db.id == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(581647)
		if !toDropIsID(i) {
			__antithesis_instrumentation__.Notify(581653)
			dbZoneConfigsToRemove.Add(db.id)
			continue
		} else {
			__antithesis_instrumentation__.Notify(581654)
		}
		__antithesis_instrumentation__.Notify(581648)
		var s stmts
		var j jobspb.SchemaChangeGCDetails
		j.ParentID = db.id
		for ; toDropIsID(i); i++ {
			__antithesis_instrumentation__.Notify(581655)
			addTableToJob(&s, &j, gj.tables[i])
		}
		__antithesis_instrumentation__.Notify(581649)
		gcJobRecords = append(gcJobRecords,
			createGCJobRecord(mkJobID(), formatStatements(&s), security.NodeUserName(), j))
	}
	{
		__antithesis_instrumentation__.Notify(581656)
		var j jobspb.SchemaChangeGCDetails
		var s stmts
		for _, t := range gj.tables {
			__antithesis_instrumentation__.Notify(581658)
			if tablesBeingDropped.Contains(t.id) {
				__antithesis_instrumentation__.Notify(581660)
				continue
			} else {
				__antithesis_instrumentation__.Notify(581661)
			}
			__antithesis_instrumentation__.Notify(581659)
			addTableToJob(&s, &j, t)
		}
		__antithesis_instrumentation__.Notify(581657)
		if len(j.Tables) > 0 {
			__antithesis_instrumentation__.Notify(581662)
			gcJobRecords = append(gcJobRecords,
				createGCJobRecord(mkJobID(), formatStatements(&s), security.NodeUserName(), j))
		} else {
			__antithesis_instrumentation__.Notify(581663)
		}
	}
	__antithesis_instrumentation__.Notify(581627)
	indexes := gj.indexes
	for len(indexes) > 0 {
		__antithesis_instrumentation__.Notify(581664)
		tableID := indexes[0].tableID
		var s stmts
		var j jobspb.SchemaChangeGCDetails
		j.ParentID = tableID
		for _, idx := range indexes {
			__antithesis_instrumentation__.Notify(581666)
			if idx.tableID != tableID {
				__antithesis_instrumentation__.Notify(581669)
				break
			} else {
				__antithesis_instrumentation__.Notify(581670)
			}
			__antithesis_instrumentation__.Notify(581667)
			indexes = indexes[1:]

			if tablesBeingDropped.Contains(tableID) {
				__antithesis_instrumentation__.Notify(581671)
				continue
			} else {
				__antithesis_instrumentation__.Notify(581672)
			}
			__antithesis_instrumentation__.Notify(581668)
			addStmt(&s, idx.statement)
			j.Indexes = append(j.Indexes, jobspb.SchemaChangeGCDetails_DroppedIndex{
				IndexID:  idx.indexID,
				DropTime: 0,
			})
		}
		__antithesis_instrumentation__.Notify(581665)
		if len(j.Indexes) > 0 {
			__antithesis_instrumentation__.Notify(581673)
			gcJobRecords = append(gcJobRecords,
				createGCJobRecord(mkJobID(), formatStatements(&s), security.NodeUserName(), j))
		} else {
			__antithesis_instrumentation__.Notify(581674)
		}
	}
	__antithesis_instrumentation__.Notify(581628)
	return dbZoneConfigsToRemove, gcJobRecords
}

func (gj gcJobs) sort() {
	__antithesis_instrumentation__.Notify(581675)
	sort.Slice(gj.dbs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581678)
		return gj.dbs[i].id < gj.dbs[j].id
	})
	__antithesis_instrumentation__.Notify(581676)
	sort.Slice(gj.tables, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581679)
		dbi, dbj := gj.tables[i].parentID, gj.tables[j].parentID
		if dbi == dbj {
			__antithesis_instrumentation__.Notify(581681)
			return gj.tables[i].id < gj.tables[j].id
		} else {
			__antithesis_instrumentation__.Notify(581682)
		}
		__antithesis_instrumentation__.Notify(581680)
		return dbi < dbj
	})
	__antithesis_instrumentation__.Notify(581677)
	sort.Slice(gj.indexes, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581683)
		dbi, dbj := gj.indexes[i].tableID, gj.indexes[j].tableID
		if dbi == dbj {
			__antithesis_instrumentation__.Notify(581685)
			return gj.indexes[i].indexID < gj.indexes[j].indexID
		} else {
			__antithesis_instrumentation__.Notify(581686)
		}
		__antithesis_instrumentation__.Notify(581684)
		return dbi < dbj
	})
}

func createGCJobRecord(
	id jobspb.JobID,
	description string,
	username security.SQLUsername,
	details jobspb.SchemaChangeGCDetails,
) jobs.Record {
	__antithesis_instrumentation__.Notify(581687)
	descriptorIDs := make([]descpb.ID, 0)
	if len(details.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(581689)
		if len(descriptorIDs) == 0 {
			__antithesis_instrumentation__.Notify(581690)
			descriptorIDs = []descpb.ID{details.ParentID}
		} else {
			__antithesis_instrumentation__.Notify(581691)
		}
	} else {
		__antithesis_instrumentation__.Notify(581692)
		for _, table := range details.Tables {
			__antithesis_instrumentation__.Notify(581694)
			descriptorIDs = append(descriptorIDs, table.ID)
		}
		__antithesis_instrumentation__.Notify(581693)
		if details.ParentID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(581695)
			descriptorIDs = append(descriptorIDs, details.ParentID)
		} else {
			__antithesis_instrumentation__.Notify(581696)
		}
	}
	__antithesis_instrumentation__.Notify(581688)
	return jobs.Record{
		JobID:         id,
		Description:   "GC for " + description,
		Username:      username,
		DescriptorIDs: descriptorIDs,
		Details:       details,
		Progress:      jobspb.SchemaChangeGCProgress{},
		RunningStatus: "waiting for GC TTL",
		NonCancelable: true,
	}
}
