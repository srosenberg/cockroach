package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

func NewInternalTrackingError(issue int, detail string) error {
	__antithesis_instrumentation__.Notify(560817)
	key := fmt.Sprintf("#%d.%s", issue, detail)
	err := errors.AssertionFailedWithDepthf(1, "%s", errors.Safe(key))
	err = errors.WithTelemetry(err, key)
	err = errors.WithIssueLink(err, errors.IssueLink{IssueURL: fmt.Sprintf("https://github.com/cockroachdb/cockroach/issues/%d", issue)})
	return err
}
