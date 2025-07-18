package fixtures

import (
	"fmt"
	"strings"
	"time"
)

// Test case: Variables with initialization values
func redundantWithInitialization() {
	var x int = 10                       // MATCH /redundant declaration of 'x'; it's redeclared via := assignment/
	x, y := computeValues()
	fmt.Println(x, y)
}

// Test case: Variable used in defer statement
func nonRedundantUsedInDefer() {
	var cleanup func()
	defer cleanup() // Used in defer before :=
	cleanup, err := getCleanupFunc()
	if err != nil {
		fmt.Println(err)
	}
}

// Test case: Variable used in go statement
func nonRedundantUsedInGo() {
	var worker func()
	go worker() // Used in go statement before :=
	worker, err := getWorkerFunc()
	if err != nil {
		fmt.Println(err)
	}
}

// Test case: Variable used as function argument
func nonRedundantUsedAsArgument() {
	var data string
	processString(data) // Used as argument before :=
	data, err := getData()
	if err != nil {
		fmt.Println(err)
	}
}

// Test case: Variable captured in closure
func nonRedundantCapturedInClosure() {
	var counter int
	fn := func() {
		fmt.Println(counter) // Captured in closure before :=
	}
	counter, err := getCounter()
	if err != nil {
		fmt.Println(err)
	}
	fn()
}

// Test case: Type assertion with :=
func redundantTypeAssertion() {
	var val interface{}                  // MATCH /redundant declaration of 'val'; it's redeclared via := assignment/
	val, ok := someInterface().(string)
	if ok {
		fmt.Println(val)
	}
}

// Test case: Channel receive with :=
func redundantChannelReceive() {
	var msg string                       // MATCH /redundant declaration of 'msg'; it's redeclared via := assignment/
	msg, ok := <-getChan()
	if ok {
		fmt.Println(msg)
	}
}

// Test case: Variable address taken
func nonRedundantAddressTaken() {
	var x int
	p := &x // Address taken before :=
	x, y := computeValues()
	fmt.Println(*p, y)
}

// Test case: Multiple variables in single declaration - mixed usage
func mixedMultipleVarDecl() {
	var a, b, c int
	fmt.Println(b) // b is used before :=
	a, x := computeValues()              // a is redundant, b is not
	b, y := computeOtherValues()         // b already used, not redundant
	c, z := processIntValues()           // c is redundant
	fmt.Println(a, b, c, x, y, z)
}

// Test case: Variable in for loop init
func redundantInForInit() {
	for i := 0; i < 10; i++ {
		var sum int                      // MATCH /redundant declaration of 'sum'; it's redeclared via := assignment/
		sum, err := calculate(i)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(sum)
	}
}

// Test case: Variable in select statement
func nonRedundantInSelect() {
	var result string
	select {
	case result = <-getChan(): // Used in select before :=
		fmt.Println(result)
	case <-time.After(time.Second):
		result, err := getDefault()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	}
}

// Test case: Variable in switch statement
func redundantInSwitch() {
	switch val := getValue(); val {
	case 1:
		var data string                  // MATCH /redundant declaration of 'data'; it's redeclared via := assignment/
		data, err := fetchData()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(data)
	}
}

// Test case: Variable in type switch
func nonRedundantInTypeSwitch() {
	var result interface{}
	switch result.(type) { // Used in type switch before :=
	case string:
		result, err := process()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	}
}

// Test case: Variable in anonymous function
func redundantInAnonymousFunc() {
	func() {
		var local int                    // MATCH /redundant declaration of 'local'; it's redeclared via := assignment/
		local, err := getLocalValue()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(local)
	}()
}

// Test case: Variable declared but never reassigned with :=
func nonRedundantNeverReassigned() {
	var unused int // Not redundant - never reassigned with :=
	fmt.Println(unused)
}

// Test case: Variable in if statement with init
func redundantInIfInit() {
	if x := 5; x > 0 {
		var y int                        // MATCH /redundant declaration of 'y'; it's redeclared via := assignment/
		y, err := computeY(x)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(y)
	}
}

// Test case: Same variable name in different branches
func redundantInDifferentBranches() {
	if condition() {
		var result string                // MATCH /redundant declaration of 'result'; it's redeclared via := assignment/
		result, err := getBranchA()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	} else {
		var result string                // MATCH /redundant declaration of 'result'; it's redeclared via := assignment/
		result, err := getBranchB()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	}
}

// Test case: Variable used in range expression
func nonRedundantUsedInRange() {
	var items []string
	for _, item := range items { // Used in range before :=
		fmt.Println(item)
	}
	items, err := getItems()
	if err != nil {
		fmt.Println(err)
	}
}

// Test case: Struct field access between declaration and :=
func nonRedundantStructFieldAccess() {
	var obj struct{ Field string }
	fmt.Println(obj.Field) // Field accessed before :=
	obj, err := getObject()
	if err != nil {
		fmt.Println(err)
	}
}

// Test case: Method call between declaration and :=
func nonRedundantMethodCall() {
	var builder strings.Builder
	builder.WriteString("test") // Method called before :=
	builder, err := getBuilder()
	if err != nil {
		fmt.Println(err)
	}
}

// Helper functions for additional tests
func getCleanupFunc() (func(), error) {
	return func() {}, nil
}

func getWorkerFunc() (func(), error) {
	return func() {}, nil
}

func processString(s string) {}

func getData() (string, error) {
	return "data", nil
}

func getCounter() (int, error) {
	return 1, nil
}

func someInterface() interface{} {
	return "test"
}

func getChan() <-chan string {
	ch := make(chan string, 1)
	ch <- "message"
	return ch
}

func processIntValues() (int, int) {
	return 7, 8
}

func calculate(i int) (int, error) {
	return i * 2, nil
}

func getValue() int {
	return 1
}

func fetchData() (string, error) {
	return "data", nil
}

func process() (interface{}, error) {
	return "processed", nil
}

func getLocalValue() (int, error) {
	return 42, nil
}

func computeY(x int) (int, error) {
	return x * 2, nil
}

func condition() bool {
	return true
}

func getBranchA() (string, error) {
	return "A", nil
}

func getBranchB() (string, error) {
	return "B", nil
}

func getItems() ([]string, error) {
	return []string{"a", "b"}, nil
}

func getObject() (struct{ Field string }, error) {
	return struct{ Field string }{Field: "value"}, nil
}

func getBuilder() (strings.Builder, error) {
	var b strings.Builder
	return b, nil
}