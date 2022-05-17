package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/security/pprompt"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/lib/pq/auth/kerberos"
)

func init() {

	pq.RegisterGSSProvider(func() (pq.GSS, error) { return kerberos.NewGSS() })
}

type sqlConn struct {
	connCtx *Context

	url          string
	conn         DriverConn
	reconnecting bool

	passwordMissing bool

	pendingNotices []*pq.Error

	delayNotices bool

	lastQueryStatsMode showLastQueryStatsMode

	dbName string

	serverVersion string
	serverBuild   string

	clusterID           string
	clusterOrganization string

	infow, errw io.Writer
}

var _ Conn = (*sqlConn)(nil)

func wrapConnError(err error) error {
	__antithesis_instrumentation__.Notify(28526)
	errMsg := err.Error()
	if errMsg == "EOF" || func() bool {
		__antithesis_instrumentation__.Notify(28528)
		return errMsg == "unexpected EOF" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28529)
		return &InitialSQLConnectionError{err}
	} else {
		__antithesis_instrumentation__.Notify(28530)
	}
	__antithesis_instrumentation__.Notify(28527)
	return err
}

func (c *sqlConn) flushNotices() {
	__antithesis_instrumentation__.Notify(28531)
	for _, notice := range c.pendingNotices {
		__antithesis_instrumentation__.Notify(28533)
		clierror.OutputError(c.errw, notice, true, false)
	}
	__antithesis_instrumentation__.Notify(28532)
	c.pendingNotices = nil
	c.delayNotices = false
}

func (c *sqlConn) handleNotice(notice *pq.Error) {
	__antithesis_instrumentation__.Notify(28534)
	c.pendingNotices = append(c.pendingNotices, notice)
	if !c.delayNotices {
		__antithesis_instrumentation__.Notify(28535)
		c.flushNotices()
	} else {
		__antithesis_instrumentation__.Notify(28536)
	}
}

func (c *sqlConn) GetURL() string {
	__antithesis_instrumentation__.Notify(28537)
	return c.url
}

func (c *sqlConn) SetURL(url string) {
	__antithesis_instrumentation__.Notify(28538)
	c.url = url
}

func (c *sqlConn) GetDriverConn() DriverConn {
	__antithesis_instrumentation__.Notify(28539)
	return c.conn
}

func (c *sqlConn) SetCurrentDatabase(dbName string) {
	__antithesis_instrumentation__.Notify(28540)
	c.dbName = dbName
}

func (c *sqlConn) SetMissingPassword(missing bool) {
	__antithesis_instrumentation__.Notify(28541)
	c.passwordMissing = missing
}

func (c *sqlConn) EnsureConn() error {
	__antithesis_instrumentation__.Notify(28542)
	if c.conn != nil {
		__antithesis_instrumentation__.Notify(28550)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(28551)
	}
	__antithesis_instrumentation__.Notify(28543)
	ctx := context.Background()

	if c.reconnecting && func() bool {
		__antithesis_instrumentation__.Notify(28552)
		return c.connCtx.IsInteractive() == true
	}() == true {
		__antithesis_instrumentation__.Notify(28553)
		fmt.Fprintf(c.errw, "warning: connection lost!\n"+
			"opening new connection: all session settings will be lost\n")
	} else {
		__antithesis_instrumentation__.Notify(28554)
	}
	__antithesis_instrumentation__.Notify(28544)
	base, err := pq.NewConnector(c.url)
	if err != nil {
		__antithesis_instrumentation__.Notify(28555)
		return wrapConnError(err)
	} else {
		__antithesis_instrumentation__.Notify(28556)
	}
	__antithesis_instrumentation__.Notify(28545)

	connector := pq.ConnectorWithNoticeHandler(base, func(notice *pq.Error) {
		__antithesis_instrumentation__.Notify(28557)
		c.handleNotice(notice)
	})
	__antithesis_instrumentation__.Notify(28546)

	conn, err := connector.Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(28558)

		errStr := strings.TrimPrefix(err.Error(), "pq: ")
		if strings.HasPrefix(errStr, "password authentication failed") && func() bool {
			__antithesis_instrumentation__.Notify(28560)
			return c.passwordMissing == true
		}() == true {
			__antithesis_instrumentation__.Notify(28561)
			if pErr := c.fillPassword(); pErr != nil {
				__antithesis_instrumentation__.Notify(28563)
				return errors.CombineErrors(err, pErr)
			} else {
				__antithesis_instrumentation__.Notify(28564)
			}
			__antithesis_instrumentation__.Notify(28562)

			return c.EnsureConn()
		} else {
			__antithesis_instrumentation__.Notify(28565)
		}
		__antithesis_instrumentation__.Notify(28559)

		return wrapConnError(err)
	} else {
		__antithesis_instrumentation__.Notify(28566)
	}
	__antithesis_instrumentation__.Notify(28547)
	if c.reconnecting && func() bool {
		__antithesis_instrumentation__.Notify(28567)
		return c.dbName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28568)

		if _, err := conn.(DriverConn).ExecContext(ctx, `SET DATABASE = $1`,
			[]driver.NamedValue{{Value: c.dbName}}); err != nil {
			__antithesis_instrumentation__.Notify(28569)
			fmt.Fprintf(c.errw, "warning: unable to restore current database: %v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(28570)
		}
	} else {
		__antithesis_instrumentation__.Notify(28571)
	}
	__antithesis_instrumentation__.Notify(28548)
	c.conn = conn.(DriverConn)
	if err := c.checkServerMetadata(ctx); err != nil {
		__antithesis_instrumentation__.Notify(28572)
		err = errors.CombineErrors(err, c.Close())
		return wrapConnError(err)
	} else {
		__antithesis_instrumentation__.Notify(28573)
	}
	__antithesis_instrumentation__.Notify(28549)
	c.reconnecting = false
	return nil
}

type showLastQueryStatsMode int

const (
	modeDisabled showLastQueryStatsMode = iota
	modeModern
	modeSimple
)

func (c *sqlConn) tryEnableServerExecutionTimings(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(28574)

	_, err := c.QueryRow(ctx, "SHOW LAST QUERY STATISTICS RETURNING x")
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(28579)
		return !clierror.IsSQLSyntaxError(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(28580)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28581)
	}
	__antithesis_instrumentation__.Notify(28575)
	if err == nil {
		__antithesis_instrumentation__.Notify(28582)
		c.connCtx.EnableServerExecutionTimings = true
		c.lastQueryStatsMode = modeModern
		return nil
	} else {
		__antithesis_instrumentation__.Notify(28583)
	}
	__antithesis_instrumentation__.Notify(28576)

	_, err = c.QueryRow(ctx, "SHOW LAST QUERY STATISTICS")
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(28584)
		return !clierror.IsSQLSyntaxError(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(28585)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28586)
	}
	__antithesis_instrumentation__.Notify(28577)
	if err == nil {
		__antithesis_instrumentation__.Notify(28587)
		c.connCtx.EnableServerExecutionTimings = true
		c.lastQueryStatsMode = modeSimple
		return nil
	} else {
		__antithesis_instrumentation__.Notify(28588)
	}
	__antithesis_instrumentation__.Notify(28578)

	fmt.Fprintln(c.errw, "warning: server does not support query statistics, cannot enable verbose timings")
	c.lastQueryStatsMode = modeDisabled
	c.connCtx.EnableServerExecutionTimings = false
	return nil
}

func (c *sqlConn) GetServerMetadata(
	ctx context.Context,
) (nodeID int32, version, clusterID string, err error) {
	__antithesis_instrumentation__.Notify(28589)

	rows, err := c.Query(ctx, "SELECT * FROM crdb_internal.node_build_info")
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28596)
		return 0, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(28597)
	}
	__antithesis_instrumentation__.Notify(28590)
	if err != nil {
		__antithesis_instrumentation__.Notify(28598)
		return 0, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(28599)
	}
	__antithesis_instrumentation__.Notify(28591)
	defer func() { __antithesis_instrumentation__.Notify(28600); _ = rows.Close() }()
	__antithesis_instrumentation__.Notify(28592)

	rowVals, err := getServerMetadataRows(rows)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(28601)
		return len(rowVals) == 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(28602)
		return len(rowVals[0]) != 3 == true
	}() == true {
		__antithesis_instrumentation__.Notify(28603)
		return 0, "", "", errors.New("incorrect data while retrieving the server version")
	} else {
		__antithesis_instrumentation__.Notify(28604)
	}
	__antithesis_instrumentation__.Notify(28593)

	var v10fields [5]string
	for _, row := range rowVals {
		__antithesis_instrumentation__.Notify(28605)
		switch row[1] {
		case "ClusterID":
			__antithesis_instrumentation__.Notify(28606)
			clusterID = row[2]
		case "Version":
			__antithesis_instrumentation__.Notify(28607)
			version = row[2]
		case "Build":
			__antithesis_instrumentation__.Notify(28608)
			c.serverBuild = row[2]
		case "Organization":
			__antithesis_instrumentation__.Notify(28609)
			c.clusterOrganization = row[2]
			id, err := strconv.Atoi(row[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(28617)
				return 0, "", "", errors.Wrap(err, "incorrect data while retrieving node id")
			} else {
				__antithesis_instrumentation__.Notify(28618)
			}
			__antithesis_instrumentation__.Notify(28610)
			nodeID = int32(id)

		case "Distribution":
			__antithesis_instrumentation__.Notify(28611)
			v10fields[0] = row[2]
		case "Tag":
			__antithesis_instrumentation__.Notify(28612)
			v10fields[1] = row[2]
		case "Platform":
			__antithesis_instrumentation__.Notify(28613)
			v10fields[2] = row[2]
		case "Time":
			__antithesis_instrumentation__.Notify(28614)
			v10fields[3] = row[2]
		case "GoVersion":
			__antithesis_instrumentation__.Notify(28615)
			v10fields[4] = row[2]
		default:
			__antithesis_instrumentation__.Notify(28616)
		}
	}
	__antithesis_instrumentation__.Notify(28594)

	if version == "" {
		__antithesis_instrumentation__.Notify(28619)

		version = "v1.0-" + v10fields[1]
		c.serverBuild = fmt.Sprintf("CockroachDB %s %s (%s, built %s, %s)",
			v10fields[0], version, v10fields[2], v10fields[3], v10fields[4])
	} else {
		__antithesis_instrumentation__.Notify(28620)
	}
	__antithesis_instrumentation__.Notify(28595)
	return nodeID, version, clusterID, nil
}

func getServerMetadataRows(rows Rows) (data [][]string, err error) {
	__antithesis_instrumentation__.Notify(28621)
	var vals []driver.Value
	cols := rows.Columns()
	if len(cols) > 0 {
		__antithesis_instrumentation__.Notify(28624)
		vals = make([]driver.Value, len(cols))
	} else {
		__antithesis_instrumentation__.Notify(28625)
	}
	__antithesis_instrumentation__.Notify(28622)
	for {
		__antithesis_instrumentation__.Notify(28626)
		err = rows.Next(vals)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(28630)
			break
		} else {
			__antithesis_instrumentation__.Notify(28631)
		}
		__antithesis_instrumentation__.Notify(28627)
		if err != nil {
			__antithesis_instrumentation__.Notify(28632)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(28633)
		}
		__antithesis_instrumentation__.Notify(28628)
		rowStrings := make([]string, len(cols))
		for i, v := range vals {
			__antithesis_instrumentation__.Notify(28634)
			rowStrings[i] = toString(v)
		}
		__antithesis_instrumentation__.Notify(28629)
		data = append(data, rowStrings)
	}
	__antithesis_instrumentation__.Notify(28623)

	return data, nil
}

func toString(v driver.Value) string {
	__antithesis_instrumentation__.Notify(28635)
	switch x := v.(type) {
	case []byte:
		__antithesis_instrumentation__.Notify(28636)
		return fmt.Sprint(string(x))
	default:
		__antithesis_instrumentation__.Notify(28637)
		return fmt.Sprint(x)
	}
}

func (c *sqlConn) checkServerMetadata(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(28638)
	if !c.connCtx.IsInteractive() {
		__antithesis_instrumentation__.Notify(28645)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(28646)
	}
	__antithesis_instrumentation__.Notify(28639)
	if c.connCtx.EmbeddedMode() {
		__antithesis_instrumentation__.Notify(28647)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(28648)
	}
	__antithesis_instrumentation__.Notify(28640)

	_, newServerVersion, newClusterID, err := c.GetServerMetadata(ctx)
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28649)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28650)
	}
	__antithesis_instrumentation__.Notify(28641)
	if err != nil {
		__antithesis_instrumentation__.Notify(28651)

		fmt.Fprintf(c.errw, "warning: unable to retrieve the server's version: %s\n", err)
	} else {
		__antithesis_instrumentation__.Notify(28652)
	}
	__antithesis_instrumentation__.Notify(28642)

	if newServerVersion != c.serverVersion {
		__antithesis_instrumentation__.Notify(28653)
		c.serverVersion = newServerVersion

		isSame := ""

		client := build.GetInfo()
		if c.serverVersion != client.Tag {
			__antithesis_instrumentation__.Notify(28655)
			fmt.Fprintln(c.infow, "# Client version:", client.Short())
		} else {
			__antithesis_instrumentation__.Notify(28656)
			isSame = " (same version as client)"
		}
		__antithesis_instrumentation__.Notify(28654)
		fmt.Fprintf(c.infow, "# Server version: %s%s\n", c.serverBuild, isSame)

		sv, err := version.Parse(c.serverVersion)
		if err == nil {
			__antithesis_instrumentation__.Notify(28657)
			cv, err := version.Parse(client.Tag)
			if err == nil {
				__antithesis_instrumentation__.Notify(28658)
				if sv.Compare(cv) == -1 {
					__antithesis_instrumentation__.Notify(28659)
					fmt.Fprintln(c.errw, "\nwarning: server version older than client! "+
						"proceed with caution; some features may not be available.\n")
				} else {
					__antithesis_instrumentation__.Notify(28660)
				}
			} else {
				__antithesis_instrumentation__.Notify(28661)
			}
		} else {
			__antithesis_instrumentation__.Notify(28662)
		}
	} else {
		__antithesis_instrumentation__.Notify(28663)
	}
	__antithesis_instrumentation__.Notify(28643)

	if old := c.clusterID; newClusterID != c.clusterID {
		__antithesis_instrumentation__.Notify(28664)
		c.clusterID = newClusterID
		if old != "" {
			__antithesis_instrumentation__.Notify(28666)
			return errors.Errorf("the cluster ID has changed!\nPrevious ID: %s\nNew ID: %s",
				old, newClusterID)
		} else {
			__antithesis_instrumentation__.Notify(28667)
		}
		__antithesis_instrumentation__.Notify(28665)
		c.clusterID = newClusterID
		fmt.Fprintln(c.infow, "# Cluster ID:", c.clusterID)
		if c.clusterOrganization != "" {
			__antithesis_instrumentation__.Notify(28668)
			fmt.Fprintln(c.infow, "# Organization:", c.clusterOrganization)
		} else {
			__antithesis_instrumentation__.Notify(28669)
		}
	} else {
		__antithesis_instrumentation__.Notify(28670)
	}
	__antithesis_instrumentation__.Notify(28644)

	return c.tryEnableServerExecutionTimings(ctx)
}

func (c *sqlConn) GetServerValue(
	ctx context.Context, what, sql string,
) (driver.Value, string, bool) {
	__antithesis_instrumentation__.Notify(28671)
	rows, err := c.Query(ctx, sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(28676)
		fmt.Fprintf(c.errw, "warning: error retrieving the %s: %v\n", what, err)
		return nil, "", false
	} else {
		__antithesis_instrumentation__.Notify(28677)
	}
	__antithesis_instrumentation__.Notify(28672)
	defer func() { __antithesis_instrumentation__.Notify(28678); _ = rows.Close() }()
	__antithesis_instrumentation__.Notify(28673)

	if len(rows.Columns()) == 0 {
		__antithesis_instrumentation__.Notify(28679)
		fmt.Fprintf(c.errw, "warning: cannot get the %s\n", what)
		return nil, "", false
	} else {
		__antithesis_instrumentation__.Notify(28680)
	}
	__antithesis_instrumentation__.Notify(28674)

	dbColType := rows.ColumnTypeDatabaseTypeName(0)
	dbVals := make([]driver.Value, len(rows.Columns()))

	err = rows.Next(dbVals[:])
	if err != nil {
		__antithesis_instrumentation__.Notify(28681)
		fmt.Fprintf(c.errw, "warning: invalid %s: %v\n", what, err)
		return nil, "", false
	} else {
		__antithesis_instrumentation__.Notify(28682)
	}
	__antithesis_instrumentation__.Notify(28675)

	return dbVals[0], dbColType, true
}

func (c *sqlConn) GetLastQueryStatistics(ctx context.Context) (results QueryStats, resErr error) {
	__antithesis_instrumentation__.Notify(28683)
	if !c.connCtx.EnableServerExecutionTimings || func() bool {
		__antithesis_instrumentation__.Notify(28688)
		return c.lastQueryStatsMode == modeDisabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(28689)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(28690)
	}
	__antithesis_instrumentation__.Notify(28684)

	stmt := `SHOW LAST QUERY STATISTICS RETURNING parse_latency, plan_latency, exec_latency, service_latency, post_commit_jobs_latency`
	if c.lastQueryStatsMode == modeSimple {
		__antithesis_instrumentation__.Notify(28691)

		stmt = `SHOW LAST QUERY STATISTICS`
	} else {
		__antithesis_instrumentation__.Notify(28692)
	}
	__antithesis_instrumentation__.Notify(28685)

	vals, cols, err := c.queryRowInternal(ctx, stmt, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(28693)
		return results, err
	} else {
		__antithesis_instrumentation__.Notify(28694)
	}
	__antithesis_instrumentation__.Notify(28686)

	for i, c := range cols {
		__antithesis_instrumentation__.Notify(28695)
		var dst *QueryStatsDuration
		switch c {
		case "parse_latency":
			__antithesis_instrumentation__.Notify(28697)
			dst = &results.Parse
		case "plan_latency":
			__antithesis_instrumentation__.Notify(28698)
			dst = &results.Plan
		case "exec_latency":
			__antithesis_instrumentation__.Notify(28699)
			dst = &results.Exec
		case "service_latency":
			__antithesis_instrumentation__.Notify(28700)
			dst = &results.Service
		case "post_commit_jobs_latency":
			__antithesis_instrumentation__.Notify(28701)
			dst = &results.PostCommitJobs
		default:
			__antithesis_instrumentation__.Notify(28702)
		}
		__antithesis_instrumentation__.Notify(28696)
		if vals[i] != nil {
			__antithesis_instrumentation__.Notify(28703)
			rawVal := toString(vals[i])
			parsedLat, err := stringToDuration(rawVal)
			if err != nil {
				__antithesis_instrumentation__.Notify(28705)
				return results, errors.Wrapf(err, "invalid interval value in SHOW LAST QUERY STATISTICS, column %q", c)
			} else {
				__antithesis_instrumentation__.Notify(28706)
			}
			__antithesis_instrumentation__.Notify(28704)
			dst.Valid = true
			dst.Value = parsedLat
		} else {
			__antithesis_instrumentation__.Notify(28707)
		}
	}
	__antithesis_instrumentation__.Notify(28687)

	results.Enabled = true
	return results, nil
}

func (c *sqlConn) ExecTxn(
	ctx context.Context, fn func(context.Context, TxBoundConn) error,
) (err error) {
	__antithesis_instrumentation__.Notify(28708)
	if err := c.Exec(ctx, `BEGIN`); err != nil {
		__antithesis_instrumentation__.Notify(28710)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28711)
	}
	__antithesis_instrumentation__.Notify(28709)
	return crdb.ExecuteInTx(ctx, sqlTxnShim{c}, func() error {
		__antithesis_instrumentation__.Notify(28712)
		return fn(ctx, c)
	})
}

func (c *sqlConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(28713)
	dVals, err := convertArgs(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(28718)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28719)
	}
	__antithesis_instrumentation__.Notify(28714)
	if err := c.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(28720)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28721)
	}
	__antithesis_instrumentation__.Notify(28715)
	if c.connCtx.Echo {
		__antithesis_instrumentation__.Notify(28722)
		fmt.Fprintln(c.errw, ">", query)
	} else {
		__antithesis_instrumentation__.Notify(28723)
	}
	__antithesis_instrumentation__.Notify(28716)
	_, err = c.conn.ExecContext(ctx, query, dVals)
	c.flushNotices()
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28724)
		c.reconnecting = true
		c.silentClose()
	} else {
		__antithesis_instrumentation__.Notify(28725)
	}
	__antithesis_instrumentation__.Notify(28717)
	return err
}

func (c *sqlConn) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	__antithesis_instrumentation__.Notify(28726)
	dVals, err := convertArgs(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(28732)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28733)
	}
	__antithesis_instrumentation__.Notify(28727)
	if err := c.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(28734)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28735)
	}
	__antithesis_instrumentation__.Notify(28728)
	if c.connCtx.Echo {
		__antithesis_instrumentation__.Notify(28736)
		fmt.Fprintln(c.errw, ">", query)
	} else {
		__antithesis_instrumentation__.Notify(28737)
	}
	__antithesis_instrumentation__.Notify(28729)
	rows, err := c.conn.QueryContext(ctx, query, dVals)
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28738)
		c.reconnecting = true
		c.silentClose()
	} else {
		__antithesis_instrumentation__.Notify(28739)
	}
	__antithesis_instrumentation__.Notify(28730)
	if err != nil {
		__antithesis_instrumentation__.Notify(28740)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28741)
	}
	__antithesis_instrumentation__.Notify(28731)
	return &sqlRows{rows: rows.(sqlRowsI), conn: c}, nil
}

func (c *sqlConn) QueryRow(
	ctx context.Context, query string, args ...interface{},
) ([]driver.Value, error) {
	__antithesis_instrumentation__.Notify(28742)
	results, _, err := c.queryRowInternal(ctx, query, args)
	return results, err
}

func (c *sqlConn) queryRowInternal(
	ctx context.Context, query string, args []interface{},
) (vals []driver.Value, colNames []string, resErr error) {
	__antithesis_instrumentation__.Notify(28743)
	rows, _, err := MakeQuery(query, args...)(ctx, c)
	if err != nil {
		__antithesis_instrumentation__.Notify(28747)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(28748)
	}
	__antithesis_instrumentation__.Notify(28744)
	defer func() {
		__antithesis_instrumentation__.Notify(28749)
		resErr = errors.CombineErrors(resErr, rows.Close())
	}()
	__antithesis_instrumentation__.Notify(28745)
	colNames = rows.Columns()
	vals = make([]driver.Value, len(colNames))
	err = rows.Next(vals)

	if err == nil {
		__antithesis_instrumentation__.Notify(28750)
		nextVals := make([]driver.Value, len(colNames))
		nextErr := rows.Next(nextVals)
		if nextErr != io.EOF {
			__antithesis_instrumentation__.Notify(28751)
			if nextErr != nil {
				__antithesis_instrumentation__.Notify(28753)
				return nil, nil, nextErr
			} else {
				__antithesis_instrumentation__.Notify(28754)
			}
			__antithesis_instrumentation__.Notify(28752)
			return nil, nil, errors.AssertionFailedf("programming error: %q: expected just 1 row of result, got more", query)
		} else {
			__antithesis_instrumentation__.Notify(28755)
		}
	} else {
		__antithesis_instrumentation__.Notify(28756)
	}
	__antithesis_instrumentation__.Notify(28746)

	return vals, colNames, err
}

func (c *sqlConn) Close() error {
	__antithesis_instrumentation__.Notify(28757)
	c.flushNotices()
	if c.conn != nil {
		__antithesis_instrumentation__.Notify(28759)
		err := c.conn.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(28761)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28762)
		}
		__antithesis_instrumentation__.Notify(28760)
		c.conn = nil
	} else {
		__antithesis_instrumentation__.Notify(28763)
	}
	__antithesis_instrumentation__.Notify(28758)
	return nil
}

func (c *sqlConn) silentClose() {
	__antithesis_instrumentation__.Notify(28764)
	if c.conn != nil {
		__antithesis_instrumentation__.Notify(28765)
		_ = c.conn.Close()
		c.conn = nil
	} else {
		__antithesis_instrumentation__.Notify(28766)
	}
}

func (connCtx *Context) MakeSQLConn(w, ew io.Writer, url string) Conn {
	__antithesis_instrumentation__.Notify(28767)
	return &sqlConn{
		connCtx: connCtx,
		url:     url,
		infow:   w,
		errw:    ew,
	}
}

func (c *sqlConn) fillPassword() error {
	__antithesis_instrumentation__.Notify(28768)
	connURL, err := url.Parse(c.url)
	if err != nil {
		__antithesis_instrumentation__.Notify(28771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28772)
	}
	__antithesis_instrumentation__.Notify(28769)

	fmt.Fprintf(c.infow, "Connecting to server %q as user %q.\n",
		connURL.Host,
		connURL.User.Username())

	pwd, err := pprompt.PromptForPassword()
	if err != nil {
		__antithesis_instrumentation__.Notify(28773)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28774)
	}
	__antithesis_instrumentation__.Notify(28770)
	connURL.User = url.UserPassword(connURL.User.Username(), pwd)
	c.url = connURL.String()
	c.passwordMissing = false
	return nil
}
