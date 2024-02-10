package main

import (
	"fmt"
	"strconv"
)

var debug = false

type Expression interface {
	Derivative() Expression
	simplify() Expression
	String() string
}

type Constant struct{ num int }

func (n *Constant) Derivative() Expression { return &Constant{0} }
func (n *Constant) String() string         { return strconv.Itoa(n.num) }
func (n *Constant) simplify() Expression   { return n }

type X struct{}

func (x *X) Derivative() Expression { return &Constant{1} }
func (x *X) String() string         { return "x" }
func (x *X) simplify() Expression   { return x }

type ExprsMultiplied struct {
	expr1 Expression
	expr2 Expression
}

func (e *ExprsMultiplied) simplify() Expression {
	e.expr1 = e.expr1.simplify()
	e.expr2 = e.expr2.simplify()
	c1, okc1 := e.expr1.(*Constant)
	//fmt.Println("c1", c1, "okc1", okc1, "expr1", e.expr1)
	c2, okc2 := e.expr2.(*Constant)
	//fmt.Println("c2", c2, "okc2", okc2, "expr2", e.expr2)
	if okc1 || okc2 {
		var constant *Constant
		var expr Expression
		if okc1 {
			constant = c1
			expr = e.expr2
		} else {
			constant = c2
			expr = e.expr1
		}
		if constant.num == 0 {
			return constant
		}
		if constant.num == 1 {
			return expr
		}
	}
	return e
}

func (e *ExprsMultiplied) Derivative() Expression {
	simp := e.simplify()
	//fmt.Println("simplified", simp, "original", e)
	if simp != e {
		return simp.Derivative()
	}
	if _, ok := e.expr1.(*Constant); ok {
		return &ExprsMultiplied{expr1: e.expr1, expr2: e.expr2.Derivative()}
	}
	if _, ok := e.expr2.(*Constant); ok {
		return &ExprsMultiplied{expr1: e.expr2, expr2: e.expr1.Derivative()}
	}
	return &ExprsAdded{
		expr1: &ExprsMultiplied{expr1: e.expr1.Derivative(), expr2: e.expr2},
		expr2: &ExprsMultiplied{expr1: e.expr1, expr2: e.expr2.Derivative()},
	}
}

func (e *ExprsMultiplied) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "mult[" + str + "]"
		}()
	}
	return e.expr1.String() + " * " + e.expr2.String()
}

type ExprsAdded struct {
	expr1 Expression
	expr2 Expression
}

func (e *ExprsAdded) simplify() Expression {
	e.expr1 = e.expr1.simplify()
	e.expr2 = e.expr2.simplify()
	c1, okc1 := e.expr1.(*Constant)
	c2, okc2 := e.expr2.(*Constant)
	if okc1 || okc2 {
		var constant *Constant
		var expr Expression
		if okc1 {
			constant = c1
			expr = e.expr2
		} else {
			constant = c2
			expr = e.expr1
		}
		if constant.num == 0 {
			return expr
		}
	}
	return e
}

func (e *ExprsAdded) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	if _, ok := e.expr1.(*Constant); ok {
		if e.expr1.(*Constant).num == 0 {
			return e.expr2.Derivative()
		}
	}
	if _, ok := e.expr2.(*Constant); ok {
		if e.expr2.(*Constant).num == 0 {
			return e.expr1.Derivative()
		}
	}
	return &ExprsAdded{expr1: e.expr1.Derivative(), expr2: e.expr2.Derivative()}
}

func (e *ExprsAdded) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "add[" + str + "]"
		}()
	}

	if _, ok := e.expr1.(*Constant); ok {
		if e.expr1.(*Constant).num == 0 {
			return e.expr2.String()
		}
	}
	if _, ok := e.expr2.(*Constant); ok {
		if e.expr2.(*Constant).num == 0 {
			return e.expr1.String()
		}
	}
	return e.expr1.String() + " + " + e.expr2.String()
}

type Polynomial struct {
	powerToCoeff map[int]int
	inside       Expression
}

func (p *Polynomial) simplify() Expression {
	p.inside = p.inside.simplify()
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return &Constant{coeff}
			}
			if power == 1 {
				return &ExprsMultiplied{expr1: &Constant{coeff}, expr2: p.inside}
			}
		}
	}
	return p
}

func (p *Polynomial) Derivative() Expression {
	simp := p.simplify()
	if simp != p {
		return simp.Derivative()
	}
	if _, ok := p.inside.(*Constant); ok {
		return &Constant{0}
	}
	if len(p.powerToCoeff) == 1 {
		for power, coeff := range p.powerToCoeff {
			if power == 0 {
				return &Constant{0}
			}
			if power == 1 {
				return &ExprsMultiplied{expr1: &Constant{coeff}, expr2: p.inside.Derivative()}
			}
		}
	}
	derivative := &Polynomial{make(map[int]int, len(p.powerToCoeff)), p.inside}
	for power, coeff := range p.powerToCoeff {
		if power > 0 {
			derivative.powerToCoeff[power-1] = power * coeff
		}
	}
	return &ExprsMultiplied{expr1: derivative, expr2: p.inside.Derivative()}
}

func (p *Polynomial) String() (str string) {
	if debug {
		defer func() {
			str = "polyn[" + str + "]"
		}()
	}

	insideStr := p.inside.String()
	if _, ok := p.inside.(*Constant); ok {
		if p.inside.(*Constant).num == 0 {
			return "0"
		}
		insideStr = "(" + insideStr + ")"
	}
	// Sort the powers in ascending order
	//powers := make([]int, 0, len(p.powerToCoeff))
	//for power := range p.powerToCoeff {
	//	powers = append(powers, power)
	//}
	//slices.Sort(powers)
	for power, coeff := range p.powerToCoeff {
		//coeff := p.powerToCoeff[power]
		if coeff == 0 {
			continue
		}
		coeffStr := ""
		if coeff != 1 {
			coeffStr = strconv.Itoa(coeff)
		}
		if power == 0 {
			if coeffStr == "" {
				coeffStr = "1"
			}
			str += coeffStr + " + "
		} else if power == 1 {
			str += coeffStr + "" + insideStr + " + "
		} else {
			str += coeffStr + "" + insideStr + "^" + strconv.Itoa(power) + " + "
		}
	}
	// Remove the trailing " + "
	if len(str) > 0 {
		str = str[:len(str)-3]
	}
	if len(p.powerToCoeff) == 1 { // if there's only a single term don't wrap it in brackets
		return str
	}
	return "(" + str + ")"
}

func main() {
	p := &Polynomial{map[int]int{0: 1, 1: 2, 2: 3},
		&Polynomial{map[int]int{30: 1, 31: 1}, &X{}},
	}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative().Derivative())

	//p := &ExprsMultiplied{
	//	&Polynomial{map[int]int{0: 1}, &X{}},
	//	&Polynomial{map[int]int{30: 1}, &X{}},
	//}
	fmt.Println(p)
	fmt.Println(p.Derivative())
	fmt.Println(p.Derivative().Derivative())
	fmt.Println(p.Derivative().Derivative().Derivative())
	fmt.Println(p.Derivative().Derivative().Derivative().Derivative())

	//p := &Polynomial{map[int]int{0: 1},
	//	&X{},
	//}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
	//fmt.Println(p.Derivative().Derivative().Derivative())

	//p := &ExprsMultiplied{&Constant{6}, &Polynomial{
	//	powerToCoeff: map[int]int{0: 1, 1: 2, 2: 3},
	//	inside:       &X{},
	//}}
	//fmt.Println(p)
	//fmt.Println(p.Derivative())
	//fmt.Println(p.Derivative().Derivative())
}
