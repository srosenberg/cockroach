package descbuilder

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func NewBuilderWithMVCCTimestamp(
	desc *descpb.Descriptor, mvccTimestamp hlc.Timestamp,
) catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(251462)
	table, database, typ, schema := descpb.FromDescriptorWithMVCCTimestamp(desc, mvccTimestamp)
	switch {
	case table != nil:
		__antithesis_instrumentation__.Notify(251463)
		return tabledesc.NewBuilder(table)
	case database != nil:
		__antithesis_instrumentation__.Notify(251464)
		return dbdesc.NewBuilder(database)
	case typ != nil:
		__antithesis_instrumentation__.Notify(251465)
		return typedesc.NewBuilder(typ)
	case schema != nil:
		__antithesis_instrumentation__.Notify(251466)
		return schemadesc.NewBuilder(schema)
	default:
		__antithesis_instrumentation__.Notify(251467)
		return nil
	}
}

func NewBuilder(desc *descpb.Descriptor) catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(251468)
	return NewBuilderWithMVCCTimestamp(desc, hlc.Timestamp{})
}

func ValidateSelf(desc catalog.Descriptor, version clusterversion.ClusterVersion) error {
	__antithesis_instrumentation__.Notify(251469)
	return validate.Self(version, desc)
}
