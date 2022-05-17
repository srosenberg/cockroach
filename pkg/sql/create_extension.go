package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

type createExtensionNode struct {
	CreateExtension tree.CreateExtension
}

func (p *planner) CreateExtension(ctx context.Context, n *tree.CreateExtension) (planNode, error) {
	__antithesis_instrumentation__.Notify(462915)
	return &createExtensionNode{
		CreateExtension: *n,
	}, nil
}

func (n *createExtensionNode) unimplementedExtensionError(issue int) error {
	__antithesis_instrumentation__.Notify(462916)
	name := n.CreateExtension.Name
	return unimplemented.NewWithIssueDetailf(
		issue,
		"CREATE EXTENSION "+name,
		"extension %q is not yet supported",
		name,
	)
}

func (n *createExtensionNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(462917)
	switch n.CreateExtension.Name {
	case "postgis":
		__antithesis_instrumentation__.Notify(462919)
		telemetry.Inc(sqltelemetry.CreateExtensionCounter(n.CreateExtension.Name))
		return nil
	case "postgis_raster",
		"postgis_topology",
		"postgis_sfcgal",
		"fuzzystrmatch",
		"address_standardizer",
		"address_standardizer_data_us",
		"postgis_tiger_geocoder":
		__antithesis_instrumentation__.Notify(462920)

		return n.unimplementedExtensionError(54514)
	case "btree_gin":
		__antithesis_instrumentation__.Notify(462921)
		return n.unimplementedExtensionError(51992)
	case "btree_gist":
		__antithesis_instrumentation__.Notify(462922)
		return n.unimplementedExtensionError(51993)
	case "citext":
		__antithesis_instrumentation__.Notify(462923)
		return n.unimplementedExtensionError(41276)
	case "postgres_fdw":
		__antithesis_instrumentation__.Notify(462924)
		return n.unimplementedExtensionError(20249)
	case "pg_trgm":
		__antithesis_instrumentation__.Notify(462925)
		return n.unimplementedExtensionError(51137)
	case "adminpack",
		"amcheck",
		"auth_delay",
		"auto_explain",
		"bloom",
		"cube",
		"dblink",
		"dict_int",
		"dict_xsyn",
		"earthdistance",
		"file_fdw",
		"hstore",
		"intagg",
		"intarray",
		"isn",
		"lo",
		"ltree",
		"pageinspect",
		"passwordcheck",
		"pg_buffercache",
		"pgcrypto",
		"pg_freespacemap",
		"pg_prewarm",
		"pgrowlocks",
		"pg_stat_statements",
		"pgstattuple",
		"pg_visibility",
		"seg",
		"sepgsql",
		"spi",
		"sslinfo",
		"tablefunc",
		"tcn",
		"test_decoding",
		"tsm_system_rows",
		"tsm_system_time",
		"unaccent",
		"uuid-ossp",
		"xml2":
		__antithesis_instrumentation__.Notify(462926)
		return n.unimplementedExtensionError(54516)
	default:
		__antithesis_instrumentation__.Notify(462927)
	}
	__antithesis_instrumentation__.Notify(462918)
	return pgerror.Newf(pgcode.UndefinedParameter, "unknown extension: %q", n.CreateExtension.Name)
}

func (n *createExtensionNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(462928)
	return false, nil
}
func (n *createExtensionNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(462929)
	return tree.Datums{}
}
func (n *createExtensionNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(462930)
}
