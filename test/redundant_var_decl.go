package test

import (
	"testing"

	"github.com/mgechev/revive/rule"
)

func TestRedundantVarDecl(t *testing.T) {
	testRule(t, "redundant_var_decl", &rule.RedundantVarDeclRule{})
}
