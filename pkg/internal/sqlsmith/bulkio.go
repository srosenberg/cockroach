package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func (s *Smither) bulkIOEnabled() bool {
	__antithesis_instrumentation__.Notify(68901)
	return s.bulkSrv != nil
}

func (s *Smither) enableBulkIO() {
	__antithesis_instrumentation__.Notify(68902)
	s.bulkSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(68904)
		s.lock.Lock()
		defer s.lock.Unlock()
		localfile := r.URL.Path
		switch r.Method {
		case "PUT":
			__antithesis_instrumentation__.Notify(68905)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				__antithesis_instrumentation__.Notify(68911)
				http.Error(w, err.Error(), 500)
				return
			} else {
				__antithesis_instrumentation__.Notify(68912)
			}
			__antithesis_instrumentation__.Notify(68906)
			s.bulkFiles[localfile] = b
			w.WriteHeader(201)
		case "GET", "HEAD":
			__antithesis_instrumentation__.Notify(68907)
			b, ok := s.bulkFiles[localfile]
			if !ok {
				__antithesis_instrumentation__.Notify(68913)
				http.Error(w, fmt.Sprintf("not found: %s", localfile), 404)
				return
			} else {
				__antithesis_instrumentation__.Notify(68914)
			}
			__antithesis_instrumentation__.Notify(68908)
			_, _ = w.Write(b)
		case "DELETE":
			__antithesis_instrumentation__.Notify(68909)
			delete(s.bulkFiles, localfile)
			w.WriteHeader(204)
		default:
			__antithesis_instrumentation__.Notify(68910)
			http.Error(w, "unsupported method", 400)
		}
	}))
	__antithesis_instrumentation__.Notify(68903)
	s.bulkFiles = map[string][]byte{}
	s.bulkBackups = map[string]tree.TargetList{}
}

func makeAsOf(s *Smither) tree.AsOfClause {
	__antithesis_instrumentation__.Notify(68915)
	var expr tree.Expr
	switch s.rnd.Intn(10) {
	case 1:
		__antithesis_instrumentation__.Notify(68917)
		expr = tree.NewStrVal("-2s")
	case 2:
		__antithesis_instrumentation__.Notify(68918)
		expr = tree.NewStrVal(timeutil.Now().Add(-2 * time.Second).Format(timeutil.FullTimeFormat))
	case 3:
		__antithesis_instrumentation__.Notify(68919)
		expr = randgen.RandDatum(s.rnd, types.Interval, false)
	case 4:
		__antithesis_instrumentation__.Notify(68920)
		datum := randgen.RandDatum(s.rnd, types.Timestamp, false)
		str := strings.TrimSuffix(datum.String(), `+00:00'`)
		str = strings.TrimPrefix(str, `'`)
		expr = tree.NewStrVal(str)
	default:
		__antithesis_instrumentation__.Notify(68921)

	}
	__antithesis_instrumentation__.Notify(68916)
	return tree.AsOfClause{
		Expr: expr,
	}
}

func makeBackup(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68922)
	name := fmt.Sprintf("%s/%s", s.bulkSrv.URL, s.name("backup"))
	var targets tree.TargetList
	seen := map[tree.TableName]bool{}
	for len(targets.Tables) < 1 || func() bool {
		__antithesis_instrumentation__.Notify(68924)
		return s.coin() == true
	}() == true {
		__antithesis_instrumentation__.Notify(68925)
		table, ok := s.getRandTable()
		if !ok {
			__antithesis_instrumentation__.Notify(68928)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(68929)
		}
		__antithesis_instrumentation__.Notify(68926)
		if seen[*table.TableName] {
			__antithesis_instrumentation__.Notify(68930)
			continue
		} else {
			__antithesis_instrumentation__.Notify(68931)
		}
		__antithesis_instrumentation__.Notify(68927)
		seen[*table.TableName] = true
		targets.Tables = append(targets.Tables, table.TableName)
	}
	__antithesis_instrumentation__.Notify(68923)
	s.lock.Lock()
	s.bulkBackups[name] = targets
	s.lock.Unlock()

	return &tree.Backup{
		Targets: &targets,
		To:      tree.StringOrPlaceholderOptList{tree.NewStrVal(name)},
		AsOf:    makeAsOf(s),
		Options: tree.BackupOptions{CaptureRevisionHistory: s.coin()},
	}, true
}

func makeRestore(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68932)
	var name string
	var targets tree.TargetList
	s.lock.Lock()
	for name, targets = range s.bulkBackups {
		__antithesis_instrumentation__.Notify(68937)
		break
	}
	__antithesis_instrumentation__.Notify(68933)

	delete(s.bulkBackups, name)
	s.lock.Unlock()

	if name == "" {
		__antithesis_instrumentation__.Notify(68938)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68939)
	}
	__antithesis_instrumentation__.Notify(68934)

	s.rnd.Shuffle(len(targets.Tables), func(i, j int) {
		__antithesis_instrumentation__.Notify(68940)
		targets.Tables[i], targets.Tables[j] = targets.Tables[j], targets.Tables[i]
	})
	__antithesis_instrumentation__.Notify(68935)
	targets.Tables = targets.Tables[:1+s.rnd.Intn(len(targets.Tables))]

	db := s.name("db")
	if _, err := s.db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, db)); err != nil {
		__antithesis_instrumentation__.Notify(68941)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68942)
	}
	__antithesis_instrumentation__.Notify(68936)

	return &tree.Restore{
		Targets: targets,
		From:    []tree.StringOrPlaceholderOptList{{tree.NewStrVal(name)}},
		AsOf:    makeAsOf(s),
		Options: tree.RestoreOptions{
			IntoDB: tree.NewStrVal("into_db"),
		},
	}, true
}

const exportSchema = "/create.sql"

func makeExport(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68943)

	if !s.bulkIOEnabled() {
		__antithesis_instrumentation__.Notify(68947)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68948)
	}
	__antithesis_instrumentation__.Notify(68944)
	table, ok := s.getRandTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68949)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68950)
	}
	__antithesis_instrumentation__.Notify(68945)

	var schema string
	if err := s.db.QueryRow(fmt.Sprintf(
		`select create_statement from [show create table %s]`,
		table.TableName.String(),
	)).Scan(&schema); err != nil {
		__antithesis_instrumentation__.Notify(68951)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68952)
	}
	__antithesis_instrumentation__.Notify(68946)
	stmt := &tree.Select{
		Select: &tree.SelectClause{
			Exprs:       tree.SelectExprs{tree.StarSelectExpr()},
			From:        tree.From{Tables: tree.TableExprs{table.TableName}},
			TableSelect: true,
		},
	}
	exp := s.name("exp")
	name := fmt.Sprintf("%s/%s", s.bulkSrv.URL, exp)
	s.lock.Lock()
	s.bulkFiles[fmt.Sprintf("/%s%s", exp, exportSchema)] = []byte(schema)
	s.bulkExports = append(s.bulkExports, string(exp))
	s.lock.Unlock()

	return &tree.Export{
		Query:      stmt,
		FileFormat: "CSV",
		File:       tree.NewStrVal(name),
	}, true
}

var importCreateTableRE = regexp.MustCompile(`CREATE TABLE (.*) \(`)

func makeImport(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68953)
	if !s.bulkIOEnabled() {
		__antithesis_instrumentation__.Notify(68959)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68960)
	}
	__antithesis_instrumentation__.Notify(68954)

	s.lock.Lock()
	if len(s.bulkExports) == 0 {
		__antithesis_instrumentation__.Notify(68961)
		s.lock.Unlock()
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68962)
	}
	__antithesis_instrumentation__.Notify(68955)
	exp := s.bulkExports[0]
	s.bulkExports = s.bulkExports[1:]

	var files tree.Exprs
	for name := range s.bulkFiles {
		__antithesis_instrumentation__.Notify(68963)
		if strings.Contains(name, exp+"/") && func() bool {
			__antithesis_instrumentation__.Notify(68964)
			return !strings.HasSuffix(name, exportSchema) == true
		}() == true {
			__antithesis_instrumentation__.Notify(68965)
			files = append(files, tree.NewStrVal(s.bulkSrv.URL+name))
		} else {
			__antithesis_instrumentation__.Notify(68966)
		}
	}
	__antithesis_instrumentation__.Notify(68956)
	s.lock.Unlock()

	if len(files) == 0 {
		__antithesis_instrumentation__.Notify(68967)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68968)
	}
	__antithesis_instrumentation__.Notify(68957)

	tab := s.name("tab")
	s.lock.Lock()
	schema := fmt.Sprintf("/%s%s", exp, exportSchema)
	tableSchema := importCreateTableRE.ReplaceAll(
		s.bulkFiles[schema],
		[]byte(fmt.Sprintf("CREATE TABLE %s (", tab)),
	)
	s.lock.Unlock()

	_, err := s.db.Exec(string(tableSchema))
	if err != nil {
		__antithesis_instrumentation__.Notify(68969)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68970)
	}
	__antithesis_instrumentation__.Notify(68958)

	return &tree.Import{
		Table:      tree.NewUnqualifiedTableName(tab),
		Into:       true,
		FileFormat: "CSV",
		Files:      files,
		Options: tree.KVOptions{
			tree.KVOption{
				Key:   "nullif",
				Value: tree.NewStrVal(""),
			},
		},
	}, true
}
