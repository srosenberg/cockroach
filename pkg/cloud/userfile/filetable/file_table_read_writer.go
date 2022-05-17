package filetable

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const ChunkDefaultSize = 1024 * 1024 * 4

var fileTableNameSuffix = "_upload_files"
var payloadTableNameSuffix = "_upload_payload"

type FileToTableExecutorRows struct {
	internalExecResultsIterator sqlutil.InternalRows
	sqlConnExecResults          driver.Rows
}

type FileToTableSystemExecutor interface {
	Query(ctx context.Context, opName, query string,
		username security.SQLUsername,
		qargs ...interface{}) (*FileToTableExecutorRows, error)
	Exec(ctx context.Context, opName, query string,
		username security.SQLUsername,
		qargs ...interface{}) error
}

type InternalFileToTableExecutor struct {
	ie sqlutil.InternalExecutor
	db *kv.DB
}

var _ FileToTableSystemExecutor = &InternalFileToTableExecutor{}

func MakeInternalFileToTableExecutor(
	ie sqlutil.InternalExecutor, db *kv.DB,
) *InternalFileToTableExecutor {
	__antithesis_instrumentation__.Notify(36721)
	return &InternalFileToTableExecutor{ie, db}
}

func (i *InternalFileToTableExecutor) Query(
	ctx context.Context, opName, query string, username security.SQLUsername, qargs ...interface{},
) (*FileToTableExecutorRows, error) {
	__antithesis_instrumentation__.Notify(36722)
	result := FileToTableExecutorRows{}
	var err error
	result.internalExecResultsIterator, err = i.ie.QueryIteratorEx(ctx, opName, nil,
		sessiondata.InternalExecutorOverride{User: username}, query, qargs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(36724)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36725)
	}
	__antithesis_instrumentation__.Notify(36723)
	return &result, nil
}

func (i *InternalFileToTableExecutor) Exec(
	ctx context.Context, opName, query string, username security.SQLUsername, qargs ...interface{},
) error {
	__antithesis_instrumentation__.Notify(36726)
	_, err := i.ie.ExecEx(ctx, opName, nil,
		sessiondata.InternalExecutorOverride{User: username}, query, qargs...)
	return err
}

type SQLConnFileToTableExecutor struct {
	executor cloud.SQLConnI
}

var _ FileToTableSystemExecutor = &SQLConnFileToTableExecutor{}

func MakeSQLConnFileToTableExecutor(executor cloud.SQLConnI) *SQLConnFileToTableExecutor {
	__antithesis_instrumentation__.Notify(36727)
	return &SQLConnFileToTableExecutor{executor: executor}
}

func (i *SQLConnFileToTableExecutor) Query(
	ctx context.Context, _, query string, _ security.SQLUsername, qargs ...interface{},
) (*FileToTableExecutorRows, error) {
	__antithesis_instrumentation__.Notify(36728)
	result := FileToTableExecutorRows{}

	argVals := make([]driver.NamedValue, len(qargs))
	for i, qarg := range qargs {
		__antithesis_instrumentation__.Notify(36731)
		namedVal := driver.NamedValue{

			Ordinal: i + 1,
			Value:   qarg,
		}
		argVals[i] = namedVal
	}
	__antithesis_instrumentation__.Notify(36729)

	var err error
	result.sqlConnExecResults, err = i.executor.QueryContext(ctx, query, argVals)
	if err != nil {
		__antithesis_instrumentation__.Notify(36732)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36733)
	}
	__antithesis_instrumentation__.Notify(36730)
	return &result, nil
}

func (i *SQLConnFileToTableExecutor) Exec(
	ctx context.Context, _, query string, _ security.SQLUsername, qargs ...interface{},
) error {
	__antithesis_instrumentation__.Notify(36734)
	argVals := make([]driver.NamedValue, len(qargs))
	for i, qarg := range qargs {
		__antithesis_instrumentation__.Notify(36736)
		namedVal := driver.NamedValue{

			Ordinal: i + 1,
			Value:   qarg,
		}
		argVals[i] = namedVal
	}
	__antithesis_instrumentation__.Notify(36735)
	_, err := i.executor.ExecContext(ctx, query, argVals)
	return err
}

type FileToTableSystem struct {
	qualifiedTableName string
	executor           FileToTableSystemExecutor
	username           security.SQLUsername
}

const fileTableSchema = `CREATE TABLE %s (filename STRING PRIMARY KEY,
file_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
file_size INT NOT NULL,
username STRING NOT NULL,
upload_time TIMESTAMP DEFAULT now())`

const payloadTableSchema = `CREATE TABLE %s (file_id UUID,
byte_offset INT,
payload BYTES,
PRIMARY KEY(file_id, byte_offset))`

func (f *FileToTableSystem) GetFQFileTableName() string {
	__antithesis_instrumentation__.Notify(36737)
	return f.qualifiedTableName + fileTableNameSuffix
}

func (f *FileToTableSystem) GetFQPayloadTableName() string {
	__antithesis_instrumentation__.Notify(36738)
	return f.qualifiedTableName + payloadTableNameSuffix
}

func (f *FileToTableSystem) GetSimpleFileTableName(prefix string) (string, error) {
	__antithesis_instrumentation__.Notify(36739)
	return prefix + fileTableNameSuffix, nil
}

func (f *FileToTableSystem) GetSimplePayloadTableName(prefix string) (string, error) {
	__antithesis_instrumentation__.Notify(36740)
	return prefix + payloadTableNameSuffix, nil
}

func (f *FileToTableSystem) GetDatabaseAndSchema() (string, error) {
	__antithesis_instrumentation__.Notify(36741)
	tableName, err := parser.ParseQualifiedTableName(f.qualifiedTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(36743)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(36744)
	}
	__antithesis_instrumentation__.Notify(36742)

	return tableName.ObjectNamePrefix.String(), nil
}

func (f *FileToTableSystem) GetTableName() (string, error) {
	__antithesis_instrumentation__.Notify(36745)
	tableName, err := parser.ParseQualifiedTableName(f.qualifiedTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(36747)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(36748)
	}
	__antithesis_instrumentation__.Notify(36746)

	return tableName.ObjectName.String(), nil
}

func resolveInternalFileToTableExecutor(
	executor FileToTableSystemExecutor,
) (*InternalFileToTableExecutor, error) {
	__antithesis_instrumentation__.Notify(36749)
	var e *InternalFileToTableExecutor
	var ok bool
	if e, ok = executor.(*InternalFileToTableExecutor); !ok {
		__antithesis_instrumentation__.Notify(36751)
		return nil, errors.Newf("unable to resolve %T to a supported executor type", executor)
	} else {
		__antithesis_instrumentation__.Notify(36752)
	}
	__antithesis_instrumentation__.Notify(36750)

	return e, nil
}

func NewFileToTableSystem(
	ctx context.Context,
	qualifiedTableName string,
	executor FileToTableSystemExecutor,
	username security.SQLUsername,
) (*FileToTableSystem, error) {
	__antithesis_instrumentation__.Notify(36753)

	_, err := parser.ParseQualifiedTableName(qualifiedTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(36758)
		return nil, errors.Wrapf(err, "unable to parse qualified table name %s supplied to userfile",
			qualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(36759)
	}
	__antithesis_instrumentation__.Notify(36754)

	f := FileToTableSystem{
		qualifiedTableName: qualifiedTableName, executor: executor, username: username,
	}

	if _, ok := executor.(*SQLConnFileToTableExecutor); ok {
		__antithesis_instrumentation__.Notify(36760)
		return &f, nil
	} else {
		__antithesis_instrumentation__.Notify(36761)
	}
	__antithesis_instrumentation__.Notify(36755)

	e, err := resolveInternalFileToTableExecutor(executor)
	if err != nil {
		__antithesis_instrumentation__.Notify(36762)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36763)
	}
	__antithesis_instrumentation__.Notify(36756)
	if err := e.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(36764)

		tablesExist, err := f.checkIfFileAndPayloadTableExist(ctx, txn, e.ie)
		if err != nil {
			__antithesis_instrumentation__.Notify(36767)
			return err
		} else {
			__antithesis_instrumentation__.Notify(36768)
		}
		__antithesis_instrumentation__.Notify(36765)

		if !tablesExist {
			__antithesis_instrumentation__.Notify(36769)
			if err := f.createFileAndPayloadTables(ctx, txn, e.ie); err != nil {
				__antithesis_instrumentation__.Notify(36772)
				return err
			} else {
				__antithesis_instrumentation__.Notify(36773)
			}
			__antithesis_instrumentation__.Notify(36770)

			if err := f.grantCurrentUserTablePrivileges(ctx, txn, e.ie); err != nil {
				__antithesis_instrumentation__.Notify(36774)
				return err
			} else {
				__antithesis_instrumentation__.Notify(36775)
			}
			__antithesis_instrumentation__.Notify(36771)

			if err := f.revokeOtherUserTablePrivileges(ctx, txn, e.ie); err != nil {
				__antithesis_instrumentation__.Notify(36776)
				return err
			} else {
				__antithesis_instrumentation__.Notify(36777)
			}
		} else {
			__antithesis_instrumentation__.Notify(36778)
		}
		__antithesis_instrumentation__.Notify(36766)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(36779)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36780)
	}
	__antithesis_instrumentation__.Notify(36757)

	return &f, nil
}

func (f *FileToTableSystem) FileSize(ctx context.Context, filename string) (int64, error) {
	__antithesis_instrumentation__.Notify(36781)
	e, err := resolveInternalFileToTableExecutor(f.executor)
	if err != nil {
		__antithesis_instrumentation__.Notify(36785)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36786)
	}
	__antithesis_instrumentation__.Notify(36782)

	getFileSizeQuery := fmt.Sprintf(`SELECT file_size FROM %s WHERE filename=$1`,
		f.GetFQFileTableName())
	rows, err := e.ie.QueryRowEx(ctx, "payload-table-storage-size", nil,
		sessiondata.InternalExecutorOverride{User: f.username},
		getFileSizeQuery, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36787)
		return 0, errors.Wrap(err, "failed to get size of file from the payload table")
	} else {
		__antithesis_instrumentation__.Notify(36788)
	}
	__antithesis_instrumentation__.Notify(36783)

	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(36789)
		return 0, errors.Newf("file %s does not exist in the UserFileStorage", filename)
	} else {
		__antithesis_instrumentation__.Notify(36790)
	}
	__antithesis_instrumentation__.Notify(36784)

	return int64(tree.MustBeDInt(rows[0])), nil
}

func (f *FileToTableSystem) ListFiles(ctx context.Context, pattern string) ([]string, error) {
	__antithesis_instrumentation__.Notify(36791)
	var files []string
	listFilesQuery := fmt.Sprintf(`SELECT filename FROM %s WHERE filename LIKE $1 ORDER BY
filename`, f.GetFQFileTableName())

	rows, err := f.executor.Query(ctx, "file-table-storage-list", listFilesQuery, f.username,
		pattern+"%")
	if err != nil {
		__antithesis_instrumentation__.Notify(36794)
		return files, errors.Wrap(err, "failed to list files from file table")
	} else {
		__antithesis_instrumentation__.Notify(36795)
	}
	__antithesis_instrumentation__.Notify(36792)

	switch f.executor.(type) {
	case *InternalFileToTableExecutor:
		__antithesis_instrumentation__.Notify(36796)

		it := rows.internalExecResultsIterator
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(36801)
			files = append(files, string(tree.MustBeDString(it.Cur()[0])))
		}
		__antithesis_instrumentation__.Notify(36797)
		if err != nil {
			__antithesis_instrumentation__.Notify(36802)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36803)
		}
	case *SQLConnFileToTableExecutor:
		__antithesis_instrumentation__.Notify(36798)
		vals := make([]driver.Value, 1)
		for {
			__antithesis_instrumentation__.Notify(36804)
			if err := rows.sqlConnExecResults.Next(vals); err == io.EOF {
				__antithesis_instrumentation__.Notify(36806)
				break
			} else {
				__antithesis_instrumentation__.Notify(36807)
				if err != nil {
					__antithesis_instrumentation__.Notify(36808)
					return files, errors.Wrap(err, "failed to list files from file table")
				} else {
					__antithesis_instrumentation__.Notify(36809)
				}
			}
			__antithesis_instrumentation__.Notify(36805)
			filename := vals[0].(string)
			files = append(files, filename)
		}
		__antithesis_instrumentation__.Notify(36799)

		if err = rows.sqlConnExecResults.Close(); err != nil {
			__antithesis_instrumentation__.Notify(36810)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36811)
		}
	default:
		__antithesis_instrumentation__.Notify(36800)
		return []string{}, errors.New("unsupported executor type in FileSize")
	}
	__antithesis_instrumentation__.Notify(36793)

	return files, nil
}

func DestroyUserFileSystem(ctx context.Context, f *FileToTableSystem) error {
	__antithesis_instrumentation__.Notify(36812)
	e, err := resolveInternalFileToTableExecutor(f.executor)
	if err != nil {
		__antithesis_instrumentation__.Notify(36815)
		return err
	} else {
		__antithesis_instrumentation__.Notify(36816)
	}
	__antithesis_instrumentation__.Notify(36813)

	if err := e.db.Txn(ctx,
		func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(36817)
			dropPayloadTableQuery := fmt.Sprintf(`DROP TABLE %s`, f.GetFQPayloadTableName())
			_, err := e.ie.ExecEx(ctx, "drop-payload-table", txn,
				sessiondata.InternalExecutorOverride{User: f.username},
				dropPayloadTableQuery)
			if err != nil {
				__antithesis_instrumentation__.Notify(36820)
				return errors.Wrap(err, "failed to drop payload table")
			} else {
				__antithesis_instrumentation__.Notify(36821)
			}
			__antithesis_instrumentation__.Notify(36818)

			dropFileTableQuery := fmt.Sprintf(`DROP TABLE %s CASCADE`, f.GetFQFileTableName())
			_, err = e.ie.ExecEx(ctx, "drop-file-table", txn,
				sessiondata.InternalExecutorOverride{User: f.username},
				dropFileTableQuery)
			if err != nil {
				__antithesis_instrumentation__.Notify(36822)
				return errors.Wrap(err, "failed to drop file table")
			} else {
				__antithesis_instrumentation__.Notify(36823)
			}
			__antithesis_instrumentation__.Notify(36819)

			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(36824)
		return err
	} else {
		__antithesis_instrumentation__.Notify(36825)
	}
	__antithesis_instrumentation__.Notify(36814)

	return nil
}

func (f *FileToTableSystem) getDeleteQuery() string {
	__antithesis_instrumentation__.Notify(36826)
	deleteFileMetadataQueryPlaceholder := `DELETE FROM %s WHERE filename=$1`
	return fmt.Sprintf(deleteFileMetadataQueryPlaceholder, f.GetFQFileTableName())
}

func (f *FileToTableSystem) getDeletePayloadQuery() string {
	__antithesis_instrumentation__.Notify(36827)
	deletePayloadQueryPlaceholder := `DELETE FROM %s WHERE file_id IN (
SELECT file_id FROM %s WHERE filename=$1)`
	return fmt.Sprintf(deletePayloadQueryPlaceholder, f.GetFQPayloadTableName(),
		f.GetFQFileTableName())
}

func (f *FileToTableSystem) deleteFileWithoutTxn(
	ctx context.Context, filename string, ie sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(36828)
	execSessionDataOverride := sessiondata.InternalExecutorOverride{User: f.username}
	_, err := ie.ExecEx(ctx, "delete-payload-table",
		nil, execSessionDataOverride, f.getDeletePayloadQuery(), filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36831)
		return errors.Wrap(err,
			"failed to delete from the payload table while preparing for overwrite")
	} else {
		__antithesis_instrumentation__.Notify(36832)
	}
	__antithesis_instrumentation__.Notify(36829)

	_, err = ie.ExecEx(ctx, "delete-file-table", nil, execSessionDataOverride,
		f.getDeleteQuery(), filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36833)
		return errors.Wrap(err, "failed to delete from the file table while preparing for overwrite")
	} else {
		__antithesis_instrumentation__.Notify(36834)
	}
	__antithesis_instrumentation__.Notify(36830)

	return nil
}

func (f *FileToTableSystem) DeleteFile(ctx context.Context, filename string) error {
	__antithesis_instrumentation__.Notify(36835)
	defer func() {
		__antithesis_instrumentation__.Notify(36840)
		_ = f.executor.Exec(ctx, "commit", `COMMIT`, f.username)
	}()
	__antithesis_instrumentation__.Notify(36836)

	txnErr := f.executor.Exec(ctx, "delete-file", `BEGIN`, f.username)
	if txnErr != nil {
		__antithesis_instrumentation__.Notify(36841)
		return txnErr
	} else {
		__antithesis_instrumentation__.Notify(36842)
	}
	__antithesis_instrumentation__.Notify(36837)

	txnErr = f.executor.Exec(ctx, "delete-payload-table", f.getDeletePayloadQuery(),
		f.username, filename)
	if txnErr != nil {
		__antithesis_instrumentation__.Notify(36843)
		return errors.Wrap(txnErr, "failed to delete from the payload table")
	} else {
		__antithesis_instrumentation__.Notify(36844)
	}
	__antithesis_instrumentation__.Notify(36838)

	txnErr = f.executor.Exec(ctx, "delete-file-table", f.getDeleteQuery(),
		f.username, filename)
	if txnErr != nil {
		__antithesis_instrumentation__.Notify(36845)
		return errors.Wrap(txnErr, "failed to delete from the file table")
	} else {
		__antithesis_instrumentation__.Notify(36846)
	}
	__antithesis_instrumentation__.Notify(36839)

	return nil
}

type payloadWriter struct {
	fileID                  tree.Datum
	ie                      sqlutil.InternalExecutor
	db                      *kv.DB
	ctx                     context.Context
	byteOffset              int
	execSessionDataOverride sessiondata.InternalExecutorOverride
	fileTableName           string
	payloadTableName        string
}

func (p *payloadWriter) WriteChunk(buf []byte, txn *kv.Txn) (int, error) {
	__antithesis_instrumentation__.Notify(36847)
	insertChunkQuery := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3)`, p.payloadTableName)
	_, err := p.ie.ExecEx(p.ctx, "insert-file-chunk", txn, p.execSessionDataOverride,
		insertChunkQuery, p.fileID, p.byteOffset, buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(36849)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36850)
	}
	__antithesis_instrumentation__.Notify(36848)

	bytesWritten := len(buf)
	p.byteOffset += bytesWritten

	return bytesWritten, nil
}

type chunkWriter struct {
	buf                     *bytes.Buffer
	pw                      *payloadWriter
	execSessionDataOverride sessiondata.InternalExecutorOverride
	fileTableName           string
	payloadTableName        string
	chunkSize               int
	filename                string
}

var _ io.WriteCloser = &chunkWriter{}

func newChunkWriter(
	ctx context.Context,
	chunkSize int,
	filename string,
	username security.SQLUsername,
	fileTableName, payloadTableName string,
	ie sqlutil.InternalExecutor,
	db *kv.DB,
) (*chunkWriter, error) {
	__antithesis_instrumentation__.Notify(36851)
	execSessionDataOverride := sessiondata.InternalExecutorOverride{User: username}

	fileNameQuery := fmt.Sprintf(`INSERT INTO %s VALUES ($1, DEFAULT, $2, $3) RETURNING file_id`,
		fileTableName)

	res, err := ie.QueryRowEx(ctx, "insert-file-name",
		nil, execSessionDataOverride, fileNameQuery, filename, 0,
		execSessionDataOverride.User)
	if err != nil {
		__antithesis_instrumentation__.Notify(36854)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36855)
	}
	__antithesis_instrumentation__.Notify(36852)
	if res == nil {
		__antithesis_instrumentation__.Notify(36856)
		return nil, errors.Newf("no UUID returned for filename %s", filename)
	} else {
		__antithesis_instrumentation__.Notify(36857)
	}
	__antithesis_instrumentation__.Notify(36853)

	pw := &payloadWriter{
		res[0], ie, db, ctx, 0,
		execSessionDataOverride, fileTableName,
		payloadTableName}
	bytesBuffer := bytes.NewBuffer(make([]byte, 0, chunkSize))
	return &chunkWriter{
		bytesBuffer, pw, execSessionDataOverride,
		fileTableName, payloadTableName,
		chunkSize, filename,
	}, nil
}

func (w *chunkWriter) fillAvailableBufferSpace(payload []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(36858)
	available := w.buf.Cap() - w.buf.Len()
	if available > len(payload) {
		__antithesis_instrumentation__.Notify(36861)
		available = len(payload)
	} else {
		__antithesis_instrumentation__.Notify(36862)
	}
	__antithesis_instrumentation__.Notify(36859)
	if _, err := w.buf.Write(payload[:available]); err != nil {
		__antithesis_instrumentation__.Notify(36863)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36864)
	}
	__antithesis_instrumentation__.Notify(36860)
	return payload[available:], nil
}

func (w *chunkWriter) Write(buf []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36865)
	bufLen := len(buf)
	for len(buf) > 0 {
		__antithesis_instrumentation__.Notify(36867)
		var err error
		buf, err = w.fillAvailableBufferSpace(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(36869)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(36870)
		}
		__antithesis_instrumentation__.Notify(36868)

		if w.buf.Len() == w.buf.Cap() {
			__antithesis_instrumentation__.Notify(36871)
			if err := w.pw.db.Txn(w.pw.ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(36873)
				if n, err := w.pw.WriteChunk(w.buf.Bytes(), txn); err != nil {
					__antithesis_instrumentation__.Notify(36875)
					return err
				} else {
					__antithesis_instrumentation__.Notify(36876)
					if n != w.buf.Len() {
						__antithesis_instrumentation__.Notify(36877)
						return errors.Wrap(io.ErrShortWrite, "error when writing in chunkWriter")
					} else {
						__antithesis_instrumentation__.Notify(36878)
					}
				}
				__antithesis_instrumentation__.Notify(36874)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(36879)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(36880)
			}
			__antithesis_instrumentation__.Notify(36872)
			w.buf.Reset()
		} else {
			__antithesis_instrumentation__.Notify(36881)
		}
	}
	__antithesis_instrumentation__.Notify(36866)

	return bufLen, nil
}

func (w *chunkWriter) Close() error {
	__antithesis_instrumentation__.Notify(36882)

	if w.buf.Len() > 0 {
		__antithesis_instrumentation__.Notify(36884)
		if err := w.pw.db.Txn(w.pw.ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(36885)
			if n, err := w.pw.WriteChunk(w.buf.Bytes(), txn); err != nil {
				__antithesis_instrumentation__.Notify(36887)
				return err
			} else {
				__antithesis_instrumentation__.Notify(36888)
				if n != w.buf.Len() {
					__antithesis_instrumentation__.Notify(36889)
					return errors.Wrap(io.ErrShortWrite, "error when closing chunkWriter")
				} else {
					__antithesis_instrumentation__.Notify(36890)
				}
			}
			__antithesis_instrumentation__.Notify(36886)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(36891)
			return err
		} else {
			__antithesis_instrumentation__.Notify(36892)
		}
	} else {
		__antithesis_instrumentation__.Notify(36893)
	}
	__antithesis_instrumentation__.Notify(36883)

	updateFileSizeQuery := fmt.Sprintf(`UPDATE %s SET file_size=$1 WHERE filename=$2`,
		w.fileTableName)
	_, err := w.pw.ie.ExecEx(w.pw.ctx, "update-file-size",
		nil, w.execSessionDataOverride, updateFileSizeQuery, w.pw.byteOffset, w.filename)

	return err
}

type reader struct {
	pos int64
	fn  func([]byte, int64) (int, error)
}

func (r *reader) Read(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36894)
	n, err := r.fn(p, r.pos)
	r.pos += int64(n)
	return n, err
}

func newFileTableReader(
	ctx context.Context,
	filename string,
	username security.SQLUsername,
	fileTableName, payloadTableName string,
	ie FileToTableSystemExecutor,
	offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36895)

	var fileID []byte
	var sz int64
	metadataQuery := fmt.Sprintf(
		`SELECT f.file_id, sum_int(length(p.payload))
		FROM %s f LEFT OUTER JOIN %s p ON p.file_id = f.file_id
		WHERE f.filename = $1 GROUP BY f.file_id`,
		fileTableName, payloadTableName)
	metaRows, err := ie.Query(ctx, "userfile-reader-info", metadataQuery, username, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36900)
		return nil, 0, errors.Wrap(err, "failed to read file info")
	} else {
		__antithesis_instrumentation__.Notify(36901)
	}
	__antithesis_instrumentation__.Notify(36896)

	switch ie.(type) {
	case *InternalFileToTableExecutor:
		__antithesis_instrumentation__.Notify(36902)
		it := metaRows.internalExecResultsIterator
		defer func() {
			__antithesis_instrumentation__.Notify(36910)
			if err := it.Close(); err != nil {
				__antithesis_instrumentation__.Notify(36911)
				log.Warningf(ctx, "failed to close %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(36912)
			}
		}()
		__antithesis_instrumentation__.Notify(36903)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(36913)
			return nil, 0, errors.Wrap(err, "failed to read file info")
		} else {
			__antithesis_instrumentation__.Notify(36914)
		}
		__antithesis_instrumentation__.Notify(36904)
		if !ok {
			__antithesis_instrumentation__.Notify(36915)
			return nil, 0, os.ErrNotExist
		} else {
			__antithesis_instrumentation__.Notify(36916)
		}
		__antithesis_instrumentation__.Notify(36905)
		fileID = it.Cur()[0].(*tree.DUuid).UUID.GetBytes()
		if it.Cur()[1] != tree.DNull {
			__antithesis_instrumentation__.Notify(36917)
			sz = int64(tree.MustBeDInt(it.Cur()[1]))
		} else {
			__antithesis_instrumentation__.Notify(36918)
		}
	case *SQLConnFileToTableExecutor:
		__antithesis_instrumentation__.Notify(36906)
		defer func() {
			__antithesis_instrumentation__.Notify(36919)
			if err := metaRows.sqlConnExecResults.Close(); err != nil {
				__antithesis_instrumentation__.Notify(36920)
				log.Warningf(ctx, "failed to close %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(36921)
			}
		}()
		__antithesis_instrumentation__.Notify(36907)
		vals := make([]driver.Value, 2)
		err := metaRows.sqlConnExecResults.Next(vals)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(36922)
			return nil, 0, os.ErrNotExist
		} else {
			__antithesis_instrumentation__.Notify(36923)
			if err != nil {
				__antithesis_instrumentation__.Notify(36924)
				return nil, 0, errors.Wrap(err, "failed to read returned file metadata")
			} else {
				__antithesis_instrumentation__.Notify(36925)
			}
		}
		__antithesis_instrumentation__.Notify(36908)
		fileID = vals[0].([]byte)
		if vals[1] != nil {
			__antithesis_instrumentation__.Notify(36926)
			sz = vals[1].(int64)
		} else {
			__antithesis_instrumentation__.Notify(36927)
		}
	default:
		__antithesis_instrumentation__.Notify(36909)
		panic("unknown executor")
	}
	__antithesis_instrumentation__.Notify(36897)

	if sz == 0 {
		__antithesis_instrumentation__.Notify(36928)
		return ioctx.NopCloser(ioctx.ReaderAdapter(bytes.NewReader(nil))), 0, nil
	} else {
		__antithesis_instrumentation__.Notify(36929)
	}
	__antithesis_instrumentation__.Notify(36898)

	const bufSize = 256 << 10

	fn := func(p []byte, pos int64) (int, error) {
		__antithesis_instrumentation__.Notify(36930)
		if pos >= sz {
			__antithesis_instrumentation__.Notify(36935)
			return 0, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(36936)
		}
		__antithesis_instrumentation__.Notify(36931)
		query := fmt.Sprintf(
			`SELECT substr(payload, $2+1-byte_offset, $3)
			FROM %s WHERE file_id=$1 AND byte_offset <= $2
			ORDER BY byte_offset DESC
			LIMIT 1`, payloadTableName)
		rows, err := ie.Query(
			ctx, "userfile-reader-payload", query, username, fileID, pos, int64(bufSize),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(36937)
			return 0, errors.Wrap(err, "reading file content")
		} else {
			__antithesis_instrumentation__.Notify(36938)
		}
		__antithesis_instrumentation__.Notify(36932)
		var block []byte
		switch ie.(type) {
		case *InternalFileToTableExecutor:
			__antithesis_instrumentation__.Notify(36939)
			it := rows.internalExecResultsIterator
			defer func() {
				__antithesis_instrumentation__.Notify(36947)
				if err := it.Close(); err != nil {
					__antithesis_instrumentation__.Notify(36948)
					log.Warningf(ctx, "failed to close %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(36949)
				}
			}()
			__antithesis_instrumentation__.Notify(36940)
			ok, err := it.Next(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(36950)
				return 0, errors.Wrap(err, "reading file content")
			} else {
				__antithesis_instrumentation__.Notify(36951)
			}
			__antithesis_instrumentation__.Notify(36941)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(36952)
				return it.Cur()[0] == tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(36953)
				return 0, io.EOF
			} else {
				__antithesis_instrumentation__.Notify(36954)
			}
			__antithesis_instrumentation__.Notify(36942)
			block = []byte(tree.MustBeDBytes(it.Cur()[0]))
		case *SQLConnFileToTableExecutor:
			__antithesis_instrumentation__.Notify(36943)
			defer func() {
				__antithesis_instrumentation__.Notify(36955)
				if err := rows.sqlConnExecResults.Close(); err != nil {
					__antithesis_instrumentation__.Notify(36956)
					log.Warningf(ctx, "failed to close %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(36957)
				}
			}()
			__antithesis_instrumentation__.Notify(36944)
			vals := make([]driver.Value, 1)
			if err := rows.sqlConnExecResults.Next(vals); err == io.EOF {
				__antithesis_instrumentation__.Notify(36958)
				return 0, io.EOF
			} else {
				__antithesis_instrumentation__.Notify(36959)
				if err != nil {
					__antithesis_instrumentation__.Notify(36960)
					return 0, errors.Wrap(err, "failed to read returned file content")
				} else {
					__antithesis_instrumentation__.Notify(36961)
				}
			}
			__antithesis_instrumentation__.Notify(36945)
			block = vals[0].([]byte)
		default:
			__antithesis_instrumentation__.Notify(36946)
			panic("unknown executor")
		}
		__antithesis_instrumentation__.Notify(36933)
		n := copy(p, block)
		if pos+int64(n) >= sz {
			__antithesis_instrumentation__.Notify(36962)
			return n, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(36963)
		}
		__antithesis_instrumentation__.Notify(36934)
		return n, nil
	}
	__antithesis_instrumentation__.Notify(36899)

	return ioctx.NopCloser(ioctx.ReaderAdapter(
			bufio.NewReaderSize(&reader{fn: fn, pos: offset}, bufSize))),
		sz, nil
}

func (f *FileToTableSystem) ReadFile(
	ctx context.Context, filename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36964)
	return newFileTableReader(
		ctx, filename, f.username, f.GetFQFileTableName(), f.GetFQPayloadTableName(), f.executor, offset,
	)
}

func (f *FileToTableSystem) checkIfFileAndPayloadTableExist(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor,
) (bool, error) {
	__antithesis_instrumentation__.Notify(36965)
	tablePrefix, err := f.GetTableName()
	if err != nil {
		__antithesis_instrumentation__.Notify(36974)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(36975)
	}
	__antithesis_instrumentation__.Notify(36966)
	if tablePrefix == "" {
		__antithesis_instrumentation__.Notify(36976)
		return false, errors.Newf("could not resolve the table name from the FQN %s", f.qualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(36977)
	}
	__antithesis_instrumentation__.Notify(36967)
	fileTableName, err := f.GetSimpleFileTableName(tablePrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(36978)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(36979)
	}
	__antithesis_instrumentation__.Notify(36968)
	payloadTableName, err := f.GetSimplePayloadTableName(tablePrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(36980)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(36981)
	}
	__antithesis_instrumentation__.Notify(36969)
	databaseSchema, err := f.GetDatabaseAndSchema()
	if err != nil {
		__antithesis_instrumentation__.Notify(36982)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(36983)
	}
	__antithesis_instrumentation__.Notify(36970)

	if databaseSchema == "" {
		__antithesis_instrumentation__.Notify(36984)
		return false, errors.Newf("could not resolve the db and schema name from %s", f.qualifiedTableName)
	} else {
		__antithesis_instrumentation__.Notify(36985)
	}
	__antithesis_instrumentation__.Notify(36971)

	tableExistenceQuery := fmt.Sprintf(
		`SELECT table_name FROM [SHOW TABLES FROM %s] WHERE table_name=$1 OR table_name=$2`,
		databaseSchema)
	numRows, err := ie.ExecEx(ctx, "tables-exist", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		tableExistenceQuery, fileTableName, payloadTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(36986)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(36987)
	}
	__antithesis_instrumentation__.Notify(36972)

	if numRows == 1 {
		__antithesis_instrumentation__.Notify(36988)
		return false, errors.New("expected both File and Payload tables to exist, " +
			"but one of them has been dropped")
	} else {
		__antithesis_instrumentation__.Notify(36989)
	}
	__antithesis_instrumentation__.Notify(36973)
	return numRows == 2, nil
}

func (f *FileToTableSystem) createFileAndPayloadTables(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(36990)

	fileTableCreateQuery := fmt.Sprintf(fileTableSchema, f.GetFQFileTableName())
	_, err := ie.ExecEx(ctx, "create-file-table", txn,
		sessiondata.InternalExecutorOverride{User: f.username},
		fileTableCreateQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(36994)
		return errors.Wrap(err, "failed to create file table to store uploaded file names")
	} else {
		__antithesis_instrumentation__.Notify(36995)
	}
	__antithesis_instrumentation__.Notify(36991)

	payloadTableCreateQuery := fmt.Sprintf(payloadTableSchema, f.GetFQPayloadTableName())
	_, err = ie.ExecEx(ctx, "create-payload-table", txn,
		sessiondata.InternalExecutorOverride{User: f.username},
		payloadTableCreateQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(36996)
		return errors.Wrap(err, "failed to create table to store chunks of uploaded files")
	} else {
		__antithesis_instrumentation__.Notify(36997)
	}
	__antithesis_instrumentation__.Notify(36992)

	addFKQuery := fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT file_id_fk FOREIGN KEY (
file_id) REFERENCES %s (file_id)`, f.GetFQPayloadTableName(), f.GetFQFileTableName())
	_, err = ie.ExecEx(ctx, "create-payload-table", txn,
		sessiondata.InternalExecutorOverride{User: f.username},
		addFKQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(36998)
		return errors.Wrap(err, "failed to add FK constraint to the payload table file_id column")
	} else {
		__antithesis_instrumentation__.Notify(36999)
	}
	__antithesis_instrumentation__.Notify(36993)

	return nil
}

func (f *FileToTableSystem) grantCurrentUserTablePrivileges(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(37000)
	grantQuery := fmt.Sprintf(`GRANT SELECT, INSERT, DROP, DELETE ON TABLE %s, %s TO %s`,
		f.GetFQFileTableName(), f.GetFQPayloadTableName(), f.username.SQLIdentifier())
	_, err := ie.ExecEx(ctx, "grant-user-file-payload-table-access", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		grantQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(37002)
		return errors.Wrap(err, "failed to grant access privileges to file and payload tables")
	} else {
		__antithesis_instrumentation__.Notify(37003)
	}
	__antithesis_instrumentation__.Notify(37001)

	return nil
}

func (f *FileToTableSystem) revokeOtherUserTablePrivileges(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(37004)
	getUsersQuery := `SELECT username FROM system.
users WHERE NOT "username" = 'root' AND NOT "username" = 'admin' AND NOT "username" = $1`
	it, err := ie.QueryIteratorEx(
		ctx, "get-users", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		getUsersQuery, f.username,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(37009)
		return errors.Wrap(err, "failed to get all the users of the cluster")
	} else {
		__antithesis_instrumentation__.Notify(37010)
	}
	__antithesis_instrumentation__.Notify(37005)

	var users []security.SQLUsername
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(37011)
		row := it.Cur()
		username := security.MakeSQLUsernameFromPreNormalizedString(string(tree.MustBeDString(row[0])))
		users = append(users, username)
	}
	__antithesis_instrumentation__.Notify(37006)
	if err != nil {
		__antithesis_instrumentation__.Notify(37012)
		return errors.Wrap(err, "failed to get all the users of the cluster")
	} else {
		__antithesis_instrumentation__.Notify(37013)
	}
	__antithesis_instrumentation__.Notify(37007)

	for _, user := range users {
		__antithesis_instrumentation__.Notify(37014)
		revokeQuery := fmt.Sprintf(`REVOKE ALL ON TABLE %s, %s FROM %s`,
			f.GetFQFileTableName(), f.GetFQPayloadTableName(), user.SQLIdentifier())
		_, err = ie.ExecEx(ctx, "revoke-user-privileges", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			revokeQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(37015)
			return errors.Wrap(err, "failed to revoke privileges")
		} else {
			__antithesis_instrumentation__.Notify(37016)
		}
	}
	__antithesis_instrumentation__.Notify(37008)

	return nil
}

func (f *FileToTableSystem) NewFileWriter(
	ctx context.Context, filename string, chunkSize int,
) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(37017)
	e, err := resolveInternalFileToTableExecutor(f.executor)
	if err != nil {
		__antithesis_instrumentation__.Notify(37020)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(37021)
	}
	__antithesis_instrumentation__.Notify(37018)

	err = f.deleteFileWithoutTxn(ctx, filename, e.ie)
	if err != nil {
		__antithesis_instrumentation__.Notify(37022)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(37023)
	}
	__antithesis_instrumentation__.Notify(37019)

	return newChunkWriter(ctx, chunkSize, filename, f.username, f.GetFQFileTableName(),
		f.GetFQPayloadTableName(), e.ie, e.db)
}
