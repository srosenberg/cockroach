// Package floatcmp provides functions for determining float values to be equal
// if they are within a tolerance. It is designed to be used in tests.
package floatcmp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	CloseFraction float64 = 1e-14

	CloseMargin float64 = CloseFraction * CloseFraction
)

func EqualApprox(expected interface{}, actual interface{}, fraction float64, margin float64) bool {
	__antithesis_instrumentation__.Notify(644233)
	return cmp.Equal(expected, actual, cmpopts.EquateApprox(fraction, margin), cmpopts.EquateNaNs())
}
