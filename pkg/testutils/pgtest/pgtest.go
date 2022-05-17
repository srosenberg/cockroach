package pgtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgproto3/v2"
)

type PGTest struct {
	fe            *pgproto3.Frontend
	conn          net.Conn
	isCockroachDB bool
}

func NewPGTest(ctx context.Context, addr, user string) (*PGTest, error) {
	__antithesis_instrumentation__.Notify(645818)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(645825)
		return nil, errors.Wrap(err, "dial")
	} else {
		__antithesis_instrumentation__.Notify(645826)
	}
	__antithesis_instrumentation__.Notify(645819)
	success := false
	defer func() {
		__antithesis_instrumentation__.Notify(645827)
		if !success {
			__antithesis_instrumentation__.Notify(645828)
			conn.Close()
		} else {
			__antithesis_instrumentation__.Notify(645829)
		}
	}()
	__antithesis_instrumentation__.Notify(645820)
	fe := pgproto3.NewFrontend(pgproto3.NewChunkReader(conn), conn)
	if err := fe.Send(&pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"user": user,
		},
	}); err != nil {
		__antithesis_instrumentation__.Notify(645830)
		return nil, errors.Wrap(err, "startup")
	} else {
		__antithesis_instrumentation__.Notify(645831)
	}
	__antithesis_instrumentation__.Notify(645821)
	if msg, err := fe.Receive(); err != nil {
		__antithesis_instrumentation__.Notify(645832)
		return nil, errors.Wrap(err, "receive")
	} else {
		__antithesis_instrumentation__.Notify(645833)
		if _, ok := msg.(*pgproto3.AuthenticationOk); !ok {
			__antithesis_instrumentation__.Notify(645834)
			return nil, errors.Errorf("unexpected: %#v", msg)
		} else {
			__antithesis_instrumentation__.Notify(645835)
		}
	}
	__antithesis_instrumentation__.Notify(645822)
	p := &PGTest{
		fe:   fe,
		conn: conn,
	}
	msgs, err := p.Until(false, &pgproto3.ReadyForQuery{})
	foundCrdb := false
	var backendKeyData *pgproto3.BackendKeyData
	for _, msg := range msgs {
		__antithesis_instrumentation__.Notify(645836)
		if s, ok := msg.(*pgproto3.ParameterStatus); ok && func() bool {
			__antithesis_instrumentation__.Notify(645838)
			return s.Name == "crdb_version" == true
		}() == true {
			__antithesis_instrumentation__.Notify(645839)
			foundCrdb = true
		} else {
			__antithesis_instrumentation__.Notify(645840)
		}
		__antithesis_instrumentation__.Notify(645837)
		if d, ok := msg.(*pgproto3.BackendKeyData); ok {
			__antithesis_instrumentation__.Notify(645841)
			backendKeyData = d
		} else {
			__antithesis_instrumentation__.Notify(645842)
		}
	}
	__antithesis_instrumentation__.Notify(645823)
	if backendKeyData == nil {
		__antithesis_instrumentation__.Notify(645843)
		return nil, errors.Errorf("did not receive BackendKeyData")
	} else {
		__antithesis_instrumentation__.Notify(645844)
	}
	__antithesis_instrumentation__.Notify(645824)
	p.isCockroachDB = foundCrdb
	success = err == nil
	return p, err
}

func (p *PGTest) Close() error {
	__antithesis_instrumentation__.Notify(645845)
	defer p.conn.Close()
	return p.fe.Send(&pgproto3.Terminate{})
}

func (p *PGTest) SendOneLine(line string) error {
	__antithesis_instrumentation__.Notify(645846)
	sp := strings.SplitN(line, " ", 2)
	msg := toMessage(sp[0])
	if len(sp) == 2 {
		__antithesis_instrumentation__.Notify(645848)
		msgBytes := []byte(sp[1])
		switch msg := msg.(type) {
		case *pgproto3.CopyData:
			__antithesis_instrumentation__.Notify(645849)
			var data struct {
				Data       string
				BinaryData []byte
			}
			if err := json.Unmarshal(msgBytes, &data); err != nil {
				__antithesis_instrumentation__.Notify(645852)
				return err
			} else {
				__antithesis_instrumentation__.Notify(645853)
			}
			__antithesis_instrumentation__.Notify(645850)
			if data.BinaryData != nil {
				__antithesis_instrumentation__.Notify(645854)
				msg.Data = data.BinaryData
			} else {
				__antithesis_instrumentation__.Notify(645855)
				msg.Data = []byte(data.Data)
			}
		default:
			__antithesis_instrumentation__.Notify(645851)
			if err := json.Unmarshal(msgBytes, msg); err != nil {
				__antithesis_instrumentation__.Notify(645856)
				return err
			} else {
				__antithesis_instrumentation__.Notify(645857)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(645858)
	}
	__antithesis_instrumentation__.Notify(645847)
	return p.Send(msg.(pgproto3.FrontendMessage))
}

func (p *PGTest) Send(msg pgproto3.FrontendMessage) error {
	__antithesis_instrumentation__.Notify(645859)
	if testing.Verbose() {
		__antithesis_instrumentation__.Notify(645861)
		fmt.Printf("SEND %T: %+[1]v\n", msg)
	} else {
		__antithesis_instrumentation__.Notify(645862)
	}
	__antithesis_instrumentation__.Notify(645860)
	return p.fe.Send(msg)
}

func (p *PGTest) Receive(
	keepErrMsg bool, typs ...pgproto3.BackendMessage,
) ([]pgproto3.BackendMessage, error) {
	__antithesis_instrumentation__.Notify(645863)
	var matched []pgproto3.BackendMessage
	for len(typs) > 0 {
		__antithesis_instrumentation__.Notify(645865)
		msgs, err := p.Until(keepErrMsg, typs[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(645867)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(645868)
		}
		__antithesis_instrumentation__.Notify(645866)
		matched = append(matched, msgs[len(msgs)-1])
		typs = typs[1:]
	}
	__antithesis_instrumentation__.Notify(645864)
	return matched, nil
}

func (p *PGTest) Until(
	keepErrMsg bool, typs ...pgproto3.BackendMessage,
) ([]pgproto3.BackendMessage, error) {
	__antithesis_instrumentation__.Notify(645869)
	var msgs []pgproto3.BackendMessage
	for len(typs) > 0 {
		__antithesis_instrumentation__.Notify(645871)
		typ := reflect.TypeOf(typs[0])

		recv, err := p.fe.Receive()
		if err != nil {
			__antithesis_instrumentation__.Notify(645879)
			return nil, errors.Wrap(err, "receive")
		} else {
			__antithesis_instrumentation__.Notify(645880)
		}
		__antithesis_instrumentation__.Notify(645872)
		if testing.Verbose() {
			__antithesis_instrumentation__.Notify(645881)
			fmt.Printf("RECV %T: %+[1]v\n", recv)
		} else {
			__antithesis_instrumentation__.Notify(645882)
		}
		__antithesis_instrumentation__.Notify(645873)
		if errmsg, ok := recv.(*pgproto3.ErrorResponse); ok {
			__antithesis_instrumentation__.Notify(645883)
			if typ != typErrorResponse {
				__antithesis_instrumentation__.Notify(645886)
				return nil, errors.Errorf("waiting for %T, got %#v", typs[0], errmsg)
			} else {
				__antithesis_instrumentation__.Notify(645887)
			}
			__antithesis_instrumentation__.Notify(645884)
			var message string
			if keepErrMsg {
				__antithesis_instrumentation__.Notify(645888)
				message = errmsg.Message
			} else {
				__antithesis_instrumentation__.Notify(645889)
			}
			__antithesis_instrumentation__.Notify(645885)

			msgs = append(msgs, &pgproto3.ErrorResponse{
				Code:           errmsg.Code,
				Message:        message,
				ConstraintName: errmsg.ConstraintName,
			})
			typs = typs[1:]
			continue
		} else {
			__antithesis_instrumentation__.Notify(645890)
		}
		__antithesis_instrumentation__.Notify(645874)

		if msg, ok := recv.(*pgproto3.ReadyForQuery); ok && func() bool {
			__antithesis_instrumentation__.Notify(645891)
			return typ != typReadyForQuery == true
		}() == true {
			__antithesis_instrumentation__.Notify(645892)
			return nil, errors.Errorf("waiting for %T, got %#v", typs[0], msg)
		} else {
			__antithesis_instrumentation__.Notify(645893)
		}
		__antithesis_instrumentation__.Notify(645875)
		if typ == reflect.TypeOf(recv) {
			__antithesis_instrumentation__.Notify(645894)
			typs = typs[1:]
		} else {
			__antithesis_instrumentation__.Notify(645895)
		}
		__antithesis_instrumentation__.Notify(645876)

		var buf bytes.Buffer
		rv := reflect.ValueOf(recv).Elem()
		x := reflect.New(rv.Type())
		if err := gob.NewEncoder(&buf).EncodeValue(rv); err != nil {
			__antithesis_instrumentation__.Notify(645896)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(645897)
		}
		__antithesis_instrumentation__.Notify(645877)
		if err := gob.NewDecoder(&buf).DecodeValue(x); err != nil {
			__antithesis_instrumentation__.Notify(645898)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(645899)
		}
		__antithesis_instrumentation__.Notify(645878)
		msg := x.Interface().(pgproto3.BackendMessage)
		msgs = append(msgs, msg)
	}
	__antithesis_instrumentation__.Notify(645870)
	return msgs, nil
}

var (
	typErrorResponse = reflect.TypeOf(&pgproto3.ErrorResponse{})
	typReadyForQuery = reflect.TypeOf(&pgproto3.ReadyForQuery{})
)
