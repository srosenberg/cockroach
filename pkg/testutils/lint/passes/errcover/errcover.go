// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package errcover

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"github.com/cockroachdb/cockroach/pkg/util/intsets"
	"github.com/cockroachdb/cockroach/pkg/util/metamorphic"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

type (
	// FunctionCountFact holds the number of functions in a package.
	FunctionCountFact struct {
		Count int
	}
)

// TODO(SR): update
// Doc documents this pass.
const (
	Doc  = `check for return of nil in a conditional which check for non-nil errors`
	name = "errcover"
)

var (
	errorType = types.Universe.Lookup("error").Type()
	numBlocks uint32
)

// Function defines the location of a function (package-level or
// method on a type).
type errConstructor struct {
	pkg      string
	receiver string // empty for package-level functions
	name     string
	errType  int
}

var errFunctions = []errConstructor{
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsTimeout", errType: metamorphic.GRPCTimeoutErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsConnectionUnavailable", errType: metamorphic.GRPCUnavailableErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsConnectionRejected", errType: metamorphic.GRPCRejectedErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsClosedConnection", errType: metamorphic.GRPCClosedErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsContextCanceled", errType: metamorphic.GRPCtxCancelledErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsAuthError", errType: metamorphic.GRPCAuthErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "IsWaitingForInit", errType: metamorphic.GRPCWaitingForInitErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/util/grpcutil", receiver: "", name: "RequestDidNotStart", errType: metamorphic.GRPCRequestDidNotStartErr},
	{pkg: "google.golang.org/grpc/codes", receiver: "", name: "PermissionDenied", errType: metamorphic.GRPCPermissionDeniedErr},
	{pkg: "google.golang.org/grpc/codes", receiver: "", name: "NotFound", errType: metamorphic.GRPCNotFoundErr},
	{pkg: "google.golang.org/grpc/codes", receiver: "", name: "AlreadyExists", errType: metamorphic.GRPCAlreadyExistsErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/sqlerrors", receiver: "", name: "IsOutOfMemoryError", errType: metamorphic.PGOutOfMemoryErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/sqlerrors", receiver: "", name: "IsRelationAlreadyExistsError", errType: metamorphic.PGRelationAlreadyExistsErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "UndefinedObject", errType: metamorphic.PGUndefinedObjectErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "UndefinedFunction", errType: metamorphic.PGUndefinedFunctionErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "UniqueViolation", errType: metamorphic.PGUniqueViolationErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "UndefinedColumn", errType: metamorphic.PGUndefinedColumnErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "InvalidSchemaName", errType: metamorphic.PGInvalidSchemaNameErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "DuplicateObject", errType: metamorphic.PGDuplicateObjectErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "UndefinedTable", errType: metamorphic.PGUndefinedTableErr},
	{pkg: "github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode", receiver: "", name: "InvalidParameterValue", errType: metamorphic.PGInvalidParameterValueErr},
}

func (f *FunctionCountFact) AFact() {}

var Analyzer = &analysis.Analyzer{
	Name: name,
	Doc:  Doc,
	// Bogus fact to force the analysis over _all_ transitive dependencies.
	//FactTypes: []analysis.Fact{(*FunctionCountFact)(nil)},
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		var f *ast.File

		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		inspect.Preorder([]ast.Node{
			(*ast.File)(nil), (*ast.IfStmt)(nil),
		}, func(n ast.Node) {
			if fileNode, ok := n.(*ast.File); ok {
				f = fileNode
				return
			}
			ifStmt := n.(*ast.IfStmt)

			// Check if `ident != nil` occurs in the condition as a binary (sub)expression, where `ident` has type `error`.
			errObj, ok := isNonNilErrCheck(pass, ifStmt.Cond)
			if !ok {
				// We bail out because we only care about the exceptional case.
				return
			}
			if len(ifStmt.Body.List) == 0 {
				// Empty body.
				fmt.Printf("empty_body|%s|%s|%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(ifStmt.Body), pass.Fset.Position(n.Pos()))
				return
			}
			if len(ifStmt.Body.List) == 1 {
				if returnStmt, ok := (ifStmt.Body.List[0]).(*ast.ReturnStmt); ok {
					if !hasErrHandling(pass, errObj, ifStmt.Body.List) {
						fmt.Printf("single_return\t%s\t%s\t%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(returnStmt), pass.Fset.Position(n.Pos()))
						return
					}
				}
			}
			usesErr := hasAcceptableAction(pass, errObj, ifStmt.Body.List) || hasErrHandling(pass, errObj, ifStmt.Body.List)
			if !usesErr {
				fmt.Printf("err_ignored\t%s\t%s\t%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(ifStmt.Body), pass.Fset.Position(n.Pos()))
				return
			}
			if hasExit(ifStmt.Body.List) {
				fmt.Printf("has_exit\t%s\t%s\t%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(ifStmt.Body), pass.Fset.Position(n.Pos()))
				return
			}
			errTypes := dedupe(extractErrTypes(pass, ifStmt.Body.List))
			if len(errTypes) > 0 {
				fmt.Printf("interesting_case\t%s\t%s\t%v\terrTypes=%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(ifStmt.Body), pass.Fset.Position(n.Pos()), errTypes)
			} else {
				fmt.Printf("interesting_case\t%s\t%s\t%v\n", passesutil.PrettyPrint(ifStmt.Cond), passesutil.PrettyPrint(ifStmt.Body), pass.Fset.Position(n.Pos()))
			}
			probability := 0.8
			injectedExpr := fmt.Sprintf("|| (metamorphic.InjectErr(%.2f, %d, &%s, []int{%s}))", probability, numBlocks, errObj.Name(),
				strings.Trim(strings.Join(strings.Fields(fmt.Sprint(errTypes)), ","), "[]"))

			numBlocks++

			fmt.Printf("filename: %s\n", pass.Fset.Position(n.Pos()).Filename)

			if pass.Fset.Position(n.Pos()).Filename == "/Users/srosenberg/workspace/go/src/github.com/cockroachdb/cockroach/pkg/sql/colcontainer/diskqueue.go" {
				var comms []string

				for _, cg := range f.Comments {
					comments := make([]string, len(cg.List))
					for i, c := range cg.List {
						comments[i] = c.Text
					}
					if ifStmt.Pos() == cg.End() || pass.Fset.Position(ifStmt.Pos()).Line-1 == pass.Fset.Position(cg.End()).Line {
						comms = append(comms, comments...)
					}
				}
				fmt.Printf("comments: %v\n", comms)

				suggestedFixes := make([]analysis.SuggestedFix, 0)
				suggestedFixes = append(suggestedFixes, analysis.SuggestedFix{
					Message: "inject error",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     ifStmt.Cond.End(),
							End:     ifStmt.Cond.End(),
							NewText: []byte(injectedExpr),
						},
						{
							Pos:     f.Name.End(),
							End:     f.Name.End(),
							NewText: []byte("; import \"github.com/cockroachdb/cockroach/pkg/util/metamorphic\""),
						},
					},
				})
				pass.Report(analysis.Diagnostic{
					Pos:            n.Pos(),
					Message:        fmt.Sprintf("Injecting error(s) with p=%.2f", probability),
					SuggestedFixes: suggestedFixes,
				})
			}
		})

		fmt.Printf("num_blocks: %d\n", numBlocks)
		return nil, nil
	},
}

// hasExit determines if there is a call to os.Exit or xxx.Fatal _anywhere_ inside the statements.
func hasExit(statements []ast.Stmt) bool {
	var seen bool
	inspect := func(n ast.Node) (wantMore bool) {
		switch n := n.(type) {
		case *ast.CallExpr:
			se, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			// N.B. we could be more precise here by checking the package path. However, "exit" functions tend to have
			// unique names.
			if se.Sel.Name == "Exit" || se.Sel.Name == "Fatal" || se.Sel.Name == "Fatalf" {
				seen = true
				return false
			}
		}
		return true
	}
	for _, stmt := range statements {
		if ast.Inspect(stmt, inspect); seen {
			return true
		}
	}
	return false
}

// hasErrHandling determines if there is some "error handling" denoted by the statements. Unlike hasAcceptableAction,
// we don't check for uses of the error object, other than in return statements. We also look at the top-level statements
// only, not the sub-statements. Intuitively, assignments, increments, decrements, and channel sends, suggest that
// some internal state is updated in response to the error. Similarly, control-flow statements suggest that an alternative
// path may be taken in response to the error. Lastly, we check that the error object is _not_ returned since that is
// indicative of _propagating_ the error to the caller rather than _handling_ it.
func hasErrHandling(pass *analysis.Pass, errObj types.Object, statements []ast.Stmt) bool {
	var seen bool

	for _, stmt := range statements {
		switch n := stmt.(type) {
		case *ast.AssignStmt:
			seen = true
		case *ast.IncDecStmt:
			seen = true
		case *ast.SendStmt:
			seen = true
		case *ast.ExprStmt:
			if _, ok := n.X.(*ast.CallExpr); ok {
				seen = true
			}
		case *ast.IfStmt:
			seen = true
		case *ast.SwitchStmt:
			seen = true
		case *ast.ForStmt:
			seen = true
		case *ast.RangeStmt:
			seen = true
		case *ast.BranchStmt:
			seen = true
		case *ast.ReturnStmt:
			if noErrReturn(pass, errObj, n) {
				seen = true
			}
		}
	}
	return seen
}

// noErrReturn determines if the return statement does _not_ return the error object.
func noErrReturn(pass *analysis.Pass, errObj types.Object, n *ast.ReturnStmt) bool {
	numRes := len(n.Results)
	if numRes == 0 {
		return true
	}
	lastRes := n.Results[numRes-1]
	id, ok := lastRes.(*ast.Ident)
	if !ok {
		return true
	}
	if pass.TypesInfo.Uses[id] == errObj {
		return false
	}
	return true
}

// hasAcceptableAction determines if there is some action in statements which
// uses errObj in a way which excuses a nil error return value. Such actions
// include assigning the error to a value, calling Error() on it, type asserting
// it, calling a function with the error as an argument, or returning the error
// in a different return statement.
// N.B. this is a variant of the function in `returnerrcheck.go`. Specifically, we removed the case of the return statement since
// that's now checked in `hasErrHandling`.
func hasAcceptableAction(pass *analysis.Pass, errObj types.Object, statements []ast.Stmt) bool {
	var seen bool
	isObj := func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if pass.TypesInfo.Uses[id] == errObj {
				seen = true
				return true
			}
		}
		return false
	}
	inspect := func(n ast.Node) (wantMore bool) {
		switch n := n.(type) {
		case *ast.CallExpr:
			// If we call a function with the non-nil error object, we probably know what we're doing.
			for _, arg := range n.Args {
				if isObj(arg) {
					return false
				}
			}
		case *ast.AssignStmt:
			// If we assign the error to some value, we probably know what we're doing.
			for _, rhs := range n.Rhs {
				if isObj(rhs) {
					return false
				}
			}
		case *ast.KeyValueExpr:
			// If we assign the error to a field or map in a literal, we probably know what we're doing.
			if isObj(n.Value) {
				return false
			}

		case *ast.SelectorExpr:
			// If we're selecting something off of the error (i.e. err.Error()), then we probably know what we're doing.
			if isObj(n.X) {
				return false
			}
		case *ast.SendStmt:
			if isObj(n.Value) {
				return false
			}
		}
		return true
	}
	for _, stmt := range statements {
		if ast.Inspect(stmt, inspect); seen {
			return true
		}
	}
	return false
}

// extractErrTypes finds designated, custom errors in CallExprs and SelectorExprs and extracts their types.
func extractErrTypes(pass *analysis.Pass, statements []ast.Stmt) []int {
	var res []int
	inspect := func(n ast.Node) (wantMore bool) {
		switch n := n.(type) {
		case *ast.CallExpr:
			callee, ok := typeutil.Callee(pass.TypesInfo, n).(*types.Func)
			if !ok {
				return false
			}
			pkg := callee.Pkg()
			if pkg == nil {
				return false
			}
			calleePkg := pkg.Path()
			calleeFunc := callee.Name()
			calleeObj := ""

			recv := callee.Type().(*types.Signature).Recv()
			if recv != nil {
				// If there is a receiver (i.e., this is a method call), get the name of the type of the receiver.
				recvType := recv.Type()
				if pointerType, ok := recvType.(*types.Pointer); ok {
					recvType = pointerType.Elem()
				}
				named, ok := recvType.(*types.Named)
				if !ok {
					return false
				}

				calleeObj = named.Obj().Name()
			}
			for _, fn := range errFunctions {
				if fn.pkg == calleePkg && fn.receiver == calleeObj && fn.name == calleeFunc {
					res = append(res, fn.errType)
					return false
				}
			}
		//TODO(srosenberg): may be further constrained to be inside IfStmt.Cond
		case *ast.SelectorExpr:
			// N.B. we're only interested in the selector name and its package.
			obj := pass.TypesInfo.Uses[n.Sel]
			if obj == nil {
				return false
			}
			pkg := obj.Pkg().Path()
			name := n.Sel.Name

			for _, fn := range errFunctions {
				if fn.pkg == pkg && fn.name == name {
					res = append(res, fn.errType)
					return false
				}
			}
		}
		return true
	}
	for _, stmt := range statements {
		ast.Inspect(stmt, inspect)
	}
	return res
}

// dedupe removes duplicates from a slice of ints.
func dedupe(in []int) []int {
	ret := intsets.Fast{}
	for _, id := range in {
		ret.Add(id)
	}
	return ret.Ordered()
}

// isNonNilErrCheck determines if the expression is a non-nil check on an error object; i.e., expr is of the form
// `ident != nil` or `nil != ident`.
//
// N.B. this is lifted from `returnerrcheck.go`.
func isNonNilErrCheck(pass *analysis.Pass, expr ast.Expr) (errObj types.Object, ok bool) {
	binaryExpr, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return nil, false
	}
	// We care about cases when errors are idents not when they're fields
	switch binaryExpr.Op {
	case token.NEQ:
		var id *ast.Ident
		if pass.TypesInfo.Types[binaryExpr.X].Type == errorType &&
			pass.TypesInfo.Types[binaryExpr.Y].IsNil() {
			id, ok = binaryExpr.X.(*ast.Ident)
		} else if pass.TypesInfo.Types[binaryExpr.Y].Type == errorType &&
			pass.TypesInfo.Types[binaryExpr.X].IsNil() {
			id, ok = binaryExpr.Y.(*ast.Ident)
		}
		if ok {
			errObj, ok := pass.TypesInfo.Uses[id]
			return errObj, ok
		}
	case token.LAND, token.LOR:
		if errObj, ok := isNonNilErrCheck(pass, binaryExpr.X); ok {
			return errObj, ok
		}
		return isNonNilErrCheck(pass, binaryExpr.Y)
	}
	return nil, false
}
