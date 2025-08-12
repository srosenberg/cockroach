// errscan.go — LCOV generator that *streams* output during analysis
// -----------------------------------------------------------------------------
//   go vet -vettool=$(go build -o errscan .) ./... > error_coverage.lcov
//   genhtml -o cov_html error_coverage.lcov
//
// This version **writes each source file’s coverage block immediately** from the
// analyzer’s Run function, so the LCOV file is produced even when the driver
// exits via os.Exit or stops after load errors.  Concurrency‑safe via a mutex.
// -----------------------------------------------------------------------------
package errscan

import (
    "flag"
    "fmt"
    "go/ast"
    "go/token"
    "go/types"
    "io"
    "os"
    "path/filepath"
    "sync"

    "golang.org/x/tools/go/analysis"
    "golang.org/x/tools/go/analysis/unitchecker"
)

// -----------------------------------------------------------------------------
// Analyzer definition ---------------------------------------------------------
// -----------------------------------------------------------------------------

var coverfile string

var Analyzer = &analysis.Analyzer{
    Name:            "errscan",
    Doc:             "streams an LCOV profile marking lines that interact with an error value",
    Flags:           *flag.NewFlagSet("errscan", flag.ExitOnError),
    Run:             run,
    RunDespiteErrors: true, // continue on packages with type errors
}

func init() {
    Analyzer.Flags.StringVar(&coverfile, "coverfile", "", "LCOV output file (default: stdout)")
}

func main() { unitchecker.Main(Analyzer) }

// -----------------------------------------------------------------------------
// Shared writer safeguarded for concurrency -----------------------------------
// -----------------------------------------------------------------------------

var (
    once       sync.Once
    mu         sync.Mutex
    lcovWriter io.Writer // stdout or the requested file
)

func openWriter() {
    if coverfile == "" {
        lcovWriter = os.Stdout
    } else {
        f, err := os.Create(coverfile) // truncate/overwrite
        if err != nil {
            fmt.Fprintf(os.Stderr, "errscan: cannot create coverfile: %v\n", err)
            os.Exit(1)
        }
        lcovWriter = f
    }
    fmt.Fprintln(lcovWriter, "TN:error-handling-coverage")
}

// emit writes one LCOV record for a single source file.
func emit(fname string, tf *token.File, hits map[int]struct{}) {
    once.Do(openWriter)

    mu.Lock()
    defer mu.Unlock()

    fmt.Fprintln(lcovWriter, "SF:"+fname)
    for line := 1; line <= tf.LineCount(); line++ {
        hit := 0
        if _, ok := hits[line]; ok {
            hit = 1
        }
        fmt.Fprintf(lcovWriter, "DA:%d,%d\n", line, hit)
    }
    fmt.Fprintln(lcovWriter, "end_of_record")
}

// -----------------------------------------------------------------------------
// Analysis logic --------------------------------------------------------------
// -----------------------------------------------------------------------------

var errorType = types.Universe.Lookup("error").Type()

var knownRetryHelpers = map[string]struct{}{
    "github.com/avast/retry-go.Do":          {},
    "github.com/cenkalti/backoff/v4.Retry":       {},
    "github.com/cenkalti/backoff/v4.RetryNotify": {},
}

func run(pass *analysis.Pass) (interface{}, error) {
    for _, file := range pass.Files {
        tf := pass.Fset.File(file.Pos())
        fname, _ := filepath.Abs(tf.Name())

        hits := make(map[int]struct{})
        var enclosing ast.Stmt

        ast.Inspect(file, func(n ast.Node) bool {
            switch node := n.(type) {
            case ast.Stmt:
                enclosing = node
                return true
            case ast.Expr:
                if enclosing == nil {
                    return true
                }
                if tv, ok := pass.TypesInfo.Types[node]; ok && isError(tv.Type) {
                    mark(hitPos(pass.Fset, node.Pos()), hits)
                }
                if call, ok := node.(*ast.CallExpr); ok {
                    if callHasError(pass, call) || isKnownRetryHelper(pass, call) {
                        mark(hitPos(pass.Fset, call.Pos()), hits)
                    }
                }
            }
            return true
        })

        emit(fname, tf, hits)
    }
    return nil, nil
}

// mark records a line‑hit in the file‑local hits map.
func mark(line int, hits map[int]struct{}) {
    if line > 0 {
        hits[line] = struct{}{}
    }
}

func hitPos(fset *token.FileSet, pos token.Pos) int {
    return fset.Position(pos).Line
}

// -----------------------------------------------------------------------------
// Helper predicates -----------------------------------------------------------
// -----------------------------------------------------------------------------

func isError(t types.Type) bool { return types.Identical(t, errorType) }

func callHasError(pass *analysis.Pass, call *ast.CallExpr) bool {
    sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
    if !ok {
        return false
    }
    return signatureHasError(sig)
}

func signatureHasError(sig *types.Signature) bool {
    for i := 0; i < sig.Params().Len(); i++ {
        if isError(sig.Params().At(i).Type()) {
            return true
        }
    }
    for i := 0; i < sig.Results().Len(); i++ {
        if isError(sig.Results().At(i).Type()) {
            return true
        }
    }
    return false
}

func isKnownRetryHelper(pass *analysis.Pass, call *ast.CallExpr) bool {
    var obj types.Object
    switch fun := call.Fun.(type) {
    case *ast.SelectorExpr:
        obj = pass.TypesInfo.ObjectOf(fun.Sel)
    case *ast.Ident:
        obj = pass.TypesInfo.ObjectOf(fun)
    default:
        return false
    }
    if obj == nil || obj.Pkg() == nil {
        return false
    }
    key := obj.Pkg().Path() + "." + obj.Name()
    _, ok := knownRetryHelpers[key]
    return ok
}

