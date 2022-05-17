package jobstest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func CheckEmittedEvents(
	t *testing.T,
	expectedStatus []string,
	startTime int64,
	jobID int64,
	expectedMessage, expectedJobType string,
) {
	__antithesis_instrumentation__.Notify(84157)

	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(84158)
		log.Flush()
		entries, err := log.FetchEntriesFromFiles(startTime,
			math.MaxInt64, 10000, cmLogRe, log.WithMarkedSensitiveData)
		if err != nil {
			__antithesis_instrumentation__.Notify(84162)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(84163)
		}
		__antithesis_instrumentation__.Notify(84159)
		foundEntry := false
		var matchingEntryIndex int
		for _, e := range entries {
			__antithesis_instrumentation__.Notify(84164)
			if !strings.Contains(e.Message, expectedMessage) {
				__antithesis_instrumentation__.Notify(84168)
				continue
			} else {
				__antithesis_instrumentation__.Notify(84169)
			}
			__antithesis_instrumentation__.Notify(84165)
			foundEntry = true

			e.Message = strings.TrimPrefix(e.Message, "Structured entry:")

			e.Message = strings.TrimPrefix(e.Message, "=")
			jsonPayload := []byte(e.Message)
			var ev eventpb.CommonJobEventDetails
			if err := json.Unmarshal(jsonPayload, &ev); err != nil {
				__antithesis_instrumentation__.Notify(84170)
				t.Errorf("unmarshalling %q: %v", e.Message, err)
			} else {
				__antithesis_instrumentation__.Notify(84171)
			}
			__antithesis_instrumentation__.Notify(84166)
			require.Equal(t, expectedJobType, ev.JobType)
			require.Equal(t, jobID, ev.JobID)
			if matchingEntryIndex >= len(expectedStatus) {
				__antithesis_instrumentation__.Notify(84172)
				return errors.New("more events fround in log than expected")
			} else {
				__antithesis_instrumentation__.Notify(84173)
			}
			__antithesis_instrumentation__.Notify(84167)
			require.Equal(t, expectedStatus[matchingEntryIndex], ev.Status)
			matchingEntryIndex++
		}
		__antithesis_instrumentation__.Notify(84160)
		if !foundEntry {
			__antithesis_instrumentation__.Notify(84174)
			return errors.New("structured entry for import not found in log")
		} else {
			__antithesis_instrumentation__.Notify(84175)
		}
		__antithesis_instrumentation__.Notify(84161)
		return nil
	})
}

var cmLogRe = regexp.MustCompile(`event_log\.go`)
