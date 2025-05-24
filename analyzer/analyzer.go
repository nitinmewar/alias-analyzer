package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "slicealias",
	Doc:      "warns when aliased slices with unknown capacity are independently appended to",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func resolveRoot(name string, alias map[string]string) string {
	for {
		next, ok := alias[name]
		if !ok || next == name {
			return name
		}
		name = next
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.AssignStmt)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.CallExpr)(nil),
	}

	// per-function state
	alias := map[string]string{}
	unknown := map[string]bool{}
	known := map[string]bool{}

	reset := func() {
		alias = map[string]string{}
		unknown = map[string]bool{}
		known = map[string]bool{}
	}

	// helper: mark unknown/known from a make() call
	markFromMake := func(id string, call *ast.CallExpr) {
		switch len(call.Args) {
		case 1:
			unknown[id] = true

		case 2: // make([]T, len)
			if lenLit, ok := call.Args[1].(*ast.BasicLit); ok && lenLit.Value == "0" {
				unknown[id] = true // len==0 ⇒ unknown cap
			} else {
				known[id] = true // len>0 ⇒ safe (cap==len)
			}

		case 3: // make([]T, len, cap)
			if lenLit, ok := call.Args[1].(*ast.BasicLit); ok && lenLit.Value == "0" {
				if capLit, ok := call.Args[2].(*ast.BasicLit); ok {
					if capLit.Value == "0" {
						unknown[id] = true // (0,0)
					} else {
						known[id] = true // (0, >0)
					}
				}
			}
		}
	}

	ins.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {

		case *ast.FuncDecl:
			reset()

		// short assignments
		case *ast.AssignStmt:
			if len(node.Lhs) != 1 || len(node.Rhs) != 1 {
				return
			}
			lhs, ok := node.Lhs[0].(*ast.Ident)
			if !ok {
				return
			}
			rhs := node.Rhs[0]
			tv := pass.TypesInfo.Types[rhs]
			if tv.Type == nil {
				return
			}
			if _, isSlice := tv.Type.(*types.Slice); !isSlice {
				return
			}

			switch e := rhs.(type) {
			case *ast.CompositeLit:
				if len(e.Elts) == 0 {
					unknown[lhs.Name] = true
				}
			case *ast.CallExpr:
				if fn, ok := e.Fun.(*ast.Ident); ok && fn.Name == "make" {
					markFromMake(lhs.Name, e)
				}
			case *ast.Ident:
				if lhs.Name != e.Name {
					alias[lhs.Name] = e.Name
				}
			}

		// var specs
		case *ast.ValueSpec:
			for i, name := range node.Names {
				if node.Type != nil && len(node.Values) == 0 {
					if _, isSlice := pass.TypesInfo.Types[node.Type].Type.(*types.Slice); isSlice {
						unknown[name.Name] = true
					}
					continue
				}
				if i >= len(node.Values) {
					continue
				}
				rhs := node.Values[i]
				tv := pass.TypesInfo.Types[rhs]
				if tv.Type == nil {
					continue
				}
				if _, isSlice := tv.Type.(*types.Slice); !isSlice {
					continue
				}
				switch e := rhs.(type) {
				case *ast.CompositeLit:
					if len(e.Elts) == 0 {
						unknown[name.Name] = true
					}
				case *ast.CallExpr:
					if fn, ok := e.Fun.(*ast.Ident); ok && fn.Name == "make" {
						markFromMake(name.Name, e)
					}
				}
			}

		// append calls
		case *ast.CallExpr:
			id, ok := node.Fun.(*ast.Ident)
			if !ok || id.Name != "append" || len(node.Args) == 0 {
				return
			}
			target, ok := node.Args[0].(*ast.Ident)
			if !ok {
				return
			}
			root := resolveRoot(target.Name, alias)
			if root == target.Name { // not aliased
				return
			}
			if known[root] { // root has safe cap
				return
			}
			if !unknown[root] { // root capacity known & safe
				return
			}
			pass.Reportf(node.Pos(),
				"append to alias '%s' of unknown-capacity slice '%s' may cause memory divergence",
				target.Name, root)
		}
	})
	return nil, nil
}
