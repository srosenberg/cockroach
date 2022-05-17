package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Status struct{}

func (s *Status) WithDetails() (*Status, error) {
	__antithesis_instrumentation__.Notify(644785)
	return s, nil
}
