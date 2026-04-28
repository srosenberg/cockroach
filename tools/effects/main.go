// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// effects derives an upper bound on the read/write heap effects of every
// function in a Go package. The derivation follows the syntax-directed
// "footprint" of region logic (Rosenberg, "Region Logic: local reasoning for
// Java programs and its automation", §2.5.1, §3.3.10).
//
// V2 scope:
//   - Read effects via ftpt(E) over expressions.
//   - Write effects from assignments, ++/--, range loop targets, send.
//   - Interprocedural inlining within a single package: each function's
//     effects are computed in reverse-topological order of the call graph,
//     and at each call site the callee's effect set is substituted (formals
//     and receiver → actuals). Recursive functions stay opaque.
//   - Builtin specials: copy, append, delete, clear, len, cap, new, make.
//   - Cross-package calls remain opaque (we only see arg footprints).
//   - No concurrency reasoning. Effects are an upper bound on what a single
//     sequential execution can read/write.
//
// Run:
//
//	./effects ./pkg/util
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type kind int

const (
	read kind = iota
	write
	send
	recv
)

func (k kind) String() string {
	switch k {
	case read:
		return "rd"
	case write:
		return "wr"
	case send:
		return "send"
	case recv:
		return "recv"
	}
	return "?"
}

// atom is a single effect: a kind and a textual location ("path").
//
// Path forms:
//
//	"x"                 local var (or unqualified package-level name)
//	"<expr>.f"          field access through value or pointer
//	"<expr>[*]"         element of slice/map/array
//	"*<expr>"           pointer dereference
//	"alloc"             the allocation set (for x := new(T) / make / &T{})
type atom struct {
	k    kind
	path string
}

func (a atom) String() string { return a.k.String() + " " + a.path }

type effectSet map[atom]struct{}

func (s effectSet) add(a atom)         { s[a] = struct{}{} }
func (s effectSet) union(o effectSet)  { for a := range o { s[a] = struct{}{} } }

// canonicalize normalizes atoms so that paths describing the same underlying
// memory region collapse into one entry, then drops any atoms still subsumed
// by another atom of the same kind.
//
// Step 1 — slice-op stripping. In Go, `s[i:j]` produces a slice header
// pointing to the same backing array as `s`; `s[i:j][*]` and `s[*]` denote
// the same memory region. We rewrite every path by removing any embedded
// slice expression `[a:b]` / `[:b]` / `[a:]` / `[:]` / `[a:b:c]`, leaving the
// underlying storage path. Index expressions `[i]` (no top-level colon
// inside the brackets) are preserved — they read a distinct element, possibly
// of a different type (e.g. `[][]int`).
//
// Step 2 — bracket-peel subsumption. Drops `<k> X[*]` when there exists
// `<k> Y[*]` in the set with `Y` reachable from `X` by peeling trailing
// balanced `[...]` segments. This catches residual nesting that step 1
// doesn't normalize away (e.g. `s[i][*]` when `s[*]` is also present, where
// `s[i]` was a real index but the element was further indexed).
//
// Both steps are per-kind: `rd` does not subsume `wr` and vice versa.
func (s effectSet) canonicalize() {
	if len(s) == 0 {
		return
	}
	// Step 1: rewrite paths by stripping embedded slice ops.
	for a := range s {
		stripped := stripSliceOps(a.path)
		if stripped == a.path {
			continue
		}
		delete(s, a)
		s.add(atom{a.k, stripped})
	}
	if len(s) < 2 {
		return
	}
	// Step 2: drop X[*] subsumed by Y[*] reachable via trailing-bracket peel.
	var toDrop []atom
	for a := range s {
		base, ok := strings.CutSuffix(a.path, "[*]")
		if !ok {
			continue
		}
		cur := base
		for {
			prefix, peeled := peelTrailingBracket(cur)
			if !peeled {
				break
			}
			if _, ok := s[atom{a.k, prefix + "[*]"}]; ok {
				toDrop = append(toDrop, a)
				break
			}
			cur = prefix
		}
	}
	for _, a := range toDrop {
		delete(s, a)
	}
}

// stripSliceOps removes every embedded slice expression from path. A bracket
// group `[...]` is a slice expression iff it contains a `:` at the top level
// (i.e. not inside a nested `[]` or `()`). Index groups (no top-level colon)
// are preserved verbatim. Bracket matching is purely textual; paths
// constructed by this tool never contain string literals.
func stripSliceOps(path string) string {
	if !strings.ContainsRune(path, '[') {
		return path
	}
	var b strings.Builder
	b.Grow(len(path))
	i := 0
	for i < len(path) {
		if path[i] != '[' {
			b.WriteByte(path[i])
			i++
			continue
		}
		// Find the matching ']' and detect a top-level ':' inside.
		depth, parenDepth := 0, 0
		hasTopColon := false
		j := i
		for j < len(path) {
			switch path[j] {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					goto done
				}
			case '(':
				parenDepth++
			case ')':
				parenDepth--
			case ':':
				if depth == 1 && parenDepth == 0 {
					hasTopColon = true
				}
			}
			j++
		}
	done:
		if j >= len(path) {
			// Unmatched bracket; emit the rest verbatim.
			b.WriteString(path[i:])
			return b.String()
		}
		if hasTopColon {
			i = j + 1 // skip the entire slice expression
		} else {
			b.WriteString(path[i : j+1])
			i = j + 1
		}
	}
	return b.String()
}

// peelTrailingBracket strips the rightmost balanced [...] from s. Returns the
// prefix and true if a bracket pair was peeled. Bracket matching is purely
// textual; paths constructed by this tool never contain string literals so
// `[` / `]` characters always denote indexing.
func peelTrailingBracket(s string) (string, bool) {
	n := len(s)
	if n == 0 || s[n-1] != ']' {
		return s, false
	}
	depth := 0
	for i := n - 1; i >= 0; i-- {
		switch s[i] {
		case ']':
			depth++
		case '[':
			depth--
			if depth == 0 {
				return s[:i], true
			}
		}
	}
	return s, false
}

func (s effectSet) sorted() []atom {
	out := make([]atom, 0, len(s))
	for a := range s {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].k != out[j].k {
			return out[i].k < out[j].k
		}
		return out[i].path < out[j].path
	})
	return out
}

// funcInfo records what we know about an intra-package function, so that
// callers can inline its effects with substitution.
type funcInfo struct {
	decl     *ast.FuncDecl
	obj      *types.Func
	recvName string   // receiver var name (e.g., "r"); "" if no receiver / unnamed
	formals  []string // ordered formal names; "" or "_" entries are unnamed
	effects  effectSet
	opaque   bool // body absent, recursive, or otherwise unanalyzable
}

type walker struct {
	fset  *token.FileSet
	info  *types.Info
	pkg   *types.Package
	funcs map[*types.Func]*funcInfo // package-scoped func index
}

// isLocalIdent reports whether an effect path is a bare local-variable name.
// "local" means a *types.Var whose enclosing scope is not the package scope.
func (w *walker) isLocalIdent(path string) bool {
	if w.info == nil || w.pkg == nil {
		return false
	}
	if path == "alloc" {
		return false // distinguished allocation atom
	}
	if strings.ContainsAny(path, ".[*") {
		return false
	}
	// Look up in package scope: anything *not* found there is a local.
	if w.pkg.Scope().Lookup(path) != nil {
		return false
	}
	return true
}

// exprText renders an expression back to source for use as a path key.
func (w *walker) exprText(e ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, w.fset, e)
	return buf.String()
}

// isPkgName reports whether id is a package qualifier (e.g., "math" in
// "math.MaxInt"). Package qualifiers don't contribute heap reads.
func (w *walker) isPkgName(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	if !ok || w.info == nil {
		return false
	}
	_, ok = w.info.Uses[id].(*types.PkgName)
	return ok
}

// isMethodSel reports whether sel resolves to a method (bound or expression
// form), not a field. Method selections don't read a heap location.
func (w *walker) isMethodSel(sel *ast.SelectorExpr) bool {
	if w.info == nil {
		return false
	}
	s := w.info.Selections[sel]
	if s == nil {
		// A nil selection means qualified ident (pkg.Name) or builtin.
		if obj := w.info.Uses[sel.Sel]; obj != nil {
			_, ok := obj.(*types.Func)
			return ok
		}
		return false
	}
	return s.Kind() == types.MethodVal || s.Kind() == types.MethodExpr
}

// nonHeapIdent reports whether id refers to something that doesn't live on the
// heap (package, type, constant, builtin, function value).
func (w *walker) nonHeapIdent(id *ast.Ident) bool {
	if w.info == nil {
		return false
	}
	switch w.info.Uses[id].(type) {
	case *types.PkgName, *types.TypeName, *types.Const, *types.Builtin, *types.Func, *types.Nil:
		return true
	}
	return false
}

// ftpt computes the read footprint of an expression (Fig 2.11 of the thesis,
// generalized to Go). Returns a fresh set.
func (w *walker) ftpt(e ast.Expr) effectSet {
	out := effectSet{}
	if e == nil {
		return out
	}
	switch e := e.(type) {

	case *ast.BasicLit:
		// no effect

	case *ast.Ident:
		if e.Name == "_" || w.nonHeapIdent(e) {
			return out
		}
		out.add(atom{read, e.Name})

	case *ast.SelectorExpr:
		if w.isPkgName(e.X) {
			// pkg.Foo: only a heap read if Sel is a package-level var.
			if obj, ok := w.info.Uses[e.Sel].(*types.Var); ok && obj.Pkg() != nil {
				out.add(atom{read, w.exprText(e)})
			}
			return out
		}
		out.union(w.ftpt(e.X))
		if !w.isMethodSel(e) {
			out.add(atom{read, w.exprText(e)})
		}

	case *ast.IndexExpr:
		// Could also be a generic instantiation (T[U]); type info distinguishes.
		if w.isTypeExpr(e) {
			return out
		}
		out.union(w.ftpt(e.X))
		out.union(w.ftpt(e.Index))
		out.add(atom{read, w.exprText(e.X) + "[*]"})

	case *ast.IndexListExpr:
		// Always a generic instantiation: T[A, B]. No runtime read.

	case *ast.SliceExpr:
		// Slicing reads the source's slice header (pointer/len/cap) to
		// validate bounds and construct a new header — but it does NOT read
		// element memory. The element read happens later if/when the sub-slice
		// is indexed or copied from. ftpt(e.X) covers the header read.
		out.union(w.ftpt(e.X))
		out.union(w.ftpt(e.Low))
		out.union(w.ftpt(e.High))
		out.union(w.ftpt(e.Max))

	case *ast.StarExpr:
		// In a value position: pointer deref.
		out.union(w.ftpt(e.X))
		out.add(atom{read, "*" + w.exprText(e.X)})

	case *ast.UnaryExpr:
		switch e.Op {
		case token.AND:
			// &x: takes address; no value read of x's *contents*. We still
			// need to "see" any subexpressions used to compute the address.
			// &T{...} on a composite literal does heap-allocate.
			if _, ok := unparen(e.X).(*ast.CompositeLit); ok {
				out.add(atom{write, "alloc"})
			}
			out.union(w.ftpt(e.X))
		case token.ARROW:
			// <-ch: channel receive.
			out.union(w.ftpt(e.X))
			out.add(atom{recv, w.exprText(e.X)})
		default:
			out.union(w.ftpt(e.X))
		}

	case *ast.BinaryExpr:
		out.union(w.ftpt(e.X))
		out.union(w.ftpt(e.Y))

	case *ast.ParenExpr:
		out.union(w.ftpt(e.X))

	case *ast.CallExpr:
		w.callEffects(e, out)

	case *ast.TypeAssertExpr:
		out.union(w.ftpt(e.X))

	case *ast.CompositeLit:
		// Allocation site (unless it's a value composite of a stack-only type;
		// we conservatively mark allocation).
		if w.isHeapComposite(e) {
			out.add(atom{write, "alloc"})
		}
		for _, el := range e.Elts {
			out.union(w.ftpt(el))
		}

	case *ast.KeyValueExpr:
		out.union(w.ftpt(e.Key))
		out.union(w.ftpt(e.Value))

	case *ast.FuncLit:
		// Closure: skip the body. A complete analysis would treat captures.

	case *ast.Ellipsis:
		out.union(w.ftpt(e.Elt))
	}
	return out
}

// callEffects accumulates the effects of a call expression into out.
// It handles three cases in order: (1) builtin specials, (2) intra-package
// callee inlining with substitution, (3) opaque fallback (just argument
// footprints).
func (w *walker) callEffects(e *ast.CallExpr, out effectSet) {
	if w.handleBuiltin(e, out) {
		return
	}
	if w.inlineCallee(e, out) {
		return
	}
	out.union(w.ftpt(e.Fun))
	for _, a := range e.Args {
		out.union(w.ftpt(a))
	}
}

// handleBuiltin recognizes a call to a Go builtin and emits the appropriate
// effect set. Returns true if the call was handled.
func (w *walker) handleBuiltin(e *ast.CallExpr, out effectSet) bool {
	id, ok := e.Fun.(*ast.Ident)
	if !ok || w.info == nil {
		return false
	}
	if _, isBuiltin := w.info.Uses[id].(*types.Builtin); !isBuiltin {
		return false
	}
	switch id.Name {
	case "new":
		out.add(atom{write, "alloc"})
	case "make":
		out.add(atom{write, "alloc"})
		for _, a := range e.Args[1:] {
			out.union(w.ftpt(a))
		}
	case "copy":
		// copy(dst, src): writes dst elements, reads src elements.
		if len(e.Args) >= 1 {
			out.union(w.ftpt(e.Args[0]))
			out.add(atom{write, w.exprText(e.Args[0]) + "[*]"})
		}
		if len(e.Args) >= 2 {
			out.union(w.ftpt(e.Args[1]))
			out.add(atom{read, w.exprText(e.Args[1]) + "[*]"})
		}
	case "append":
		// append(s, x...): reads s elements; may grow (alloc); reads varargs.
		if len(e.Args) >= 1 {
			out.union(w.ftpt(e.Args[0]))
			out.add(atom{read, w.exprText(e.Args[0]) + "[*]"})
		}
		for _, a := range e.Args[1:] {
			out.union(w.ftpt(a))
		}
		out.add(atom{write, "alloc"})
	case "delete":
		if len(e.Args) >= 1 {
			out.union(w.ftpt(e.Args[0]))
			out.add(atom{write, w.exprText(e.Args[0]) + "[*]"})
		}
		if len(e.Args) >= 2 {
			out.union(w.ftpt(e.Args[1]))
		}
	case "clear":
		if len(e.Args) >= 1 {
			out.union(w.ftpt(e.Args[0]))
			out.add(atom{write, w.exprText(e.Args[0]) + "[*]"})
		}
	case "len", "cap", "min", "max", "complex", "real", "imag", "panic", "print", "println", "recover", "close":
		// No heap effects beyond the argument footprints.
		for _, a := range e.Args {
			out.union(w.ftpt(a))
		}
	default:
		return false
	}
	return true
}

// inlineCallee tries to resolve an intra-package call to a known funcInfo and
// inline its precomputed effect set with formal→actual substitution.
// Returns true if inlining was performed.
func (w *walker) inlineCallee(e *ast.CallExpr, out effectSet) bool {
	if w.funcs == nil {
		return false
	}
	obj := w.callTarget(e)
	if obj == nil {
		return false
	}
	fi, ok := w.funcs[obj]
	if !ok || fi.opaque || fi.effects == nil {
		return false
	}
	subst := map[string]string{}
	if fi.recvName != "" {
		if sel, ok := unparen(e.Fun).(*ast.SelectorExpr); ok {
			subst[fi.recvName] = w.exprText(sel.X)
		}
	}
	for i, name := range fi.formals {
		if name == "" || name == "_" || i >= len(e.Args) {
			continue
		}
		subst[name] = w.exprText(e.Args[i])
	}
	for a := range fi.effects {
		out.add(substAtom(a, subst))
	}
	// Argument-evaluation footprints (the args may themselves contain
	// nontrivial sub-expressions, e.g., method calls).
	for _, a := range e.Args {
		out.union(w.ftpt(a))
	}
	if sel, ok := unparen(e.Fun).(*ast.SelectorExpr); ok {
		out.union(w.ftpt(sel.X))
	}
	return true
}

// callTarget resolves a call expression to the *types.Func it invokes, or nil
// if the target can't be determined statically (function value, interface
// method, etc.). Returned funcs are canonicalized via Origin() so that
// instantiated generic methods map back to their declaration.
func (w *walker) callTarget(call *ast.CallExpr) *types.Func {
	if w.info == nil {
		return nil
	}
	var f *types.Func
	switch fun := unparen(call.Fun).(type) {
	case *ast.Ident:
		f, _ = w.info.Uses[fun].(*types.Func)
	case *ast.SelectorExpr:
		if sel := w.info.Selections[fun]; sel != nil {
			if sel.Kind() != types.MethodVal {
				return nil
			}
			f, _ = sel.Obj().(*types.Func)
		} else {
			// Package-qualified plain function call: pkg.F.
			f, _ = w.info.Uses[fun.Sel].(*types.Func)
		}
	}
	if f == nil {
		return nil
	}
	return f.Origin()
}

// substAtom applies subst to an atom's path.
func substAtom(a atom, subst map[string]string) atom {
	if len(subst) == 0 || a.path == "alloc" {
		return a
	}
	return atom{a.k, substPath(a.path, subst)}
}

// substPath rewrites the leading identifier of a path according to subst.
// Path forms: "name", "name.f.g", "name[*]", "*name", "name.f[*]", etc.
func substPath(path string, subst map[string]string) string {
	if path == "" {
		return path
	}
	star := ""
	if path[0] == '*' {
		star = "*"
		path = path[1:]
	}
	i := 0
	for i < len(path) && isIdentByte(path[i]) {
		i++
	}
	if i == 0 {
		return star + path
	}
	root, rest := path[:i], path[i:]
	if r, ok := subst[root]; ok {
		return star + r + rest
	}
	return star + path
}

func isIdentByte(c byte) bool {
	return c == '_' ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

// isTypeExpr returns true if e (used in an index position) is a generic type
// instantiation rather than an indexing operation.
func (w *walker) isTypeExpr(e *ast.IndexExpr) bool {
	if w.info == nil {
		return false
	}
	tv, ok := w.info.Types[e.X]
	if !ok {
		return false
	}
	return tv.IsType()
}

// isHeapComposite reports whether a composite literal allocates on the heap.
// Slice/map/chan literals always do. Struct/array literals don't, in
// isolation; if their address is taken (&T{...}) the UnaryExpr handler emits
// the alloc effect.
func (w *walker) isHeapComposite(e *ast.CompositeLit) bool {
	if w.info == nil {
		return false
	}
	tv, ok := w.info.Types[e]
	if !ok {
		return false
	}
	switch tv.Type.Underlying().(type) {
	case *types.Slice, *types.Map, *types.Chan:
		return true
	}
	return false
}

// lvalEffect returns the write atom for an l-value, plus any read effects
// needed to *resolve* the l-value's path (e.g., to write x.f.g we must read x
// and x.f).
func (w *walker) lvalEffect(e ast.Expr) (atom, effectSet) {
	out := effectSet{}
	switch e := e.(type) {
	case *ast.Ident:
		if e.Name == "_" {
			return atom{}, out
		}
		return atom{write, e.Name}, out
	case *ast.SelectorExpr:
		out.union(w.ftpt(e.X))
		return atom{write, w.exprText(e)}, out
	case *ast.IndexExpr:
		out.union(w.ftpt(e.X))
		out.union(w.ftpt(e.Index))
		return atom{write, w.exprText(e.X) + "[*]"}, out
	case *ast.StarExpr:
		out.union(w.ftpt(e.X))
		return atom{write, "*" + w.exprText(e.X)}, out
	case *ast.ParenExpr:
		return w.lvalEffect(e.X)
	}
	return atom{write, w.exprText(e)}, out
}

// stmt accumulates effects of a statement into out.
func (w *walker) stmt(s ast.Stmt, out effectSet) {
	if s == nil {
		return
	}
	switch s := s.(type) {

	case *ast.AssignStmt:
		for _, r := range s.Rhs {
			out.union(w.ftpt(r))
		}
		for _, l := range s.Lhs {
			switch s.Tok {
			case token.DEFINE:
				if id, ok := l.(*ast.Ident); ok && id.Name != "_" {
					out.add(atom{write, id.Name})
				}
			case token.ASSIGN:
				wr, reads := w.lvalEffect(l)
				if wr.path != "" {
					out.add(wr)
				}
				out.union(reads)
			default:
				// op-assign (+=, etc.): both reads and writes the lhs.
				out.union(w.ftpt(l))
				wr, reads := w.lvalEffect(l)
				if wr.path != "" {
					out.add(wr)
				}
				out.union(reads)
			}
		}

	case *ast.IncDecStmt:
		out.union(w.ftpt(s.X))
		wr, reads := w.lvalEffect(s.X)
		if wr.path != "" {
			out.add(wr)
		}
		out.union(reads)

	case *ast.ExprStmt:
		out.union(w.ftpt(s.X))

	case *ast.ReturnStmt:
		for _, r := range s.Results {
			out.union(w.ftpt(r))
		}

	case *ast.IfStmt:
		w.stmt(s.Init, out)
		out.union(w.ftpt(s.Cond))
		w.stmt(s.Body, out)
		w.stmt(s.Else, out)

	case *ast.ForStmt:
		w.stmt(s.Init, out)
		out.union(w.ftpt(s.Cond))
		w.stmt(s.Post, out)
		w.stmt(s.Body, out)

	case *ast.RangeStmt:
		out.union(w.ftpt(s.X))
		out.add(atom{read, w.exprText(s.X) + "[*]"})
		if s.Key != nil && !isBlankIdent(s.Key) {
			if s.Tok == token.DEFINE {
				if id, ok := s.Key.(*ast.Ident); ok {
					out.add(atom{write, id.Name})
				}
			} else {
				wr, reads := w.lvalEffect(s.Key)
				if wr.path != "" {
					out.add(wr)
				}
				out.union(reads)
			}
		}
		if s.Value != nil && !isBlankIdent(s.Value) {
			if s.Tok == token.DEFINE {
				if id, ok := s.Value.(*ast.Ident); ok {
					out.add(atom{write, id.Name})
				}
			} else {
				wr, reads := w.lvalEffect(s.Value)
				if wr.path != "" {
					out.add(wr)
				}
				out.union(reads)
			}
		}
		w.stmt(s.Body, out)

	case *ast.SwitchStmt:
		w.stmt(s.Init, out)
		out.union(w.ftpt(s.Tag))
		w.stmt(s.Body, out)

	case *ast.TypeSwitchStmt:
		w.stmt(s.Init, out)
		w.stmt(s.Assign, out)
		w.stmt(s.Body, out)

	case *ast.CaseClause:
		for _, e := range s.List {
			out.union(w.ftpt(e))
		}
		for _, st := range s.Body {
			w.stmt(st, out)
		}

	case *ast.SelectStmt:
		w.stmt(s.Body, out)

	case *ast.CommClause:
		w.stmt(s.Comm, out)
		for _, st := range s.Body {
			w.stmt(st, out)
		}

	case *ast.BlockStmt:
		for _, st := range s.List {
			w.stmt(st, out)
		}

	case *ast.DeclStmt:
		gd, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			return
		}
		for _, sp := range gd.Specs {
			vs, ok := sp.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name != "_" {
					out.add(atom{write, name.Name})
				}
			}
			for _, val := range vs.Values {
				out.union(w.ftpt(val))
			}
		}

	case *ast.DeferStmt:
		// Defer's call body executes at return. Including its arg footprints
		// here keeps the result a sound upper bound.
		out.union(w.ftpt(s.Call))

	case *ast.GoStmt:
		out.union(w.ftpt(s.Call))

	case *ast.SendStmt:
		out.union(w.ftpt(s.Chan))
		out.union(w.ftpt(s.Value))
		out.add(atom{send, w.exprText(s.Chan)})

	case *ast.LabeledStmt:
		w.stmt(s.Stmt, out)

	case *ast.BranchStmt, *ast.EmptyStmt:
		// no effect
	}
}

func isBlankIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "_"
}

func unparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}

// effectsForFunc returns the upper-bound effect set for a function.
func (w *walker) effectsForFunc(fd *ast.FuncDecl) effectSet {
	out := effectSet{}
	if fd.Body == nil {
		return out
	}
	w.stmt(fd.Body, out)
	out.canonicalize()
	return out
}

func funcSig(fset *token.FileSet, fd *ast.FuncDecl) string {
	var b strings.Builder
	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		var rb bytes.Buffer
		_ = printer.Fprint(&rb, fset, fd.Recv.List[0].Type)
		b.WriteString("(")
		b.WriteString(rb.String())
		b.WriteString(") ")
	}
	b.WriteString(fd.Name.Name)
	return b.String()
}

// collectFuncs builds the funcInfo index for a package.
func collectFuncs(pkg *packages.Package) map[*types.Func]*funcInfo {
	out := map[*types.Func]*funcInfo{}
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			obj, _ := pkg.TypesInfo.Defs[fd.Name].(*types.Func)
			if obj == nil {
				continue
			}
			obj = obj.Origin()
			fi := &funcInfo{decl: fd, obj: obj, opaque: fd.Body == nil}
			if fd.Recv != nil && len(fd.Recv.List) > 0 {
				if names := fd.Recv.List[0].Names; len(names) > 0 {
					fi.recvName = names[0].Name
				}
			}
			if fd.Type.Params != nil {
				for _, field := range fd.Type.Params.List {
					if len(field.Names) == 0 {
						fi.formals = append(fi.formals, "")
						continue
					}
					for _, n := range field.Names {
						fi.formals = append(fi.formals, n.Name)
					}
				}
			}
			out[obj] = fi
		}
	}
	return out
}

// calleesOf returns the set of intra-package functions called from fi's body.
func calleesOf(fi *funcInfo, w *walker) map[*types.Func]struct{} {
	out := map[*types.Func]struct{}{}
	if fi.decl.Body == nil {
		return out
	}
	ast.Inspect(fi.decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if obj := w.callTarget(call); obj != nil {
			if _, intra := w.funcs[obj]; intra {
				out[obj] = struct{}{}
			}
		}
		return true
	})
	return out
}

// topoOrder returns funcs in reverse-topological order (callees before
// callers). Functions in cycles are appended at the end and marked opaque.
func topoOrder(funcs map[*types.Func]*funcInfo, w *walker) []*funcInfo {
	callees := map[*types.Func]map[*types.Func]struct{}{}
	callers := map[*types.Func]map[*types.Func]struct{}{}
	for obj, fi := range funcs {
		cs := calleesOf(fi, w)
		// Drop self-edges; they don't help Kahn's, and we mark self-recursive
		// funcs as opaque below.
		if _, self := cs[obj]; self {
			fi.opaque = true
			delete(cs, obj)
		}
		callees[obj] = cs
		for c := range cs {
			if callers[c] == nil {
				callers[c] = map[*types.Func]struct{}{}
			}
			callers[c][obj] = struct{}{}
		}
	}
	inDeg := map[*types.Func]int{}
	for obj := range funcs {
		inDeg[obj] = len(callees[obj])
	}
	var queue []*types.Func
	for obj, d := range inDeg {
		if d == 0 {
			queue = append(queue, obj)
		}
	}
	done := map[*types.Func]bool{}
	var order []*funcInfo
	for len(queue) > 0 {
		obj := queue[0]
		queue = queue[1:]
		done[obj] = true
		order = append(order, funcs[obj])
		for caller := range callers[obj] {
			if done[caller] {
				continue
			}
			inDeg[caller]--
			if inDeg[caller] == 0 {
				queue = append(queue, caller)
			}
		}
	}
	// Funcs left in cycles: process them last, with bodies treated as opaque
	// for inlining purposes (their effects still get computed using arg
	// footprints + non-recursive callees).
	for obj, fi := range funcs {
		if !done[obj] {
			fi.opaque = true
			order = append(order, fi)
		}
	}
	return order
}

// renderEffects formats a sorted slice of atoms for display.
//
// In compact mode (the default), atoms sharing a leading identifier (paths of
// the form `<id>.<tail>`) are grouped: e.g. `rd r.buffer, rd r.head, rd r.tail`
// becomes `rd r.{buffer, head, tail}`. Atoms whose path doesn't start with
// `<id>.` (singletons like `wr alloc`, `rd buf[*]`, `*p`) print on their own
// line. Within a group, tails are sorted lexicographically; if a group has
// more than maxItems entries, it is truncated with `... (+N more)`.
//
// In full mode, every atom prints on its own line.
//
// Both modes prefix each line with two spaces and return the lines (no
// trailing newlines).
func renderEffects(atoms []atom, full bool) []string {
	if len(atoms) == 0 {
		return []string{"  (none)"}
	}
	if full {
		out := make([]string, len(atoms))
		for i, a := range atoms {
			out[i] = "  " + a.String()
		}
		return out
	}
	const maxItems = 8
	type gkey struct {
		k     kind
		pivot string
	}
	groups := map[gkey][]string{}
	var singletons []atom
	for _, a := range atoms {
		pivot, tail, ok := splitPivot(a.path)
		if !ok {
			singletons = append(singletons, a)
			continue
		}
		groups[gkey{a.k, pivot}] = append(groups[gkey{a.k, pivot}], tail)
	}
	type item struct {
		k    kind
		name string
		line string
	}
	items := make([]item, 0, len(groups)+len(singletons))
	for gk, tails := range groups {
		if len(tails) == 1 {
			items = append(items, item{
				k: gk.k, name: gk.pivot,
				line: fmt.Sprintf("%s %s.%s", gk.k, gk.pivot, tails[0]),
			})
			continue
		}
		sort.Strings(tails)
		if len(tails) > maxItems {
			extra := len(tails) - (maxItems - 1)
			tails = append(tails[:maxItems-1:maxItems-1],
				fmt.Sprintf("... (+%d more)", extra))
		}
		items = append(items, item{
			k: gk.k, name: gk.pivot,
			line: fmt.Sprintf("%s %s.{%s}", gk.k, gk.pivot, strings.Join(tails, ", ")),
		})
	}
	for _, a := range singletons {
		items = append(items, item{k: a.k, name: a.path, line: a.String()})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].k != items[j].k {
			return items[i].k < items[j].k
		}
		return items[i].name < items[j].name
	})
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = "  " + it.line
	}
	return out
}

// splitPivot splits a path into (pivot, tail) at the first '.' if the part
// before the dot is a bare identifier. Returns ok=false otherwise (path has
// no top-level field selection — singleton).
func splitPivot(path string) (pivot, tail string, ok bool) {
	i := strings.IndexByte(path, '.')
	if i <= 0 {
		return "", "", false
	}
	for j := 0; j < i; j++ {
		if !isIdentByte(path[j]) {
			return "", "", false
		}
	}
	return path[:i], path[i+1:], true
}

func main() {
	heapOnly := flag.Bool("heap", true, "filter out reads/writes of bare local variables")
	full := flag.Bool("full", false, "print one atom per line instead of grouping by leading identifier")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-heap=false] [-full] <pkg-pattern>...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps |
			packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		log.Fatalf("load: %v", err)
	}
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			fmt.Fprintln(os.Stderr, e)
		}
		w := &walker{fset: pkg.Fset, info: pkg.TypesInfo, pkg: pkg.Types}
		w.funcs = collectFuncs(pkg)
		order := topoOrder(w.funcs, w)
		// Compute effects in reverse-topological order (callees first).
		for _, fi := range order {
			fi.effects = w.effectsForFunc(fi.decl)
		}
		// Print results in source order: walk files and decls.
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				obj, _ := pkg.TypesInfo.Defs[fd.Name].(*types.Func)
				if obj != nil {
					obj = obj.Origin()
				}
				fi := w.funcs[obj]
				if fi == nil {
					continue
				}
				suffix := ""
				if fi.opaque && fd.Body != nil {
					suffix = "  // recursive: opaque"
				}
				fmt.Printf("%s: %s%s\n", pkg.Fset.Position(fd.Pos()), funcSig(pkg.Fset, fd), suffix)
				atoms := fi.effects.sorted()
				if *heapOnly {
					filtered := atoms[:0]
					for _, a := range atoms {
						if w.isLocalIdent(a.path) {
							continue
						}
						filtered = append(filtered, a)
					}
					atoms = filtered
				}
				for _, line := range renderEffects(atoms, *full) {
					fmt.Println(line)
				}
			}
		}
	}
}
