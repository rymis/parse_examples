package main

import (
	"github.com/rymis/parse"
	"fmt"
)

// You can use very simple grammar for calc with parse module.
// The grammar itself in PEG form looks like:
// Expression <- Expression [+-] Production / Production
// Production <- Production [*/%] Atom / Atom
// Atom <- '(' Expression ')' / Number
// Number is floating point number.

// First I will define Atom type:
type Atom struct {
	parse.FirstOf  // It is variant type
	Expr struct { // Braced expression
		_ string `literal:"("`
		Expr *Expression
		_ string `literal:")"`
	}
	Number float64
}

// Next productions (multiplicative operations):
type Production struct {
	parse.FirstOf
	Prod struct {
		First *Production
		Op    string `regexp:"[/%*]"`
		Second Atom
	}
	Atom Atom
}

// And Expression itself:
type Expression struct {
	parse.FirstOf
	Expr struct {
		First *Expression
		Op    string `regexp:"[-+]"`
		Second Production
	}
	Prod Production
}

// Interface for all calculatables:
type Calc interface {
	Calc() float64
}

// Easy to write calc functions:
func (self *Atom) Calc() float64 {
	if self.Field == "Number" {
		return self.Number
	} else {
		return self.Expr.Expr.Calc()
	}
}

func (self *Production) Calc() float64 {
	if self.Field == "Atom" {
		return self.Atom.Calc()
	} else {
		return op(self.Prod.Op[0], self.Prod.First.Calc(), self.Prod.Second.Calc())
	}
}

func (self *Expression) Calc() float64 {
	if self.Field == "Prod" {
		return self.Prod.Calc()
	} else {
		return op(self.Expr.Op[0], self.Expr.First.Calc(), self.Expr.Second.Calc())
	}
}

func op(op byte, a, b float64) float64 {
	switch op {
	case '+':
		return a + b
	case '-':
		return a - b
	case '*':
		return a * b
	case '/':
		return a / b
	case '%':
		return float64(uint64(a) % uint64(b))
	}

	panic("Unknown operation")
}

// Function to calculate value of expression:
func calc(expr string) float64 {
	var e Expression
	nl, err := parse.Parse(&e, []byte(expr), nil)
	if err != nil {
		println("Error: ", err)
		return 0.0
	} else if nl < len(expr) {
		println("WARNING: '", expr[nl:], "' was not parsed!")
	}

	return e.Calc()
}

func test(e string, v float64) {
	f := calc(e)
	fmt.Printf("%s == %f (error = %f)\n", e, f, f - v)
}

func main() {
	test("2 * 2", 2.0 * 2.0)
	test("1 + 2 * 3 - 4 / 5.0 / .333e-1", 1 + 2 * 3 - 4 / 5.0 / .333e-1)
}

