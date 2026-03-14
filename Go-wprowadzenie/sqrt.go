package main

import (
	"fmt"
	"math"
)

func Sqrt(x float64) float64 {
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
		fmt.Println(z)
	}
	return z
}

func SqrtConvergence(x float64) float64 {
	z := 2.0
	iterations := 0
	for {
		prev := z
		z -= (z*z - x) / (2 * z)
		iterations++
		if math.Abs(z-prev) < 1e-10 {
			break
		}
	}
	fmt.Printf("Converged in %d iterations\n", iterations)
	return z
}

func SqrtWithGuess(x, guess float64) float64 {
	z := guess
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func main() {
	x := 2.0

	fmt.Println("=== Part 1: 10 iterations ===")
	result1 := Sqrt(x)
	fmt.Printf("Result: %v\n\n", result1)

	fmt.Println("=== Part 2: Convergence test ===")
	result2 := SqrtConvergence(x)
	fmt.Printf("Result: %v\n\n", result2)

	fmt.Println("=== Testing different initial guesses ===")
	fmt.Printf("Guess = 1:   %v\n", SqrtWithGuess(x, 1.0))
	fmt.Printf("Guess = x:   %v\n", SqrtWithGuess(x, x))
	fmt.Printf("Guess = x/2: %v\n", SqrtWithGuess(x, x/2))

	fmt.Println("\n=== Comparison with math.Sqrt ===")
	fmt.Printf("My Sqrt:    %v\n", result2)
	fmt.Printf("math.Sqrt:  %v\n", math.Sqrt(x))
	fmt.Printf("Difference: %v\n", math.Abs(result2-math.Sqrt(x)))
}
