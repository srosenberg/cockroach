package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

func RunNemesis(f TestFeedFactory, db *gosql.DB, isSinkless bool) (Validator, error) {
	__antithesis_instrumentation__.Notify(14728)

	ctx := context.Background()
	rng, _ := randutil.NewPseudoRand()

	eventPauseCount := 10
	if isSinkless {
		__antithesis_instrumentation__.Notify(14741)

		eventPauseCount = 0
	} else {
		__antithesis_instrumentation__.Notify(14742)
	}
	__antithesis_instrumentation__.Notify(14729)
	ns := &nemeses{
		maxTestColumnCount: 10,
		rowCount:           4,
		db:                 db,

		eventMix: map[fsm.Event]int{

			eventFinished{}: 0,

			eventFeedMessage{}: 50,

			eventSplit{}: 5,

			eventOpenTxn{}: 10,

			eventCommit{}: 5,

			eventRollback{}: 5,

			eventPush{}: 5,

			eventAbort{}: 5,

			eventPause{}: eventPauseCount,

			eventResume{}: 50,

			eventAddColumn{
				CanAddColumnAfter: fsm.True,
			}: 5,

			eventAddColumn{
				CanAddColumnAfter: fsm.False,
			}: 5,

			eventRemoveColumn{
				CanRemoveColumnAfter: fsm.True,
			}: 5,

			eventRemoveColumn{
				CanRemoveColumnAfter: fsm.False,
			}: 5,

			eventCreateEnum{}: 5,
		},
	}

	if _, err := db.Exec(`CREATE TABLE foo (id INT PRIMARY KEY, ts STRING DEFAULT '0')`); err != nil {
		__antithesis_instrumentation__.Notify(14743)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14744)
	}
	__antithesis_instrumentation__.Notify(14730)
	if _, err := db.Exec(`SET CLUSTER SETTING kv.range_merge.queue_enabled = false`); err != nil {
		__antithesis_instrumentation__.Notify(14745)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14746)
	}
	__antithesis_instrumentation__.Notify(14731)
	if _, err := db.Exec(`ALTER TABLE foo SPLIT AT VALUES ($1)`, ns.rowCount/2); err != nil {
		__antithesis_instrumentation__.Notify(14747)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14748)
	}
	__antithesis_instrumentation__.Notify(14732)

	for i := 0; i < ns.rowCount*5; i++ {
		__antithesis_instrumentation__.Notify(14749)
		payload, err := newOpenTxnPayload(ns)
		if err != nil {
			__antithesis_instrumentation__.Notify(14752)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14753)
		}
		__antithesis_instrumentation__.Notify(14750)
		if err := openTxn(fsm.Args{Ctx: ctx, Extended: ns, Payload: payload}); err != nil {
			__antithesis_instrumentation__.Notify(14754)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14755)
		}
		__antithesis_instrumentation__.Notify(14751)

		if rand.Intn(3) < 2 || func() bool {
			__antithesis_instrumentation__.Notify(14756)
			return i == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(14757)
			if err := commit(fsm.Args{Ctx: ctx, Extended: ns}); err != nil {
				__antithesis_instrumentation__.Notify(14758)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14759)
			}
		} else {
			__antithesis_instrumentation__.Notify(14760)
			if err := rollback(fsm.Args{Ctx: ctx, Extended: ns}); err != nil {
				__antithesis_instrumentation__.Notify(14761)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14762)
			}
		}
	}
	__antithesis_instrumentation__.Notify(14733)

	foo, err := f.Feed(`CREATE CHANGEFEED FOR foo WITH updated, resolved, diff`)
	if err != nil {
		__antithesis_instrumentation__.Notify(14763)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14764)
	}
	__antithesis_instrumentation__.Notify(14734)
	ns.f = foo
	defer func() { __antithesis_instrumentation__.Notify(14765); _ = foo.Close() }()
	__antithesis_instrumentation__.Notify(14735)

	scratchTableName := `fprint`
	var createFprintStmtBuf bytes.Buffer
	fmt.Fprintf(&createFprintStmtBuf, `CREATE TABLE %s (id INT PRIMARY KEY, ts STRING)`, scratchTableName)
	if _, err := db.Exec(createFprintStmtBuf.String()); err != nil {
		__antithesis_instrumentation__.Notify(14766)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14767)
	}
	__antithesis_instrumentation__.Notify(14736)
	baV, err := NewBeforeAfterValidator(db, `foo`)
	if err != nil {
		__antithesis_instrumentation__.Notify(14768)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14769)
	}
	__antithesis_instrumentation__.Notify(14737)
	fprintV, err := NewFingerprintValidator(db, `foo`, scratchTableName, foo.Partitions(), ns.maxTestColumnCount)
	if err != nil {
		__antithesis_instrumentation__.Notify(14770)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14771)
	}
	__antithesis_instrumentation__.Notify(14738)
	ns.v = MakeCountValidator(Validators{
		NewOrderValidator(`foo`),
		baV,
		fprintV,
	})

	if err := db.QueryRow(`SELECT count(*) FROM foo`).Scan(&ns.availableRows); err != nil {
		__antithesis_instrumentation__.Notify(14772)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14773)
	}
	__antithesis_instrumentation__.Notify(14739)

	txnOpenBeforeInitialScan := false

	if rand.Intn(2) < 1 {
		__antithesis_instrumentation__.Notify(14774)
		txnOpenBeforeInitialScan = true
		payload, err := newOpenTxnPayload(ns)
		if err != nil {
			__antithesis_instrumentation__.Notify(14776)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14777)
		}
		__antithesis_instrumentation__.Notify(14775)
		if err := openTxn(fsm.Args{Ctx: ctx, Extended: ns, Payload: payload}); err != nil {
			__antithesis_instrumentation__.Notify(14778)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14779)
		}
	} else {
		__antithesis_instrumentation__.Notify(14780)
	}
	__antithesis_instrumentation__.Notify(14740)

	initialState := stateRunning{
		FeedPaused:      fsm.False,
		TxnOpen:         fsm.FromBool(txnOpenBeforeInitialScan),
		CanAddColumn:    fsm.True,
		CanRemoveColumn: fsm.False,
	}
	m := fsm.MakeMachine(compiledStateTransitions, initialState, ns)
	for {
		__antithesis_instrumentation__.Notify(14781)
		state := m.CurState()
		if _, ok := state.(stateDone); ok {
			__antithesis_instrumentation__.Notify(14784)
			return ns.v, nil
		} else {
			__antithesis_instrumentation__.Notify(14785)
		}
		__antithesis_instrumentation__.Notify(14782)
		event, eventPayload, err := ns.nextEvent(rng, state, &m)
		if err != nil {
			__antithesis_instrumentation__.Notify(14786)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14787)
		}
		__antithesis_instrumentation__.Notify(14783)
		if err := m.ApplyWithPayload(ctx, event, eventPayload); err != nil {
			__antithesis_instrumentation__.Notify(14788)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14789)
		}
	}
}

type openTxnType string

const (
	openTxnTypeUpsert openTxnType = `UPSERT`
	openTxnTypeDelete openTxnType = `DELETE`
)

type openTxnPayload struct {
	openTxnType openTxnType

	rowID int
}

type addColumnType string

const (
	addColumnTypeString addColumnType = "string"
	addColumnTypeEnum   addColumnType = "enum"
)

type addColumnPayload struct {
	columnType addColumnType

	enum int
}

type nemeses struct {
	rowCount           int
	maxTestColumnCount int
	eventMix           map[fsm.Event]int

	v  *CountValidator
	db *gosql.DB
	f  TestFeed

	availableRows          int
	currentTestColumnCount int
	txn                    *gosql.Tx
	openTxnType            openTxnType
	openTxnID              int
	openTxnTs              string

	enumCount int
}

func (ns *nemeses) nextEvent(
	rng *rand.Rand, state fsm.State, m *fsm.Machine,
) (se fsm.Event, payload fsm.EventPayload, err error) {
	__antithesis_instrumentation__.Notify(14790)
	var noPayload interface{}
	if ns.v.NumResolvedWithRows >= 6 && func() bool {
		__antithesis_instrumentation__.Notify(14795)
		return ns.v.NumResolvedRows >= 10 == true
	}() == true {
		__antithesis_instrumentation__.Notify(14796)
		return eventFinished{}, noPayload, nil
	} else {
		__antithesis_instrumentation__.Notify(14797)
	}
	__antithesis_instrumentation__.Notify(14791)
	possibleEvents, ok := compiledStateTransitions.GetExpanded()[state]
	if !ok {
		__antithesis_instrumentation__.Notify(14798)
		return nil, noPayload, errors.Errorf(`unknown state: %T %s`, state, state)
	} else {
		__antithesis_instrumentation__.Notify(14799)
	}
	__antithesis_instrumentation__.Notify(14792)
	mixTotal := 0
	for event := range possibleEvents {
		__antithesis_instrumentation__.Notify(14800)
		weight, ok := ns.eventMix[event]
		if !ok {
			__antithesis_instrumentation__.Notify(14802)
			return nil, noPayload, errors.Errorf(`unknown event: %T`, event)
		} else {
			__antithesis_instrumentation__.Notify(14803)
		}
		__antithesis_instrumentation__.Notify(14801)
		mixTotal += weight
	}
	__antithesis_instrumentation__.Notify(14793)
	r, t := rng.Intn(mixTotal), 0
	for event := range possibleEvents {
		__antithesis_instrumentation__.Notify(14804)
		t += ns.eventMix[event]
		if r >= t {
			__antithesis_instrumentation__.Notify(14810)
			continue
		} else {
			__antithesis_instrumentation__.Notify(14811)
		}
		__antithesis_instrumentation__.Notify(14805)
		if _, ok := event.(eventFeedMessage); ok {
			__antithesis_instrumentation__.Notify(14812)

			if ns.availableRows < 1 {
				__antithesis_instrumentation__.Notify(14814)
				s := state.(stateRunning)
				if s.TxnOpen.Get() {
					__antithesis_instrumentation__.Notify(14817)
					return eventCommit{}, noPayload, nil
				} else {
					__antithesis_instrumentation__.Notify(14818)
				}
				__antithesis_instrumentation__.Notify(14815)
				payload, err := newOpenTxnPayload(ns)
				if err != nil {
					__antithesis_instrumentation__.Notify(14819)
					return eventOpenTxn{}, noPayload, err
				} else {
					__antithesis_instrumentation__.Notify(14820)
				}
				__antithesis_instrumentation__.Notify(14816)
				return eventOpenTxn{}, payload, nil
			} else {
				__antithesis_instrumentation__.Notify(14821)
			}
			__antithesis_instrumentation__.Notify(14813)
			return eventFeedMessage{}, noPayload, nil
		} else {
			__antithesis_instrumentation__.Notify(14822)
		}
		__antithesis_instrumentation__.Notify(14806)
		if _, ok := event.(eventOpenTxn); ok {
			__antithesis_instrumentation__.Notify(14823)
			payload, err := newOpenTxnPayload(ns)
			if err != nil {
				__antithesis_instrumentation__.Notify(14825)
				return eventOpenTxn{}, noPayload, err
			} else {
				__antithesis_instrumentation__.Notify(14826)
			}
			__antithesis_instrumentation__.Notify(14824)
			return eventOpenTxn{}, payload, nil
		} else {
			__antithesis_instrumentation__.Notify(14827)
		}
		__antithesis_instrumentation__.Notify(14807)
		if e, ok := event.(eventAddColumn); ok {
			__antithesis_instrumentation__.Notify(14828)
			e.CanAddColumnAfter = fsm.FromBool(ns.currentTestColumnCount < ns.maxTestColumnCount-1)
			payload := addColumnPayload{}
			if ns.enumCount > 0 && func() bool {
				__antithesis_instrumentation__.Notify(14830)
				return rng.Intn(4) < 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(14831)
				payload.columnType = addColumnTypeEnum
				payload.enum = rng.Intn(ns.enumCount)
			} else {
				__antithesis_instrumentation__.Notify(14832)
				payload.columnType = addColumnTypeString
			}
			__antithesis_instrumentation__.Notify(14829)
			return e, payload, nil
		} else {
			__antithesis_instrumentation__.Notify(14833)
		}
		__antithesis_instrumentation__.Notify(14808)
		if e, ok := event.(eventRemoveColumn); ok {
			__antithesis_instrumentation__.Notify(14834)
			e.CanRemoveColumnAfter = fsm.FromBool(ns.currentTestColumnCount > 1)
			return e, noPayload, nil
		} else {
			__antithesis_instrumentation__.Notify(14835)
		}
		__antithesis_instrumentation__.Notify(14809)
		return event, noPayload, nil
	}
	__antithesis_instrumentation__.Notify(14794)

	panic(`unreachable`)
}

func newOpenTxnPayload(ns *nemeses) (openTxnPayload, error) {
	__antithesis_instrumentation__.Notify(14836)
	payload := openTxnPayload{}
	if rand.Intn(10) == 0 {
		__antithesis_instrumentation__.Notify(14838)
		rows, err := ns.db.Query(`SELECT id FROM foo ORDER BY random() LIMIT 1`)
		if err != nil {
			__antithesis_instrumentation__.Notify(14842)
			return payload, err
		} else {
			__antithesis_instrumentation__.Notify(14843)
		}
		__antithesis_instrumentation__.Notify(14839)
		defer func() { __antithesis_instrumentation__.Notify(14844); _ = rows.Close() }()
		__antithesis_instrumentation__.Notify(14840)
		if rows.Next() {
			__antithesis_instrumentation__.Notify(14845)
			var deleteID int
			if err := rows.Scan(&deleteID); err != nil {
				__antithesis_instrumentation__.Notify(14847)
				return payload, err
			} else {
				__antithesis_instrumentation__.Notify(14848)
			}
			__antithesis_instrumentation__.Notify(14846)
			payload.rowID = deleteID
			payload.openTxnType = openTxnTypeDelete
			return payload, nil
		} else {
			__antithesis_instrumentation__.Notify(14849)
		}
		__antithesis_instrumentation__.Notify(14841)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(14850)
			return payload, err
		} else {
			__antithesis_instrumentation__.Notify(14851)
		}

	} else {
		__antithesis_instrumentation__.Notify(14852)
	}
	__antithesis_instrumentation__.Notify(14837)

	payload.rowID = rand.Intn(ns.rowCount)
	payload.openTxnType = openTxnTypeUpsert
	return payload, nil
}

type stateRunning struct {
	FeedPaused      fsm.Bool
	TxnOpen         fsm.Bool
	CanRemoveColumn fsm.Bool
	CanAddColumn    fsm.Bool
}
type stateDone struct{}

func (stateRunning) State() { __antithesis_instrumentation__.Notify(14853) }
func (stateDone) State()    { __antithesis_instrumentation__.Notify(14854) }

type eventOpenTxn struct{}
type eventFeedMessage struct{}
type eventPause struct{}
type eventResume struct{}
type eventCommit struct{}
type eventPush struct{}
type eventAbort struct{}
type eventRollback struct{}
type eventSplit struct{}
type eventAddColumn struct {
	CanAddColumnAfter fsm.Bool
}
type eventRemoveColumn struct {
	CanRemoveColumnAfter fsm.Bool
}
type eventCreateEnum struct{}
type eventFinished struct{}

func (eventOpenTxn) Event()      { __antithesis_instrumentation__.Notify(14855) }
func (eventFeedMessage) Event()  { __antithesis_instrumentation__.Notify(14856) }
func (eventPause) Event()        { __antithesis_instrumentation__.Notify(14857) }
func (eventResume) Event()       { __antithesis_instrumentation__.Notify(14858) }
func (eventCommit) Event()       { __antithesis_instrumentation__.Notify(14859) }
func (eventPush) Event()         { __antithesis_instrumentation__.Notify(14860) }
func (eventAbort) Event()        { __antithesis_instrumentation__.Notify(14861) }
func (eventRollback) Event()     { __antithesis_instrumentation__.Notify(14862) }
func (eventSplit) Event()        { __antithesis_instrumentation__.Notify(14863) }
func (eventAddColumn) Event()    { __antithesis_instrumentation__.Notify(14864) }
func (eventRemoveColumn) Event() { __antithesis_instrumentation__.Notify(14865) }
func (eventCreateEnum) Event()   { __antithesis_instrumentation__.Notify(14866) }
func (eventFinished) Event()     { __antithesis_instrumentation__.Notify(14867) }

var stateTransitions = fsm.Pattern{
	stateRunning{
		FeedPaused:      fsm.Var("FeedPaused"),
		TxnOpen:         fsm.Var("TxnOpen"),
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventSplit{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.Var("TxnOpen"),
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(split),
		},
		eventFinished{}: {
			Next:   stateDone{},
			Action: logEvent(cleanup),
		},
	},
	stateRunning{
		FeedPaused:      fsm.Var("FeedPaused"),
		TxnOpen:         fsm.False,
		CanAddColumn:    fsm.True,
		CanRemoveColumn: fsm.Any,
	}: {
		eventAddColumn{
			CanAddColumnAfter: fsm.Var("CanAddColumnAfter"),
		}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.Var("CanAddColumnAfter"),
				CanRemoveColumn: fsm.True},
			Action: logEvent(addColumn),
		},
	},
	stateRunning{
		FeedPaused:      fsm.Var("FeedPaused"),
		TxnOpen:         fsm.False,
		CanAddColumn:    fsm.Any,
		CanRemoveColumn: fsm.True,
	}: {
		eventRemoveColumn{
			CanRemoveColumnAfter: fsm.Var("CanRemoveColumnAfter"),
		}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.True,
				CanRemoveColumn: fsm.Var("CanRemoveColumnAfter")},
			Action: logEvent(removeColumn),
		},
	},
	stateRunning{
		FeedPaused:      fsm.False,
		TxnOpen:         fsm.False,
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventFeedMessage{}: {
			Next: stateRunning{
				FeedPaused:      fsm.False,
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(noteFeedMessage),
		},
	},
	stateRunning{
		FeedPaused:      fsm.Var("FeedPaused"),
		TxnOpen:         fsm.False,
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventOpenTxn{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.True,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(openTxn),
		},
		eventCreateEnum{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn"),
			},
			Action: logEvent(createEnum),
		},
	},
	stateRunning{
		FeedPaused:      fsm.Var("FeedPaused"),
		TxnOpen:         fsm.True,
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventCommit{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(commit),
		},
		eventRollback{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.False,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(rollback),
		},
		eventAbort{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.True,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(abort),
		},
		eventPush{}: {
			Next: stateRunning{
				FeedPaused:      fsm.Var("FeedPaused"),
				TxnOpen:         fsm.True,
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(push),
		},
	},
	stateRunning{
		FeedPaused:      fsm.False,
		TxnOpen:         fsm.Var("TxnOpen"),
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventPause{}: {
			Next: stateRunning{
				FeedPaused:      fsm.True,
				TxnOpen:         fsm.Var("TxnOpen"),
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(pause),
		},
	},
	stateRunning{
		FeedPaused:      fsm.True,
		TxnOpen:         fsm.Var("TxnOpen"),
		CanAddColumn:    fsm.Var("CanAddColumn"),
		CanRemoveColumn: fsm.Var("CanRemoveColumn"),
	}: {
		eventResume{}: {
			Next: stateRunning{
				FeedPaused:      fsm.False,
				TxnOpen:         fsm.Var("TxnOpen"),
				CanAddColumn:    fsm.Var("CanAddColumn"),
				CanRemoveColumn: fsm.Var("CanRemoveColumn")},
			Action: logEvent(resume),
		},
	},
}

var compiledStateTransitions = fsm.Compile(stateTransitions)

func logEvent(fn func(fsm.Args) error) func(fsm.Args) error {
	__antithesis_instrumentation__.Notify(14868)
	return func(a fsm.Args) error {
		__antithesis_instrumentation__.Notify(14869)
		log.Infof(a.Ctx, "Event: %#v, Payload: %#v\n", a.Event, a.Payload)
		return fn(a)
	}
}

func cleanup(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14870)
	if txn := a.Extended.(*nemeses).txn; txn != nil {
		__antithesis_instrumentation__.Notify(14872)
		return txn.Rollback()
	} else {
		__antithesis_instrumentation__.Notify(14873)
	}
	__antithesis_instrumentation__.Notify(14871)
	return nil
}

func openTxn(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14874)
	ns := a.Extended.(*nemeses)
	payload := a.Payload.(openTxnPayload)

	txn, err := ns.db.Begin()
	if err != nil {
		__antithesis_instrumentation__.Notify(14877)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14878)
	}
	__antithesis_instrumentation__.Notify(14875)
	switch payload.openTxnType {
	case openTxnTypeUpsert:
		__antithesis_instrumentation__.Notify(14879)
		if err := txn.QueryRow(
			`UPSERT INTO foo VALUES ($1, cluster_logical_timestamp()::string) RETURNING id, ts`,
			payload.rowID,
		).Scan(&ns.openTxnID, &ns.openTxnTs); err != nil {
			__antithesis_instrumentation__.Notify(14882)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14883)
		}
	case openTxnTypeDelete:
		__antithesis_instrumentation__.Notify(14880)
		if err := txn.QueryRow(
			`DELETE FROM foo WHERE id = $1 RETURNING id, ts`,
			payload.rowID,
		).Scan(&ns.openTxnID, &ns.openTxnTs); err != nil {
			__antithesis_instrumentation__.Notify(14884)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14885)
		}
	default:
		__antithesis_instrumentation__.Notify(14881)
		panic("unreachable")
	}
	__antithesis_instrumentation__.Notify(14876)
	ns.openTxnType = payload.openTxnType
	ns.txn = txn
	return nil
}

func commit(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14886)
	ns := a.Extended.(*nemeses)
	defer func() { __antithesis_instrumentation__.Notify(14889); ns.txn = nil }()
	__antithesis_instrumentation__.Notify(14887)
	if err := ns.txn.Commit(); err != nil {
		__antithesis_instrumentation__.Notify(14890)

		if strings.Contains(err.Error(), `restart transaction`) {
			__antithesis_instrumentation__.Notify(14891)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(14892)
		}
	} else {
		__antithesis_instrumentation__.Notify(14893)
	}
	__antithesis_instrumentation__.Notify(14888)
	ns.availableRows++
	return nil
}

func rollback(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14894)
	ns := a.Extended.(*nemeses)
	defer func() { __antithesis_instrumentation__.Notify(14896); ns.txn = nil }()
	__antithesis_instrumentation__.Notify(14895)
	return ns.txn.Rollback()
}

func createEnum(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14897)
	ns := a.Extended.(*nemeses)

	if _, err := ns.db.Exec(fmt.Sprintf(`CREATE TYPE enum%d AS ENUM ('hello')`, ns.enumCount)); err != nil {
		__antithesis_instrumentation__.Notify(14899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14900)
	}
	__antithesis_instrumentation__.Notify(14898)
	ns.enumCount++
	return nil
}

func addColumn(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14901)
	ns := a.Extended.(*nemeses)
	payload := a.Payload.(addColumnPayload)

	if ns.currentTestColumnCount >= ns.maxTestColumnCount {
		__antithesis_instrumentation__.Notify(14905)
		return errors.AssertionFailedf(`addColumn should be called when`+
			`there are less than %d columns.`, ns.maxTestColumnCount)
	} else {
		__antithesis_instrumentation__.Notify(14906)
	}
	__antithesis_instrumentation__.Notify(14902)

	switch payload.columnType {
	case addColumnTypeEnum:
		__antithesis_instrumentation__.Notify(14907)

		enum := payload.enum
		if _, err := ns.db.Exec(fmt.Sprintf(`ALTER TABLE foo ADD COLUMN test%d enum%d DEFAULT 'hello'`,
			ns.currentTestColumnCount, enum)); err != nil {
			__antithesis_instrumentation__.Notify(14910)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14911)
		}
	case addColumnTypeString:
		__antithesis_instrumentation__.Notify(14908)
		if _, err := ns.db.Exec(fmt.Sprintf(`ALTER TABLE foo ADD COLUMN test%d STRING DEFAULT 'x'`,
			ns.currentTestColumnCount)); err != nil {
			__antithesis_instrumentation__.Notify(14912)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14913)
		}
	default:
		__antithesis_instrumentation__.Notify(14909)
	}
	__antithesis_instrumentation__.Notify(14903)

	ns.currentTestColumnCount++
	var rows int

	if err := ns.db.QueryRow(`SELECT count(*) FROM foo`).Scan(&rows); err != nil {
		__antithesis_instrumentation__.Notify(14914)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14915)
	}
	__antithesis_instrumentation__.Notify(14904)

	ns.availableRows += 2 * rows
	return nil
}

func removeColumn(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14916)
	ns := a.Extended.(*nemeses)

	if ns.currentTestColumnCount == 0 {
		__antithesis_instrumentation__.Notify(14920)
		return errors.AssertionFailedf(`removeColumn should be called with` +
			`at least one test column.`)
	} else {
		__antithesis_instrumentation__.Notify(14921)
	}
	__antithesis_instrumentation__.Notify(14917)
	if _, err := ns.db.Exec(fmt.Sprintf(`ALTER TABLE foo DROP COLUMN test%d`,
		ns.currentTestColumnCount-1)); err != nil {
		__antithesis_instrumentation__.Notify(14922)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14923)
	}
	__antithesis_instrumentation__.Notify(14918)
	ns.currentTestColumnCount--
	var rows int

	if err := ns.db.QueryRow(`SELECT count(*) FROM foo`).Scan(&rows); err != nil {
		__antithesis_instrumentation__.Notify(14924)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14925)
	}
	__antithesis_instrumentation__.Notify(14919)

	ns.availableRows += 2 * rows
	return nil
}

func noteFeedMessage(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14926)
	ns := a.Extended.(*nemeses)

	if ns.availableRows <= 0 {
		__antithesis_instrumentation__.Notify(14928)
		return errors.AssertionFailedf(`noteFeedMessage should be called with at` +
			`least one available row.`)
	} else {
		__antithesis_instrumentation__.Notify(14929)
	}
	__antithesis_instrumentation__.Notify(14927)
	for {
		__antithesis_instrumentation__.Notify(14930)
		m, err := ns.f.Next()
		if err != nil {
			__antithesis_instrumentation__.Notify(14932)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14933)
			if m == nil {
				__antithesis_instrumentation__.Notify(14934)
				return errors.Errorf(`expected another message`)
			} else {
				__antithesis_instrumentation__.Notify(14935)
			}
		}
		__antithesis_instrumentation__.Notify(14931)

		if len(m.Resolved) > 0 {
			__antithesis_instrumentation__.Notify(14936)
			_, ts, err := ParseJSONValueTimestamps(m.Resolved)
			if err != nil {
				__antithesis_instrumentation__.Notify(14938)
				return err
			} else {
				__antithesis_instrumentation__.Notify(14939)
			}
			__antithesis_instrumentation__.Notify(14937)
			log.Infof(a.Ctx, "%v", string(m.Resolved))
			err = ns.v.NoteResolved(m.Partition, ts)
			if err != nil {
				__antithesis_instrumentation__.Notify(14940)
				return err
			} else {
				__antithesis_instrumentation__.Notify(14941)
			}

		} else {
			__antithesis_instrumentation__.Notify(14942)
			ts, _, err := ParseJSONValueTimestamps(m.Value)
			if err != nil {
				__antithesis_instrumentation__.Notify(14944)
				return err
			} else {
				__antithesis_instrumentation__.Notify(14945)
			}
			__antithesis_instrumentation__.Notify(14943)
			ns.availableRows--
			log.Infof(a.Ctx, "%s->%s", m.Key, m.Value)
			return ns.v.NoteRow(m.Partition, string(m.Key), string(m.Value), ts)
		}
	}
}

func pause(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14946)
	return a.Extended.(*nemeses).f.(EnterpriseTestFeed).Pause()
}

func resume(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14947)
	return a.Extended.(*nemeses).f.(EnterpriseTestFeed).Resume()
}

func abort(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14948)
	ns := a.Extended.(*nemeses)
	const delete = `BEGIN TRANSACTION PRIORITY HIGH; ` +
		`SELECT count(*) FROM [DELETE FROM foo RETURNING *]; ` +
		`COMMIT`
	var deletedRows int
	if err := ns.db.QueryRow(delete).Scan(&deletedRows); err != nil {
		__antithesis_instrumentation__.Notify(14950)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14951)
	}
	__antithesis_instrumentation__.Notify(14949)
	ns.availableRows += deletedRows
	return nil
}

func push(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14952)
	ns := a.Extended.(*nemeses)
	_, err := ns.db.Exec(`BEGIN TRANSACTION PRIORITY HIGH; SELECT * FROM foo; COMMIT`)
	return err
}

func split(a fsm.Args) error {
	__antithesis_instrumentation__.Notify(14953)
	ns := a.Extended.(*nemeses)
	_, err := ns.db.Exec(`ALTER TABLE foo SPLIT AT VALUES ((random() * $1)::int)`, ns.rowCount)
	return err
}
