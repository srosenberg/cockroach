package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
)

func NewWithDepthf(depth int, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644575)
	return fmt.Errorf(format, args)
}

func New(code pgcode.Code, msg string) error {
	__antithesis_instrumentation__.Notify(644576)
	return fmt.Errorf(msg)

}

func Newf(code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644577)
	return fmt.Errorf(format, args)
}

func Wrapf(err error, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644578)
	return fmt.Errorf(format, args)
}

func WrapWithDepthf(
	depth int, err error, code pgcode.Code, format string, args ...interface{},
) error {
	__antithesis_instrumentation__.Notify(644579)
	return fmt.Errorf(format, args)

}

func Wrap(err error, code pgcode.Code, msg string) error {
	__antithesis_instrumentation__.Notify(644580)
	return fmt.Errorf(msg)
}
