// Package pgconnect provides a way to get byte encodings from a simple query.
package pgconnect

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgproto3/v2"
)

func Connect(
	ctx context.Context, input, addr, user string, code pgwirebase.FormatCode,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(37830)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(37837)
		return nil, errors.Wrap(err, "dail")
	} else {
		__antithesis_instrumentation__.Notify(37838)
	}
	__antithesis_instrumentation__.Notify(37831)
	defer conn.Close()

	fe := pgproto3.NewFrontend(pgproto3.NewChunkReader(conn), conn)

	send := make(chan pgproto3.FrontendMessage)
	recv := make(chan pgproto3.BackendMessage)
	var res []byte

	g := ctxgroup.WithContext(ctx)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(37839)
		defer close(send)
		for {
			__antithesis_instrumentation__.Notify(37840)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(37841)
				return ctx.Err()
			case msg := <-send:
				__antithesis_instrumentation__.Notify(37842)
				err := fe.Send(msg)
				if err != nil {
					__antithesis_instrumentation__.Notify(37843)
					return errors.Wrap(err, "send")
				} else {
					__antithesis_instrumentation__.Notify(37844)
				}
			}
		}
	})
	__antithesis_instrumentation__.Notify(37832)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(37845)
		defer close(recv)
		for {
			__antithesis_instrumentation__.Notify(37846)
			msg, err := fe.Receive()
			if err != nil {
				__antithesis_instrumentation__.Notify(37848)
				return errors.Wrap(err, "receive")
			} else {
				__antithesis_instrumentation__.Notify(37849)
			}
			__antithesis_instrumentation__.Notify(37847)

			x := reflect.ValueOf(msg)
			starX := x.Elem()
			y := reflect.New(starX.Type())
			starY := y.Elem()
			starY.Set(starX)
			dup := y.Interface().(pgproto3.BackendMessage)

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(37850)
				return ctx.Err()
			case recv <- dup:
				__antithesis_instrumentation__.Notify(37851)
			}
		}
	})
	__antithesis_instrumentation__.Notify(37833)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(37852)
		send <- &pgproto3.StartupMessage{
			ProtocolVersion: 196608,
			Parameters: map[string]string{
				"user": user,
			},
		}
		{
			__antithesis_instrumentation__.Notify(37857)
			r := <-recv
			if _, ok := r.(*pgproto3.AuthenticationOk); !ok {
				__antithesis_instrumentation__.Notify(37858)
				return errors.Errorf("unexpected: %#v\n", r)
			} else {
				__antithesis_instrumentation__.Notify(37859)
			}
		}
		__antithesis_instrumentation__.Notify(37853)
	WaitConnLoop:
		for {
			__antithesis_instrumentation__.Notify(37860)
			msg := <-recv
			switch msg.(type) {
			case *pgproto3.ReadyForQuery:
				__antithesis_instrumentation__.Notify(37861)
				break WaitConnLoop
			}
		}
		__antithesis_instrumentation__.Notify(37854)
		send <- &pgproto3.Parse{
			Query: input,
		}
		send <- &pgproto3.Describe{
			ObjectType: 'S',
		}
		send <- &pgproto3.Sync{}
		r := <-recv
		if _, ok := r.(*pgproto3.ParseComplete); !ok {
			__antithesis_instrumentation__.Notify(37862)
			return errors.Errorf("unexpected: %#v", r)
		} else {
			__antithesis_instrumentation__.Notify(37863)
		}
		__antithesis_instrumentation__.Notify(37855)
		send <- &pgproto3.Bind{
			ResultFormatCodes: []int16{int16(code)},
		}
		send <- &pgproto3.Execute{}
		send <- &pgproto3.Sync{}
	WaitExecuteLoop:
		for {
			__antithesis_instrumentation__.Notify(37864)
			msg := <-recv
			switch msg := msg.(type) {
			case *pgproto3.DataRow:
				__antithesis_instrumentation__.Notify(37865)
				if res != nil {
					__antithesis_instrumentation__.Notify(37869)
					return errors.New("already saw a row")
				} else {
					__antithesis_instrumentation__.Notify(37870)
				}
				__antithesis_instrumentation__.Notify(37866)
				if len(msg.Values) != 1 {
					__antithesis_instrumentation__.Notify(37871)
					return errors.Errorf("unexpected: %#v\n", msg)
				} else {
					__antithesis_instrumentation__.Notify(37872)
				}
				__antithesis_instrumentation__.Notify(37867)
				res = msg.Values[0]
			case *pgproto3.CommandComplete,
				*pgproto3.EmptyQueryResponse,
				*pgproto3.ErrorResponse:
				__antithesis_instrumentation__.Notify(37868)
				break WaitExecuteLoop
			}
		}
		__antithesis_instrumentation__.Notify(37856)

		cancel()
		return nil
	})
	__antithesis_instrumentation__.Notify(37834)
	err = g.Wait()

	if res != nil {
		__antithesis_instrumentation__.Notify(37873)
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(37874)
	}
	__antithesis_instrumentation__.Notify(37835)
	if err == nil {
		__antithesis_instrumentation__.Notify(37875)
		return nil, errors.New("unexpected")
	} else {
		__antithesis_instrumentation__.Notify(37876)
	}
	__antithesis_instrumentation__.Notify(37836)
	return nil, err
}
