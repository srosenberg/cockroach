package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"sync/atomic"
)

type RoundRobinDB struct {
	handles []*gosql.DB
	current uint32
}

func NewRoundRobinDB(urls []string) (*RoundRobinDB, error) {
	__antithesis_instrumentation__.Notify(695715)
	r := &RoundRobinDB{current: 0, handles: make([]*gosql.DB, 0, len(urls))}
	for _, url := range urls {
		__antithesis_instrumentation__.Notify(695717)
		db, err := gosql.Open(`cockroach`, url)
		if err != nil {
			__antithesis_instrumentation__.Notify(695719)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(695720)
		}
		__antithesis_instrumentation__.Notify(695718)
		r.handles = append(r.handles, db)
	}
	__antithesis_instrumentation__.Notify(695716)
	return r, nil
}

func (db *RoundRobinDB) next() *gosql.DB {
	__antithesis_instrumentation__.Notify(695721)
	return db.handles[(atomic.AddUint32(&db.current, 1)-1)%uint32(len(db.handles))]
}

func (db *RoundRobinDB) QueryRow(query string, args ...interface{}) *gosql.Row {
	__antithesis_instrumentation__.Notify(695722)
	return db.next().QueryRow(query, args...)
}

func (db *RoundRobinDB) Exec(query string, args ...interface{}) (gosql.Result, error) {
	__antithesis_instrumentation__.Notify(695723)
	return db.next().Exec(query, args...)
}
