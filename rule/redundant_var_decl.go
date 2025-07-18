package rule

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/mgechev/revive/lint"
)

// RedundantVarDeclRule detects redundant variable declarations
// that are immediately reassigned using short variable declaration.
type RedundantVarDeclRule struct{}

// Name returns the rule name.
func (*RedundantVarDeclRule) Name() string {
	return "redundant-var-decl"
}

// Apply applies the rule to given file.
func (*RedundantVarDeclRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	// Walk through each function and function literal
	ast.Inspect(file.AST, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Skip init functions as they have special semantics
			if node.Name.Name == "init" {
				return false
			}
			if node.Body != nil {
				failures = append(failures, analyzeFunc(node.Body, file, file.Pkg.TypesInfo())...)
			}
		case *ast.FuncLit:
			// Check function literals
			if node.Body != nil {
				failures = append(failures, analyzeFunc(node.Body, file, file.Pkg.TypesInfo())...)
			}
		}
		return true
	})

	return failures
}

func analyzeFunc(body *ast.BlockStmt, file *lint.File, typesInfo *types.Info) []lint.Failure {
	var failures []lint.Failure
	tracker := &varTracker{
		vars:      map[string]*varDecl{},
		file:      file,
		typesInfo: typesInfo,
	}

	// Process all statements in the function body
	analyzeBlock(body, tracker, &failures)

	return failures
}

type varDecl struct {
	node      ast.Node
	pos       token.Pos
	used      bool
	redefined bool
}

type varTracker struct {
	vars      map[string]*varDecl
	file      *lint.File
	typesInfo *types.Info
}

// singleVarSpec represents a single variable from a multi-variable declaration.
type singleVarSpec struct {
	spec  *ast.ValueSpec
	index int
	name  *ast.Ident
}

func (s *singleVarSpec) Pos() token.Pos { return s.name.Pos() }
func (s *singleVarSpec) End() token.Pos { return s.name.End() }

func analyzeBlock(block *ast.BlockStmt, tracker *varTracker, failures *[]lint.Failure) {
	// Create a new scope for this block
	blockTracker := &varTracker{
		vars:      map[string]*varDecl{},
		file:      tracker.file,
		typesInfo: tracker.typesInfo,
	}

	// Copy parent scope variables
	for k, v := range tracker.vars {
		blockTracker.vars[k] = v
	}

	// Process statements in order
	for _, stmt := range block.List {
		processStatement(stmt, blockTracker, failures)
	}
}

func processStatement(stmt ast.Stmt, tracker *varTracker, failures *[]lint.Failure) {
	switch s := stmt.(type) {
	case *ast.DeclStmt:
		// Handle variable declarations
		if genDecl, ok := s.Decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// Skip multiple variable declarations to avoid false positives
					// It's complex to determine which specific variables in a multi-var
					// declaration are redundant when some might be used before :=
					if len(valueSpec.Names) > 1 {
						continue
					}
					for i, name := range valueSpec.Names {
						if name.Name == "_" {
							continue
						}
						// Create a synthetic node for each variable to track them individually
						tracker.vars[name.Name] = &varDecl{
							node: &singleVarSpec{
								spec:  valueSpec,
								index: i,
								name:  name,
							},
							pos: name.Pos(),
						}
					}
				}
			}
		}

	case *ast.AssignStmt:
		// Mark variables in RHS as used first
		markUsedInExpr(s.Rhs, tracker)

		if s.Tok == token.DEFINE { // := assignment
			// Check each LHS identifier
			for _, lhs := range s.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
					// Check if this := actually defines a new variable or reuses existing one
					if tracker.typesInfo != nil {
						if obj := tracker.typesInfo.Defs[ident]; obj != nil {
							// This identifier defines a new variable in a new scope
							continue
						}
					}

					if decl, exists := tracker.vars[ident.Name]; exists && !decl.used && !decl.redefined {
						// Found redundant declaration
						position := tracker.file.ToPosition(decl.node.Pos())
						failure := lint.Failure{
							Confidence: 1,
							Node:       decl.node,
							Failure:    fmt.Sprintf("redundant declaration of '%s'; it's redeclared via := assignment", ident.Name),
							Position:   lint.FailurePosition{Start: position},
						}
						*failures = append(*failures, failure)
						decl.redefined = true
					}
				}
			}
		} else if s.Tok == token.ASSIGN { // = assignment
			// Mark variables as used
			for _, lhs := range s.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					if decl, exists := tracker.vars[ident.Name]; exists {
						decl.used = true
					}
				}
			}
		}

	case *ast.BlockStmt:
		// Recursively process nested blocks
		analyzeBlock(s, tracker, failures)

	case *ast.IfStmt:
		// Mark variables in init statement
		if s.Init != nil {
			processStatement(s.Init, tracker, failures)
		}
		// Mark variables in condition as used
		if s.Cond != nil {
			markUsedInExpr([]ast.Expr{s.Cond}, tracker)
		}
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}
		// Process else
		if s.Else != nil {
			processStatement(s.Else, tracker, failures)
		}

	case *ast.ForStmt:
		// Mark variables in init statement
		if s.Init != nil {
			processStatement(s.Init, tracker, failures)
		}
		// Mark variables in condition as used
		if s.Cond != nil {
			markUsedInExpr([]ast.Expr{s.Cond}, tracker)
		}
		// Mark variables in post statement
		if s.Post != nil {
			processStatement(s.Post, tracker, failures)
		}
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}

	case *ast.RangeStmt:
		// Mark variables in range expression as used
		if s.X != nil {
			markUsedInExpr([]ast.Expr{s.X}, tracker)
		}
		// Handle the key/value assignments
		if s.Tok == token.DEFINE {
			// Check if key or value are redundant
			checkRedundantInAssignment([]ast.Expr{s.Key, s.Value}, tracker, failures)
		}
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}

	case *ast.SwitchStmt:
		// Mark variables in init statement
		if s.Init != nil {
			processStatement(s.Init, tracker, failures)
		}
		// Mark variables in tag as used
		if s.Tag != nil {
			markUsedInExpr([]ast.Expr{s.Tag}, tracker)
		}
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}

	case *ast.TypeSwitchStmt:
		// Mark variables in init statement
		if s.Init != nil {
			processStatement(s.Init, tracker, failures)
		}
		// Handle assign statement specially
		if s.Assign != nil {
			// Mark the expression being type-switched as used
			if assign, ok := s.Assign.(*ast.AssignStmt); ok && len(assign.Rhs) > 0 {
				markUsedInExpr(assign.Rhs, tracker)
			}
			processStatement(s.Assign, tracker, failures)
		}
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}

	case *ast.SelectStmt:
		// Process body
		if s.Body != nil {
			analyzeBlock(s.Body, tracker, failures)
		}

	case *ast.CaseClause:
		// Mark variables in case expressions as used
		markUsedInExpr(s.List, tracker)
		// Process case body statements
		for _, caseStmt := range s.Body {
			processStatement(caseStmt, tracker, failures)
		}

	case *ast.ExprStmt:
		// Mark any variables used in expressions
		markUsedInExpr([]ast.Expr{s.X}, tracker)

	case *ast.ReturnStmt:
		// Mark any variables used in return
		markUsedInExpr(s.Results, tracker)

	case *ast.DeferStmt:
		// Mark any variables used in defer
		markUsedInCallExpr(s.Call, tracker)

	case *ast.GoStmt:
		// Mark any variables used in go statement
		markUsedInCallExpr(s.Call, tracker)

	case *ast.SendStmt:
		// Mark variables used in channel send
		markUsedInExpr([]ast.Expr{s.Chan, s.Value}, tracker)

	case *ast.IncDecStmt:
		// Mark variable as used
		markUsedInExpr([]ast.Expr{s.X}, tracker)

	case *ast.LabeledStmt:
		// Process the labeled statement
		processStatement(s.Stmt, tracker, failures)

	case *ast.BranchStmt:
		// break, continue, goto, fallthrough - no variables to track
	}
}

func markUsedInExpr(exprs []ast.Expr, tracker *varTracker) {
	for _, expr := range exprs {
		if expr == nil {
			continue
		}

		ast.Inspect(expr, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.Ident:
				if decl, exists := tracker.vars[node.Name]; exists {
					decl.used = true
				}
			case *ast.UnaryExpr:
				if node.Op == token.AND {
					// Handle address-of operator specially
					if ident, ok := node.X.(*ast.Ident); ok {
						if decl, exists := tracker.vars[ident.Name]; exists {
							decl.used = true
						}
					}
				}
			}
			return true
		})
	}
}

func markUsedInCallExpr(call *ast.CallExpr, tracker *varTracker) {
	// Mark function as used
	markUsedInExpr([]ast.Expr{call.Fun}, tracker)
	// Mark arguments as used
	markUsedInExpr(call.Args, tracker)
}

func checkRedundantInAssignment(exprs []ast.Expr, tracker *varTracker, failures *[]lint.Failure) {
	for _, expr := range exprs {
		if expr == nil {
			continue
		}
		if ident, ok := expr.(*ast.Ident); ok && ident.Name != "_" {
			if decl, exists := tracker.vars[ident.Name]; exists && !decl.used && !decl.redefined {
				// Found redundant declaration
				position := tracker.file.ToPosition(decl.node.Pos())
				failure := lint.Failure{
					Confidence: 1,
					Node:       decl.node,
					Failure:    fmt.Sprintf("redundant declaration of '%s'; it's redeclared via := assignment", ident.Name),
					Position:   lint.FailurePosition{Start: position},
				}
				*failures = append(*failures, failure)
				decl.redefined = true
			}
		}
	}
}
