package test

import (
	"testing"

	"github.com/mgechev/revive/rule"
)

func TestRedundantVarDeclEdgeCases(t *testing.T) {
	testRule(t, "redundant_var_decl_edge_cases", &rule.RedundantVarDeclRule{})
}
