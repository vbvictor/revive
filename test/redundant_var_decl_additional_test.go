package test

import (
	"testing"

	"github.com/mgechev/revive/rule"
)

func TestRedundantVarDeclAdditional(t *testing.T) {
	testRule(t, "redundant_var_decl_additional", &rule.RedundantVarDeclRule{})
}
