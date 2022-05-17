package workloadsql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"golang.org/x/time/rate"
)

func Setup(
	ctx context.Context, db *gosql.DB, gen workload.Generator, l workload.InitialDataLoader,
) (int64, error) {
	__antithesis_instrumentation__.Notify(699149)
	var hooks workload.Hooks
	if h, ok := gen.(workload.Hookser); ok {
		__antithesis_instrumentation__.Notify(699155)
		hooks = h.Hooks()
	} else {
		__antithesis_instrumentation__.Notify(699156)
	}
	__antithesis_instrumentation__.Notify(699150)

	if hooks.PreCreate != nil {
		__antithesis_instrumentation__.Notify(699157)
		if err := hooks.PreCreate(db); err != nil {
			__antithesis_instrumentation__.Notify(699158)
			return 0, errors.Wrapf(err, "Could not pre-create")
		} else {
			__antithesis_instrumentation__.Notify(699159)
		}
	} else {
		__antithesis_instrumentation__.Notify(699160)
	}
	__antithesis_instrumentation__.Notify(699151)

	bytes, err := l.InitialDataLoad(ctx, db, gen)
	if err != nil {
		__antithesis_instrumentation__.Notify(699161)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(699162)
	}
	__antithesis_instrumentation__.Notify(699152)

	const splitConcurrency = 384
	for _, table := range gen.Tables() {
		__antithesis_instrumentation__.Notify(699163)
		if err := Split(ctx, db, table, splitConcurrency); err != nil {
			__antithesis_instrumentation__.Notify(699164)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(699165)
		}
	}
	__antithesis_instrumentation__.Notify(699153)

	if hooks.PostLoad != nil {
		__antithesis_instrumentation__.Notify(699166)
		if err := hooks.PostLoad(db); err != nil {
			__antithesis_instrumentation__.Notify(699167)
			return 0, errors.Wrapf(err, "Could not postload")
		} else {
			__antithesis_instrumentation__.Notify(699168)
		}
	} else {
		__antithesis_instrumentation__.Notify(699169)
	}
	__antithesis_instrumentation__.Notify(699154)

	return bytes, nil
}

func maybeDisableMergeQueue(db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(699170)
	var ok bool
	if err := db.QueryRow(
		`SELECT count(*) > 0 FROM [ SHOW ALL CLUSTER SETTINGS ] AS _ (v) WHERE v = 'kv.range_merge.queue_enabled'`,
	).Scan(&ok); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(699174)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(699175)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699176)
	}
	__antithesis_instrumentation__.Notify(699171)
	var versionStr string
	err := db.QueryRow(
		`SELECT value FROM crdb_internal.node_build_info WHERE field = 'Version'`,
	).Scan(&versionStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(699177)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699178)
	}
	__antithesis_instrumentation__.Notify(699172)
	v, err := version.Parse(versionStr)

	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(699179)
		return (v.Major() > 19 || func() bool {
			__antithesis_instrumentation__.Notify(699180)
			return (v.Major() == 19 && func() bool {
				__antithesis_instrumentation__.Notify(699181)
				return v.Minor() >= 2 == true
			}() == true) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(699182)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(699183)
	}
	__antithesis_instrumentation__.Notify(699173)
	_, err = db.Exec("SET CLUSTER SETTING kv.range_merge.queue_enabled = false")
	return err
}

func Split(ctx context.Context, db *gosql.DB, table workload.Table, concurrency int) error {
	__antithesis_instrumentation__.Notify(699184)

	if err := maybeDisableMergeQueue(db); err != nil {
		__antithesis_instrumentation__.Notify(699191)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699192)
	}
	__antithesis_instrumentation__.Notify(699185)

	_, err := db.Exec("SHOW RANGES FROM TABLE system.descriptor")
	if err != nil {
		__antithesis_instrumentation__.Notify(699193)
		if strings.Contains(err.Error(), "not fully contained in tenant") || func() bool {
			__antithesis_instrumentation__.Notify(699195)
			return strings.Contains(err.Error(), "operation is unsupported in multi-tenancy mode") == true
		}() == true {
			__antithesis_instrumentation__.Notify(699196)
			log.Infof(ctx, `skipping workload splits; can't split on tenants'`)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(699197)
		}
		__antithesis_instrumentation__.Notify(699194)
		return err
	} else {
		__antithesis_instrumentation__.Notify(699198)
	}
	__antithesis_instrumentation__.Notify(699186)

	if table.Splits.NumBatches <= 0 {
		__antithesis_instrumentation__.Notify(699199)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(699200)
	}
	__antithesis_instrumentation__.Notify(699187)
	splitPoints := make([][]interface{}, 0, table.Splits.NumBatches)
	for splitIdx := 0; splitIdx < table.Splits.NumBatches; splitIdx++ {
		__antithesis_instrumentation__.Notify(699201)
		splitPoints = append(splitPoints, table.Splits.BatchRows(splitIdx)...)
	}
	__antithesis_instrumentation__.Notify(699188)
	sort.Sort(sliceSliceInterface(splitPoints))

	type pair struct {
		lo, hi int
	}
	splitCh := make(chan pair, len(splitPoints)/2+1)
	splitCh <- pair{0, len(splitPoints)}
	doneCh := make(chan struct{})

	log.Infof(ctx, `starting %d splits`, len(splitPoints))
	g := ctxgroup.WithContext(ctx)

	r := rate.NewLimiter(128, 1)
	for i := 0; i < concurrency; i++ {
		__antithesis_instrumentation__.Notify(699202)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(699203)
			var buf bytes.Buffer
			for {
				__antithesis_instrumentation__.Notify(699204)
				select {
				case p, ok := <-splitCh:
					__antithesis_instrumentation__.Notify(699205)
					if !ok {
						__antithesis_instrumentation__.Notify(699213)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(699214)
					}
					__antithesis_instrumentation__.Notify(699206)
					if err := r.Wait(ctx); err != nil {
						__antithesis_instrumentation__.Notify(699215)
						return err
					} else {
						__antithesis_instrumentation__.Notify(699216)
					}
					__antithesis_instrumentation__.Notify(699207)
					m := (p.lo + p.hi) / 2
					split := strings.Join(StringTuple(splitPoints[m]), `,`)

					buf.Reset()
					fmt.Fprintf(&buf, `ALTER TABLE %s SPLIT AT VALUES (%s)`, table.Name, split)

					stmt := buf.String()
					if _, err := db.Exec(stmt); err != nil {
						__antithesis_instrumentation__.Notify(699217)
						mtErr := errorutil.UnsupportedWithMultiTenancy(0)
						if strings.Contains(err.Error(), mtErr.Error()) {
							__antithesis_instrumentation__.Notify(699219)

							break
						} else {
							__antithesis_instrumentation__.Notify(699220)
						}
						__antithesis_instrumentation__.Notify(699218)
						return errors.Wrapf(err, "executing %s", stmt)
					} else {
						__antithesis_instrumentation__.Notify(699221)
					}
					__antithesis_instrumentation__.Notify(699208)

					buf.Reset()
					fmt.Fprintf(&buf, `ALTER TABLE %s SCATTER FROM (%s) TO (%s)`,
						table.Name, split, split)
					stmt = buf.String()
					if _, err := db.Exec(stmt); err != nil {
						__antithesis_instrumentation__.Notify(699222)

						log.Warningf(ctx, `%s: %v`, stmt, err)
					} else {
						__antithesis_instrumentation__.Notify(699223)
					}
					__antithesis_instrumentation__.Notify(699209)

					select {
					case doneCh <- struct{}{}:
						__antithesis_instrumentation__.Notify(699224)
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(699225)
						return ctx.Err()
					}
					__antithesis_instrumentation__.Notify(699210)

					if p.lo < m {
						__antithesis_instrumentation__.Notify(699226)
						splitCh <- pair{p.lo, m}
					} else {
						__antithesis_instrumentation__.Notify(699227)
					}
					__antithesis_instrumentation__.Notify(699211)
					if m+1 < p.hi {
						__antithesis_instrumentation__.Notify(699228)
						splitCh <- pair{m + 1, p.hi}
					} else {
						__antithesis_instrumentation__.Notify(699229)
					}
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(699212)
					return ctx.Err()
				}

			}
		})
	}
	__antithesis_instrumentation__.Notify(699189)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(699230)
		finished := 0
		for finished < len(splitPoints) {
			__antithesis_instrumentation__.Notify(699232)
			select {
			case <-doneCh:
				__antithesis_instrumentation__.Notify(699233)
				finished++
				if finished%1000 == 0 {
					__antithesis_instrumentation__.Notify(699235)
					log.Infof(ctx, "finished %d of %d splits", finished, len(splitPoints))
				} else {
					__antithesis_instrumentation__.Notify(699236)
				}
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(699234)
				return ctx.Err()
			}
		}
		__antithesis_instrumentation__.Notify(699231)
		close(splitCh)
		return nil
	})
	__antithesis_instrumentation__.Notify(699190)
	return g.Wait()
}

func StringTuple(datums []interface{}) []string {
	__antithesis_instrumentation__.Notify(699237)
	s := make([]string, len(datums))
	for i, datum := range datums {
		__antithesis_instrumentation__.Notify(699239)
		if datum == nil {
			__antithesis_instrumentation__.Notify(699241)
			s[i] = `NULL`
			continue
		} else {
			__antithesis_instrumentation__.Notify(699242)
		}
		__antithesis_instrumentation__.Notify(699240)
		switch x := datum.(type) {
		case int:
			__antithesis_instrumentation__.Notify(699243)
			s[i] = strconv.Itoa(x)
		case int64:
			__antithesis_instrumentation__.Notify(699244)
			s[i] = strconv.FormatInt(x, 10)
		case uint64:
			__antithesis_instrumentation__.Notify(699245)
			s[i] = strconv.FormatUint(x, 10)
		case string:
			__antithesis_instrumentation__.Notify(699246)
			s[i] = lexbase.EscapeSQLString(x)
		case float64:
			__antithesis_instrumentation__.Notify(699247)
			s[i] = fmt.Sprintf(`%f`, x)
		case []byte:
			__antithesis_instrumentation__.Notify(699248)

			s[i] = lexbase.EscapeSQLString(string(x))
		default:
			__antithesis_instrumentation__.Notify(699249)
			panic(errors.AssertionFailedf("unsupported type %T: %v", x, x))
		}
	}
	__antithesis_instrumentation__.Notify(699238)
	return s
}

type sliceSliceInterface [][]interface{}

func (s sliceSliceInterface) Len() int { __antithesis_instrumentation__.Notify(699250); return len(s) }
func (s sliceSliceInterface) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(699251)
	s[i], s[j] = s[j], s[i]
}
func (s sliceSliceInterface) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(699252)
	for offset := 0; ; offset++ {
		__antithesis_instrumentation__.Notify(699253)
		iLen, jLen := len(s[i]), len(s[j])
		if iLen <= offset || func() bool {
			__antithesis_instrumentation__.Notify(699256)
			return jLen <= offset == true
		}() == true {
			__antithesis_instrumentation__.Notify(699257)
			return iLen < jLen
		} else {
			__antithesis_instrumentation__.Notify(699258)
		}
		__antithesis_instrumentation__.Notify(699254)
		var cmp int
		switch x := s[i][offset].(type) {
		case int:
			__antithesis_instrumentation__.Notify(699259)
			if y := s[j][offset].(int); x < y {
				__antithesis_instrumentation__.Notify(699270)
				return true
			} else {
				__antithesis_instrumentation__.Notify(699271)
				if x > y {
					__antithesis_instrumentation__.Notify(699272)
					return false
				} else {
					__antithesis_instrumentation__.Notify(699273)
				}
			}
			__antithesis_instrumentation__.Notify(699260)
			continue
		case int64:
			__antithesis_instrumentation__.Notify(699261)
			if y := s[j][offset].(int64); x < y {
				__antithesis_instrumentation__.Notify(699274)
				return true
			} else {
				__antithesis_instrumentation__.Notify(699275)
				if x > y {
					__antithesis_instrumentation__.Notify(699276)
					return false
				} else {
					__antithesis_instrumentation__.Notify(699277)
				}
			}
			__antithesis_instrumentation__.Notify(699262)
			continue
		case float64:
			__antithesis_instrumentation__.Notify(699263)
			if y := s[j][offset].(float64); x < y {
				__antithesis_instrumentation__.Notify(699278)
				return true
			} else {
				__antithesis_instrumentation__.Notify(699279)
				if x > y {
					__antithesis_instrumentation__.Notify(699280)
					return false
				} else {
					__antithesis_instrumentation__.Notify(699281)
				}
			}
			__antithesis_instrumentation__.Notify(699264)
			continue
		case uint64:
			__antithesis_instrumentation__.Notify(699265)
			if y := s[j][offset].(uint64); x < y {
				__antithesis_instrumentation__.Notify(699282)
				return true
			} else {
				__antithesis_instrumentation__.Notify(699283)
				if x > y {
					__antithesis_instrumentation__.Notify(699284)
					return false
				} else {
					__antithesis_instrumentation__.Notify(699285)
				}
			}
			__antithesis_instrumentation__.Notify(699266)
			continue
		case string:
			__antithesis_instrumentation__.Notify(699267)
			cmp = strings.Compare(x, s[j][offset].(string))
		case []byte:
			__antithesis_instrumentation__.Notify(699268)
			cmp = bytes.Compare(x, s[j][offset].([]byte))
		default:
			__antithesis_instrumentation__.Notify(699269)
			panic(errors.AssertionFailedf("unsupported type %T: %v", x, x))
		}
		__antithesis_instrumentation__.Notify(699255)
		if cmp < 0 {
			__antithesis_instrumentation__.Notify(699286)
			return true
		} else {
			__antithesis_instrumentation__.Notify(699287)
			if cmp > 0 {
				__antithesis_instrumentation__.Notify(699288)
				return false
			} else {
				__antithesis_instrumentation__.Notify(699289)
			}
		}
	}
}
