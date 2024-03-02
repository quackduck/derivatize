package main

import (
	"reflect"
)

type ExprsDivided struct {
	high Expression
	low  Expression
}

// may edit things within it
func (e *ExprsDivided) simplify() (ret Expression) {
	e.high = e.high.simplify()
	e.low = e.low.simplify()

	m1, okm1 := e.high.(*ExprsMultiplied)
	m2, okm2 := e.low.(*ExprsMultiplied)
	if okm1 && okm2 {
		e.high, e.low = removeCommonMul(m1, m2)
	}

	//fmt.Println("simplify", e.high, e.low)

	if mul1, ok := e.high.(*ExprsMultiplied); ok {
		// (a * b) / c = (a * b * 1/c)
		//fmt.Println("HELLO", mul1.es, e.low)
		mul1.es = append(mul1.es, div(num(1), e.low).simplify())
		//fmt.Println("HELLO", mul1.es, e.low)
		e.low = num(1)
		return mul1.simplify()
	}

	if p1, ok := e.high.(*Polynomial); ok {
		if v, ok := e.low.(*Var); ok {
			if pv, ok := p1.inside.(*Var); ok && reflect.DeepEqual(pv, v) {
				newPowerToCoeff := make(map[float64]float64, len(p1.powerToCoeff))
				for power, coeff := range p1.powerToCoeff {
					newPowerToCoeff[power-1] = coeff
				}
				p1.powerToCoeff = newPowerToCoeff
				e.low = num(1)
				return p1.simplify()
			}
		}
	}

	c1, okc1 := e.high.(*Constant)
	c2, okc2 := e.low.(*Constant)
	if okc1 && okc2 {
		return &Constant{c1.num / c2.num}
	}
	if okc2 {
		if c2.num == 1 {
			return e.high
		}
		if c2.num == 0 {
			panic("division by zero")
		}
		if c2.num == -1 {
			return mul(num(-1), e.high)
		}
	}
	if okc1 {
		//fmt.Println("c1", c1.num)
		if c1.num == 0 {
			//fmt.Println("zero")
			return &Constant{0}
		}
		if polyn, ok := e.low.(*Polynomial); ok { // constant divided by polynomial
			return (poly(map[float64]float64{-1: c1.num}, polyn)).simplify()
		}
		if v, ok := e.low.(*Var); ok {
			return (poly(map[float64]float64{-1: c1.num}, v)).simplify()
		}
	}

	// check if the high is ExprsDivided
	if d, ok := e.high.(*ExprsDivided); ok {
		// (a/b) / c = a / (b*c)
		return div(d.high, mul(d.low, e.low)).simplify()
	}
	return e
}

// caution: edits high and low
func removeCommonMul(high *ExprsMultiplied, low *ExprsMultiplied) (Expression, Expression) {
	for i := 0; i < len(high.es); i++ {
		for j := 0; j < len(low.es) && i >= 0; j++ {
			if reflect.DeepEqual(high.es[i], low.es[j]) {
				high.es = append(high.es[:i], high.es[i+1:]...)
				low.es = append(low.es[:j], low.es[j+1:]...)
				i--
				j--
			}
		}
	}
	var r1 Expression
	switch len(high.es) {
	case 0:
		r1 = num(1)
	case 1:
		r1 = high.es[0]
	default:
		r1 = high
	}
	switch len(low.es) {
	case 0:
		return r1, num(1)
	case 1:
		return r1, low.es[0]
	default:
		return r1, low
	}
}

func (e *ExprsDivided) Derivative() Expression {
	simp := e.simplify()
	if simp != e {
		return simp.Derivative()
	}
	return &ExprsDivided{
		high: &ExprsSubtracted{
			expr1: mul(e.low, e.high.Derivative()),
			expr2: mul(e.high, e.low.Derivative()),
		},
		low: poly(map[float64]float64{2: 1}, e.low),
	}
}

func (e *ExprsDivided) String() (str string) {
	simp := e.simplify()
	if simp != e {
		return simp.String()
	}
	if debug {
		defer func() {
			str = "div[" + str + "]"
		}()
	}
	highStr := e.high.String()
	lowStr := e.low.String()
	//switch e.high.(type) {
	//case *ExprsAdded, *ExprsSubtracted:
	//	highStr = "(" + highStr + ")"
	//}
	switch e.low.(type) {
	//case *ExprsAdded, *ExprsSubtracted, *ExprsMultiplied, *ExprsDivided:
	case *ExprsMultiplied, *ExprsDivided:
		lowStr = "[ " + lowStr + " ]"
	}
	return highStr + " / " + lowStr
}

func div(e1 Expression, e2 Expression) Expression {
	//fmt.Println("div", e1, e2)
	return &ExprsDivided{e1, e2}
}

func (e *ExprsDivided) structure() string {
	return "div{" + e.high.structure() + " / " + e.low.structure() + "}"
}
