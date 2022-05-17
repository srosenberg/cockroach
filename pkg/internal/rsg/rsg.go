package rsg

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/internal/rsg/yacc"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type RSG struct {
	Rnd *rand.Rand

	lock  syncutil.Mutex
	seen  map[string]bool
	prods map[string][]*yacc.ExpressionNode
}

func NewRSG(seed int64, y string, allowDuplicates bool) (*RSG, error) {
	__antithesis_instrumentation__.Notify(68517)
	tree, err := yacc.Parse("sql", y)
	if err != nil {
		__antithesis_instrumentation__.Notify(68521)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(68522)
	}
	__antithesis_instrumentation__.Notify(68518)
	rsg := RSG{
		Rnd:   rand.New(&lockedSource{src: rand.NewSource(seed).(rand.Source64)}),
		prods: make(map[string][]*yacc.ExpressionNode),
	}
	if !allowDuplicates {
		__antithesis_instrumentation__.Notify(68523)
		rsg.seen = make(map[string]bool)
	} else {
		__antithesis_instrumentation__.Notify(68524)
	}
	__antithesis_instrumentation__.Notify(68519)
	for _, prod := range tree.Productions {
		__antithesis_instrumentation__.Notify(68525)
		rsg.prods[prod.Name] = prod.Expressions
	}
	__antithesis_instrumentation__.Notify(68520)
	return &rsg, nil
}

func (r *RSG) Generate(root string, depth int) string {
	__antithesis_instrumentation__.Notify(68526)
	for i := 0; i < 100000; i++ {
		__antithesis_instrumentation__.Notify(68528)
		s := strings.Join(r.generate(root, depth), " ")
		if r.seen != nil {
			__antithesis_instrumentation__.Notify(68530)
			r.lock.Lock()
			if !r.seen[s] {
				__antithesis_instrumentation__.Notify(68532)
				r.seen[s] = true
			} else {
				__antithesis_instrumentation__.Notify(68533)
				s = ""
			}
			__antithesis_instrumentation__.Notify(68531)
			r.lock.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(68534)
		}
		__antithesis_instrumentation__.Notify(68529)
		if s != "" {
			__antithesis_instrumentation__.Notify(68535)
			s = strings.Replace(s, "_LA", "", -1)
			s = strings.Replace(s, " AS OF SYSTEM TIME \"string\"", "", -1)
			return s
		} else {
			__antithesis_instrumentation__.Notify(68536)
		}
	}
	__antithesis_instrumentation__.Notify(68527)
	panic("couldn't find unique string")
}

func (r *RSG) generate(root string, depth int) []string {
	__antithesis_instrumentation__.Notify(68537)

	ret := make([]string, 0)
	prods := r.prods[root]
	if len(prods) == 0 {
		__antithesis_instrumentation__.Notify(68540)
		return []string{root}
	} else {
		__antithesis_instrumentation__.Notify(68541)
	}
	__antithesis_instrumentation__.Notify(68538)
	prod := prods[r.Intn(len(prods))]
	for _, item := range prod.Items {
		__antithesis_instrumentation__.Notify(68542)
		switch item.Typ {
		case yacc.TypLiteral:
			__antithesis_instrumentation__.Notify(68543)
			v := item.Value[1 : len(item.Value)-1]
			ret = append(ret, v)
		case yacc.TypToken:
			__antithesis_instrumentation__.Notify(68544)
			var v []string
			switch item.Value {
			case "IDENT":
				__antithesis_instrumentation__.Notify(68548)
				v = []string{"ident"}
			case "c_expr":
				__antithesis_instrumentation__.Notify(68549)
				v = r.generate(item.Value, 30)
			case "SCONST":
				__antithesis_instrumentation__.Notify(68550)
				v = []string{`'string'`}
			case "ICONST":
				__antithesis_instrumentation__.Notify(68551)
				v = []string{fmt.Sprint(r.Intn(1000) - 500)}
			case "FCONST":
				__antithesis_instrumentation__.Notify(68552)
				v = []string{fmt.Sprint(r.Float64())}
			case "BCONST":
				__antithesis_instrumentation__.Notify(68553)
				v = []string{`b'bytes'`}
			case "BITCONST":
				__antithesis_instrumentation__.Notify(68554)
				v = []string{`B'10010'`}
			case "substr_from":
				__antithesis_instrumentation__.Notify(68555)
				v = []string{"FROM", `'string'`}
			case "substr_for":
				__antithesis_instrumentation__.Notify(68556)
				v = []string{"FOR", `'string'`}
			case "overlay_placing":
				__antithesis_instrumentation__.Notify(68557)
				v = []string{"PLACING", `'string'`}
			default:
				__antithesis_instrumentation__.Notify(68558)
				if depth == 0 {
					__antithesis_instrumentation__.Notify(68560)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(68561)
				}
				__antithesis_instrumentation__.Notify(68559)
				v = r.generate(item.Value, depth-1)
			}
			__antithesis_instrumentation__.Notify(68545)
			if v == nil {
				__antithesis_instrumentation__.Notify(68562)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(68563)
			}
			__antithesis_instrumentation__.Notify(68546)
			ret = append(ret, v...)
		default:
			__antithesis_instrumentation__.Notify(68547)
			panic("unknown item type")
		}
	}
	__antithesis_instrumentation__.Notify(68539)
	return ret
}

func (r *RSG) Intn(n int) int {
	__antithesis_instrumentation__.Notify(68564)
	return r.Rnd.Intn(n)
}

func (r *RSG) Int63() int64 {
	__antithesis_instrumentation__.Notify(68565)
	return r.Rnd.Int63()
}

func (r *RSG) Float64() float64 {
	__antithesis_instrumentation__.Notify(68566)
	v := r.Rnd.Float64()*2 - 1
	switch r.Rnd.Intn(10) {
	case 0:
		__antithesis_instrumentation__.Notify(68568)
		v = 0
	case 1:
		__antithesis_instrumentation__.Notify(68569)
		v = math.Inf(1)
	case 2:
		__antithesis_instrumentation__.Notify(68570)
		v = math.Inf(-1)
	case 3:
		__antithesis_instrumentation__.Notify(68571)
		v = math.NaN()
	case 4, 5:
		__antithesis_instrumentation__.Notify(68572)
		i := r.Rnd.Intn(50)
		v *= math.Pow10(i)
	case 6, 7:
		__antithesis_instrumentation__.Notify(68573)
		i := r.Rnd.Intn(50)
		v *= math.Pow10(-i)
	default:
		__antithesis_instrumentation__.Notify(68574)
	}
	__antithesis_instrumentation__.Notify(68567)
	return v
}

func (r *RSG) GenerateRandomArg(typ *types.T) string {
	__antithesis_instrumentation__.Notify(68575)
	switch r.Intn(20) {
	case 0:
		__antithesis_instrumentation__.Notify(68577)
		return "NULL"
	case 1:
		__antithesis_instrumentation__.Notify(68578)
		return fmt.Sprintf("NULL::%s", typ)
	case 2:
		__antithesis_instrumentation__.Notify(68579)
		return fmt.Sprintf("(SELECT NULL)::%s", typ)
	default:
		__antithesis_instrumentation__.Notify(68580)
	}
	__antithesis_instrumentation__.Notify(68576)

	r.lock.Lock()
	datum := randgen.RandDatumWithNullChance(r.Rnd, typ, 0)
	r.lock.Unlock()

	return tree.Serialize(datum)
}

type lockedSource struct {
	lk  syncutil.Mutex
	src rand.Source64
}

func (r *lockedSource) Int63() (n int64) {
	__antithesis_instrumentation__.Notify(68581)
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Uint64() (n uint64) {
	__antithesis_instrumentation__.Notify(68582)
	r.lk.Lock()
	n = r.src.Uint64()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	__antithesis_instrumentation__.Notify(68583)
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}
