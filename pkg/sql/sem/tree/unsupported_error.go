package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

var _ error = &UnsupportedError{}

type UnsupportedError struct {
	Err         error
	FeatureName string
}

func (u *UnsupportedError) Error() string {
	__antithesis_instrumentation__.Notify(615869)
	return u.Err.Error()
}

func (u *UnsupportedError) Cause() error { __antithesis_instrumentation__.Notify(615870); return u.Err }

func (u *UnsupportedError) Unwrap() error {
	__antithesis_instrumentation__.Notify(615871)
	return u.Err
}
