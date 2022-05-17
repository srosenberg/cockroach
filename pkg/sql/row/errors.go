package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type singleKVFetcher struct {
	kvs  [1]roachpb.KeyValue
	done bool
}

func (f *singleKVFetcher) nextBatch(
	ctx context.Context,
) (ok bool, kvs []roachpb.KeyValue, batchResponse []byte, err error) {
	__antithesis_instrumentation__.Notify(567369)
	if f.done {
		__antithesis_instrumentation__.Notify(567371)
		return false, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(567372)
	}
	__antithesis_instrumentation__.Notify(567370)
	f.done = true
	return true, f.kvs[:], nil, nil
}

func ConvertBatchError(ctx context.Context, tableDesc catalog.TableDescriptor, b *kv.Batch) error {
	__antithesis_instrumentation__.Notify(567373)
	origPErr := b.MustPErr()
	switch v := origPErr.GetDetail().(type) {
	case *roachpb.MinTimestampBoundUnsatisfiableError:
		__antithesis_instrumentation__.Notify(567375)
		return pgerror.WithCandidateCode(
			origPErr.GoError(),
			pgcode.UnsatisfiableBoundedStaleness,
		)

	case *roachpb.ConditionFailedError:
		__antithesis_instrumentation__.Notify(567376)
		if origPErr.Index == nil {
			__antithesis_instrumentation__.Notify(567382)
			break
		} else {
			__antithesis_instrumentation__.Notify(567383)
		}
		__antithesis_instrumentation__.Notify(567377)
		j := origPErr.Index.Index
		if j >= int32(len(b.Results)) {
			__antithesis_instrumentation__.Notify(567384)
			return errors.AssertionFailedf("index %d outside of results: %+v", j, b.Results)
		} else {
			__antithesis_instrumentation__.Notify(567385)
		}
		__antithesis_instrumentation__.Notify(567378)
		result := b.Results[j]
		if len(result.Rows) == 0 {
			__antithesis_instrumentation__.Notify(567386)
			break
		} else {
			__antithesis_instrumentation__.Notify(567387)
		}
		__antithesis_instrumentation__.Notify(567379)
		key := result.Rows[0].Key
		return NewUniquenessConstraintViolationError(ctx, tableDesc, key, v.ActualValue)

	case *roachpb.WriteIntentError:
		__antithesis_instrumentation__.Notify(567380)
		key := v.Intents[0].Key
		decodeKeyFn := func() (tableName string, indexName string, colNames []string, values []string, err error) {
			__antithesis_instrumentation__.Notify(567388)
			codec, index, err := decodeKeyCodecAndIndex(tableDesc, key)
			if err != nil {
				__antithesis_instrumentation__.Notify(567391)
				return "", "", nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(567392)
			}
			__antithesis_instrumentation__.Notify(567389)
			var spec descpb.IndexFetchSpec
			if err := rowenc.InitIndexFetchSpec(&spec, codec, tableDesc, index, nil); err != nil {
				__antithesis_instrumentation__.Notify(567393)
				return "", "", nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(567394)
			}
			__antithesis_instrumentation__.Notify(567390)

			colNames, values, err = decodeKeyValsUsingSpec(&spec, key)
			return spec.TableName, spec.IndexName, colNames, values, err
		}
		__antithesis_instrumentation__.Notify(567381)
		return newLockNotAvailableError(v.Reason, decodeKeyFn)
	}
	__antithesis_instrumentation__.Notify(567374)
	return origPErr.GoError()
}

func ConvertFetchError(spec *descpb.IndexFetchSpec, err error) error {
	__antithesis_instrumentation__.Notify(567395)
	var errs struct {
		wi *roachpb.WriteIntentError
		bs *roachpb.MinTimestampBoundUnsatisfiableError
	}
	switch {
	case errors.As(err, &errs.wi):
		__antithesis_instrumentation__.Notify(567397)
		key := errs.wi.Intents[0].Key
		decodeKeyFn := func() (tableName string, indexName string, colNames []string, values []string, err error) {
			__antithesis_instrumentation__.Notify(567401)
			colNames, values, err = decodeKeyValsUsingSpec(spec, key)
			return spec.TableName, spec.IndexName, colNames, values, err
		}
		__antithesis_instrumentation__.Notify(567398)
		return newLockNotAvailableError(errs.wi.Reason, decodeKeyFn)

	case errors.As(err, &errs.bs):
		__antithesis_instrumentation__.Notify(567399)
		return pgerror.WithCandidateCode(
			err,
			pgcode.UnsatisfiableBoundedStaleness,
		)
	default:
		__antithesis_instrumentation__.Notify(567400)
	}
	__antithesis_instrumentation__.Notify(567396)
	return err
}

func NewUniquenessConstraintViolationError(
	ctx context.Context, tableDesc catalog.TableDescriptor, key roachpb.Key, value *roachpb.Value,
) error {
	__antithesis_instrumentation__.Notify(567402)
	index, names, values, err := DecodeRowInfo(ctx, tableDesc, key, value, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(567404)
		return pgerror.Wrap(err, pgcode.UniqueViolation,
			"duplicate key value got decoding error")
	} else {
		__antithesis_instrumentation__.Notify(567405)
	}
	__antithesis_instrumentation__.Notify(567403)

	skipCols := index.ExplicitColumnStartIdx()
	return errors.WithDetail(
		pgerror.WithConstraintName(pgerror.Newf(pgcode.UniqueViolation,
			"duplicate key value violates unique constraint %q", index.GetName(),
		), index.GetName()),
		fmt.Sprintf(
			"Key (%s)=(%s) already exists.",
			strings.Join(names[skipCols:], ","),
			strings.Join(values[skipCols:], ","),
		),
	)
}

func decodeKeyValsUsingSpec(
	spec *descpb.IndexFetchSpec, key roachpb.Key,
) (colNames []string, values []string, err error) {
	__antithesis_instrumentation__.Notify(567406)

	keyCols := spec.KeyColumns()
	keyVals := make([]rowenc.EncDatum, len(keyCols))
	if len(key) < int(spec.KeyPrefixLength) {
		__antithesis_instrumentation__.Notify(567410)
		return nil, nil, errors.AssertionFailedf("invalid table key")
	} else {
		__antithesis_instrumentation__.Notify(567411)
	}
	__antithesis_instrumentation__.Notify(567407)
	if _, _, err := rowenc.DecodeKeyValsUsingSpec(keyCols, key[spec.KeyPrefixLength:], keyVals); err != nil {
		__antithesis_instrumentation__.Notify(567412)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567413)
	}
	__antithesis_instrumentation__.Notify(567408)
	colNames = make([]string, len(keyCols))
	values = make([]string, len(keyCols))
	for i := range keyCols {
		__antithesis_instrumentation__.Notify(567414)
		colNames[i] = keyCols[i].Name
		values[i] = keyVals[i].String(keyCols[i].Type)
	}
	__antithesis_instrumentation__.Notify(567409)
	return colNames, values, nil
}

func newLockNotAvailableError(
	reason roachpb.WriteIntentError_Reason,
	decodeKeyFn func() (tableName string, indexName string, colNames []string, values []string, err error),
) error {
	__antithesis_instrumentation__.Notify(567415)
	baseMsg := "could not obtain lock on row"
	if reason == roachpb.WriteIntentError_REASON_LOCK_TIMEOUT {
		__antithesis_instrumentation__.Notify(567418)
		baseMsg = "canceling statement due to lock timeout on row"
	} else {
		__antithesis_instrumentation__.Notify(567419)
	}
	__antithesis_instrumentation__.Notify(567416)
	tableName, indexName, colNames, values, err := decodeKeyFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(567420)
		return pgerror.Wrapf(err, pgcode.LockNotAvailable, "%s: got decoding error", baseMsg)
	} else {
		__antithesis_instrumentation__.Notify(567421)
	}
	__antithesis_instrumentation__.Notify(567417)
	return pgerror.Newf(pgcode.LockNotAvailable,
		"%s (%s)=(%s) in %s@%s",
		baseMsg,
		strings.Join(colNames, ","),
		strings.Join(values, ","),
		tableName,
		indexName,
	)
}

func decodeKeyCodecAndIndex(
	tableDesc catalog.TableDescriptor, key roachpb.Key,
) (keys.SQLCodec, catalog.Index, error) {
	__antithesis_instrumentation__.Notify(567422)
	_, tenantID, err := keys.DecodeTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(567426)
		return keys.SQLCodec{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567427)
	}
	__antithesis_instrumentation__.Notify(567423)
	codec := keys.MakeSQLCodec(tenantID)
	indexID, _, err := rowenc.DecodeIndexKeyPrefix(codec, tableDesc.GetID(), key)
	if err != nil {
		__antithesis_instrumentation__.Notify(567428)
		return keys.SQLCodec{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567429)
	}
	__antithesis_instrumentation__.Notify(567424)
	index, err := tableDesc.FindIndexWithID(indexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(567430)
		return keys.SQLCodec{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567431)
	}
	__antithesis_instrumentation__.Notify(567425)

	return keys.MakeSQLCodec(tenantID), index, nil
}

func DecodeRowInfo(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	key roachpb.Key,
	value *roachpb.Value,
	allColumns bool,
) (_ catalog.Index, columnNames []string, columnValues []string, _ error) {
	__antithesis_instrumentation__.Notify(567432)
	codec, index, err := decodeKeyCodecAndIndex(tableDesc, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(567442)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567443)
	}
	__antithesis_instrumentation__.Notify(567433)
	var rf Fetcher

	var colIDs []descpb.ColumnID
	if !allColumns {
		__antithesis_instrumentation__.Notify(567444)
		colIDs = make([]descpb.ColumnID, index.NumKeyColumns())
		for i := 0; i < index.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(567445)
			colIDs[i] = index.GetKeyColumnID(i)
		}
	} else {
		__antithesis_instrumentation__.Notify(567446)
		if index.Primary() {
			__antithesis_instrumentation__.Notify(567447)
			publicColumns := tableDesc.PublicColumns()
			colIDs = make([]descpb.ColumnID, len(publicColumns))
			for i, col := range publicColumns {
				__antithesis_instrumentation__.Notify(567448)
				colIDs[i] = col.GetID()
			}
		} else {
			__antithesis_instrumentation__.Notify(567449)
			maxNumIDs := index.NumKeyColumns() + index.NumKeySuffixColumns() + index.NumSecondaryStoredColumns()
			colIDs = make([]descpb.ColumnID, 0, maxNumIDs)
			for i := 0; i < index.NumKeyColumns(); i++ {
				__antithesis_instrumentation__.Notify(567452)
				colIDs = append(colIDs, index.GetKeyColumnID(i))
			}
			__antithesis_instrumentation__.Notify(567450)
			for i := 0; i < index.NumKeySuffixColumns(); i++ {
				__antithesis_instrumentation__.Notify(567453)
				colIDs = append(colIDs, index.GetKeySuffixColumnID(i))
			}
			__antithesis_instrumentation__.Notify(567451)
			for i := 0; i < index.NumSecondaryStoredColumns(); i++ {
				__antithesis_instrumentation__.Notify(567454)
				colIDs = append(colIDs, index.GetStoredColumnID(i))
			}
		}
	}
	__antithesis_instrumentation__.Notify(567434)
	cols := make([]catalog.Column, len(colIDs))
	for i, colID := range colIDs {
		__antithesis_instrumentation__.Notify(567455)
		col, err := tableDesc.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(567457)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(567458)
		}
		__antithesis_instrumentation__.Notify(567456)
		cols[i] = col
	}
	__antithesis_instrumentation__.Notify(567435)
	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(&spec, codec, tableDesc, index, colIDs); err != nil {
		__antithesis_instrumentation__.Notify(567459)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567460)
	}
	__antithesis_instrumentation__.Notify(567436)
	rf.IgnoreUnexpectedNulls = true
	if err := rf.Init(
		ctx,
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		&tree.DatumAlloc{},
		nil,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(567461)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567462)
	}
	__antithesis_instrumentation__.Notify(567437)
	f := singleKVFetcher{kvs: [1]roachpb.KeyValue{{Key: key}}}
	if value != nil {
		__antithesis_instrumentation__.Notify(567463)
		f.kvs[0].Value = *value
	} else {
		__antithesis_instrumentation__.Notify(567464)
	}
	__antithesis_instrumentation__.Notify(567438)

	if err := rf.StartScanFrom(ctx, &f, false); err != nil {
		__antithesis_instrumentation__.Notify(567465)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567466)
	}
	__antithesis_instrumentation__.Notify(567439)
	datums, err := rf.NextRowDecoded(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567467)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567468)
	}
	__antithesis_instrumentation__.Notify(567440)

	names := make([]string, len(cols))
	values := make([]string, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(567469)
		if cols[i].IsExpressionIndexColumn() {
			__antithesis_instrumentation__.Notify(567472)
			names[i] = cols[i].GetComputeExpr()
		} else {
			__antithesis_instrumentation__.Notify(567473)
			names[i] = cols[i].GetName()
		}
		__antithesis_instrumentation__.Notify(567470)
		if datums[i] == tree.DNull {
			__antithesis_instrumentation__.Notify(567474)
			continue
		} else {
			__antithesis_instrumentation__.Notify(567475)
		}
		__antithesis_instrumentation__.Notify(567471)
		values[i] = datums[i].String()
	}
	__antithesis_instrumentation__.Notify(567441)
	return index, names, values, nil
}

func (f *singleKVFetcher) close(context.Context) { __antithesis_instrumentation__.Notify(567476) }
