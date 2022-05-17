package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

type ValidationLevel uint32

const (
	NoValidation ValidationLevel = 0

	ValidationLevelSelfOnly = 1<<(iota+1) - 1

	ValidationLevelCrossReferences

	ValidationLevelNamespace

	ValidationLevelAllPreTxnCommit
)

type ValidationTelemetry int

const (
	NoValidationTelemetry ValidationTelemetry = iota

	ValidationReadTelemetry

	ValidationWriteTelemetry
)

type ValidationErrorAccumulator interface {
	Report(err error)

	IsActive(version clusterversion.Key) bool
}

type ValidationErrors []error

func (ve ValidationErrors) CombinedError() error {
	__antithesis_instrumentation__.Notify(272366)
	var combinedErr error
	var extraTelemetryKeys []string
	for i := len(ve) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(272370)
		combinedErr = errors.CombineErrors(ve[i], combinedErr)
	}
	__antithesis_instrumentation__.Notify(272367)

	for _, err := range ve {
		__antithesis_instrumentation__.Notify(272371)
		for _, key := range errors.GetTelemetryKeys(err) {
			__antithesis_instrumentation__.Notify(272372)
			if strings.HasPrefix(key, telemetry.ValidationTelemetryKeyPrefix) {
				__antithesis_instrumentation__.Notify(272373)
				extraTelemetryKeys = append(extraTelemetryKeys, key)
			} else {
				__antithesis_instrumentation__.Notify(272374)
			}
		}
	}
	__antithesis_instrumentation__.Notify(272368)
	if extraTelemetryKeys == nil {
		__antithesis_instrumentation__.Notify(272375)
		return combinedErr
	} else {
		__antithesis_instrumentation__.Notify(272376)
	}
	__antithesis_instrumentation__.Notify(272369)
	return errors.WithTelemetry(combinedErr, extraTelemetryKeys...)
}

type ValidationDescGetter interface {
	GetDatabaseDescriptor(id descpb.ID) (DatabaseDescriptor, error)

	GetSchemaDescriptor(id descpb.ID) (SchemaDescriptor, error)

	GetTableDescriptor(id descpb.ID) (TableDescriptor, error)

	GetTypeDescriptor(id descpb.ID) (TypeDescriptor, error)
}
