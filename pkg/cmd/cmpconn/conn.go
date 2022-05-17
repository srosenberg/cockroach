// Package cmpconn assists in comparing results from DB connections.
package cmpconn

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
)

type Conn interface {
	DB() *gosql.DB

	PGX() *pgx.Conn

	Values(ctx context.Context, prep, exec string) (rows pgx.Rows, err error)

	Exec(ctx context.Context, s string) error

	Ping(ctx context.Context) error

	Close(ctx context.Context)
}

type conn struct {
	db  *gosql.DB
	pgx *pgx.Conn
}

var _ Conn = &conn{}

func (c *conn) DB() *gosql.DB {
	__antithesis_instrumentation__.Notify(38051)
	return c.db
}

func (c *conn) PGX() *pgx.Conn {
	__antithesis_instrumentation__.Notify(38052)
	return c.pgx
}

func (c *conn) Values(ctx context.Context, prep, exec string) (rows pgx.Rows, err error) {
	__antithesis_instrumentation__.Notify(38053)
	if prep != "" {
		__antithesis_instrumentation__.Notify(38055)
		rows, err = c.pgx.Query(ctx, prep, pgx.QuerySimpleProtocol(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(38057)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(38058)
		}
		__antithesis_instrumentation__.Notify(38056)
		rows.Close()
	} else {
		__antithesis_instrumentation__.Notify(38059)
	}
	__antithesis_instrumentation__.Notify(38054)
	return c.pgx.Query(ctx, exec, pgx.QuerySimpleProtocol(true))
}

func (c *conn) Exec(ctx context.Context, s string) error {
	__antithesis_instrumentation__.Notify(38060)
	_, err := c.pgx.Exec(ctx, s, pgx.QuerySimpleProtocol(true))
	return errors.Wrap(err, "exec")
}

func (c *conn) Ping(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(38061)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.pgx.Ping(ctx)
}

func (c *conn) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(38062)
	_ = c.db.Close()
	_ = c.pgx.Close(ctx)
}

func NewConn(ctx context.Context, uri string, initSQL ...string) (Conn, error) {
	__antithesis_instrumentation__.Notify(38063)
	c := conn{}

	{
		__antithesis_instrumentation__.Notify(38066)
		connector, err := pq.NewConnector(uri)
		if err != nil {
			__antithesis_instrumentation__.Notify(38068)
			return nil, errors.Wrap(err, "pq conn")
		} else {
			__antithesis_instrumentation__.Notify(38069)
		}
		__antithesis_instrumentation__.Notify(38067)
		c.db = gosql.OpenDB(connector)
	}

	{
		__antithesis_instrumentation__.Notify(38070)
		config, err := pgx.ParseConfig(uri)
		if err != nil {
			__antithesis_instrumentation__.Notify(38073)
			return nil, errors.Wrap(err, "pgx parse")
		} else {
			__antithesis_instrumentation__.Notify(38074)
		}
		__antithesis_instrumentation__.Notify(38071)
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			__antithesis_instrumentation__.Notify(38075)
			return nil, errors.Wrap(err, "pgx conn")
		} else {
			__antithesis_instrumentation__.Notify(38076)
		}
		__antithesis_instrumentation__.Notify(38072)
		c.pgx = conn
	}
	__antithesis_instrumentation__.Notify(38064)

	for _, s := range initSQL {
		__antithesis_instrumentation__.Notify(38077)
		if s == "" {
			__antithesis_instrumentation__.Notify(38079)
			continue
		} else {
			__antithesis_instrumentation__.Notify(38080)
		}
		__antithesis_instrumentation__.Notify(38078)

		if _, err := c.pgx.Exec(ctx, s); err != nil {
			__antithesis_instrumentation__.Notify(38081)
			return nil, errors.Wrap(err, "init SQL")
		} else {
			__antithesis_instrumentation__.Notify(38082)
		}
	}
	__antithesis_instrumentation__.Notify(38065)

	return &c, nil
}

type connWithMutators struct {
	Conn
	rng         *rand.Rand
	sqlMutators []randgen.Mutator
}

var _ Conn = &connWithMutators{}

func NewConnWithMutators(
	ctx context.Context, uri string, rng *rand.Rand, sqlMutators []randgen.Mutator, initSQL ...string,
) (Conn, error) {
	__antithesis_instrumentation__.Notify(38083)
	mutatedInitSQL := make([]string, len(initSQL))
	for i, s := range initSQL {
		__antithesis_instrumentation__.Notify(38086)
		mutatedInitSQL[i] = s
		if s == "" {
			__antithesis_instrumentation__.Notify(38088)
			continue
		} else {
			__antithesis_instrumentation__.Notify(38089)
		}
		__antithesis_instrumentation__.Notify(38087)

		mutatedInitSQL[i], _ = randgen.ApplyString(rng, s, sqlMutators...)
	}
	__antithesis_instrumentation__.Notify(38084)
	conn, err := NewConn(ctx, uri, mutatedInitSQL...)
	if err != nil {
		__antithesis_instrumentation__.Notify(38090)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(38091)
	}
	__antithesis_instrumentation__.Notify(38085)
	return &connWithMutators{
		Conn:        conn,
		rng:         rng,
		sqlMutators: sqlMutators,
	}, nil
}

func CompareConns(
	ctx context.Context,
	timeout time.Duration,
	conns map[string]Conn,
	prep, exec string,
	ignoreSQLErrors bool,
) (ignoredErr bool, err error) {
	__antithesis_instrumentation__.Notify(38092)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	connRows := make(map[string]pgx.Rows)
	connExecs := make(map[string]string)
	for name, conn := range conns {
		__antithesis_instrumentation__.Notify(38095)
		connExecs[name] = exec
		if cwm, withMutators := conn.(*connWithMutators); withMutators {
			__antithesis_instrumentation__.Notify(38098)
			connExecs[name], _ = randgen.ApplyString(cwm.rng, exec, cwm.sqlMutators...)
		} else {
			__antithesis_instrumentation__.Notify(38099)
		}
		__antithesis_instrumentation__.Notify(38096)
		rows, err := conn.Values(ctx, prep, connExecs[name])
		if err != nil {
			__antithesis_instrumentation__.Notify(38100)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(38101)
		}
		__antithesis_instrumentation__.Notify(38097)
		defer rows.Close()
		connRows[name] = rows
	}
	__antithesis_instrumentation__.Notify(38093)

	defer func() {
		__antithesis_instrumentation__.Notify(38102)
		if err == nil {
			__antithesis_instrumentation__.Notify(38105)
			return
		} else {
			__antithesis_instrumentation__.Notify(38106)
		}
		__antithesis_instrumentation__.Notify(38103)
		var sb strings.Builder
		prev := ""
		for name, mutated := range connExecs {
			__antithesis_instrumentation__.Notify(38107)
			fmt.Fprintf(&sb, "\n%s:", name)
			if prev == mutated {
				__antithesis_instrumentation__.Notify(38109)
				sb.WriteString(" [same as previous]\n")
			} else {
				__antithesis_instrumentation__.Notify(38110)
				fmt.Fprintf(&sb, "\n%s;\n", mutated)
			}
			__antithesis_instrumentation__.Notify(38108)
			prev = mutated
		}
		__antithesis_instrumentation__.Notify(38104)
		err = fmt.Errorf("%w%s", err, sb.String())
	}()
	__antithesis_instrumentation__.Notify(38094)

	return compareRows(connRows, ignoreSQLErrors)
}

func compareRows(
	connRows map[string]pgx.Rows, ignoreSQLErrors bool,
) (ignoredErr bool, retErr error) {
	__antithesis_instrumentation__.Notify(38111)
	var first []interface{}
	var firstName string
	var minCount int
	rowCounts := make(map[string]int)
ReadRows:
	for {
		__antithesis_instrumentation__.Notify(38115)
		first = nil
		firstName = ""
		for name, rows := range connRows {
			__antithesis_instrumentation__.Notify(38116)
			if !rows.Next() {
				__antithesis_instrumentation__.Notify(38119)
				minCount = rowCounts[name]
				break ReadRows
			} else {
				__antithesis_instrumentation__.Notify(38120)
			}
			__antithesis_instrumentation__.Notify(38117)
			rowCounts[name]++
			vals, err := rows.Values()
			if err != nil {
				__antithesis_instrumentation__.Notify(38121)
				if ignoreSQLErrors {
					__antithesis_instrumentation__.Notify(38123)

					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(38124)
				}
				__antithesis_instrumentation__.Notify(38122)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(38125)
			}
			__antithesis_instrumentation__.Notify(38118)
			if firstName == "" {
				__antithesis_instrumentation__.Notify(38126)
				firstName = name
				first = vals
			} else {
				__antithesis_instrumentation__.Notify(38127)
				if err := CompareVals(first, vals); err != nil {
					__antithesis_instrumentation__.Notify(38128)
					return false, errors.Wrapf(err, "compare %s to %s", firstName, name)
				} else {
					__antithesis_instrumentation__.Notify(38129)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(38112)

	for name, rows := range connRows {
		__antithesis_instrumentation__.Notify(38130)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(38132)
			rowCounts[name]++
		}
		__antithesis_instrumentation__.Notify(38131)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(38133)
			if ignoreSQLErrors {
				__antithesis_instrumentation__.Notify(38135)

				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(38136)
			}
			__antithesis_instrumentation__.Notify(38134)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(38137)
		}
	}
	__antithesis_instrumentation__.Notify(38113)

	for name, count := range rowCounts {
		__antithesis_instrumentation__.Notify(38138)
		if minCount != count {
			__antithesis_instrumentation__.Notify(38139)
			return false, fmt.Errorf("%s had %d rows, expected %d", name, count, minCount)
		} else {
			__antithesis_instrumentation__.Notify(38140)
		}
	}
	__antithesis_instrumentation__.Notify(38114)
	return false, nil
}
