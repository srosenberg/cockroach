package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
)

func WithCandidateCode(err error, code pgcode.Code) error {
	__antithesis_instrumentation__.Notify(560818)
	if err == nil {
		__antithesis_instrumentation__.Notify(560820)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(560821)
	}
	__antithesis_instrumentation__.Notify(560819)

	return &withCandidateCode{cause: err, code: code.String()}
}

func HasCandidateCode(err error) bool {
	__antithesis_instrumentation__.Notify(560822)
	return errors.HasType(err, (*withCandidateCode)(nil))
}

func GetPGCodeInternal(
	err error, computeDefaultCode func(err error) (code pgcode.Code),
) (code pgcode.Code) {
	__antithesis_instrumentation__.Notify(560823)
	code = pgcode.Uncategorized
	if c, ok := err.(*withCandidateCode); ok {
		__antithesis_instrumentation__.Notify(560826)
		code = pgcode.MakeCode(c.code)
	} else {
		__antithesis_instrumentation__.Notify(560827)
		if newCode := computeDefaultCode(err); newCode.String() != "" {
			__antithesis_instrumentation__.Notify(560828)
			code = newCode
		} else {
			__antithesis_instrumentation__.Notify(560829)
		}
	}
	__antithesis_instrumentation__.Notify(560824)

	if c := errors.UnwrapOnce(err); c != nil {
		__antithesis_instrumentation__.Notify(560830)
		innerCode := GetPGCodeInternal(c, computeDefaultCode)
		code = combineCodes(innerCode, code)
	} else {
		__antithesis_instrumentation__.Notify(560831)
	}
	__antithesis_instrumentation__.Notify(560825)

	return code
}

func ComputeDefaultCode(err error) pgcode.Code {
	__antithesis_instrumentation__.Notify(560832)
	switch e := err.(type) {

	case *Error:
		__antithesis_instrumentation__.Notify(560836)
		return pgcode.MakeCode(e.Code)

	case ClientVisibleRetryError:
		__antithesis_instrumentation__.Notify(560837)
		return pgcode.SerializationFailure
	case ClientVisibleAmbiguousError:
		__antithesis_instrumentation__.Notify(560838)
		return pgcode.StatementCompletionUnknown
	}
	__antithesis_instrumentation__.Notify(560833)

	if errors.IsAssertionFailure(err) {
		__antithesis_instrumentation__.Notify(560839)
		return pgcode.Internal
	} else {
		__antithesis_instrumentation__.Notify(560840)
	}
	__antithesis_instrumentation__.Notify(560834)
	if errors.IsUnimplementedError(err) {
		__antithesis_instrumentation__.Notify(560841)
		return pgcode.FeatureNotSupported
	} else {
		__antithesis_instrumentation__.Notify(560842)
	}
	__antithesis_instrumentation__.Notify(560835)
	return pgcode.Code{}
}

type ClientVisibleRetryError interface {
	ClientVisibleRetryError()
}

type ClientVisibleAmbiguousError interface {
	ClientVisibleAmbiguousError()
}

func combineCodes(innerCode, outerCode pgcode.Code) pgcode.Code {
	__antithesis_instrumentation__.Notify(560843)
	if outerCode == pgcode.Uncategorized {
		__antithesis_instrumentation__.Notify(560847)
		return innerCode
	} else {
		__antithesis_instrumentation__.Notify(560848)
	}
	__antithesis_instrumentation__.Notify(560844)
	if strings.HasPrefix(outerCode.String(), "XX") {
		__antithesis_instrumentation__.Notify(560849)
		return outerCode
	} else {
		__antithesis_instrumentation__.Notify(560850)
	}
	__antithesis_instrumentation__.Notify(560845)
	if innerCode != pgcode.Uncategorized {
		__antithesis_instrumentation__.Notify(560851)
		return innerCode
	} else {
		__antithesis_instrumentation__.Notify(560852)
	}
	__antithesis_instrumentation__.Notify(560846)
	return outerCode
}
