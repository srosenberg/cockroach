package extract

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/internal/rsg/yacc"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const (
	rrAddr = "http://bottlecaps.de/rr/ui"
)

var (
	reIsExpr  = regexp.MustCompile("^[a-z_0-9]+$")
	reIsIdent = regexp.MustCompile("^[A-Z_0-9]+$")
	rrLock    syncutil.Mutex
)

func GenerateRRJar(jar string, bnf []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(39495)

	cmd := exec.Command(
		"java",
		"-jar", jar,
		"-suppressebnf",
		"-color:#ffffff",
		"-width:760",
		"-")
	cmd.Stdin = bytes.NewReader(bnf)

	out, err := cmd.CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(39497)
		return nil, fmt.Errorf("%w: %s", err, out)
	} else {
		__antithesis_instrumentation__.Notify(39498)
	}
	__antithesis_instrumentation__.Notify(39496)
	return out, nil
}

func GenerateRRNet(bnf []byte, railroadAPITimeout time.Duration) ([]byte, error) {
	__antithesis_instrumentation__.Notify(39499)
	rrLock.Lock()
	defer rrLock.Unlock()

	v := url.Values{}
	v.Add("color", "#ffffff")
	v.Add("frame", "diagram")

	v.Add("text", string(bnf))
	v.Add("width", "760")
	v.Add("options", "eliminaterecursion")
	v.Add("options", "factoring")
	v.Add("options", "inline")

	httpClient := httputil.NewClientWithTimeout(railroadAPITimeout)
	resp, err := httpClient.Post(context.TODO(), rrAddr, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	if err != nil {
		__antithesis_instrumentation__.Notify(39503)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(39504)
	}
	__antithesis_instrumentation__.Notify(39500)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		__antithesis_instrumentation__.Notify(39505)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(39506)
	}
	__antithesis_instrumentation__.Notify(39501)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		__antithesis_instrumentation__.Notify(39507)
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	} else {
		__antithesis_instrumentation__.Notify(39508)
	}
	__antithesis_instrumentation__.Notify(39502)
	return body, nil
}

func GenerateBNF(addr string, bnfAPITimeout time.Duration) (ebnf []byte, err error) {
	__antithesis_instrumentation__.Notify(39509)
	var b []byte
	if strings.HasPrefix(addr, "http") {
		__antithesis_instrumentation__.Notify(39515)
		httpClient := httputil.NewClientWithTimeout(bnfAPITimeout)
		resp, err := httpClient.Get(context.TODO(), addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(39518)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(39519)
		}
		__antithesis_instrumentation__.Notify(39516)
		b, err = func() ([]byte, error) {
			__antithesis_instrumentation__.Notify(39520)
			defer resp.Body.Close()
			return ioutil.ReadAll(resp.Body)
		}()
		__antithesis_instrumentation__.Notify(39517)
		if err != nil {
			__antithesis_instrumentation__.Notify(39521)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(39522)
		}
	} else {
		__antithesis_instrumentation__.Notify(39523)
		body, err := ioutil.ReadFile(addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(39525)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(39526)
		}
		__antithesis_instrumentation__.Notify(39524)
		b = body
	}
	__antithesis_instrumentation__.Notify(39510)
	t, err := yacc.Parse(addr, string(b))
	if err != nil {
		__antithesis_instrumentation__.Notify(39527)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(39528)
	}
	__antithesis_instrumentation__.Notify(39511)
	buf := new(bytes.Buffer)

	prods := make(map[string][][]yacc.Item)
	for _, p := range t.Productions {
		__antithesis_instrumentation__.Notify(39529)
		var impl [][]yacc.Item
		for _, e := range p.Expressions {
			__antithesis_instrumentation__.Notify(39531)
			if strings.Contains(e.Command, "unimplemented") && func() bool {
				__antithesis_instrumentation__.Notify(39534)
				return !strings.Contains(e.Command, "FORCE DOC") == true
			}() == true {
				__antithesis_instrumentation__.Notify(39535)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39536)
			}
			__antithesis_instrumentation__.Notify(39532)
			if strings.Contains(e.Command, "SKIP DOC") {
				__antithesis_instrumentation__.Notify(39537)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39538)
			}
			__antithesis_instrumentation__.Notify(39533)
			impl = append(impl, e.Items)
		}
		__antithesis_instrumentation__.Notify(39530)
		prods[p.Name] = impl
	}
	__antithesis_instrumentation__.Notify(39512)

	for {
		__antithesis_instrumentation__.Notify(39539)
		changed := false
		for name, exprs := range prods {
			__antithesis_instrumentation__.Notify(39541)
			var next [][]yacc.Item
			for _, expr := range exprs {
				__antithesis_instrumentation__.Notify(39543)
				add := true
				var items []yacc.Item
				for _, item := range expr {
					__antithesis_instrumentation__.Notify(39545)
					p := prods[item.Value]
					if item.Typ == yacc.TypToken && func() bool {
						__antithesis_instrumentation__.Notify(39548)
						return !isUpper(item.Value) == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(39549)
						return len(p) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(39550)
						add = false
						changed = true
						break
					} else {
						__antithesis_instrumentation__.Notify(39551)
					}
					__antithesis_instrumentation__.Notify(39546)

					if len(p) == 1 && func() bool {
						__antithesis_instrumentation__.Notify(39552)
						return len(p[0]) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(39553)
						changed = true
						continue
					} else {
						__antithesis_instrumentation__.Notify(39554)
					}
					__antithesis_instrumentation__.Notify(39547)
					items = append(items, item)
				}
				__antithesis_instrumentation__.Notify(39544)
				if add {
					__antithesis_instrumentation__.Notify(39555)
					next = append(next, items)
				} else {
					__antithesis_instrumentation__.Notify(39556)
				}
			}
			__antithesis_instrumentation__.Notify(39542)
			prods[name] = next
		}
		__antithesis_instrumentation__.Notify(39540)
		if !changed {
			__antithesis_instrumentation__.Notify(39557)
			break
		} else {
			__antithesis_instrumentation__.Notify(39558)
		}
	}
	__antithesis_instrumentation__.Notify(39513)

	start := true
	for _, prod := range t.Productions {
		__antithesis_instrumentation__.Notify(39559)
		p := prods[prod.Name]
		if len(p) == 0 {
			__antithesis_instrumentation__.Notify(39562)
			continue
		} else {
			__antithesis_instrumentation__.Notify(39563)
		}
		__antithesis_instrumentation__.Notify(39560)
		if start {
			__antithesis_instrumentation__.Notify(39564)
			start = false
		} else {
			__antithesis_instrumentation__.Notify(39565)
			buf.WriteString("\n")
		}
		__antithesis_instrumentation__.Notify(39561)
		fmt.Fprintf(buf, "%s ::=\n", prod.Name)
		for i, items := range p {
			__antithesis_instrumentation__.Notify(39566)
			buf.WriteString("\t")
			if i > 0 {
				__antithesis_instrumentation__.Notify(39569)
				buf.WriteString("| ")
			} else {
				__antithesis_instrumentation__.Notify(39570)
			}
			__antithesis_instrumentation__.Notify(39567)
			for j, item := range items {
				__antithesis_instrumentation__.Notify(39571)
				if j > 0 {
					__antithesis_instrumentation__.Notify(39573)
					buf.WriteString(" ")
				} else {
					__antithesis_instrumentation__.Notify(39574)
				}
				__antithesis_instrumentation__.Notify(39572)
				buf.WriteString(item.Value)
			}
			__antithesis_instrumentation__.Notify(39568)
			buf.WriteString("\n")
		}
	}
	__antithesis_instrumentation__.Notify(39514)
	return buf.Bytes(), nil
}

func isUpper(s string) bool {
	__antithesis_instrumentation__.Notify(39575)
	return s == strings.ToUpper(s)
}

func ParseGrammar(r io.Reader) (Grammar, error) {
	__antithesis_instrumentation__.Notify(39576)
	g := make(Grammar)

	var name string
	var prods productions
	scan := bufio.NewScanner(r)
	i := 0
	for scan.Scan() {
		__antithesis_instrumentation__.Notify(39580)
		s := scan.Text()
		i++
		f := strings.Fields(s)
		if len(f) == 0 {
			__antithesis_instrumentation__.Notify(39585)
			if len(prods) > 0 {
				__antithesis_instrumentation__.Notify(39587)
				g[name] = prods
			} else {
				__antithesis_instrumentation__.Notify(39588)
			}
			__antithesis_instrumentation__.Notify(39586)
			continue
		} else {
			__antithesis_instrumentation__.Notify(39589)
		}
		__antithesis_instrumentation__.Notify(39581)
		if !unicode.IsSpace(rune(s[0])) {
			__antithesis_instrumentation__.Notify(39590)
			if len(f) != 2 {
				__antithesis_instrumentation__.Notify(39592)
				return nil, fmt.Errorf("bad line: %v: %s", i, s)
			} else {
				__antithesis_instrumentation__.Notify(39593)
			}
			__antithesis_instrumentation__.Notify(39591)
			name = f[0]
			prods = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(39594)
		}
		__antithesis_instrumentation__.Notify(39582)
		if f[0] == "|" {
			__antithesis_instrumentation__.Notify(39595)
			f = f[1:]
		} else {
			__antithesis_instrumentation__.Notify(39596)
		}
		__antithesis_instrumentation__.Notify(39583)
		var seq sequence
		for _, v := range f {
			__antithesis_instrumentation__.Notify(39597)
			if reIsIdent.MatchString(v) {
				__antithesis_instrumentation__.Notify(39598)
				seq = append(seq, literal(v))
			} else {
				__antithesis_instrumentation__.Notify(39599)
				if reIsExpr.MatchString(v) {
					__antithesis_instrumentation__.Notify(39600)
					seq = append(seq, token(v))
				} else {
					__antithesis_instrumentation__.Notify(39601)
					if strings.HasPrefix(v, `'`) && func() bool {
						__antithesis_instrumentation__.Notify(39602)
						return strings.HasSuffix(v, `'`) == true
					}() == true {
						__antithesis_instrumentation__.Notify(39603)
						seq = append(seq, literal(v[1:len(v)-1]))
					} else {
						__antithesis_instrumentation__.Notify(39604)
						if strings.HasPrefix(v, `/*`) && func() bool {
							__antithesis_instrumentation__.Notify(39605)
							return strings.HasSuffix(v, `*/`) == true
						}() == true {
							__antithesis_instrumentation__.Notify(39606)
							seq = append(seq, comment(v))
						} else {
							__antithesis_instrumentation__.Notify(39607)
							panic(v)
						}
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(39584)
		prods = append(prods, seq)
	}
	__antithesis_instrumentation__.Notify(39577)
	if err := scan.Err(); err != nil {
		__antithesis_instrumentation__.Notify(39608)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(39609)
	}
	__antithesis_instrumentation__.Notify(39578)
	if len(prods) > 0 {
		__antithesis_instrumentation__.Notify(39610)
		g[name] = prods
	} else {
		__antithesis_instrumentation__.Notify(39611)
	}
	__antithesis_instrumentation__.Notify(39579)
	g.simplify()
	return g, nil
}

type Grammar map[string]productions

func (g Grammar) ExtractProduction(
	name string, descend, nosplit bool, match, exclude []*regexp.Regexp,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(39612)
	names := []token{token(name)}
	b := new(bytes.Buffer)
	done := map[token]bool{token(name): true}
	for i := 0; i < len(names); i++ {
		__antithesis_instrumentation__.Notify(39614)
		if i > 0 {
			__antithesis_instrumentation__.Notify(39618)
			b.WriteString("\n")
		} else {
			__antithesis_instrumentation__.Notify(39619)
		}
		__antithesis_instrumentation__.Notify(39615)
		n := names[i]
		prods := g[string(n)]
		if len(prods) == 0 {
			__antithesis_instrumentation__.Notify(39620)
			return nil, fmt.Errorf("couldn't find %s", n)
		} else {
			__antithesis_instrumentation__.Notify(39621)
		}
		__antithesis_instrumentation__.Notify(39616)
		walkToken(prods, func(t token) {
			__antithesis_instrumentation__.Notify(39622)
			if !done[t] && func() bool {
				__antithesis_instrumentation__.Notify(39623)
				return descend == true
			}() == true {
				__antithesis_instrumentation__.Notify(39624)
				names = append(names, t)
				done[t] = true
			} else {
				__antithesis_instrumentation__.Notify(39625)
			}
		})
		__antithesis_instrumentation__.Notify(39617)
		fmt.Fprintf(b, "%s ::=\n", n)
		b.WriteString(prods.Match(nosplit, match, exclude))
	}
	__antithesis_instrumentation__.Notify(39613)
	return b.Bytes(), nil
}

func (g Grammar) Inline(names ...string) error {
	__antithesis_instrumentation__.Notify(39626)
	for _, name := range names {
		__antithesis_instrumentation__.Notify(39628)
		p, ok := g[name]
		if !ok {
			__antithesis_instrumentation__.Notify(39630)
			return fmt.Errorf("unknown name: %s", name)
		} else {
			__antithesis_instrumentation__.Notify(39631)
		}
		__antithesis_instrumentation__.Notify(39629)
		grp := group(p)
		for _, prods := range g {
			__antithesis_instrumentation__.Notify(39632)
			replaceToken(prods, func(t token) expression {
				__antithesis_instrumentation__.Notify(39633)
				if string(t) == name {
					__antithesis_instrumentation__.Notify(39635)
					return grp
				} else {
					__antithesis_instrumentation__.Notify(39636)
				}
				__antithesis_instrumentation__.Notify(39634)
				return nil
			})
		}
	}
	__antithesis_instrumentation__.Notify(39627)
	return nil
}

func (g Grammar) simplify() {
	__antithesis_instrumentation__.Notify(39637)
	for name, prods := range g {
		__antithesis_instrumentation__.Notify(39638)
		p := simplify(name, prods)
		if p != nil {
			__antithesis_instrumentation__.Notify(39639)
			g[name] = p
		} else {
			__antithesis_instrumentation__.Notify(39640)
		}
	}
}

func simplify(name string, prods productions) productions {
	__antithesis_instrumentation__.Notify(39641)
	funcs := []func(string, productions) productions{
		simplifySelfRefList,
	}
	for _, f := range funcs {
		__antithesis_instrumentation__.Notify(39643)
		if e := f(name, prods); e != nil {
			__antithesis_instrumentation__.Notify(39644)
			return e
		} else {
			__antithesis_instrumentation__.Notify(39645)
		}
	}
	__antithesis_instrumentation__.Notify(39642)
	return nil
}

func simplifySelfRefList(name string, prods productions) productions {
	__antithesis_instrumentation__.Notify(39646)

	var group1, group2 group
	for _, p := range prods {
		__antithesis_instrumentation__.Notify(39649)
		s, ok := p.(sequence)
		if !ok {
			__antithesis_instrumentation__.Notify(39651)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(39652)
		}
		__antithesis_instrumentation__.Notify(39650)
		if len(s) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(39653)
			return s[0] == token(name) == true
		}() == true {
			__antithesis_instrumentation__.Notify(39654)
			group2 = append(group2, s[1:])
		} else {
			__antithesis_instrumentation__.Notify(39655)
			group1 = append(group1, s)
		}
	}
	__antithesis_instrumentation__.Notify(39647)
	if len(group2) == 0 {
		__antithesis_instrumentation__.Notify(39656)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(39657)
	}
	__antithesis_instrumentation__.Notify(39648)
	return productions{
		sequence{group1, repeat{group2}},
	}
}

func replaceToken(p productions, f func(token) expression) {
	__antithesis_instrumentation__.Notify(39658)
	replacetoken(p, f)
}

func replacetoken(e expression, f func(token) expression) expression {
	__antithesis_instrumentation__.Notify(39659)
	switch e := e.(type) {
	case sequence:
		__antithesis_instrumentation__.Notify(39661)
		for i, v := range e {
			__antithesis_instrumentation__.Notify(39668)
			n := replacetoken(v, f)
			if n != nil {
				__antithesis_instrumentation__.Notify(39669)
				e[i] = n
			} else {
				__antithesis_instrumentation__.Notify(39670)
			}
		}
	case token:
		__antithesis_instrumentation__.Notify(39662)
		return f(e)
	case group:
		__antithesis_instrumentation__.Notify(39663)
		for i, v := range e {
			__antithesis_instrumentation__.Notify(39671)
			n := replacetoken(v, f)
			if n != nil {
				__antithesis_instrumentation__.Notify(39672)
				e[i] = n
			} else {
				__antithesis_instrumentation__.Notify(39673)
			}
		}
	case productions:
		__antithesis_instrumentation__.Notify(39664)
		for i, v := range e {
			__antithesis_instrumentation__.Notify(39674)
			n := replacetoken(v, f)
			if n != nil {
				__antithesis_instrumentation__.Notify(39675)
				e[i] = n
			} else {
				__antithesis_instrumentation__.Notify(39676)
			}
		}
	case repeat:
		__antithesis_instrumentation__.Notify(39665)
		return replacetoken(e.expression, f)
	case literal, comment:
		__antithesis_instrumentation__.Notify(39666)

	default:
		__antithesis_instrumentation__.Notify(39667)
		panic(fmt.Errorf("unknown type: %T", e))
	}
	__antithesis_instrumentation__.Notify(39660)
	return nil
}

func walkToken(e expression, f func(token)) {
	__antithesis_instrumentation__.Notify(39677)
	switch e := e.(type) {
	case sequence:
		__antithesis_instrumentation__.Notify(39678)
		for _, v := range e {
			__antithesis_instrumentation__.Notify(39685)
			walkToken(v, f)
		}
	case token:
		__antithesis_instrumentation__.Notify(39679)
		f(e)
	case group:
		__antithesis_instrumentation__.Notify(39680)
		for _, v := range e {
			__antithesis_instrumentation__.Notify(39686)
			walkToken(v, f)
		}
	case repeat:
		__antithesis_instrumentation__.Notify(39681)
		walkToken(e.expression, f)
	case productions:
		__antithesis_instrumentation__.Notify(39682)
		for _, v := range e {
			__antithesis_instrumentation__.Notify(39687)
			walkToken(v, f)
		}
	case literal, comment:
		__antithesis_instrumentation__.Notify(39683)

	default:
		__antithesis_instrumentation__.Notify(39684)
		panic(fmt.Errorf("unknown type: %T", e))
	}
}

type productions []expression

func (p productions) Match(nosplit bool, match, exclude []*regexp.Regexp) string {
	__antithesis_instrumentation__.Notify(39688)
	b := new(bytes.Buffer)
	first := true
	for _, e := range p {
		__antithesis_instrumentation__.Notify(39690)
		if nosplit {
			__antithesis_instrumentation__.Notify(39692)
			b.WriteString("\t")
			if !first {
				__antithesis_instrumentation__.Notify(39694)
				b.WriteString("| ")
			} else {
				__antithesis_instrumentation__.Notify(39695)
				first = false
			}
			__antithesis_instrumentation__.Notify(39693)
			b.WriteString(e.String())
			b.WriteString("\n")
			continue
		} else {
			__antithesis_instrumentation__.Notify(39696)
		}
		__antithesis_instrumentation__.Notify(39691)
	Loop:
		for _, s := range split(e) {
			__antithesis_instrumentation__.Notify(39697)
			for _, ex := range exclude {
				__antithesis_instrumentation__.Notify(39701)
				if ex.MatchString(s) {
					__antithesis_instrumentation__.Notify(39702)
					continue Loop
				} else {
					__antithesis_instrumentation__.Notify(39703)
				}
			}
			__antithesis_instrumentation__.Notify(39698)
			for _, ma := range match {
				__antithesis_instrumentation__.Notify(39704)
				if !ma.MatchString(s) {
					__antithesis_instrumentation__.Notify(39705)
					continue Loop
				} else {
					__antithesis_instrumentation__.Notify(39706)
				}
			}
			__antithesis_instrumentation__.Notify(39699)
			b.WriteString("\t")
			if !first {
				__antithesis_instrumentation__.Notify(39707)
				b.WriteString("| ")
			} else {
				__antithesis_instrumentation__.Notify(39708)
				first = false
			}
			__antithesis_instrumentation__.Notify(39700)
			b.WriteString(s)
			b.WriteString("\n")
		}
	}
	__antithesis_instrumentation__.Notify(39689)
	return b.String()
}

func (p productions) String() string {
	__antithesis_instrumentation__.Notify(39709)
	b := new(bytes.Buffer)
	for i, e := range p {
		__antithesis_instrumentation__.Notify(39711)
		b.WriteString("\t")
		if i > 0 {
			__antithesis_instrumentation__.Notify(39713)
			b.WriteString("| ")
		} else {
			__antithesis_instrumentation__.Notify(39714)
		}
		__antithesis_instrumentation__.Notify(39712)
		b.WriteString(e.String())
		b.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(39710)
	return b.String()
}

type expression interface {
	String() string
}

type sequence []expression

func (s sequence) String() string {
	__antithesis_instrumentation__.Notify(39715)
	b := new(bytes.Buffer)
	for i, e := range s {
		__antithesis_instrumentation__.Notify(39717)
		if i > 0 {
			__antithesis_instrumentation__.Notify(39719)
			b.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(39720)
		}
		__antithesis_instrumentation__.Notify(39718)
		b.WriteString(e.String())
	}
	__antithesis_instrumentation__.Notify(39716)
	return b.String()
}

type token string

func (t token) String() string {
	__antithesis_instrumentation__.Notify(39721)
	return string(t)
}

type literal string

func (l literal) String() string {
	__antithesis_instrumentation__.Notify(39722)
	return fmt.Sprintf("'%s'", string(l))
}

type group []expression

func (g group) String() string {
	__antithesis_instrumentation__.Notify(39723)
	b := new(bytes.Buffer)
	b.WriteString("( ")
	for i, e := range g {
		__antithesis_instrumentation__.Notify(39725)
		if i > 0 {
			__antithesis_instrumentation__.Notify(39727)
			b.WriteString(" | ")
		} else {
			__antithesis_instrumentation__.Notify(39728)
		}
		__antithesis_instrumentation__.Notify(39726)
		b.WriteString(e.String())
	}
	__antithesis_instrumentation__.Notify(39724)
	b.WriteString(" )")
	return b.String()
}

type repeat struct {
	expression
}

func (r repeat) String() string {
	__antithesis_instrumentation__.Notify(39729)
	return fmt.Sprintf("( %s )*", r.expression)
}

type comment string

func (c comment) String() string {
	__antithesis_instrumentation__.Notify(39730)
	return string(c)
}

func split(e expression) []string {
	__antithesis_instrumentation__.Notify(39731)
	appendRet := func(cur, add []string) []string {
		__antithesis_instrumentation__.Notify(39734)
		if len(cur) == 0 {
			__antithesis_instrumentation__.Notify(39737)
			if len(add) == 0 {
				__antithesis_instrumentation__.Notify(39739)
				return []string{""}
			} else {
				__antithesis_instrumentation__.Notify(39740)
			}
			__antithesis_instrumentation__.Notify(39738)
			return add
		} else {
			__antithesis_instrumentation__.Notify(39741)
		}
		__antithesis_instrumentation__.Notify(39735)
		var next []string
		for _, r := range cur {
			__antithesis_instrumentation__.Notify(39742)
			for _, s := range add {
				__antithesis_instrumentation__.Notify(39743)
				next = append(next, r+" "+s)
			}
		}
		__antithesis_instrumentation__.Notify(39736)
		return next
	}
	__antithesis_instrumentation__.Notify(39732)
	var ret []string
	switch e := e.(type) {
	case sequence:
		__antithesis_instrumentation__.Notify(39744)
		for _, v := range e {
			__antithesis_instrumentation__.Notify(39749)
			ret = appendRet(ret, split(v))
		}
	case group:
		__antithesis_instrumentation__.Notify(39745)
		var next []string
		for _, v := range e {
			__antithesis_instrumentation__.Notify(39750)
			next = append(next, appendRet(ret, split(v))...)
		}
		__antithesis_instrumentation__.Notify(39746)
		ret = next
	case literal, comment, repeat, token:
		__antithesis_instrumentation__.Notify(39747)
		ret = append(ret, e.String())
	default:
		__antithesis_instrumentation__.Notify(39748)
		panic(fmt.Errorf("unknown type: %T", e))
	}
	__antithesis_instrumentation__.Notify(39733)
	return ret
}
