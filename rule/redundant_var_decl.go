package rule

import (
	"fmt"
	"go/ast"
	"go/token"

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
	funcVisitor := &functionVisitor{
		file:     file,
		failures: []lint.Failure{},
	}
	ast.Walk(funcVisitor, file.AST)
	return funcVisitor.failures
}

// functionVisitor visits function declarations
type functionVisitor struct {
	file     *lint.File
	failures []lint.Failure
}

func (v *functionVisitor) Visit(node ast.Node) ast.Visitor {
	funcDecl, ok := node.(*ast.FuncDecl)
	if !ok || funcDecl.Body == nil {
		return v
	}

	// Process each function with its own variable tracking
	bodyVisitor := &bodyVisitor{
		file:         v.file,
		vars:         make(map[string]*varInfo),
		failures:     &v.failures,
		currentScope: nil,
	}
	ast.Walk(bodyVisitor, funcDecl.Body)
	
	return nil // Skip further traversal of this function
}

// bodyVisitor handles the function body analysis
type bodyVisitor struct {
	file         *lint.File
	vars         map[string]*varInfo
	failures     *[]lint.Failure
	currentScope ast.Node // For tracking the current block scope
}

// varInfo tracks information about a variable
type varInfo struct {
	declNode  ast.Node    // The declaration node
	declPos   token.Pos   // Position of declaration
	usedPos   []token.Pos // Positions where the variable is used
	isTracked bool        // Whether we're tracking this variable
}

func (v *bodyVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch x := node.(type) {
	case *ast.BlockStmt:
		// Create a new visitor for nested blocks to handle scoping
		if v.currentScope != nil { // If we're already in a nested scope
			nestedVisitor := &bodyVisitor{
				file:         v.file,
				vars:         make(map[string]*varInfo),
				failures:     v.failures, // Share the failures collection
				currentScope: x,
			}
			ast.Walk(nestedVisitor, x)
			return nil // Skip further traversal of this block
		}
		v.currentScope = x // Set current scope for root block

	case *ast.GenDecl:
		if x.Tok != token.VAR {
			return v
		}

		// Process var declarations
		for _, spec := range x.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}
				
				// Store declaration info
				v.vars[name.Name] = &varInfo{
					declNode:  spec,
					declPos:   name.Pos(),
					usedPos:   []token.Pos{},
					isTracked: true,
				}
			}
		}

	case *ast.AssignStmt:
		if x.Tok == token.DEFINE { // := assignment
			for _, lhs := range x.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name == "_" {
					continue
				}

				info, exists := v.vars[ident.Name]
				if !exists || !info.isTracked {
					continue
				}

				// Check if the variable was used between declaration and := assignment
				usedBetween := false
				for _, pos := range info.usedPos {
					if info.declPos < pos && pos < x.Pos() {
						usedBetween = true
						break
					}
				}

				if !usedBetween {
					// Variable not used between declaration and := assignment
					position := v.file.ToPosition(info.declNode.Pos())
					failure := lint.Failure{
						Confidence: 1,
						Node:       info.declNode,
						Failure:    fmt.Sprintf("redundant declaration of '%s'; it's redeclared via := assignment", ident.Name),
						Position:   lint.FailurePosition{Start: position},
					}
					*v.failures = append(*v.failures, failure)
				}

				// Stop tracking this variable after := assignment
				info.isTracked = false
			}
		} else if x.Tok == token.ASSIGN { // = assignment
			// Record variable usage in regular assignment
			for _, lhs := range x.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name == "_" {
					continue
				}

				info, exists := v.vars[ident.Name]
				if !exists || !info.isTracked {
					continue
				}

				// Record this usage
				info.usedPos = append(info.usedPos, ident.Pos())
			}
		}

	case *ast.Ident:
		// Record any usage of tracked variables
		info, exists := v.vars[x.Name]
		if !exists || !info.isTracked {
			return v
		}

		// Skip identifiers in their own declaration (part of a ValueSpec)
		if isPartOfValueSpec(x) {
			return v
		}

		// Record this usage
		info.usedPos = append(info.usedPos, x.Pos())
	}

	return v
}

// Helper function to check if an identifier is part of its own declaration
func isPartOfValueSpec(ident *ast.Ident) bool {
	// This is a simplification. In a real implementation, you'd need 
	// parent node tracking to determine if the identifier is in a ValueSpec.
	return false
}
