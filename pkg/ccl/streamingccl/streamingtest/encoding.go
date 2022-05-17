package streamingtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/stretchr/testify/require"
)

func EncodeKV(
	t *testing.T, codec keys.SQLCodec, descr catalog.TableDescriptor, pkeyVals ...interface{},
) roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(25619)
	require.Equal(t, 1, descr.NumFamilies(), "there can be only one")
	primary := descr.GetPrimaryIndex()
	require.LessOrEqual(t, primary.NumKeyColumns(), len(pkeyVals))

	var datums tree.Datums
	var colMap catalog.TableColMap
	for i, val := range pkeyVals {
		__antithesis_instrumentation__.Notify(25621)
		datums = append(datums, nativeToDatum(t, val))
		col, err := descr.FindColumnWithID(descpb.ColumnID(i + 1))
		require.NoError(t, err)
		colMap.Set(col.GetID(), col.Ordinal())
	}
	__antithesis_instrumentation__.Notify(25620)

	const includeEmpty = true
	indexEntries, err := rowenc.EncodePrimaryIndex(codec, descr, primary,
		colMap, datums, includeEmpty)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexEntries))
	indexEntries[0].Value.InitChecksum(indexEntries[0].Key)
	return roachpb.KeyValue{Key: indexEntries[0].Key, Value: indexEntries[0].Value}
}

func nativeToDatum(t *testing.T, native interface{}) tree.Datum {
	__antithesis_instrumentation__.Notify(25622)
	t.Helper()
	switch v := native.(type) {
	case bool:
		__antithesis_instrumentation__.Notify(25623)
		return tree.MakeDBool(tree.DBool(v))
	case int:
		__antithesis_instrumentation__.Notify(25624)
		return tree.NewDInt(tree.DInt(v))
	case string:
		__antithesis_instrumentation__.Notify(25625)
		return tree.NewDString(v)
	case nil:
		__antithesis_instrumentation__.Notify(25626)
		return tree.DNull
	case tree.Datum:
		__antithesis_instrumentation__.Notify(25627)
		return v
	default:
		__antithesis_instrumentation__.Notify(25628)
		t.Fatalf("unexpected value type %T", v)
		return nil
	}
}
