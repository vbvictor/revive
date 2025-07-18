package fixtures

import "fmt"

// Package-level variables (current implementation doesn't check these)
var pkgErr error

func usePkgVar() {
	result, pkgErr := someFunction() // pkgErr is package-level, not function-level
	if pkgErr != nil {
		fmt.Println(pkgErr)
	}
	fmt.Println(result)
}

// Test case: init function
func init() {
	var initVar int // Current implementation should check this
	initVar, err := computeValues()
	if err != nil {
		panic(err)
	}
	fmt.Println(initVar)
}

// Test case: Nested function literals
func nestedFunctionLiterals() {
	fn := func() {
		var outer int // MATCH /redundant declaration of 'outer'; it's redeclared via := assignment/
		outer, err := getCounter()
		if err != nil {
			innerFn := func() {
				var inner int // MATCH /redundant declaration of 'inner'; it's redeclared via := assignment/
				inner, err2 := getCounter()
				if err2 != nil {
					fmt.Println(err2)
				}
				fmt.Println(inner, outer)
			}
			innerFn()
		}
		fmt.Println(outer)
	}
	fn()
}

// Test case: Variable in labeled statement
func redundantWithLabel() {
outer:
	for i := 0; i < 10; i++ {
		var x int // MATCH /redundant declaration of 'x'; it's redeclared via := assignment/
		x, err := calculate(i)
		if err != nil {
			break outer
		}
		fmt.Println(x)
	}
}

// Test case: Multiple := in sequence
func multipleShortDeclarations() {
	var x int // MATCH /redundant declaration of 'x'; it's redeclared via := assignment/
	x, y := computeValues()
	x, z := computeValues() // Second := on same variable
	fmt.Println(x, y, z)
}

// Test case: Variable in very nested scopes
func veryNestedScopes() {
	{
		{
			{
				var deep int // MATCH /redundant declaration of 'deep'; it's redeclared via := assignment/
				deep, err := getCounter()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(deep)
			}
		}
	}
}

// Test case: Interface method implementations
type MyInterface interface {
	Method() (int, error)
}

type MyStruct struct{}

func (m MyStruct) Method() (int, error) {
	var result int // MATCH /redundant declaration of 'result'; it's redeclared via := assignment/
	result, err := computeValues()
	return result, err
}

// Test case: Pointer receiver method
func (m *MyStruct) PointerMethod() {
	var data string // MATCH /redundant declaration of 'data'; it's redeclared via := assignment/
	data, err := getData()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}

// Test case: Variable with complex type
func redundantComplexType() {
	var fn func(int) (string, error) // MATCH /redundant declaration of 'fn'; it's redeclared via := assignment/
	fn, err := getFunctionValue()
	if err != nil {
		fmt.Println(err)
	}
	result, _ := fn(42)
	fmt.Println(result)
}

// Test case: Map indexing with comma-ok
func redundantMapIndex() {
	m := make(map[string]int)
	var val int // MATCH /redundant declaration of 'val'; it's redeclared via := assignment/
	val, ok := m["key"]
	if ok {
		fmt.Println(val)
	}
}

// Test case: Variable declared in multiple var groups
func multipleVarGroups() {
	var a int // MATCH /redundant declaration of 'a'; it's redeclared via := assignment/

	var b string // MATCH /redundant declaration of 'b'; it's redeclared via := assignment/

	a, x := computeValues()
	b, err := getData()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a, b, x)
}

// Test case: Const shadowing (should not be flagged)
func constShadowing() {
	const x = 10
	x, y := computeValues() // x here shadows the const, creates new variable
	fmt.Println(x, y)
}

// Helper function
func getFunctionValue() (func(int) (string, error), error) {
	return func(i int) (string, error) {
		return fmt.Sprintf("%d", i), nil
	}, nil
}
