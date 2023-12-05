package main

import (
	"fmt"

	ppkg "github.com/bitwombat/gps-tags/poly"
)

// main function to test the code
func main() {
	// Example usage
	poly := []ppkg.Point{{X: 0.0, Y: 0.0}, {X: 10.0, Y: 3.0}, {X: 10.0, Y: 10.0}, {X: 0.0, Y: 10.0}}
	// Small algorithm problem... if the infinity line hits a vertex, it gets
	// counted as intersecting both segments that make up that vertex, giving the
	// wrong answer. In reality, this will rarely happen due to the number of
	// digits in GPS coordinates.

	//	p := Point{X:8.0, Y:3.0}  // Returns 'false', which is the wrong answer.
	p := ppkg.Point{X: 8.0, Y: 3.0000001} // Returns 'true'

	if ppkg.IsInside(poly, p) {
		fmt.Println("The point is inside the polygon.")
	} else {
		fmt.Println("The point is not inside the polygon.")
	}
}
