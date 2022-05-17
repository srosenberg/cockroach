package bench

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os/exec"

	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
)

const tpcbQuery = `BEGIN;
UPDATE pgbench_accounts SET abalance = abalance + %[1]d WHERE aid = %[2]d;
SELECT abalance FROM pgbench_accounts WHERE aid = %[2]d;
UPDATE pgbench_tellers SET tbalance = tbalance + %[1]d WHERE tid = %[3]d;
UPDATE pgbench_branches SET bbalance = bbalance + %[1]d WHERE bid = %[4]d;
INSERT INTO pgbench_history (tid, bid, aid, delta, mtime) VALUES (%[3]d, %[4]d, %[2]d, %[1]d, CURRENT_TIMESTAMP);
END;`

func RunOne(db sqlutils.DBHandle, r *rand.Rand, accounts int) error {
	__antithesis_instrumentation__.Notify(1814)
	account := r.Intn(accounts)
	delta := r.Intn(5000)
	teller := r.Intn(tellers)
	branch := 1

	q := fmt.Sprintf(tpcbQuery, delta, account, teller, branch)
	_, err := db.ExecContext(context.TODO(), q)
	return err
}

func ExecPgbench(pgURL url.URL, dbname string, count int) (*exec.Cmd, error) {
	__antithesis_instrumentation__.Notify(1815)
	host, port, err := net.SplitHostPort(pgURL.Host)
	if err != nil {
		__antithesis_instrumentation__.Notify(1818)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(1819)
	}
	__antithesis_instrumentation__.Notify(1816)

	args := []string{
		"-n",
		"-r",
		fmt.Sprintf("--transactions=%d", count),
		"-h", host,
		"-p", port,
	}

	if pgURL.User != nil {
		__antithesis_instrumentation__.Notify(1820)
		if user := pgURL.User.Username(); user != "" {
			__antithesis_instrumentation__.Notify(1821)
			args = append(args, "-U", user)
		} else {
			__antithesis_instrumentation__.Notify(1822)
		}
	} else {
		__antithesis_instrumentation__.Notify(1823)
	}
	__antithesis_instrumentation__.Notify(1817)
	args = append(args, dbname)

	return exec.Command("pgbench", args...), nil
}
