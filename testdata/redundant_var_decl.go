package fixtures

import (
	"fmt"
	"io"
)

func redundantErrDecl() {
	var err error                        // MATCH /redundant declaration of 'err'; it's redeclared via := assignment/
	result, err := someFunction()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func redundantIntDecl() {
	var foo int                          // MATCH /redundant declaration of 'foo'; it's redeclared via := assignment/
	foo, bar := 2, 3
	fmt.Println(foo, bar)
}

func redundantMultipleVars() {
	var x int                            // MATCH /redundant declaration of 'x'; it's redeclared via := assignment/
	var y int                            // MATCH /redundant declaration of 'y'; it's redeclared via := assignment/
	x, z := computeValues()              
	y, w := computeOtherValues()
	fmt.Println(x, y, z, w)
}

func nonRedundantUsedBetween() {
	var err error
	fmt.Println("Error is nil:", err == nil) // Used between declaration and :=
	result, err := someFunction()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func nonRedundantRegularAssignment() {
	var err error
	result := someFunction2()
	err = result.Error() // Regular = assignment, not :=
	if err != nil {
		fmt.Println(err)
	}
}

func redundantActuallyUsed() {
	var count int                        // MATCH /redundant declaration of 'count'; it's redeclared via := assignment/
	count, err := processData() // This reuses the existing count variable
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(count)
}

func nonRedundantDifferentScope() {
	var count int
	fmt.Println("Initial count:", count)
	if true {
		count := 10 // Different scope, shadows outer count
		fmt.Println("Inner count:", count)
	}
	fmt.Println("Outer count:", count)
}

func redundantReaderDecl() {
	var reader io.Reader                 // MATCH /redundant declaration of 'reader'; it's redeclared via := assignment/
	reader, err := openFile("test.txt")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(reader)
}

func redundantComplexCase() {
	var a int                            // MATCH /redundant declaration of 'a'; it's redeclared via := assignment/
	var b int                            // MATCH /redundant declaration of 'b'; it's redeclared via := assignment/
	var c int                            // MATCH /redundant declaration of 'c'; it's redeclared via := assignment/
	a, d := 1, 2                         
	b, e := 3, 4                         
	c, f := 5, 6
	fmt.Println(a, b, c, d, e, f)
}

func nonRedundantBlankIdentifier() {
	var _ error
	result, _ := someFunction() // Blank identifier, no issue
	fmt.Println(result)
}

// Helper functions for tests
func someFunction() (string, error) {
	return "result", nil
}

func someFunction2() struct{ Error func() error } {
	return struct{ Error func() error }{
		Error: func() error { return nil },
	}
}

func computeValues() (int, int) {
	return 1, 2
}

func computeOtherValues() (int, int) {
	return 3, 4
}

func processData() (int, error) {
	return 42, nil
}

func openFile(name string) (io.Reader, error) {
	return nil, fmt.Errorf("not implemented")
}