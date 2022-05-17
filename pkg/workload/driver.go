package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"database/sql/driver"
	"strings"
	"sync/atomic"

	"github.com/lib/pq"
)

type cockroachDriver struct {
	idx uint32
}

func (d *cockroachDriver) Open(name string) (driver.Conn, error) {
	__antithesis_instrumentation__.Notify(694051)
	urls := strings.Split(name, " ")
	i := atomic.AddUint32(&d.idx, 1) - 1
	return pq.Open(urls[i%uint32(len(urls))])
}

func init() {
	gosql.Register("cockroach", &cockroachDriver{})
}
