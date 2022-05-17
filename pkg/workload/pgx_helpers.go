package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/sync/errgroup"
)

type MultiConnPool struct {
	Pools []*pgxpool.Pool

	counter uint32

	mu struct {
		syncutil.RWMutex

		preparedStatements map[string]string
	}
}

type MultiConnPoolCfg struct {
	MaxTotalConnections int

	MaxConnsPerPool int
}

type pgxLogger struct{}

var _ pgx.Logger = pgxLogger{}

func (p pgxLogger) Log(
	ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{},
) {
	__antithesis_instrumentation__.Notify(694964)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(694968)

		return
	} else {
		__antithesis_instrumentation__.Notify(694969)
	}
	__antithesis_instrumentation__.Notify(694965)
	if strings.Contains(msg, "restart transaction") {
		__antithesis_instrumentation__.Notify(694970)

		return
	} else {
		__antithesis_instrumentation__.Notify(694971)
	}
	__antithesis_instrumentation__.Notify(694966)

	if data != nil {
		__antithesis_instrumentation__.Notify(694972)
		ev := data["err"]
		if err, ok := ev.(error); ok && func() bool {
			__antithesis_instrumentation__.Notify(694973)
			return strings.Contains(err.Error(), "restart transaction") == true
		}() == true {
			__antithesis_instrumentation__.Notify(694974)
			return
		} else {
			__antithesis_instrumentation__.Notify(694975)
		}
	} else {
		__antithesis_instrumentation__.Notify(694976)
	}
	__antithesis_instrumentation__.Notify(694967)
	log.Infof(ctx, "pgx logger [%s]: %s logParams=%v", level.String(), msg, data)
}

func NewMultiConnPool(
	ctx context.Context, cfg MultiConnPoolCfg, urls ...string,
) (*MultiConnPool, error) {
	__antithesis_instrumentation__.Notify(694977)
	m := &MultiConnPool{}
	m.mu.preparedStatements = map[string]string{}

	connsPerURL := distribute(cfg.MaxTotalConnections, len(urls))
	maxConnsPerPool := cfg.MaxConnsPerPool
	if maxConnsPerPool == 0 {
		__antithesis_instrumentation__.Notify(694983)
		maxConnsPerPool = cfg.MaxTotalConnections
	} else {
		__antithesis_instrumentation__.Notify(694984)
	}
	__antithesis_instrumentation__.Notify(694978)

	var warmupConns [][]*pgxpool.Conn
	for i := range urls {
		__antithesis_instrumentation__.Notify(694985)
		connsPerPool := distributeMax(connsPerURL[i], maxConnsPerPool)
		for _, numConns := range connsPerPool {
			__antithesis_instrumentation__.Notify(694986)
			connCfg, err := pgxpool.ParseConfig(urls[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(694990)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(694991)
			}
			__antithesis_instrumentation__.Notify(694987)

			connCfg.ConnConfig.BuildStatementCache = nil
			connCfg.ConnConfig.LogLevel = pgx.LogLevelWarn
			connCfg.ConnConfig.Logger = pgxLogger{}
			connCfg.MaxConns = int32(numConns)
			connCfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
				__antithesis_instrumentation__.Notify(694992)
				m.mu.RLock()
				defer m.mu.RUnlock()
				for name, sql := range m.mu.preparedStatements {
					__antithesis_instrumentation__.Notify(694994)

					if _, err := conn.Prepare(ctx, name, sql); err != nil {
						__antithesis_instrumentation__.Notify(694995)
						log.Warningf(ctx, "error preparing statement. name=%s sql=%s %v", name, sql, err)
						return false
					} else {
						__antithesis_instrumentation__.Notify(694996)
					}
				}
				__antithesis_instrumentation__.Notify(694993)
				return true
			}
			__antithesis_instrumentation__.Notify(694988)
			p, err := pgxpool.ConnectConfig(ctx, connCfg)
			if err != nil {
				__antithesis_instrumentation__.Notify(694997)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(694998)
			}
			__antithesis_instrumentation__.Notify(694989)

			warmupConns = append(warmupConns, make([]*pgxpool.Conn, numConns))
			m.Pools = append(m.Pools, p)
		}
	}
	__antithesis_instrumentation__.Notify(694979)

	var g errgroup.Group

	sem := make(chan struct{}, 100)
	for i, p := range m.Pools {
		__antithesis_instrumentation__.Notify(694999)
		p := p
		conns := warmupConns[i]
		for j := range conns {
			__antithesis_instrumentation__.Notify(695000)
			j := j
			sem <- struct{}{}
			g.Go(func() error {
				__antithesis_instrumentation__.Notify(695001)
				var err error
				conns[j], err = p.Acquire(ctx)
				<-sem
				return err
			})
		}
	}
	__antithesis_instrumentation__.Notify(694980)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(695002)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(695003)
	}
	__antithesis_instrumentation__.Notify(694981)
	for i := range m.Pools {
		__antithesis_instrumentation__.Notify(695004)
		for _, c := range warmupConns[i] {
			__antithesis_instrumentation__.Notify(695005)
			c.Release()
		}
	}
	__antithesis_instrumentation__.Notify(694982)

	return m, nil
}

func (m *MultiConnPool) AddPreparedStatement(name string, statement string) {
	__antithesis_instrumentation__.Notify(695006)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mu.preparedStatements[name] = statement
}

func (m *MultiConnPool) Get() *pgxpool.Pool {
	__antithesis_instrumentation__.Notify(695007)
	if len(m.Pools) == 1 {
		__antithesis_instrumentation__.Notify(695009)
		return m.Pools[0]
	} else {
		__antithesis_instrumentation__.Notify(695010)
	}
	__antithesis_instrumentation__.Notify(695008)
	i := atomic.AddUint32(&m.counter, 1) - 1
	return m.Pools[i%uint32(len(m.Pools))]
}

func (m *MultiConnPool) Close() {
	__antithesis_instrumentation__.Notify(695011)
	for _, p := range m.Pools {
		__antithesis_instrumentation__.Notify(695012)
		p.Close()
	}
}

func distribute(total, num int) []int {
	__antithesis_instrumentation__.Notify(695013)
	res := make([]int, num)
	for i := range res {
		__antithesis_instrumentation__.Notify(695015)

		div := len(res) - i
		res[i] = (total + div/2) / div
		total -= res[i]
	}
	__antithesis_instrumentation__.Notify(695014)
	return res
}

func distributeMax(total, max int) []int {
	__antithesis_instrumentation__.Notify(695016)
	return distribute(total, (total+max-1)/max)
}
