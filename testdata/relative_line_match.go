package fixtures

import "fmt"

func exampleFunction() {
	var unused int // This variable is unused
	// MATCH:[-1] /exported function \w+ should have comment or be unexported/
	
	fmt.Println("hello") // This line is fine
	var another int      // Another unused variable
	// MATCH:[-1] /exported function \w+ should have comment or be unexported/
	
	var third int // Third unused variable
	var fourth int
	// MATCH:[+1] /exported function \w+ should have comment or be unexported/
	var fifth int // Fifth unused variable
}