package jobutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func WaitForJobToSucceed(t testing.TB, db *sqlutils.SQLRunner, jobID jobspb.JobID) {
	__antithesis_instrumentation__.Notify(644325)
	t.Helper()
	waitForJobToHaveStatus(t, db, jobID, jobs.StatusSucceeded)
}

func WaitForJobToPause(t testing.TB, db *sqlutils.SQLRunner, jobID jobspb.JobID) {
	__antithesis_instrumentation__.Notify(644326)
	t.Helper()
	waitForJobToHaveStatus(t, db, jobID, jobs.StatusPaused)
}

func WaitForJobToCancel(t testing.TB, db *sqlutils.SQLRunner, jobID jobspb.JobID) {
	__antithesis_instrumentation__.Notify(644327)
	t.Helper()
	waitForJobToHaveStatus(t, db, jobID, jobs.StatusCanceled)
}

func waitForJobToHaveStatus(
	t testing.TB, db *sqlutils.SQLRunner, jobID jobspb.JobID, expectedStatus jobs.Status,
) {
	__antithesis_instrumentation__.Notify(644328)
	t.Helper()
	if err := retry.ForDuration(time.Minute*2, func() error {
		__antithesis_instrumentation__.Notify(644329)
		var status string
		var payloadBytes []byte
		db.QueryRow(
			t, `SELECT status, payload FROM system.jobs WHERE id = $1`, jobID,
		).Scan(&status, &payloadBytes)
		if jobs.Status(status) == jobs.StatusFailed {
			__antithesis_instrumentation__.Notify(644332)
			payload := &jobspb.Payload{}
			if err := protoutil.Unmarshal(payloadBytes, payload); err == nil {
				__antithesis_instrumentation__.Notify(644334)
				t.Fatalf("job failed: %s", payload.Error)
			} else {
				__antithesis_instrumentation__.Notify(644335)
			}
			__antithesis_instrumentation__.Notify(644333)
			t.Fatalf("job failed")
		} else {
			__antithesis_instrumentation__.Notify(644336)
		}
		__antithesis_instrumentation__.Notify(644330)
		if e, a := expectedStatus, jobs.Status(status); e != a {
			__antithesis_instrumentation__.Notify(644337)
			return errors.Errorf("expected job status %s, but got %s", e, a)
		} else {
			__antithesis_instrumentation__.Notify(644338)
		}
		__antithesis_instrumentation__.Notify(644331)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(644339)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(644340)
	}
}

func RunJob(
	t *testing.T,
	db *sqlutils.SQLRunner,
	allowProgressIota *chan struct{},
	ops []string,
	query string,
	args ...interface{},
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(644341)
	*allowProgressIota = make(chan struct{})
	errCh := make(chan error)
	go func() {
		__antithesis_instrumentation__.Notify(644345)
		_, err := db.DB.ExecContext(context.TODO(), query, args...)
		errCh <- err
	}()
	__antithesis_instrumentation__.Notify(644342)
	select {
	case *allowProgressIota <- struct{}{}:
		__antithesis_instrumentation__.Notify(644346)
	case err := <-errCh:
		__antithesis_instrumentation__.Notify(644347)
		return 0, errors.Wrapf(err, "query returned before expected: %s", query)
	}
	__antithesis_instrumentation__.Notify(644343)
	var jobID jobspb.JobID
	db.QueryRow(t, `SELECT id FROM system.jobs ORDER BY created DESC LIMIT 1`).Scan(&jobID)
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(644348)
		db.Exec(t, fmt.Sprintf("%s JOB %d", op, jobID))
		*allowProgressIota <- struct{}{}
	}
	__antithesis_instrumentation__.Notify(644344)
	close(*allowProgressIota)
	return jobID, <-errCh
}

func BulkOpResponseFilter(allowProgressIota *chan struct{}) kvserverbase.ReplicaResponseFilter {
	__antithesis_instrumentation__.Notify(644349)
	return func(_ context.Context, ba roachpb.BatchRequest, br *roachpb.BatchResponse) *roachpb.Error {
		__antithesis_instrumentation__.Notify(644350)
		for _, ru := range br.Responses {
			__antithesis_instrumentation__.Notify(644352)
			switch ru.GetInner().(type) {
			case *roachpb.ExportResponse, *roachpb.AddSSTableResponse:
				__antithesis_instrumentation__.Notify(644353)
				<-*allowProgressIota
			}
		}
		__antithesis_instrumentation__.Notify(644351)
		return nil
	}
}

type logT struct{ testing.TB }

func (n logT) Errorf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(644354)
	n.Logf(format, args...)
}
func (n logT) FailNow() { __antithesis_instrumentation__.Notify(644355) }

func verifySystemJob(
	t testing.TB,
	db *sqlutils.SQLRunner,
	offset int,
	filterType jobspb.Type,
	expectedStatus string,
	expectedRunningStatus string,
	expected jobs.Record,
) error {
	__antithesis_instrumentation__.Notify(644356)
	var actual jobs.Record
	var rawDescriptorIDs pq.Int64Array
	var statusString string
	var runningStatus gosql.NullString
	var runningStatusString string
	var usernameString string

	db.QueryRow(t, `
		SELECT description, user_name, descriptor_ids, status, running_status
		FROM crdb_internal.jobs WHERE job_type = $1 ORDER BY created LIMIT 1 OFFSET $2`,
		filterType.String(),
		offset,
	).Scan(
		&actual.Description, &usernameString, &rawDescriptorIDs,
		&statusString, &runningStatus,
	)
	actual.Username = security.MakeSQLUsernameFromPreNormalizedString(usernameString)
	if runningStatus.Valid {
		__antithesis_instrumentation__.Notify(644362)
		runningStatusString = runningStatus.String
	} else {
		__antithesis_instrumentation__.Notify(644363)
	}
	__antithesis_instrumentation__.Notify(644357)

	for _, id := range rawDescriptorIDs {
		__antithesis_instrumentation__.Notify(644364)
		actual.DescriptorIDs = append(actual.DescriptorIDs, descpb.ID(id))
	}
	__antithesis_instrumentation__.Notify(644358)
	sort.Sort(actual.DescriptorIDs)
	sort.Sort(expected.DescriptorIDs)
	expected.Details = nil
	if e, a := expected, actual; !assert.Equal(logT{t}, e, a) {
		__antithesis_instrumentation__.Notify(644365)
		return errors.Errorf("job %d did not match:\n%s",
			offset, sqlutils.MatrixToStr(db.QueryStr(t, "SELECT * FROM crdb_internal.jobs")))
	} else {
		__antithesis_instrumentation__.Notify(644366)
	}
	__antithesis_instrumentation__.Notify(644359)
	if expectedStatus != statusString {
		__antithesis_instrumentation__.Notify(644367)
		return errors.Errorf("job %d: expected status %v, got %v", offset, expectedStatus, statusString)
	} else {
		__antithesis_instrumentation__.Notify(644368)
	}
	__antithesis_instrumentation__.Notify(644360)
	if expectedRunningStatus != "" && func() bool {
		__antithesis_instrumentation__.Notify(644369)
		return expectedRunningStatus != runningStatusString == true
	}() == true {
		__antithesis_instrumentation__.Notify(644370)
		return errors.Errorf("job %d: expected running status %v, got %v",
			offset, expectedRunningStatus, runningStatusString)
	} else {
		__antithesis_instrumentation__.Notify(644371)
	}
	__antithesis_instrumentation__.Notify(644361)

	return nil
}

func VerifyRunningSystemJob(
	t testing.TB,
	db *sqlutils.SQLRunner,
	offset int,
	filterType jobspb.Type,
	expectedRunningStatus jobs.RunningStatus,
	expected jobs.Record,
) error {
	__antithesis_instrumentation__.Notify(644372)
	return verifySystemJob(t, db, offset, filterType, "running", string(expectedRunningStatus), expected)
}

func VerifySystemJob(
	t testing.TB,
	db *sqlutils.SQLRunner,
	offset int,
	filterType jobspb.Type,
	expectedStatus jobs.Status,
	expected jobs.Record,
) error {
	__antithesis_instrumentation__.Notify(644373)
	return verifySystemJob(t, db, offset, filterType, string(expectedStatus), "", expected)
}

func GetJobID(t testing.TB, db *sqlutils.SQLRunner, offset int) jobspb.JobID {
	__antithesis_instrumentation__.Notify(644374)
	var jobID jobspb.JobID
	db.QueryRow(t, `
	SELECT job_id FROM crdb_internal.jobs ORDER BY created LIMIT 1 OFFSET $1`, offset,
	).Scan(&jobID)
	return jobID
}

func GetLastJobID(t testing.TB, db *sqlutils.SQLRunner) jobspb.JobID {
	__antithesis_instrumentation__.Notify(644375)
	var jobID jobspb.JobID
	db.QueryRow(
		t, `SELECT id FROM system.jobs ORDER BY created DESC LIMIT 1`,
	).Scan(&jobID)
	return jobID
}

func GetJobProgress(t *testing.T, db *sqlutils.SQLRunner, jobID jobspb.JobID) *jobspb.Progress {
	__antithesis_instrumentation__.Notify(644376)
	ret := &jobspb.Progress{}
	var buf []byte
	db.QueryRow(t, `SELECT progress FROM system.jobs WHERE id = $1`, jobID).Scan(&buf)
	if err := protoutil.Unmarshal(buf, ret); err != nil {
		__antithesis_instrumentation__.Notify(644378)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(644379)
	}
	__antithesis_instrumentation__.Notify(644377)
	return ret
}
